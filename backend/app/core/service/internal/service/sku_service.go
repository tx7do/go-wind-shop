package service

import (
	"context"

	"go-wind-shop/pkg/logx"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type SkuService struct {
	catalogV1.UnimplementedSkuServiceServer

	log *logx.Helper

	repo *data.SkuRepo
}

func NewSkuService(ctx *bootstrap.Context, repo *data.SkuRepo) *SkuService {
	return &SkuService{
		log:  logx.New(ctx.NewLoggerHelper("sku/service/core-service")),
		repo: repo,
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
