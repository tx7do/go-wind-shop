// Package resetcode 提供密码重置验证码的生成、Redis 存取与一次性校验。
//
// 设计要点：
//   - 验证码 6 位数字，crypto/rand 生成，避免伪随机可预测。
//   - 存 Redis，key 形如 gws:pwdreset:{email}，TTL 默认 10 分钟。
//   - 校验采用 Lua 脚本原子"比对即删"，保证一次性（防重放）。
//   - 爆破防护：每个验证码最多 MaxAttempts 次错误尝试，超限则作废验证码，
//     强制重新获取。尝试计数独立 Redis key，与验证码同生命周期。
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
	keyPrefix  = "gws:pwdreset:"
	attemptSuffix = ":attempts"
	defaultTTL   = 10 * time.Minute
	// MaxAttempts 单个验证码允许的最大错误尝试次数，超限作废验证码。
	MaxAttempts = 5
)

// verifyScript 原子比对并删除：仅当 key 存在且值等于传入 code 时返回 1 并删除，
// 否则返回 0（不删，允许剩余次数重试，但仍受 TTL 自然过期约束）。
const verifyScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
else
	return 0
end
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

// Verify 校验验证码：
//   - 正确：删除验证码与尝试计数（一次性），返回 (true, nil)。
//   - 错误：尝试计数 +1；达到 MaxAttempts 则作废验证码并返回 ErrTooManyAttempts。
//   - 验证码已过期/不存在：返回 (false, nil)。
//
// 调用方应将 ErrTooManyAttempts 与普通错误码区别对待（提示用户重新获取）。
func (s *Store) Verify(ctx context.Context, email, code string) (bool, error) {
	// 先检查尝试次数是否已超限（防御快速并发爆破）。
	attempts, err := s.rdb.Get(ctx, s.attemptsKey(email)).Int()
	if err != nil && err != redis.Nil {
		return false, fmt.Errorf("get attempts: %w", err)
	}
	if attempts >= MaxAttempts {
		// 已超限，确保验证码作废。
		s.rdb.Del(ctx, s.codeKey(email), s.attemptsKey(email))
		return false, ErrTooManyAttempts
	}

	// 原子校验（比对即删）。
	res, err := s.rdb.Eval(ctx, verifyScript, []string{s.codeKey(email)}, code).Int64()
	if err != nil {
		return false, fmt.Errorf("verify reset code: %w", err)
	}
	if res == 1 {
		// 校验通过，清理尝试计数。
		s.rdb.Del(ctx, s.attemptsKey(email))
		return true, nil
	}

	// 校验失败：累加尝试次数。首次失败时设置与验证码同 TTL。
	attemptsKey := s.attemptsKey(email)
	incr := s.rdb.Incr(ctx, attemptsKey)
	if err := incr.Err(); err != nil {
		return false, fmt.Errorf("incr attempts: %w", err)
	}
	// 仅在首次（值变为 1）时设置过期，避免每次重置 TTL 延长窗口。
	if incr.Val() == 1 {
		s.rdb.Expire(ctx, attemptsKey, s.ttl)
	}
	// 达到上限则作废验证码，切断后续尝试。
	if int(incr.Val()) >= MaxAttempts {
		s.rdb.Del(ctx, s.codeKey(email))
		return false, ErrTooManyAttempts
	}
	return false, nil
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
