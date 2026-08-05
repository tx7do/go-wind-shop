package service

import (
	"context"
	"fmt"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"github.com/tx7do/go-utils/trans"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	paymentV1 "go-wind-shop/api/gen/go/payment/service/v1"
)

type PaymentRefundService struct {
	paymentV1.UnimplementedPaymentRefundServiceServer

	log                   *log.Helper
	repo                  *data.PaymentRefundRepo
	transactionRepo       *data.PaymentTransactionRepo
	transactionService    *PaymentTransactionService
}

func NewPaymentRefundService(
	ctx *bootstrap.Context,
	repo *data.PaymentRefundRepo,
	transactionRepo *data.PaymentTransactionRepo,
	transactionService *PaymentTransactionService,
) *PaymentRefundService {
	return &PaymentRefundService{
		log:                ctx.NewLoggerHelper("payment-refund/service/core-service"),
		repo:               repo,
		transactionRepo:    transactionRepo,
		transactionService: transactionService,
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
	txTenantId := tx.GetTenantId()
	req.Data.TenantId = &txTenantId

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
			if r.GetTransactionId() == txId && r.GetStatus() == paymentV1.PaymentRefund_SUCCEEDED {
				return nil, paymentV1.ErrorBadRequest("payment transaction already refunded")
			}
		}
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
		txId := req.Data.GetTransactionId()
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
				// 当前 repo 模式不支持跨实体事务，故用补偿回滚兜底。
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
