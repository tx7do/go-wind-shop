package service

import (
	"context"

	"go-wind-shop/pkg/logx"
	"time"

	"github.com/tx7do/go-utils/trans"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	entCrud "github.com/tx7do/go-crud/entgo"

	"go-wind-shop/app/core/service/internal/data"
	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/sku"
	"go-wind-shop/app/core/service/internal/data/ent/stockoplog"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type SkuService struct {
	catalogV1.UnimplementedSkuServiceServer

	log *logx.Helper

	repo      *data.SkuRepo
	entClient *entCrud.EntClient[*ent.Client]
}

func NewSkuService(ctx *bootstrap.Context, repo *data.SkuRepo, entClient *entCrud.EntClient[*ent.Client]) *SkuService {
	return &SkuService{
		log:       logx.New(ctx.NewLoggerHelper("sku/service/core-service")),
		repo:      repo,
		entClient: entClient,
	}
}

func (s *SkuService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListSkuResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *SkuService) Get(ctx context.Context, req *catalogV1.GetSkuRequest) (*catalogV1.Sku, error) {
	return s.repo.Get(ctx, req)
}

func (s *SkuService) Create(ctx context.Context, req *catalogV1.CreateSkuRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *SkuService) Update(ctx context.Context, req *catalogV1.UpdateSkuRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	// 库存变更必须走 AdjustStock（行锁 + append-only 流水）；通用更新里带 stockQty
	// 会绕过流水审计，直接拒绝。
	for _, path := range req.GetUpdateMask().GetPaths() {
		if path == "stockQty" || path == "stock_qty" {
			return nil, catalogV1.ErrorBadRequest("库存变更请使用库存调整接口（不记审计流水）")
		}
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *SkuService) Delete(ctx context.Context, req *catalogV1.DeleteSkuRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// AdjustStock 手动调整库存（admin）。
// 事务内行锁读改写，并追加库存操作流水（MANUAL_ADJUST，操作者取 viewer 用户）。
// 调整后库存 < 0 拒绝（BadRequest），整事务回滚。
func (s *SkuService) AdjustStock(ctx context.Context, req *catalogV1.AdjustSkuStockRequest) (*emptypb.Empty, error) {
	if req.GetId() == 0 {
		return nil, catalogV1.ErrorBadRequest("id is required")
	}
	if req.GetDelta() == 0 {
		return nil, catalogV1.ErrorBadRequest("delta must not be zero")
	}

	operator, _ := viewerUserIDFromContext(ctx)

	tx, err := s.entClient.Client().Tx(ctx)
	if err != nil {
		s.log.Errorf("begin adjust-stock tx failed: %v", err)
		return nil, catalogV1.ErrorInternalServerError("begin tx failed")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	skuEnt, qErr := tx.Sku.Query().
		Where(sku.IDEQ(req.GetId())).
		ForUpdate().
		Only(ctx)
	if qErr != nil || skuEnt == nil {
		err = catalogV1.ErrorBadRequest("sku not found")
		return nil, err
	}

	before := int32(0)
	if skuEnt.StockQty != nil {
		before = *skuEnt.StockQty
	}
	after := before + req.GetDelta()
	if after < 0 {
		err = catalogV1.ErrorBadRequest("insufficient stock for adjustment: before=%d delta=%d", before, req.GetDelta())
		return nil, err
	}

	if uErr := tx.Sku.UpdateOneID(req.GetId()).
		SetStockQty(after).
		Exec(ctx); uErr != nil {
		err = uErr
		s.log.Errorf("adjust sku [%d] stock failed: %v", req.GetId(), uErr)
		return nil, catalogV1.ErrorInternalServerError("adjust stock failed")
	}

	var productId uint32
	if skuEnt.ProductID != nil {
		productId = *skuEnt.ProductID
	}
	if lErr := tx.StockOpLog.Create().
		SetNillableSkuID(&skuEnt.ID).
		SetNillableProductID(&productId).
		SetNillableDelta(trans.Ptr(req.GetDelta())).
		SetNillableStockBefore(&before).
		SetNillableStockAfter(&after).
		SetReason(stockoplog.ReasonStockOpReasonManualAdjust).
		SetNillableCreatedBy(&operator).
		SetNillableRemark(req.Remark).
		SetCreatedAt(time.Now()).
		Exec(ctx); lErr != nil {
		// 流水失败不回滚库存变更（审计可后补），但记日志告警。
		s.log.Errorf("insert stock op log for sku [%d] adjust failed: %v", req.GetId(), lErr)
	}

	if cErr := tx.Commit(); cErr != nil {
		err = cErr
		return nil, catalogV1.ErrorInternalServerError("commit failed")
	}
	err = nil

	s.log.Infof("sku [%d] stock adjusted by [%d]: %d -> %d (delta=%d)", req.GetId(), operator, before, after, req.GetDelta())
	return &emptypb.Empty{}, nil
}
