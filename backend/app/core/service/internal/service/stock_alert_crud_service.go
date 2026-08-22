package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

// StockAlertService 库存预警记录管理服务（core，gRPC）。
//
// 裁剪版：仅 List/Get/Update。告警由 StockAlertScannerService 周期任务落库
// （OPEN 态），本服务仅提供 admin 查询与处置。
//
// Update 唯一允许的状态转移是 OPEN→RESOLVED：core 层强制校验 req.Data.Status
// 必须为 RESOLVED，否则 BadRequest。无 Create/Delete/Count（告警仅由扫描器
// 创建，Delete 不提供以保留审计留痕，Count 不需要——分页用 total 字段）。
type StockAlertService struct {
	catalogV1.UnimplementedStockAlertServiceServer

	log *log.Helper

	repo *data.StockAlertRepo
}

func NewStockAlertService(ctx *bootstrap.Context, repo *data.StockAlertRepo) *StockAlertService {
	return &StockAlertService{
		log:  ctx.NewLoggerHelper("stock-alert/service/core-service"),
		repo: repo,
	}
}

func (s *StockAlertService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListStockAlertResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *StockAlertService) Get(ctx context.Context, req *catalogV1.GetStockAlertRequest) (*catalogV1.StockAlert, error) {
	return s.repo.Get(ctx, req)
}

// Update 强制 req.Data.Status == RESOLVED（唯一允许的处置转移）。
// 其余字段（sku_id/stock_qty_at_trigger/threshold/created_by/created_at）均
// 不可由本接口变更——update_mask 仅含 "status"（admin BFF 注入）+ "updated_by"。
func (s *StockAlertService) Update(ctx context.Context, req *catalogV1.UpdateStockAlertRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if req.Data.Status == nil || *req.Data.Status != catalogV1.StockAlert_RESOLVED {
		return nil, catalogV1.ErrorBadRequest("only resolution is permitted")
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}
