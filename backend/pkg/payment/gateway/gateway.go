// Package gateway 支付网关适配器：为「部署方自带资质」的直连商户模式提供
// 统一下单/回调验签抽象。适配器只负责与渠道网关交互，资金流向为
// 买家 → 支付渠道 → 部署方商户号，本系统不经手任何交易资金。
package gateway

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	paymentV1 "go-wind-shop/api/gen/go/payment/service/v1"
)

// PayRequest 网关下单请求
type PayRequest struct {
	// 商户交易号：本项目固定为支付流水 ID 的十进制字符串（渠道侧需唯一）
	OutTradeNo string
	// 应付金额（分）
	AmountCents int64
	// 商品标题
	Subject string
	// 服务端回调地址（必须公网可达）
	NotifyURL string
}

// PayResult 网关下单结果
type PayResult struct {
	// 支付跳转链接（支付宝电脑网站支付）
	RedirectURL string
	// 二维码内容（微信 Native 的 code_url）
	QRCode string
}

// NotifyResult 网关回调解析结果（验签通过后）
type NotifyResult struct {
	OutTradeNo     string
	GatewayTradeNo string
	PaidCents      int64
	Success        bool
}

// Gateway 支付网关适配器接口
type Gateway interface {
	CreatePayment(ctx context.Context, req PayRequest) (*PayResult, error)
	// ParseNotify 验签并解析异步通知。验签失败必须返回错误（调用方拒绝应答）。
	ParseNotify(ctx context.Context, r *http.Request) (*NotifyResult, error)
}

// New 按渠道类型构造网关适配器。credentialsJSON 为解密后的凭据明文 JSON。
func New(channel paymentV1.PaymentChannelType, appID, mchID string, sandbox bool, credentialsJSON string) (Gateway, error) {
	switch channel {
	case paymentV1.PaymentChannelType_PAYMENT_CHANNEL_MOCK:
		return NewMockGateway(), nil
	case paymentV1.PaymentChannelType_PAYMENT_CHANNEL_WECHAT_NATIVE:
		return NewWechatNativeGateway(appID, mchID, sandbox, credentialsJSON)
	case paymentV1.PaymentChannelType_PAYMENT_CHANNEL_ALIPAY_PAGE:
		return NewAlipayPageGateway(appID, sandbox, credentialsJSON)
	default:
		return nil, fmt.Errorf("unsupported payment channel: %v", channel)
	}
}

// CentsToYuanString 分 → 元字符串（支付宝金额格式，两位小数）。
func CentsToYuanString(cents int64) string {
	return fmt.Sprintf("%d.%02d", cents/100, cents%100)
}

// YuanStringToCents 元字符串 → 分。解析失败返回错误（金额口径必须严谨）。
func YuanStringToCents(s string) (int64, error) {
	parts := strings.SplitN(strings.TrimSpace(s), ".", 2)
	yuan, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount string: %q", s)
	}
	if len(parts) == 1 {
		return yuan * 100, nil
	}
	fracStr := parts[1]
	if len(fracStr) == 1 {
		// "12.3" → 1230 分口径为 12.30
		fracStr += "0"
	}
	if len(fracStr) < 2 {
		return 0, fmt.Errorf("invalid amount fraction: %q", s)
	}
	frac, err := strconv.ParseInt(fracStr[:2], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid amount string: %q", s)
	}
	return yuan*100 + frac, nil
}
