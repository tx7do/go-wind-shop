package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	orderV1 "go-wind-shop/api/gen/go/order/service/v1"
	paymentV1 "go-wind-shop/api/gen/go/payment/service/v1"
)

type PaymentTransactionService struct {
	paymentV1.UnimplementedPaymentTransactionServiceServer

	log          *log.Helper
	repo         *data.PaymentTransactionRepo
	orderService *OrderService
}

func NewPaymentTransactionService(
	ctx *bootstrap.Context,
	repo *data.PaymentTransactionRepo,
	orderService *OrderService,
) *PaymentTransactionService {
	return &PaymentTransactionService{
		log:          ctx.NewLoggerHelper("payment-transaction/service/core-service"),
		repo:         repo,
		orderService: orderService,
	}
}

func (s *PaymentTransactionService) List(ctx context.Context, req *paginationV1.PagingRequest) (*paymentV1.ListPaymentTransactionResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *PaymentTransactionService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*paymentV1.CountPaymentTransactionResponse, error) {
	count, err := s.repo.Count(ctx, req)
	if err != nil {
		return nil, err
	}

	return &paymentV1.CountPaymentTransactionResponse{
		Count: uint64(count),
	}, nil
}

func (s *PaymentTransactionService) Get(ctx context.Context, req *paymentV1.GetPaymentTransactionRequest) (*paymentV1.PaymentTransaction, error) {
	return s.repo.Get(ctx, req)
}

// Create 创建支付流水。
//
// MVP 阶段支付网关为 STUB：创建即"成功"——流水状态直接置 SUCCEEDED，
// 并在进程内推进关联订单状态 PENDING_PAYMENT→PAID（带 expected_status 前置条件，
// 乐观并发防覆盖）。真实网关（Stripe/PayPal webhook）接入后，此处改为创建 PENDING 流水，
// 由 webhook 回调更新 SUCCEEDED 再触发 MarkPaid。
func (s *PaymentTransactionService) Create(ctx context.Context, req *paymentV1.CreatePaymentTransactionRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, paymentV1.ErrorBadRequest("invalid parameter")
	}

	// 强制置 SUCCEEDED（MVP stub，忽略客户端传入）。
	succeeded := paymentV1.PaymentTransaction_SUCCEEDED
	req.Data.Status = &succeeded

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	// 进程内推进关联订单为 PAID（MarkPaid 路径）。
	// 仅当流水携带 order_id 时触发。expected_status 前置条件防并发覆盖终态。
	orderId := req.Data.GetOrderId()
	if orderId != 0 {
		paid := orderV1.Order_PAID
		updateReq := &orderV1.UpdateOrderRequest{
			Id: orderId,
			Data: &orderV1.Order{
				Status: &paid,
			},
			ExpectedStatus: []orderV1.Order_Status{orderV1.Order_PENDING_PAYMENT},
		}
		if _, uErr := s.orderService.Update(ctx, updateReq); uErr != nil {
			s.log.Errorf("mark order [%d] as PAID after payment stub failed: %v", orderId, uErr)
			// 支付已"成功"但订单状态推进失败。MVP 不退款（stub 无真实扣款），
			// 订单保持 PENDING_PAYMENT 由超时任务兜底取消。
		}
	}

	return &emptypb.Empty{}, nil
}

func (s *PaymentTransactionService) Update(ctx context.Context, req *paymentV1.UpdatePaymentTransactionRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, paymentV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *PaymentTransactionService) Delete(ctx context.Context, req *paymentV1.DeletePaymentTransactionRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
