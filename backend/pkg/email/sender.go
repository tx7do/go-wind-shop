// Package email 提供轻量邮件发送能力，用于找回密码等通知场景。
//
// 配置通过环境变量读取（12-factor 风格，零侵入 config proto）：
//
//	SMTP_HOST   SMTP 服务器地址（如 smtp.example.com）
//	SMTP_PORT   SMTP 端口（如 25 / 465 / 587）
//	SMTP_USER   登录账号
//	SMTP_PASS   登录密码/授权码
//	SMTP_FROM   发件人地址（如 noreply@example.com）
//	SMTP_FROM_NAME 发件人名称（可选，默认 "GoWind Shop"）
//
// 当 SMTP_HOST 未配置时，Sender 处于降级模式：不真正发邮件，而是把验证码
// 写入日志（WARN 级）。这样开发/演示环境无需 SMTP 即可跑通完整找回流程；
// 生产环境配置好环境变量后自动切换为真实发送。
//
// SMTP 传输层（TLS 模式/认证/报文拼装）由 github.com/tx7do/go-utils/mailer 承担；
// 本包只管环境变量解析、降级策略与脱敏工具。
package email

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/go-utils/mailer"
)

const (
	defaultFromName = "GoWind Shop"

	// implicitTLSPort 全程 TLS 的端口；其余端口由 mailer 明文建连后自动 STARTTLS 升级。
	implicitTLSPort = 465
)

// Config 从环境变量构建的 SMTP 配置。Host 为空表示未配置（降级模式）。
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
	FromName string
}

// LoadConfig 从环境变量读取 SMTP 配置。Host 为空表示未配置。
func LoadConfig() Config {
	cfg := Config{
		Host:     strings.TrimSpace(os.Getenv("SMTP_HOST")),
		User:     os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASS"),
		From:     strings.TrimSpace(os.Getenv("SMTP_FROM")),
		FromName: strings.TrimSpace(os.Getenv("SMTP_FROM_NAME")),
	}
	if cfg.Port, _ = strconv.Atoi(os.Getenv("SMTP_PORT")); cfg.Port == 0 {
		cfg.Port = 587
	}
	if cfg.FromName == "" {
		cfg.FromName = defaultFromName
	}
	if cfg.From == "" && cfg.User != "" {
		cfg.From = cfg.User
	}
	return cfg
}

// Configured 返回是否配置了 SMTP（决定真实发送还是降级日志）。
func (c Config) Configured() bool {
	return c.Host != "" && c.From != ""
}

// Sender 邮件发送器。
type Sender struct {
	cfg Config
	log *log.Helper
}

func NewSender(cfg Config, logger *log.Helper) *Sender {
	return &Sender{cfg: cfg, log: logger}
}

// IsConfigured 是否处于真实发送模式。
func (s *Sender) IsConfigured() bool {
	return s.cfg.Configured()
}

// Send 发送一封纯文本邮件到 to。
// 未配置 SMTP 时降级：不发送，仅记录收件人与主题（不记录正文，避免验证码等
// 敏感信息泄露到日志）。开发环境如需查看验证码，应配置真实 SMTP 或在调用方
// 自行按需打印。
func (s *Sender) Send(to, subject, body string) error {
	if !s.IsConfigured() {
		s.log.Warnf("[email degraded] SMTP not configured, skip sending. to=%s subject=%s", to, subject)
		return nil
	}
	// mailer.SendMail 的 ctx 贯穿 TCP 建连：SMTP 不可达时不会卡到 OS 级连接超时。
	// 本包对外维持无 ctx 的既有签名，调用方是找回密码的同步链路，Background 足够。
	return mailer.SendMail(context.Background(), mailer.SmtpConfig{
		Host:     s.cfg.Host,
		Port:     uint32(s.cfg.Port),
		Username: s.cfg.User,
		Password: s.cfg.Password,
		From:     s.cfg.From,
		FromName: s.cfg.FromName,
		TlsMode:  tlsModeOf(s.cfg.Port),
	}, []string{to}, subject, body)
}

// tlsModeOf 按端口推断加密方式：465 全程 TLS，其余由 mailer 自动 STARTTLS 升级。
func tlsModeOf(port int) string {
	if port == implicitTLSPort {
		return "SSL"
	}
	return "START_TLS"
}

// MaskEmail 邮箱脱敏：保留首字符与域名，中间用 *** 替代。
// 如 alice@example.com → a***@example.com
func MaskEmail(email string) string {
	at := strings.LastIndex(email, "@")
	if at <= 0 || at == len(email)-1 {
		return email
	}
	local := email[:at]
	domain := email[at:]
	if len(local) <= 1 {
		return local + "***" + domain
	}
	return string(local[0]) + "***" + domain
}

// ErrNotConfigured SMTP 未配置时调用真实发送会返回此错误（内部守卫）。
var ErrNotConfigured = errors.New("email: SMTP not configured")
