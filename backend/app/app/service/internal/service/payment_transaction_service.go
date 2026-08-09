package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	paymentV1 "go-wind-shop/api/gen/go/payment/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

type PaymentTransactionService struct {
	appV1.PaymentTransactionServiceHTTPServer

	log *log.Helper

	paymentTransactionServiceClient paymentV1.PaymentTransactionServiceClient
}

func NewPaymentTransactionService(
	ctx *bootstrap.Context,
	paymentTransactionServiceClient paymentV1.PaymentTransactionServiceClient,
) *PaymentTransactionService {
	return &PaymentTransactionService{
		log:                             ctx.NewLoggerHelper("payment-transaction/service/app-service"),
		paymentTransactionServiceClient: paymentTransactionServiceClient,
	}
}

func (s *PaymentTransactionService) List(ctx context.Context, req *paginationV1.PagingRequest) (*paymentV1.ListPaymentTransactionResponse, error) {
	return s.paymentTransactionServiceClient.List(ctx, req)
}

func (s *PaymentTransactionService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*paymentV1.CountPaymentTransactionResponse, error) {
	return s.paymentTransactionServiceClient.Count(ctx, req)
}

func (s *PaymentTransactionService) Get(ctx context.Context, req *paymentV1.GetPaymentTransactionRequest) (*paymentV1.PaymentTransaction, error) {
	return s.paymentTransactionServiceClient.Get(ctx, req)
}

func (s *PaymentTransactionService) Create(ctx context.Context, req *paymentV1.CreatePaymentTransactionRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, appV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	return s.paymentTransactionServiceClient.Create(ctx, req)
}
