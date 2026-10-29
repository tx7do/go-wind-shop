package service

import (
	"context"

	"go-wind-shop/pkg/logx"

	"github.com/tx7do/kratos-bootstrap/bootstrap"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	paymentV1 "go-wind-shop/api/gen/go/payment/service/v1"
)

// PaymentChannelService 支付渠道（app BFF，C 端只读）。
// 仅暴露启用渠道的公开视图（名称/类型/沙箱标记），不含任何配置与凭据细节。
type PaymentChannelService struct {
	appV1.PaymentChannelServiceHTTPServer

	log *logx.Helper

	channelConfigServiceClient paymentV1.PaymentChannelConfigServiceClient
}

func NewPaymentChannelService(
	ctx *bootstrap.Context,
	channelConfigServiceClient paymentV1.PaymentChannelConfigServiceClient,
) *PaymentChannelService {
	return &PaymentChannelService{
		log:                        logx.New(ctx.NewLoggerHelper("payment-channel/app-service")),
		channelConfigServiceClient: channelConfigServiceClient,
	}
}

// ListEnabledChannels 当前启用渠道（结算页渠道选择；登录态——结算本身要求登录）。
func (s *PaymentChannelService) ListEnabledChannels(ctx context.Context, req *paymentV1.ListEnabledChannelsRequest) (*paymentV1.ListEnabledChannelsResponse, error) {
	return s.channelConfigServiceClient.ListEnabledChannels(ctx, req)
}
