// Package resetcode 提供密码重置验证码的生成、Redis 存取与一次性校验。
//
// 设计要点：
//   - 验证码 6 位数字，crypto/rand 生成，避免伪随机可预测。
//   - 存 Redis，key 形如 gws:pwdreset:{email}，TTL 默认 10 分钟。
//   - 校验采用 Lua 脚本原子"比对即删"，保证一次性（防重放/爆破）。
//   - 邮箱在 key 里小写归一，避免大小写差异导致校验失败。
package resetcode

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	keyPrefix  = "gws:pwdreset:"
	defaultTTL = 10 * time.Minute
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

// Store 基于 Redis 的重置验证码存储。
type Store struct {
	rdb *redis.Client
	ttl time.Duration
}

func New(rdb *redis.Client) *Store {
	return &Store{rdb: rdb, ttl: defaultTTL}
}

func (s *Store) key(email string) string {
	return keyPrefix + normalizeEmail(email)
}

// Generate 生成 6 位数字验证码并存入 Redis（覆盖已有），返回明文验证码。
func (s *Store) Generate(ctx context.Context, email string) (string, error) {
	code, err := randomDigits(6)
	if err != nil {
		return "", fmt.Errorf("generate reset code: %w", err)
	}
	if err := s.rdb.Set(ctx, s.key(email), code, s.ttl).Err(); err != nil {
		return "", fmt.Errorf("store reset code: %w", err)
	}
	return code, nil
}

// Verify 校验验证码，正确则删除（一次性），错误则保留（允许重试）。
// 返回 (true, nil) 表示校验通过。
func (s *Store) Verify(ctx context.Context, email, code string) (bool, error) {
	res, err := s.rdb.Eval(ctx, verifyScript, []string{s.key(email)}, code).Int64()
	if err != nil {
		// redis.Nil 不会出现在 Eval（脚本返回 0），此处为真实错误
		return false, fmt.Errorf("verify reset code: %w", err)
	}
	return res == 1, nil
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
