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

	"github.com/tx7do/go-crud/viewer"
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

// RestoreOrderStockForRefund 转发到 OrderService.RestoreStockForRefund。
//
// 退款成功后需回补下单时扣减的库存。PaymentRefundService 已注入本 service，
// 由本方法转发到 OrderService，避免 PaymentRefundService 直接依赖 OrderService
// （OrderService 已被 PaymentTransactionService 依赖，本转发维持原依赖方向，
// 不引入循环）。
func (s *PaymentTransactionService) RestoreOrderStockForRefund(ctx context.Context, orderId uint32) error {
	return s.orderService.RestoreStockForRefund(ctx, orderId)
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

	// 支付流水状态机校验：状态变更须先查库确认当前状态，仅允许白名单内转换。
	//   SUCCEEDED → REFUNDED   （退款联动，由 PaymentRefundService.Update 调用）
	//   PENDING → FAILED/SUCCEEDED  （支付回调，真实网关接入后）
	//   SUCCEEDED → FAILED  （仅系统上下文：ExpireOrderByTimeout 的孤儿支付清理，
	//     消除"支付 stub 成功但订单已取消"的遗留流水。普通调用方拒绝此转换。）
	// 其他转换（含 REFUNDED→任意、FAILED→任意）拒绝。
	// 这防止退款联动把流水翻 REFUNDED 后，又被并发 Update 翻回 SUCCEEDED。
	// 系统上下文判断由 isAllowedTxTransition 内 viewer.FromContext 完成，
	// 消除"状态机只守 service.Update 入口、而清理/联动走 ent builder 旁路"的缺口。
	targetStatus := req.Data.GetStatus()
	if targetStatus != paymentV1.PaymentTransaction_STATUS_UNSPECIFIED && req.GetId() != 0 {
		existing, gErr := s.repo.Get(ctx, &paymentV1.GetPaymentTransactionRequest{
			QueryBy: &paymentV1.GetPaymentTransactionRequest_Id{Id: req.GetId()},
		})
		if gErr != nil || existing == nil {
			return nil, paymentV1.ErrorBadRequest("payment transaction not found")
		}
		if !isAllowedTxTransition(ctx, existing.GetStatus(), targetStatus) {
			s.log.Warnf("illegal payment transaction status transition: %v -> %v", existing.GetStatus(), targetStatus)
			return nil, paymentV1.ErrorBadRequest("illegal payment transaction status transition: %v -> %v", existing.GetStatus(), targetStatus)
		}
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// isAllowedTxTransition 校验支付流水状态转换是否在允许列表内。
// 普通调用方仅允许：
//   PENDING → SUCCEEDED / FAILED   （支付回调）
//   SUCCEEDED → REFUNDED            （退款联动）
// 系统上下文（ExpireOrderByTimeout 注入的 SystemViewer）额外允许：
//   SUCCEEDED → FAILED  （孤儿支付清理）
// 终态 REFUNDED/FAILED 对所有调用方均为吸收态无出边。
func isAllowedTxTransition(ctx context.Context, from, to paymentV1.PaymentTransaction_Status) bool {
	allowed := map[paymentV1.PaymentTransaction_Status]map[paymentV1.PaymentTransaction_Status]bool{
		paymentV1.PaymentTransaction_PENDING: {
			paymentV1.PaymentTransaction_SUCCEEDED: true,
			paymentV1.PaymentTransaction_FAILED:    true,
		},
		paymentV1.PaymentTransaction_SUCCEEDED: {
			paymentV1.PaymentTransaction_REFUNDED: true,
		},
	}
	// 系统上下文口子：ExpireOrderByTimeout 的孤儿支付清理需 SUCCEEDED→FAILED。
	vc, exist := viewer.FromContext(ctx)
	if exist && vc.IsSystemContext() {
		if allowed[paymentV1.PaymentTransaction_SUCCEEDED] == nil {
			allowed[paymentV1.PaymentTransaction_SUCCEEDED] = map[paymentV1.PaymentTransaction_Status]bool{}
		}
		allowed[paymentV1.PaymentTransaction_SUCCEEDED][paymentV1.PaymentTransaction_FAILED] = true
	}
	allowedTo, ok := allowed[from]
	if !ok {
		return false
	}
	return allowedTo[to]
}

func (s *PaymentTransactionService) Delete(ctx context.Context, req *paymentV1.DeletePaymentTransactionRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
