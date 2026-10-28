package service

import (
	"go-wind-shop/pkg/logx"
	"context"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	adminV1 "go-wind-shop/api/gen/go/admin/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

// StockOpLogService 库存操作流水查询（admin 只读透传）。
type StockOpLogService struct {
	adminV1.StockOpLogServiceHTTPServer

	log *logx.Helper

	stockOpLogServiceClient catalogV1.StockOpLogServiceClient
}

func NewStockOpLogService(
	ctx *bootstrap.Context,
	stockOpLogServiceClient catalogV1.StockOpLogServiceClient,
) *StockOpLogService {
	return &StockOpLogService{
		log:                     logx.New(ctx.NewLoggerHelper("stock-op-log/service/admin-service")),
		stockOpLogServiceClient: stockOpLogServiceClient,
	}
}

func (s *StockOpLogService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListStockOpLogResponse, error) {
	return s.stockOpLogServiceClient.List(ctx, req)
}

func (s *StockOpLogService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.CountStockOpLogResponse, error) {
	return s.stockOpLogServiceClient.Count(ctx, req)
}

func (s *StockOpLogService) Get(ctx context.Context, req *catalogV1.GetStockOpLogRequest) (*catalogV1.StockOpLog, error) {
	return s.stockOpLogServiceClient.Get(ctx, req)
}
