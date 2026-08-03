package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	paymentV1 "go-wind-shop/api/gen/go/payment/service/v1"
)

type PaymentRefundService struct {
	paymentV1.UnimplementedPaymentRefundServiceServer

	log  *log.Helper
	repo *data.PaymentRefundRepo
}

func NewPaymentRefundService(ctx *bootstrap.Context, repo *data.PaymentRefundRepo) *PaymentRefundService {
	return &PaymentRefundService{
		log:  ctx.NewLoggerHelper("payment-refund/service/core-service"),
		repo: repo,
	}
}

func (s *PaymentRefundService) List(ctx context.Context, req *paginationV1.PagingRequest) (*paymentV1.ListPaymentRefundResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *PaymentRefundService) Get(ctx context.Context, req *paymentV1.GetPaymentRefundRequest) (*paymentV1.PaymentRefund, error) {
	return s.repo.Get(ctx, req)
}

func (s *PaymentRefundService) Create(ctx context.Context, req *paymentV1.CreatePaymentRefundRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, paymentV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *PaymentRefundService) Update(ctx context.Context, req *paymentV1.UpdatePaymentRefundRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, paymentV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *PaymentRefundService) Delete(ctx context.Context, req *paymentV1.DeletePaymentRefundRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
