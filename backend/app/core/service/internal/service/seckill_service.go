package service

import (
	"context"
	"time"

	"go-wind-shop/pkg/logx"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	entCrud "github.com/tx7do/go-crud/entgo"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"
	"github.com/tx7do/go-utils/trans"

	"go-wind-shop/app/core/service/internal/data"
	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/order"
	"go-wind-shop/app/core/service/internal/data/ent/orderitem"
	"go-wind-shop/app/core/service/internal/data/ent/seckillactivity"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type SeckillActivityService struct {
	catalogV1.UnimplementedSeckillActivityServiceServer

	log *logx.Helper

	repo *data.SeckillActivityRepo

	entClient *entCrud.EntClient[*ent.Client]
}

func NewSeckillActivityService(ctx *bootstrap.Context, repo *data.SeckillActivityRepo, entClient *entCrud.EntClient[*ent.Client]) *SeckillActivityService {
	return &SeckillActivityService{
		log:       logx.New(ctx.NewLoggerHelper("seckill-activity/service/core-service")),
		repo:      repo,
		entClient: entClient,
	}
}

func (s *SeckillActivityService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListSeckillActivityResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *SeckillActivityService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.CountSeckillActivityResponse, error) {
	count, err := s.repo.Count(ctx, req)
	if err != nil {
		return nil, err
	}

	return &catalogV1.CountSeckillActivityResponse{
		Count: uint64(count),
	}, nil
}

func (s *SeckillActivityService) Get(ctx context.Context, req *catalogV1.GetSeckillActivityRequest) (*catalogV1.SeckillActivity, error) {
	return s.repo.Get(ctx, req)
}

func (s *SeckillActivityService) ListActiveSeckills(ctx context.Context, req *catalogV1.ListActiveSeckillsRequest) (*catalogV1.ListSeckillActivityResponse, error) {
	return s.repo.ListActive(ctx)
}

func (s *SeckillActivityService) Create(ctx context.Context, req *catalogV1.CreateSeckillActivityRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.validateSeckill(req.Data); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *SeckillActivityService) Update(ctx context.Context, req *catalogV1.UpdateSeckillActivityRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.validateSeckill(req.Data); err != nil {
		return nil, err
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *SeckillActivityService) Delete(ctx context.Context, req *catalogV1.DeleteSeckillActivityRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// validateSeckill 活动参数校验：必须指定 SKU、秒杀价为正、限量/限购合法、
// 窗口起止有序。
func (s *SeckillActivityService) validateSeckill(data *catalogV1.SeckillActivity) error {
	if data.GetSkuId() == 0 {
		return catalogV1.ErrorBadRequest("seckill requires sku_id")
	}
	if data.GetSeckillPrice() <= 0 {
		return catalogV1.ErrorBadRequest("seckill_price must be positive")
	}
	if data.GetTotalStock() <= 0 {
		return catalogV1.ErrorBadRequest("total_stock must be positive")
	}
	if data.GetPerUserLimit() <= 0 {
		data.PerUserLimit = trans.Ptr(int32(1))
	}
	if data.GetValidFrom() != nil && data.GetValidUntil() != nil &&
		data.GetValidFrom().AsTime().After(data.GetValidUntil().AsTime()) {
		return catalogV1.ErrorBadRequest("valid_from must not be after valid_until")
	}
	return nil
}

// activeSeckillInTx 事务内查询某 SKU 当前可用的秒杀（status=ACTIVE、处于有效
// 窗口、未售罄），取优先级最高者。仅供订单创建事务使用； sold_count 的行锁
// 扣减由调用方（deductSeckillInTx）在同一事务内完成。
//
// 返回 (活动, 是否命中)。租户隔离由隐私层按 viewer 注入。
func activeSeckillInTx(ctx context.Context, tx *ent.Tx, skuId uint32, now time.Time) (*ent.SeckillActivity, bool) {
	now = time.Now()
	list, err := tx.SeckillActivity.Query().
		Where(
			seckillactivity.SkuIDEQ(skuId),
			seckillactivity.StatusEQ(seckillactivity.StatusSeckillStatusActive),
			seckillactivity.ValidFromLTE(now),
			seckillactivity.ValidUntilGTE(now),
		).
		Order(ent.Desc(seckillactivity.FieldPriority)).
		All(ctx)
	if err != nil || len(list) == 0 {
		return nil, false
	}
	for _, it := range list {
		total := int32(0)
		if it.TotalStock != nil {
			total = *it.TotalStock
		}
		if it.SoldCount < total {
			return it, true
		}
	}
	return nil, false
}

// deductSeckillInTx 行锁扣减秒杀已售数（与订单创建同事务，防超卖）。
// 成功返回 nil；售罄/并发竞争失败返回错误（整事务回滚）。
func deductSeckillInTx(ctx context.Context, tx *ent.Tx, activityId uint32) error {
	row, err := tx.SeckillActivity.Query().
		Where(seckillactivity.IDEQ(activityId)).
		ForUpdate().
		Only(ctx)
	if err != nil || row == nil {
		return catalogV1.ErrorBadRequest("seckill activity not found")
	}
	total := int32(0)
	if row.TotalStock != nil {
		total = *row.TotalStock
	}
	if row.SoldCount >= total {
		return catalogV1.ErrorBadRequest("seckill sold out")
	}
	sold := row.SoldCount + 1
	return tx.SeckillActivity.UpdateOneID(activityId).SetSoldCount(sold).Exec(ctx)
}

// checkSeckillUserLimitInTx 校验用户在活动窗口内对该 SKU 的累计购买量
// （含本次数量）不超 per_user_limit。
func checkSeckillUserLimitInTx(ctx context.Context, tx *ent.Tx, activity *ent.SeckillActivity, userId uint32, quantity int32) error {
	limit := int32(1)
	if activity.PerUserLimit > 0 {
		limit = activity.PerUserLimit
	}
	from := time.Unix(0, 0)
	if activity.ValidFrom != nil {
		from = *activity.ValidFrom
	}
	until := time.Now().Add(time.Hour * 24 * 365)
	if activity.ValidUntil != nil {
		until = *activity.ValidUntil
	}

	// 该用户在窗口内对该 SKU 的已购件数（已取消订单不计）。
	// order_item 与 order 无 ent Edge，走两段查询：先取用户窗口内非取消订单 ID，
	// 再按 (order_id IN, sku_id) 计数。
	if activity.SkuID == nil {
		return nil
	}
	orderIds, err := tx.Order.Query().
		Where(
			order.UserIDEQ(userId),
			order.StatusNEQ(order.StatusOrderStatusCancelled),
			order.CreatedAtGTE(from),
			order.CreatedAtLTE(until),
		).
		IDs(ctx)
	if err != nil {
		return catalogV1.ErrorInternalServerError("check seckill user limit failed")
	}
	if len(orderIds) == 0 {
		return nil
	}
	bought, err := tx.OrderItem.Query().
		Where(
			orderitem.SkuIDEQ(*activity.SkuID),
			orderitem.OrderIDIn(orderIds...),
		).
		Count(ctx)
	if err != nil {
		return catalogV1.ErrorInternalServerError("check seckill user limit failed")
	}
	if int32(bought)+quantity > limit {
		return catalogV1.ErrorBadRequest("seckill per-user limit exceeded (limit %d)", limit)
	}
	return nil
}
