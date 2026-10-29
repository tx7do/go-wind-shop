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

// paymentMethodChannels 支付方式 → 渠道类型候选列表（按优先序）。
// WECHAT 对应 Native/H5 两种渠道配置（同商户资质），任一启用即可服务
// 关单/退款/查单；与 channelPaymentMethod 互逆。
func paymentMethodChannels(pm paymentV1.PaymentMethod) []paymentV1.PaymentChannelType {
	switch pm {
	case paymentV1.PaymentMethod_WECHAT:
		return []paymentV1.PaymentChannelType{
			paymentV1.PaymentChannelType_PAYMENT_CHANNEL_WECHAT_NATIVE,
			paymentV1.PaymentChannelType_PAYMENT_CHANNEL_WECHAT_H5,
		}
	case paymentV1.PaymentMethod_ALIPAY:
		return []paymentV1.PaymentChannelType{paymentV1.PaymentChannelType_PAYMENT_CHANNEL_ALIPAY_PAGE}
	case paymentV1.PaymentMethod_MOCK:
		return []paymentV1.PaymentChannelType{paymentV1.PaymentChannelType_PAYMENT_CHANNEL_MOCK}
	default:
		// BALANCE/CREDIT/OFFLINE 等记账型流水无网关语义
		return nil
	}
}

// ByPaymentMethod 按支付方式解析网关。非网关渠道返回 (nil, nil)，调用方跳过联动。
// 渠道候选按序取第一条启用配置。解析失败（渠道未配置/未启用/凭据不可用）返回
// 错误——取消关单按 best-effort 记日志，退款联动则向调用方报错（不得在拿不到
// 渠道时静默标记退款成功）。
func (r *PaymentGatewayResolver) ByPaymentMethod(ctx context.Context, pm paymentV1.PaymentMethod) (gateway.Gateway, error) {
	candidates := paymentMethodChannels(pm)
	if len(candidates) == 0 {
		return nil, nil
	}

	var lastErr error
	for _, channel := range candidates {
		cfg, err := r.channelConfigService.ResolveEnabledChannel(ctx, &paymentV1.ResolveEnabledChannelRequest{
			Channel: channel,
		})
		if err != nil {
			lastErr = err
			continue
		}

		return gateway.New(cfg.GetChannel(), cfg.GetAppId(), cfg.GetMchId(), cfg.GetSandbox(), cfg.GetCredentials())
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, paymentV1.ErrorNotFound("no enabled payment channel for method %v", pm)
}
