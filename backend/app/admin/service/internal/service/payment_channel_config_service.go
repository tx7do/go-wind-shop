package service

import (
	"context"

	"go-wind-shop/pkg/logx"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	adminV1 "go-wind-shop/api/gen/go/admin/service/v1"
	paymentV1 "go-wind-shop/api/gen/go/payment/service/v1"
)

// PaymentChannelConfigService 支付渠道配置管理（admin BFF，REST → gRPC 转发）。
//
// 凭据安全口径：credentials 写入即加密落库；读取路径（core 层）恒不回填，
// BFF 无需额外处理——这里保持透传即可，密文/明文都不会出现在 REST 响应里。
type PaymentChannelConfigService struct {
	adminV1.PaymentChannelConfigServiceHTTPServer

	log *logx.Helper

	channelConfigServiceClient paymentV1.PaymentChannelConfigServiceClient
}

func NewPaymentChannelConfigService(
	ctx *bootstrap.Context,
	channelConfigServiceClient paymentV1.PaymentChannelConfigServiceClient,
) *PaymentChannelConfigService {
	return &PaymentChannelConfigService{
		log:                        logx.New(ctx.NewLoggerHelper("payment-channel-config/service/admin-service")),
		channelConfigServiceClient: channelConfigServiceClient,
	}
}

func (s *PaymentChannelConfigService) List(ctx context.Context, req *paginationV1.PagingRequest) (*paymentV1.ListPaymentChannelConfigResponse, error) {
	return s.channelConfigServiceClient.List(ctx, req)
}

func (s *PaymentChannelConfigService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*paymentV1.CountPaymentChannelConfigResponse, error) {
	return s.channelConfigServiceClient.Count(ctx, req)
}

func (s *PaymentChannelConfigService) Get(ctx context.Context, req *paymentV1.GetPaymentChannelConfigRequest) (*paymentV1.PaymentChannelConfig, error) {
	return s.channelConfigServiceClient.Get(ctx, req)
}

func (s *PaymentChannelConfigService) Create(ctx context.Context, req *paymentV1.CreatePaymentChannelConfigRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}
	return s.channelConfigServiceClient.Create(ctx, req)
}

func (s *PaymentChannelConfigService) Update(ctx context.Context, req *paymentV1.UpdatePaymentChannelConfigRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}
	if req.GetId() == 0 {
		return nil, adminV1.ErrorBadRequest("id is required")
	}
	return s.channelConfigServiceClient.Update(ctx, req)
}

func (s *PaymentChannelConfigService) Delete(ctx context.Context, req *paymentV1.DeletePaymentChannelConfigRequest) (*emptypb.Empty, error) {
	return s.channelConfigServiceClient.Delete(ctx, req)
}
