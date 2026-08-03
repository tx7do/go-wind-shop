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

type SkuPriceService struct {
	catalogV1.UnimplementedSkuPriceServiceServer

	log *log.Helper

	repo *data.SkuPriceRepo
}

func NewSkuPriceService(ctx *bootstrap.Context, repo *data.SkuPriceRepo) *SkuPriceService {
	return &SkuPriceService{
		log:  ctx.NewLoggerHelper("sku-price/service/core-service"),
		repo: repo,
	}
}

func (s *SkuPriceService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListSkuPriceResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *SkuPriceService) Get(ctx context.Context, req *catalogV1.GetSkuPriceRequest) (*catalogV1.SkuPrice, error) {
	return s.repo.Get(ctx, req)
}

func (s *SkuPriceService) Create(ctx context.Context, req *catalogV1.CreateSkuPriceRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *SkuPriceService) Update(ctx context.Context, req *catalogV1.UpdateSkuPriceRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *SkuPriceService) Delete(ctx context.Context, req *catalogV1.DeleteSkuPriceRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
