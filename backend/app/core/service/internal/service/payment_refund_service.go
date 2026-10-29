package service

import (
	"context"
	"fmt"
	"strconv"

	"go-wind-shop/pkg/logx"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"github.com/tx7do/go-utils/trans"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	paymentV1 "go-wind-shop/api/gen/go/payment/service/v1"
	"go-wind-shop/pkg/payment/gateway"
)

type PaymentRefundService struct {
	paymentV1.UnimplementedPaymentRefundServiceServer

	log *logx.Helper
	repo               *data.PaymentRefundRepo
	transactionRepo    *data.PaymentTransactionRepo
	transactionService *PaymentTransactionService
	gatewayResolver    *PaymentGatewayResolver
}

func NewPaymentRefundService(
	ctx *bootstrap.Context,
	repo *data.PaymentRefundRepo,
	transactionRepo *data.PaymentTransactionRepo,
	transactionService *PaymentTransactionService,
	gatewayResolver *PaymentGatewayResolver,
) *PaymentRefundService {
	return &PaymentRefundService{
		log:                logx.New(ctx.NewLoggerHelper("payment-refund/service/core-service")),
		repo:               repo,
		transactionRepo:    transactionRepo,
		transactionService: transactionService,
		gatewayResolver:    gatewayResolver,
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

	// 退款记录必须关联一条具体的支付流水（transaction_id），且该流水须存在、
	// 当前状态为 SUCCEEDED（只能对已成功的支付发起退款）。
	// 退款记录的 tenant_id 强制取自该流水——退款归属到交易所在租户，
	// 而非调用方 token 的 tenant（平台 admin token tenant=0，若用 token 值会
	// 导致跨租户运营退款被误拒、且退款记录 tenant_id=0 归属错乱）。
	txId := req.Data.GetTransactionId()
	if txId == 0 {
		return nil, paymentV1.ErrorBadRequest("transaction_id is required for refund")
	}
	tx, gErr := s.transactionRepo.Get(ctx, &paymentV1.GetPaymentTransactionRequest{
		QueryBy: &paymentV1.GetPaymentTransactionRequest_Id{Id: txId},
	})
	if gErr != nil || tx == nil {
		return nil, paymentV1.ErrorBadRequest("linked payment transaction not found")
	}
	if tx.GetStatus() != paymentV1.PaymentTransaction_SUCCEEDED {
		return nil, paymentV1.ErrorBadRequest("cannot refund a non-succeeded payment transaction")
	}
	// 退款金额必须为正且不超过实付金额（支持部分退款，累计不超额由渠道侧
	// 幂等与人工对账兜底）。
	if req.Data.GetAmount() <= 0 {
		return nil, paymentV1.ErrorBadRequest("refund amount must be positive")
	}
	if req.Data.GetAmount() > tx.GetAmount() {
		return nil, paymentV1.ErrorBadRequest("refund amount exceeds payment amount")
	}
	txTenantId := tx.GetTenantId()
	req.Data.TenantId = &txTenantId

	// 退款记录的 user_id 强制取自关联支付流水的 user_id。退款归属到交易所属用户，
	// 与调用方 token 的用户无关（平台 admin token user_id=0，若用 token 值会导致
	// 退款记录归属错乱、且 UserPrivacy 会把运营侧的退款拒掉）。
	txUserId := tx.GetUserId()
	req.Data.UserId = &txUserId

	// 重复退款保护：同一 transaction 已存在 SUCCEEDED 退款记录则拒绝。
	// 通过 List refund 过滤 transaction_id=txId 检查（退款表的 transaction_id
	// 仅有普通索引，需遍历结果过滤状态）。NoPaging=true 取全量匹配记录。
	// List 失败时拒绝（fail-closed），不静默降级跳过校验。
	existingRefunds, lErr := s.repo.List(ctx, &paginationV1.PagingRequest{
		NoPaging: trans.Ptr(true),
		FilteringType: &paginationV1.PagingRequest_Query{
			Query: s.buildRefundFilterQuery(txId),
		},
	})
	if lErr != nil {
		return nil, paymentV1.ErrorInternalServerError("duplicate refund check failed")
	}
	if existingRefunds != nil {
		for _, r := range existingRefunds.Items {
			if r.GetTransactionId() != txId {
				continue
			}
			// 已有成功退款 → 拒绝；已有处理中申请 → 拒绝重复提交
			if r.GetStatus() == paymentV1.PaymentRefund_SUCCEEDED {
				return nil, paymentV1.ErrorBadRequest("payment transaction already refunded")
			}
			if r.GetStatus() == paymentV1.PaymentRefund_PENDING {
				return nil, paymentV1.ErrorBadRequest("a pending refund request already exists")
			}
		}
	}

	// 状态机只认 PENDING 起步（Update 白名单仅 PENDING→SUCCEEDED/FAILED），
	// 裸建记录缺省 STATUS_UNSPECIFIED 会永远无法推进，故强制补默认初始态。
	if req.Data.GetStatus() == paymentV1.PaymentRefund_STATUS_UNSPECIFIED {
		req.Data.Status = paymentV1.PaymentRefund_PENDING.Enum()
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// buildRefundFilterQuery 构造按 transaction_id 过滤的 query JSON。
// go-crud 的 query-string filter 把 JSON 里的字段映射到实体列（transaction_id 列存在）。
func (s *PaymentRefundService) buildRefundFilterQuery(txId uint32) string {
	return fmt.Sprintf(`{"transaction_id":%d}`, txId)
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

	// 网关退款联动（推进 SUCCEEDED 的前置门槛）：真实渠道流水必须先拿到渠道
	// 受理才允许推进退款状态；渠道拒绝则整体失败、退款记录停留 PENDING 可重试。
	// 记账型渠道（BALANCE/OFFLINE）与 MOCK 渠道无真实资金，直接放行/受理即成功。
	if targetStatus == paymentV1.PaymentRefund_SUCCEEDED {
		if gErr := s.gatewayRefund(ctx, req.GetId()); gErr != nil {
			return nil, gErr
		}
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	// 退款成功联动：将关联的支付流水翻为 REFUNDED。
	// 仅当目标状态为 SUCCEEDED 且退款记录携带 transaction_id 时执行。
	// 联动前先校验该 transaction 仍为 SUCCEEDED（防已被孤儿清理翻为 FAILED 或
	// 已被并发退款翻为 REFUNDED 时仍盲目翻状态）。
	// 联动走 PaymentTransactionService.Update（经支付流水状态机校验
	// SUCCEEDED→REFUNDED），消除"联动直接打 repo 绕过状态机"的旁路。
	if targetStatus == paymentV1.PaymentRefund_SUCCEEDED {
		// transaction_id 以库中退款记录为准兜底：裸状态更新（仅带 status）也必须
		// 触发联动，不能因请求体缺字段而静默跳过翻 REFUNDED/回补库存。
		txId := req.Data.GetTransactionId()
		if txId == 0 {
			rec, gErr := s.repo.Get(ctx, &paymentV1.GetPaymentRefundRequest{
				QueryBy: &paymentV1.GetPaymentRefundRequest_Id{Id: req.GetId()},
			})
			if gErr == nil && rec != nil {
				txId = rec.GetTransactionId()
			}
		}
		if txId != 0 {
			tx, gErr := s.transactionRepo.Get(ctx, &paymentV1.GetPaymentTransactionRequest{
				QueryBy: &paymentV1.GetPaymentTransactionRequest_Id{Id: txId},
			})
			if gErr != nil || tx == nil {
				s.log.Errorf("linked payment transaction [%d] not found during refund linkage: %v", txId, gErr)
				return nil, paymentV1.ErrorBadRequest("linked payment transaction not found")
			}
			if tx.GetStatus() != paymentV1.PaymentTransaction_SUCCEEDED {
				s.log.Warnf("linked payment transaction [%d] status %v, cannot flip to REFUNDED", txId, tx.GetStatus())
				return nil, paymentV1.ErrorBadRequest("linked payment transaction is not succeeded")
			}
			refundedStatus := paymentV1.PaymentTransaction_REFUNDED
			updateTxReq := &paymentV1.UpdatePaymentTransactionRequest{
				Id: txId,
				Data: &paymentV1.PaymentTransaction{
					Status: &refundedStatus,
				},
			}
			if _, uErr := s.transactionService.Update(ctx, updateTxReq); uErr != nil {
				// 联动失败补偿：退款记录已在前面 repo.Update 翻为 SUCCEEDED，
				// 但关联流水未能翻 REFUNDED。为避免"退款成功但流水未 REFUNDED"
				// 的持久不一致，回滚退款记录为 PENDING。
				//
				// 注意：此补偿调 s.repo.Update 直连 repo 层，有意绕过退款状态机
				// （退款白名单只允许 PENDING→SUCCEEDED/FAILED，SUCCEEDED→PENDING
				// 不在白名单内）。这是联动失败兜底的特权操作——目标是回退到初始态
				// 而非越权提升终态，故可接受绕状态机。当前 repo 模式不支持跨实体
				// 事务（退款记录 Update 与流水翻 REFUNDED 无法同事务），故用此
				// 补偿回滚兜底；补偿本身失败时仅记日志，残留不一致需监控告警补救。
				s.log.Errorf("flip payment_transaction [%d] to REFUNDED via service layer failed: %v", txId, uErr)
				pendingStatus := paymentV1.PaymentRefund_PENDING
				rollbackReq := &paymentV1.UpdatePaymentRefundRequest{
					Id: req.GetId(),
					Data: &paymentV1.PaymentRefund{
						Status: &pendingStatus,
					},
				}
				if rErr := s.repo.Update(ctx, rollbackReq); rErr != nil {
					s.log.Errorf("compensating rollback refund [%d] to PENDING failed: %v", req.GetId(), rErr)
				}
				return nil, uErr
			}

			// 库存回补：退款成功后，把下单时扣减的 SKU 库存加回。
			// 取关联支付流水携带的 order_id，经 transactionService 转发到
			// OrderService.RestoreStockForRefund（事务内遍历 OrderItem +
			// AddStockQty，与超时取消同模式，不改订单状态）。
			//
			// 回补失败补偿：与流水翻 REFUNDED 失败同模式——退款记录已 SUCCEEDED
			// 但库存未回补属持久不一致，回滚退款记录为 PENDING。orderId 为 0
			// （流水无关联订单）时跳过回补。
			orderId := tx.GetOrderId()
			if orderId != 0 {
				if rErr := s.transactionService.RestoreOrderStockForRefund(ctx, orderId); rErr != nil {
					s.log.Errorf("restore stock for order [%d] during refund [%d] failed: %v", orderId, req.GetId(), rErr)
					pendingStatus := paymentV1.PaymentRefund_PENDING
					rollbackReq := &paymentV1.UpdatePaymentRefundRequest{
						Id: req.GetId(),
						Data: &paymentV1.PaymentRefund{
							Status: &pendingStatus,
						},
					}
					if rbErr := s.repo.Update(ctx, rollbackReq); rbErr != nil {
						s.log.Errorf("compensating rollback refund [%d] to PENDING after stock-restore failure failed: %v", req.GetId(), rbErr)
					}
					return nil, rErr
				}

				// 退款成功通知（best-effort）：站内信告知买家退款已处理。
				s.transactionService.NotifyOrderRefundSuccess(ctx, orderId, tx.GetAmount())

				// 积分回冲（best-effort，幂等）：对订单已入账积分对冲。
				s.transactionService.ForwardRefundPointsDeduct(ctx, orderId)
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

// gatewayRefund 对网关渠道流水发起渠道退款（推进 SUCCEEDED 的前置门槛）。
// - 渠道拒绝 → 返回错误，退款记录停留 PENDING（可修数后重试，out_request_no
//   幂等保证渠道侧不重复退）；
// - 记账型渠道（BALANCE/OFFLINE 等）无网关语义 → 直接放行；
// - MOCK 渠道受理即成功（不出真实资金，本地闭环可演示）。
func (s *PaymentRefundService) gatewayRefund(ctx context.Context, refundId uint32) error {
	refund, gErr := s.repo.Get(ctx, &paymentV1.GetPaymentRefundRequest{
		QueryBy: &paymentV1.GetPaymentRefundRequest_Id{Id: refundId},
	})
	if gErr != nil || refund == nil {
		return paymentV1.ErrorBadRequest("refund record not found")
	}

	txId := refund.GetTransactionId()
	tx, txErr := s.transactionRepo.Get(ctx, &paymentV1.GetPaymentTransactionRequest{
		QueryBy: &paymentV1.GetPaymentTransactionRequest_Id{Id: txId},
	})
	if txErr != nil || tx == nil {
		return paymentV1.ErrorBadRequest("linked payment transaction not found")
	}

	gw, rErr := s.gatewayResolver.ByPaymentMethod(ctx, tx.GetPaymentMethod())
	if rErr != nil {
		s.log.Errorf("resolve gateway for refund [%d] failed: %v", refundId, rErr)
		return paymentV1.ErrorInternalServerError("payment gateway unavailable")
	}
	if gw == nil {
		// 记账型渠道：无真实资金流转，放行走既有联动
		return nil
	}

	if rErr := gw.Refund(ctx, gateway.RefundRequest{
		OutTradeNo:   strconv.FormatUint(uint64(txId), 10),
		OutRequestNo: strconv.FormatUint(uint64(refundId), 10),
		RefundCents:  refund.GetAmount(),
		TotalCents:   tx.GetAmount(),
		Reason:       "订单退款",
	}); rErr != nil {
		s.log.Errorf("gateway refund rejected: refund=[%d] tx=[%d] err=%v", refundId, txId, rErr)
		return paymentV1.ErrorBadRequest("gateway rejected refund: %v", rErr.Error())
	}

	s.log.Infof("gateway refund accepted: refund=[%d] tx=[%d] amount=%d", refundId, txId, refund.GetAmount())
	return nil
}