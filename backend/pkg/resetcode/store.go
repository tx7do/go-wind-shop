// Package resetcode 提供密码重置验证码的生成、Redis 存取与一次性校验。
//
// 设计要点：
//   - 验证码 6 位数字，crypto/rand 生成，避免伪随机可预测。
//   - 存 Redis，key 形如 gws:pwdreset:{email}，TTL 默认 10 分钟。
//   - 校验与尝试计数全部在单条 Lua 脚本内原子完成（check + 比对 + INCR + 删除），
//     消除"GET 检查"与"INCR"分离导致的 TOCTOU 竞态，确保并发下尝试次数严格受限。
//   - 爆破防护：每个验证码最多 MaxAttempts 次错误尝试，超限则作废验证码，
//     强制重新获取。
//   - 邮箱在 key 里小写归一，避免大小写差异导致校验失败。
package resetcode

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	keyPrefix    = "gws:pwdreset:"
	attemptSuffix = ":attempts"
	defaultTTL   = 10 * time.Minute
	// MaxAttempts 单个验证码允许的最大错误尝试次数，超限作废验证码。
	MaxAttempts = 5
)

// verifyScript 原子地完成：尝试次数检查 + 验证码比对 + 失败累加 + 达限删除。
//
// KEYS[1] = codeKey, KEYS[2] = attemptsKey
// ARGV[1] = 用户输入的验证码, ARGV[2] = MaxAttempts, ARGV[3] = TTL(秒)
//
// 返回值：
//   "1"           验证码正确，已删除 codeKey 与 attemptsKey
//   "0"           验证码错误，已 INCR attemptsKey（未达上限）
//   "TOO_MANY"    尝试次数已达上限，codeKey 已删除
//   "NO_CODE"     codeKey 不存在（已过期或从未生成）
//
// 所有读写在 Redis 单线程内原子执行，并发请求不会越过 MaxAttempts 上限。
const verifyScript = `
local code = redis.call("GET", KEYS[1])
if not code then
	return "NO_CODE"
end
local attempts = tonumber(redis.call("GET", KEYS[2]) or "0")
if attempts >= tonumber(ARGV[2]) then
	-- 已超限，确保 codeKey 作废并返回 TOO_MANY
	redis.call("DEL", KEYS[1])
	return "TOO_MANY"
end
if code == ARGV[1] then
	-- 校验通过：清理 codeKey 与 attemptsKey
	redis.call("DEL", KEYS[1])
	redis.call("DEL", KEYS[2])
	return "1"
end
-- 校验失败：累加尝试次数。首次失败时设置与 codeKey 同 TTL，避免计数键永生。
attempts = redis.call("INCR", KEYS[2])
if attempts == 1 then
	redis.call("EXPIRE", KEYS[2], ARGV[3])
end
if attempts >= tonumber(ARGV[2]) then
	-- 达上限，作废验证码，切断后续尝试
	redis.call("DEL", KEYS[1])
	return "TOO_MANY"
end
return "0"
`

// ErrTooManyAttempts 错误尝试超限，验证码已作废，需重新获取。
var ErrTooManyAttempts = errors.New("resetcode: too many failed attempts, code invalidated")

// Store 基于 Redis 的重置验证码存储。
type Store struct {
	rdb *redis.Client
	ttl time.Duration
}

func New(rdb *redis.Client) *Store {
	return &Store{rdb: rdb, ttl: defaultTTL}
}

func (s *Store) codeKey(email string) string {
	return keyPrefix + normalizeEmail(email)
}

func (s *Store) attemptsKey(email string) string {
	return keyPrefix + normalizeEmail(email) + attemptSuffix
}

// Generate 生成 6 位数字验证码并存入 Redis（覆盖已有），同时重置尝试计数。
// 返回明文验证码。
func (s *Store) Generate(ctx context.Context, email string) (string, error) {
	code, err := randomDigits(6)
	if err != nil {
		return "", fmt.Errorf("generate reset code: %w", err)
	}
	pipe := s.rdb.TxPipeline()
	pipe.Set(ctx, s.codeKey(email), code, s.ttl)
	// 尝试计数与验证码同 TTL，重新生成时归零。
	pipe.Del(ctx, s.attemptsKey(email))
	if _, err := pipe.Exec(ctx); err != nil {
		return "", fmt.Errorf("store reset code: %w", err)
	}
	return code, nil
}

// Verify 校验验证码（全部在单条 Lua 脚本内原子完成）：
//   - 正确：删除验证码与尝试计数（一次性），返回 (true, nil)。
//   - 错误且未达上限：尝试计数 +1，返回 (false, nil)。
//   - 错误且达上限：作废验证码，返回 (false, ErrTooManyAttempts)。
//   - 验证码已过期/不存在：返回 (false, nil)。
//
// 由于 check + incr + del 原子执行，并发请求无法越过 MaxAttempts 上限。
func (s *Store) Verify(ctx context.Context, email, code string) (bool, error) {
	res, err := s.rdb.Eval(ctx, verifyScript,
		[]string{s.codeKey(email), s.attemptsKey(email)},
		code, MaxAttempts, int(s.ttl.Seconds()),
	).Text()
	if err != nil {
		return false, fmt.Errorf("verify reset code: %w", err)
	}
	switch res {
	case "1":
		return true, nil
	case "TOO_MANY":
		return false, ErrTooManyAttempts
	case "0", "NO_CODE":
		return false, nil
	default:
		// 未知返回值，保守按失败处理。
		return false, fmt.Errorf("verify reset code: unexpected script result %q", res)
	}
}

// TTL 返回当前配置的有效期（供上层回传给前端做倒计时）。
func (s *Store) TTL() time.Duration {
	return s.ttl
}

// randomDigits 生成 n 位数字字符串（前导零补齐），使用 crypto/rand。
func randomDigits(n int) (string, error) {
	max := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(n)), nil)
	num, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%0*d", n, num), nil
}

func normalizeEmail(email string) string {
	b := []byte(email)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		}
	}
	return string(b)
}
