package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	adminV1 "go-wind-shop/api/gen/go/admin/service/v1"
	paymentV1 "go-wind-shop/api/gen/go/payment/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

type PaymentRefundService struct {
	adminV1.PaymentRefundServiceHTTPServer

	log *log.Helper

	paymentRefundServiceClient paymentV1.PaymentRefundServiceClient
}

func NewPaymentRefundService(
	ctx *bootstrap.Context,
	paymentRefundServiceClient paymentV1.PaymentRefundServiceClient,
) *PaymentRefundService {
	return &PaymentRefundService{
		log:                        ctx.NewLoggerHelper("payment-refund/service/admin-service"),
		paymentRefundServiceClient: paymentRefundServiceClient,
	}
}

func (s *PaymentRefundService) List(ctx context.Context, req *paginationV1.PagingRequest) (*paymentV1.ListPaymentRefundResponse, error) {
	return s.paymentRefundServiceClient.List(ctx, req)
}

func (s *PaymentRefundService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*paymentV1.CountPaymentRefundResponse, error) {
	return s.paymentRefundServiceClient.Count(ctx, req)
}

func (s *PaymentRefundService) Get(ctx context.Context, req *paymentV1.GetPaymentRefundRequest) (*paymentV1.PaymentRefund, error) {
	return s.paymentRefundServiceClient.Get(ctx, req)
}

func (s *PaymentRefundService) Create(ctx context.Context, req *paymentV1.CreatePaymentRefundRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	return s.paymentRefundServiceClient.Create(ctx, req)
}

func (s *PaymentRefundService) Update(ctx context.Context, req *paymentV1.UpdatePaymentRefundRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.Id = trans.Ptr(req.GetId())

	req.Data.UpdatedBy = trans.Ptr(operator.GetUserId())
	if req.UpdateMask != nil {
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "updated_by")
	}

	return s.paymentRefundServiceClient.Update(ctx, req)
}

func (s *PaymentRefundService) Delete(ctx context.Context, req *paymentV1.DeletePaymentRefundRequest) (*emptypb.Empty, error) {
	return s.paymentRefundServiceClient.Delete(ctx, req)
}
