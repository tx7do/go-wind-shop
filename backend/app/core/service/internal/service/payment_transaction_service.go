package service

import (
	"context"
	"fmt"

	"go-wind-shop/pkg/logx"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	orderV1 "go-wind-shop/api/gen/go/order/service/v1"
	paymentV1 "go-wind-shop/api/gen/go/payment/service/v1"

	"github.com/tx7do/go-crud/viewer"
)

type PaymentTransactionService struct {
	paymentV1.UnimplementedPaymentTransactionServiceServer

	log *logx.Helper
	repo         *data.PaymentTransactionRepo
	orderService *OrderService
	channelConfigService *PaymentChannelConfigService
}

func NewPaymentTransactionService(
	ctx *bootstrap.Context,
	repo *data.PaymentTransactionRepo,
	orderService *OrderService,
	channelConfigService *PaymentChannelConfigService,
) *PaymentTransactionService {
	return &PaymentTransactionService{
		log:                  logx.New(ctx.NewLoggerHelper("payment-transaction/service/core-service")),
		repo:                 repo,
		orderService:         orderService,
		channelConfigService: channelConfigService,
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

// NotifyOrderRefundSuccess 转发到 OrderService 发送退款成功站内通知（best-effort，
// 通知失败不影响退款状态）。amount 为最小货币单位（分）。
func (s *PaymentTransactionService) NotifyOrderRefundSuccess(ctx context.Context, orderId uint32, amount int64) {
	s.orderService.notifyOrderRefundSuccess(ctx, orderId, amount)
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

// channelPaymentMethod 渠道类型 → 流水支付方式枚举映射。
func channelPaymentMethod(channel paymentV1.PaymentChannelType) paymentV1.PaymentMethod {
	switch channel {
	case paymentV1.PaymentChannelType_PAYMENT_CHANNEL_WECHAT_NATIVE:
		return paymentV1.PaymentMethod_WECHAT
	case paymentV1.PaymentChannelType_PAYMENT_CHANNEL_ALIPAY_PAGE:
		return paymentV1.PaymentMethod_ALIPAY
	case paymentV1.PaymentChannelType_PAYMENT_CHANNEL_MOCK:
		return paymentV1.PaymentMethod_MOCK
	default:
		return paymentV1.PaymentMethod_PAYMENT_METHOD_UNSPECIFIED
	}
}

// CreatePaymentIntent 创建支付意图：校验订单归属与待支付态 → 幂等落 PENDING 流水
// → 带回渠道配置（凭据解密，仅经 core→BFF gRPC 内网传递）。
//
// 网关下单（二维码/跳转链接）由 BFF 执行；本方法不外联任何网关。
// 金额取服务端订单 total_amount，不信任客户端。
func (s *PaymentTransactionService) CreatePaymentIntent(ctx context.Context, req *paymentV1.CreatePaymentIntentRequest) (*paymentV1.CreatePaymentIntentResponse, error) {
	if req == nil || req.GetOrderId() == 0 || req.GetChannelId() == 0 || req.GetUserId() == 0 {
		return nil, paymentV1.ErrorBadRequest("order_id, channel_id, user_id are required")
	}
	if req.GetIdempotencyKey() == "" {
		return nil, paymentV1.ErrorBadRequest("idempotency_key is required")
	}

	cfg, credJSON, err := s.channelConfigService.GetDecryptedCredentials(ctx, req.GetChannelId())
	if err != nil {
		return nil, err
	}
	// 凭据以内网明文形式经 gRPC 下发 BFF（C 端响应模型不含 credentials 字段）
	cfg.Credentials = trans.Ptr(credJSON)

	order, gErr := s.orderService.Get(ctx, &orderV1.GetOrderRequest{
		QueryBy: &orderV1.GetOrderRequest_Id{Id: req.GetOrderId()},
	})
	if gErr != nil || order == nil {
		return nil, paymentV1.ErrorBadRequest("order not found")
	}
	if order.GetStatus() != orderV1.Order_PENDING_PAYMENT {
		return nil, paymentV1.ErrorBadRequest("order [%d] is not pending payment (current: %v)", order.GetId(), order.GetStatus())
	}
	if order.GetUserId() != req.GetUserId() {
		return nil, paymentV1.ErrorForbidden("order does not belong to current user")
	}

	// 幂等：同租户同幂等键已有流水（重试/回调竞态）直接返回既有流水。
	if existing := s.findByIdempotencyKey(ctx, req.GetIdempotencyKey()); existing != nil {
		return &paymentV1.CreatePaymentIntentResponse{
			Transaction: existing,
			Channel:     cfg,
			Subject:     paymentSubject(order),
		}, nil
	}

	currency := order.GetCurrency()
	if currency == "" {
		currency = "CNY"
	}
	status := paymentV1.PaymentTransaction_PENDING
	tx := &paymentV1.PaymentTransaction{
		OrderId:        order.Id,
		UserId:         trans.Ptr(req.GetUserId()),
		Amount:         trans.Ptr(order.GetTotalAmount()),
		Status:         &status,
		Currency:       &currency,
		BusinessRefId:  order.BusinessRefId,
		IdempotencyKey: trans.Ptr(req.GetIdempotencyKey()),
		PaymentMethod:  channelPaymentMethod(cfg.GetChannel()).Enum(),
		BusinessType:   paymentV1.BusinessType_BUSINESS_TYPE_CONSUME.Enum(),
		TenantId:       order.TenantId,
		CreatedBy:      trans.Ptr(req.GetUserId()),
	}
	if err := s.repo.Create(ctx, &paymentV1.CreatePaymentTransactionRequest{Data: tx}); err != nil {
		return nil, err
	}

	// Create 返回 empty，按幂等键回查拿自增 ID（并发下也天然取到唯一行）。
	created := s.findByIdempotencyKey(ctx, req.GetIdempotencyKey())
	if created == nil {
		return nil, paymentV1.ErrorInternalServerError("payment transaction lookup after create failed")
	}

	s.log.Infof("payment intent created: tx=%d order=%d channel=%s amount=%d user=%d",
		created.GetId(), order.GetId(), cfg.GetChannel(), created.GetAmount(), req.GetUserId())

	return &paymentV1.CreatePaymentIntentResponse{
		Transaction: created,
		Channel:     cfg,
		Subject:     paymentSubject(order),
	}, nil
}

// ConfirmPayment 确认支付：BFF 网关回调验签后 / MOCK 渠道确认时调用。
// 幂等：已 SUCCEEDED 直接幂等成功；金额不符拒绝；成功后推进订单 PAID。
func (s *PaymentTransactionService) ConfirmPayment(ctx context.Context, req *paymentV1.ConfirmPaymentRequest) (*paymentV1.ConfirmPaymentResponse, error) {
	if req == nil || req.GetTransactionId() == 0 {
		return nil, paymentV1.ErrorBadRequest("transaction_id is required")
	}

	tx, gErr := s.repo.Get(ctx, &paymentV1.GetPaymentTransactionRequest{
		QueryBy: &paymentV1.GetPaymentTransactionRequest_Id{Id: req.GetTransactionId()},
	})
	if gErr != nil || tx == nil {
		return nil, paymentV1.ErrorBadRequest("payment transaction not found")
	}

	// 幂等重放：网关会重复回调，已成功的流水直接幂等成功。
	if tx.GetStatus() == paymentV1.PaymentTransaction_SUCCEEDED {
		return &paymentV1.ConfirmPaymentResponse{
			TransactionId: tx.GetId(),
			Status:        tx.GetStatus(),
			OrderId:       tx.GetOrderId(),
			FirstSuccess:  false,
		}, nil
	}
	if tx.GetStatus() != paymentV1.PaymentTransaction_PENDING {
		return nil, paymentV1.ErrorBadRequest("payment transaction [%d] is not confirmable (current: %v)", tx.GetId(), tx.GetStatus())
	}

	// 金额核验：回调金额 ≠ 应付金额 → 拒绝确认（保留 PENDING 供人工排查）。
	if req.GetPaidAmount() != tx.GetAmount() {
		s.log.Warnf("payment amount mismatch: tx=%d expect=%d got=%d gateway_no=%s",
			tx.GetId(), tx.GetAmount(), req.GetPaidAmount(), req.GetGatewayTradeNo())
		return nil, paymentV1.ErrorBadRequest("paid amount mismatch")
	}

	succeeded := paymentV1.PaymentTransaction_SUCCEEDED
	if _, uErr := s.Update(ctx, &paymentV1.UpdatePaymentTransactionRequest{
		Id: req.GetTransactionId(),
		Data: &paymentV1.PaymentTransaction{
			Status:         &succeeded,
			GatewayTradeNo: trans.Ptr(req.GetGatewayTradeNo()),
		},
	}); uErr != nil {
		return nil, uErr
	}

	// 推进关联订单 PENDING_PAYMENT→PAID（expected_status 乐观并发防覆盖，
	// 与超时取消/确认收货的并发路径互斥）。失败不回滚流水——钱已收，订单状态
	// 由人工/对账处理（如订单已被超时取消，走退款流程）。
	orderId := tx.GetOrderId()
	if orderId != 0 {
		paid := orderV1.Order_PAID
		if _, uErr := s.orderService.Update(ctx, &orderV1.UpdateOrderRequest{
			Id:             orderId,
			Data:           &orderV1.Order{Status: &paid},
			ExpectedStatus: []orderV1.Order_Status{orderV1.Order_PENDING_PAYMENT},
		}); uErr != nil {
			s.log.Errorf("mark order [%d] as PAID after payment confirm failed: %v", orderId, uErr)
		}
	}

	s.log.Infof("payment confirmed: tx=%d order=%d amount=%d gateway_no=%s first=true",
		tx.GetId(), orderId, tx.GetAmount(), req.GetGatewayTradeNo())

	return &paymentV1.ConfirmPaymentResponse{
		TransactionId: tx.GetId(),
		Status:        succeeded,
		OrderId:       tx.GetOrderId(),
		FirstSuccess:  true,
	}, nil
}

// findByIdempotencyKey 按幂等键查流水（租户隔离由隐私层按 viewer 注入）。
func (s *PaymentTransactionService) findByIdempotencyKey(ctx context.Context, key string) *paymentV1.PaymentTransaction {
	ret, err := s.repo.List(ctx, &paginationV1.PagingRequest{
		NoPaging: trans.Ptr(true),
		FilteringType: &paginationV1.PagingRequest_Query{
			Query: `{"idempotency_key":"` + key + `"}`,
		},
	})
	if err != nil || ret == nil || len(ret.GetItems()) == 0 {
		return nil
	}
	return ret.Items[0]
}

// paymentSubject 网关商品标题（订单摘要）。
func paymentSubject(order *orderV1.Order) string {
	ref := order.GetBusinessRefId()
	if ref == "" {
		ref = fmt.Sprintf("%d", order.GetId())
	}
	return fmt.Sprintf("订单支付 %s", ref)
}
