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
package email

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-kratos/kratos/v2/log"
)

const (
	defaultFromName = "GoWind Shop"
	defaultTimeout  = 15 * time.Second
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
// 未配置 SMTP 时降级：把正文写入日志（含验证码），方便开发环境查看。
func (s *Sender) Send(to, subject, body string) error {
	if !s.IsConfigured() {
		// 降级模式：不发送，仅记录日志。验证码会出现在日志里，便于开发/演示。
		s.log.Warnf("[email degraded] SMTP not configured, skip sending. to=%s subject=%s\n%s", to, subject, body)
		return nil
	}
	return s.sendSMTP(to, subject, body)
}

func (s *Sender) sendSMTP(to, subject, body string) error {
	addr := net.JoinHostPort(s.cfg.Host, strconv.Itoa(s.cfg.Port))
	auth := smtp.PlainAuth("", s.cfg.User, s.cfg.Password, s.cfg.Host)

	headers := map[string]string{
		"From":         fmt.Sprintf("%s <%s>", s.cfg.FromName, s.cfg.From),
		"To":           to,
		"Subject":      subject,
		"MIME-Version": "1.0",
		"Content-Type": "text/plain; charset=UTF-8",
		"Date":         time.Now().Format(time.RFC1123Z),
	}

	var msg strings.Builder
	for k, v := range headers {
		fmt.Fprintf(&msg, "%s: %s\r\n", k, v)
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)

	// 端口 465 用隐式 TLS，其余用 STARTTLS（PLAIN AUTH 前尝试升级）。
	if s.cfg.Port == 465 {
		return s.sendImplicitTLS(addr, auth, []string{to}, []byte(msg.String()))
	}
	return smtp.SendMail(addr, auth, s.cfg.From, []string{to}, []byte(msg.String()))
}

func (s *Sender) sendImplicitTLS(addr string, auth smtp.Auth, to []string, msg []byte) error {
	dialer := &net.Dialer{Timeout: defaultTimeout}
	conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: s.cfg.Host})
	if err != nil {
		return fmt.Errorf("smtp tls dial: %w", err)
	}
	defer conn.Close()

	c, err := smtp.NewClient(conn, s.cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer c.Close()

	if err = c.Auth(auth); err != nil {
		return fmt.Errorf("smtp auth: %w", err)
	}
	if err = c.Mail(s.cfg.From); err != nil {
		return fmt.Errorf("smtp mail: %w", err)
	}
	for _, rcpt := range to {
		if err = c.Rcpt(rcpt); err != nil {
			return fmt.Errorf("smtp rcpt: %w", err)
		}
	}
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("smtp close: %w", err)
	}
	return c.Quit()
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
