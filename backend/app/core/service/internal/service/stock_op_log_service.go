package service

import (
	"go-wind-shop/pkg/logx"
	"context"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	"go-wind-shop/app/core/service/internal/data"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

// StockOpLogService 库存操作流水查询（append-only 审计，只读）。
type StockOpLogService struct {
	catalogV1.UnimplementedStockOpLogServiceServer

	log *logx.Helper

	repo *data.StockOpLogRepo
}

func NewStockOpLogService(ctx *bootstrap.Context, repo *data.StockOpLogRepo) *StockOpLogService {
	return &StockOpLogService{
		log:  logx.New(ctx.NewLoggerHelper("stock-op-log/service/core-service")),
		repo: repo,
	}
}

func (s *StockOpLogService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListStockOpLogResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *StockOpLogService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.CountStockOpLogResponse, error) {
	count, err := s.repo.Count(ctx, req)
	if err != nil {
		return nil, err
	}

	return &catalogV1.CountStockOpLogResponse{
		Count: uint64(count),
	}, nil
}

func (s *StockOpLogService) Get(ctx context.Context, req *catalogV1.GetStockOpLogRequest) (*catalogV1.StockOpLog, error) {
	return s.repo.Get(ctx, req)
}
