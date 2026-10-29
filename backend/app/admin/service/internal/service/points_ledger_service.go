package service

import (
	"context"

	"go-wind-shop/pkg/logx"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	adminV1 "go-wind-shop/api/gen/go/admin/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

// PointsLedgerService 积分台账管理（admin BFF 只读转发）。
type PointsLedgerService struct {
	adminV1.PointsLedgerServiceHTTPServer

	log *logx.Helper

	pointsLedgerServiceClient catalogV1.PointsLedgerServiceClient
}

func NewPointsLedgerService(
	ctx *bootstrap.Context,
	pointsLedgerServiceClient catalogV1.PointsLedgerServiceClient,
) *PointsLedgerService {
	return &PointsLedgerService{
		log:                       logx.New(ctx.NewLoggerHelper("points-ledger/service/admin-service")),
		pointsLedgerServiceClient: pointsLedgerServiceClient,
	}
}

func (s *PointsLedgerService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListPointsLedgerResponse, error) {
	return s.pointsLedgerServiceClient.List(ctx, req)
}

func (s *PointsLedgerService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.CountPointsLedgerResponse, error) {
	return s.pointsLedgerServiceClient.Count(ctx, req)
}

func (s *PointsLedgerService) Get(ctx context.Context, req *catalogV1.GetPointsLedgerRequest) (*catalogV1.PointsLedger, error) {
	return s.pointsLedgerServiceClient.Get(ctx, req)
}
