package service

import (
	"context"
	"time"

	"go-wind-shop/pkg/logx"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	entCrud "github.com/tx7do/go-crud/entgo"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	"go-wind-shop/app/core/service/internal/data"
	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/pointsledger"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

// PointsAwardRatePerYuan 消费入账比率：1 积分 / 1 元（totalAmount 分 → /100）。
const PointsAwardRatePerYuan = 1

type PointsLedgerService struct {
	catalogV1.UnimplementedPointsLedgerServiceServer

	log *logx.Helper

	repo      *data.PointsLedgerRepo
	entClient *entCrud.EntClient[*ent.Client]
}

func NewPointsLedgerService(ctx *bootstrap.Context, repo *data.PointsLedgerRepo, entClient *entCrud.EntClient[*ent.Client]) *PointsLedgerService {
	return &PointsLedgerService{
		log:       logx.New(ctx.NewLoggerHelper("points-ledger/service/core-service")),
		repo:      repo,
		entClient: entClient,
	}
}

func (s *PointsLedgerService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListPointsLedgerResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *PointsLedgerService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.CountPointsLedgerResponse, error) {
	count, err := s.repo.Count(ctx, req)
	if err != nil {
		return nil, err
	}
	return &catalogV1.CountPointsLedgerResponse{Count: uint64(count)}, nil
}

func (s *PointsLedgerService) Get(ctx context.Context, req *catalogV1.GetPointsLedgerRequest) (*catalogV1.PointsLedger, error) {
	return s.repo.Get(ctx, req)
}

// GetUserPointsBalance 用户积分余额（delta 求和）。
func (s *PointsLedgerService) GetUserPointsBalance(ctx context.Context, req *catalogV1.GetUserPointsBalanceRequest) (*catalogV1.GetUserPointsBalanceResponse, error) {
	balance, err := s.userBalance(ctx, req.GetUserId())
	if err != nil {
		return nil, catalogV1.ErrorInternalServerError("query balance failed")
	}
	return &catalogV1.GetUserPointsBalanceResponse{Balance: balance}, nil
}

func (s *PointsLedgerService) userBalance(ctx context.Context, userId uint32) (int64, error) {
	var rows []struct{ Sum int64 }
	err := s.entClient.Client().PointsLedger.Query().
		Where(pointsledger.UserIDEQ(userId)).
		Aggregate(ent.As(ent.Sum(pointsledger.FieldDelta), "sum")).
		Scan(ctx, &rows)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	return rows[0].Sum, nil
}

// AwardForOrder 订单支付成功入账（幂等：同订单已有 EARN 台账则跳过）。
// best-effort 由调用方保证（积分失败不影响支付主流程，仅记日志）。
func (s *PointsLedgerService) AwardForOrder(ctx context.Context, userId uint32, orderId uint32, totalAmount int64) error {
	// 幂等：同订单已入账跳过
	exist, err := s.entClient.Client().PointsLedger.Query().
		Where(
			pointsledger.RefOrderIDEQ(orderId),
			pointsledger.TypeEQ(pointsledger.TypePointsTypeEarn),
		).
		Exist(ctx)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	award := totalAmount / 100 * PointsAwardRatePerYuan
	if award <= 0 {
		return nil
	}

	balance, err := s.userBalance(ctx, userId)
	if err != nil {
		return err
	}
	balance += award

	return s.entClient.Client().PointsLedger.Create().
		SetUserID(userId).
		SetDelta(award).
		SetBalanceAfter(balance).
		SetType(pointsledger.TypePointsTypeEarn).
		SetReason("订单消费入账").
		SetRefOrderID(orderId).
		SetCreatedAt(time.Now()).
		Exec(ctx)
}

// DeductForOrder 退款成功回冲（对订单已入账积分对冲，余额拖底为 0 不产生负积分）。
// 幂等：同订单已有 DEDUCT 台账则跳过。
func (s *PointsLedgerService) DeductForOrder(ctx context.Context, orderId uint32) error {
	// 幂等
	exist, err := s.entClient.Client().PointsLedger.Query().
		Where(
			pointsledger.RefOrderIDEQ(orderId),
			pointsledger.TypeEQ(pointsledger.TypePointsTypeDeduct),
		).
		Exist(ctx)
	if err != nil {
		return err
	}
	if exist {
		return nil
	}

	// 订单此前入账了多少
	var rows []struct{ Sum int64 }
	err = s.entClient.Client().PointsLedger.Query().
		Where(
			pointsledger.RefOrderIDEQ(orderId),
			pointsledger.TypeEQ(pointsledger.TypePointsTypeEarn),
		).
		Aggregate(ent.As(ent.Sum(pointsledger.FieldDelta), "sum")).
		Scan(ctx, &rows)
	if err != nil {
		return err
	}
	var awarded int64
	if len(rows) > 0 {
		awarded = rows[0].Sum
	}
	if awarded <= 0 {
		return nil
	}

	// 找到订单归属用户（从 EARN 台账取）
	led, err := s.entClient.Client().PointsLedger.Query().
		Where(
			pointsledger.RefOrderIDEQ(orderId),
			pointsledger.TypeEQ(pointsledger.TypePointsTypeEarn),
		).
		First(ctx)
	if err != nil || led == nil || led.UserID == nil {
		return nil
	}
	userId := *led.UserID

	balance, err := s.userBalance(ctx, userId)
	if err != nil {
		return err
	}
	deduct := awarded
	if balance-deduct < 0 {
		deduct = balance // 拖底为 0
	}
	if deduct <= 0 {
		return nil
	}
	balance -= deduct

	return s.entClient.Client().PointsLedger.Create().
		SetUserID(userId).
		SetDelta(-deduct).
		SetBalanceAfter(balance).
		SetType(pointsledger.TypePointsTypeDeduct).
		SetReason("退款回冲").
		SetRefOrderID(orderId).
		SetCreatedAt(time.Now()).
		Exec(ctx)
}

