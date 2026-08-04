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

	log              *log.Helper
	repo             *data.PaymentRefundRepo
	transactionRepo  *data.PaymentTransactionRepo
}

func NewPaymentRefundService(
	ctx *bootstrap.Context,
	repo *data.PaymentRefundRepo,
	transactionRepo *data.PaymentTransactionRepo,
) *PaymentRefundService {
	return &PaymentRefundService{
		log:             ctx.NewLoggerHelper("payment-refund/service/core-service"),
		repo:            repo,
		transactionRepo: transactionRepo,
	}
}

func (s *PaymentRefundService) List(ctx context.Context, req *paginationV1.PagingRequest) (*paymentV1.ListPaymentRefundResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *PaymentRefundService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*paymentV1.CountPaymentRefundResponse, error) {
	count, err := s.repo.Count(ctx, req)
	if err != nil {
		return nil, err
	}

	return &paymentV1.CountPaymentRefundResponse{
		Count: uint64(count),
	}, nil
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

	// 退款状态机校验：仅允许 PENDING→SUCCEEDED 与 PENDING→FAILED。
	// 退款记录当前状态需先查库。
	targetStatus := req.Data.GetStatus()
	if targetStatus != paymentV1.PaymentRefund_STATUS_UNSPECIFIED {
		existing, gErr := s.repo.Get(ctx, &paymentV1.GetPaymentRefundRequest{
			QueryBy: &paymentV1.GetPaymentRefundRequest_Id{Id: req.GetId()},
		})
		if gErr != nil {
			return nil, gErr
		}
		if existing.GetStatus() != paymentV1.PaymentRefund_PENDING {
			return nil, paymentV1.ErrorBadRequest("refund cannot transition from %v to %v", existing.GetStatus(), targetStatus)
		}
		if targetStatus != paymentV1.PaymentRefund_SUCCEEDED && targetStatus != paymentV1.PaymentRefund_FAILED {
			return nil, paymentV1.ErrorBadRequest("illegal refund status transition to %v", targetStatus)
		}
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	// 退款成功联动：将关联的支付流水翻为 REFUNDED。
	// 仅当目标状态为 SUCCEEDED 且退款记录携带 transaction_id 时执行。
	if targetStatus == paymentV1.PaymentRefund_SUCCEEDED {
		txId := req.Data.GetTransactionId()
		if txId != 0 {
			txData := &paymentV1.PaymentTransaction{
				Status: paymentV1.PaymentTransaction_REFUNDED.Enum(),
			}
			updateTxReq := &paymentV1.UpdatePaymentTransactionRequest{
				Id:   txId,
				Data: txData,
			}
			if uErr := s.transactionRepo.Update(ctx, updateTxReq); uErr != nil {
				s.log.Errorf("flip payment_transaction [%d] to REFUNDED failed: %v", txId, uErr)
				return nil, uErr
			}
		}
	}

	return &emptypb.Empty{}, nil
}

func (s *PaymentRefundService) Delete(ctx context.Context, req *paymentV1.DeletePaymentRefundRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
