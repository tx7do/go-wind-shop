package service

import (
	"context"

	"go-wind-shop/pkg/logx"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	adminV1 "go-wind-shop/api/gen/go/admin/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

// SeckillService 秒杀活动管理（admin BFF，REST → gRPC 转发）。
type SeckillService struct {
	adminV1.SeckillActivityServiceHTTPServer

	log *logx.Helper

	seckillActivityServiceClient catalogV1.SeckillActivityServiceClient
}

func NewSeckillService(
	ctx *bootstrap.Context,
	seckillActivityServiceClient catalogV1.SeckillActivityServiceClient,
) *SeckillService {
	return &SeckillService{
		log:                          logx.New(ctx.NewLoggerHelper("seckill/service/admin-service")),
		seckillActivityServiceClient: seckillActivityServiceClient,
	}
}

func (s *SeckillService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListSeckillActivityResponse, error) {
	return s.seckillActivityServiceClient.List(ctx, req)
}

func (s *SeckillService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.CountSeckillActivityResponse, error) {
	return s.seckillActivityServiceClient.Count(ctx, req)
}

func (s *SeckillService) Get(ctx context.Context, req *catalogV1.GetSeckillActivityRequest) (*catalogV1.SeckillActivity, error) {
	return s.seckillActivityServiceClient.Get(ctx, req)
}

func (s *SeckillService) Create(ctx context.Context, req *catalogV1.CreateSeckillActivityRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}
	return s.seckillActivityServiceClient.Create(ctx, req)
}

func (s *SeckillService) Update(ctx context.Context, req *catalogV1.UpdateSeckillActivityRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}
	if req.GetId() == 0 {
		return nil, adminV1.ErrorBadRequest("id is required")
	}
	return s.seckillActivityServiceClient.Update(ctx, req)
}

func (s *SeckillService) Delete(ctx context.Context, req *catalogV1.DeleteSeckillActivityRequest) (*emptypb.Empty, error) {
	return s.seckillActivityServiceClient.Delete(ctx, req)
}
