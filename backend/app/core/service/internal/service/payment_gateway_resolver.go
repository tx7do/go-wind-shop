package service

import (
	"context"

	"github.com/tx7do/kratos-bootstrap/bootstrap"

	"go-wind-shop/pkg/logx"
	"go-wind-shop/pkg/payment/gateway"

	paymentV1 "go-wind-shop/api/gen/go/payment/service/v1"
)

// PaymentGatewayResolver 把支付流水/支付方式解析为可执行的网关适配器。
// 供 core 侧取消关单、退款联动使用（与 app BFF 的下单/回调网关共用同一套适配器）。
type PaymentGatewayResolver struct {
	log *logx.Helper

	channelConfigService *PaymentChannelConfigService
}

func NewPaymentGatewayResolver(
	ctx *bootstrap.Context,
	channelConfigService *PaymentChannelConfigService,
) *PaymentGatewayResolver {
	return &PaymentGatewayResolver{
		log:                  logx.New(ctx.NewLoggerHelper("payment-gateway-resolver/service/core-service")),
		channelConfigService: channelConfigService,
	}
}

// paymentMethodChannel 支付方式 → 渠道类型映射（与 channelPaymentMethod 互逆）。
func paymentMethodChannel(pm paymentV1.PaymentMethod) paymentV1.PaymentChannelType {
	switch pm {
	case paymentV1.PaymentMethod_WECHAT:
		return paymentV1.PaymentChannelType_PAYMENT_CHANNEL_WECHAT_NATIVE
	case paymentV1.PaymentMethod_ALIPAY:
		return paymentV1.PaymentChannelType_PAYMENT_CHANNEL_ALIPAY_PAGE
	case paymentV1.PaymentMethod_MOCK:
		return paymentV1.PaymentChannelType_PAYMENT_CHANNEL_MOCK
	default:
		// BALANCE/CREDIT/OFFLINE 等记账型流水无网关语义
		return paymentV1.PaymentChannelType_PAYMENT_CHANNEL_UNSPECIFIED
	}
}

// ByPaymentMethod 按支付方式解析网关。非网关渠道返回 (nil, nil)，调用方跳过联动。
// 解析失败（渠道未配置/未启用/凭据不可用）返回错误——取消关单按 best-effort
// 记日志，退款联动则向调用方报错（不得在拿不到渠道时静默标记退款成功）。
func (r *PaymentGatewayResolver) ByPaymentMethod(ctx context.Context, pm paymentV1.PaymentMethod) (gateway.Gateway, error) {
	channel := paymentMethodChannel(pm)
	if channel == paymentV1.PaymentChannelType_PAYMENT_CHANNEL_UNSPECIFIED {
		return nil, nil
	}

	cfg, err := r.channelConfigService.ResolveEnabledChannel(ctx, &paymentV1.ResolveEnabledChannelRequest{
		Channel: channel,
	})
	if err != nil {
		return nil, err
	}

	return gateway.New(cfg.GetChannel(), cfg.GetAppId(), cfg.GetMchId(), cfg.GetSandbox(), cfg.GetCredentials())
}
