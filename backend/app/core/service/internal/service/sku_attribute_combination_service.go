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

type SkuAttributeCombinationService struct {
	catalogV1.UnimplementedSkuAttributeCombinationServiceServer

	log *log.Helper

	repo *data.SkuAttributeCombinationRepo
}

func NewSkuAttributeCombinationService(ctx *bootstrap.Context, repo *data.SkuAttributeCombinationRepo) *SkuAttributeCombinationService {
	return &SkuAttributeCombinationService{
		log:  ctx.NewLoggerHelper("sku-attribute-combination/service/core-service"),
		repo: repo,
	}
}

func (s *SkuAttributeCombinationService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListSkuAttributeCombinationResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *SkuAttributeCombinationService) Get(ctx context.Context, req *catalogV1.GetSkuAttributeCombinationRequest) (*catalogV1.SkuAttributeCombination, error) {
	return s.repo.Get(ctx, req)
}

func (s *SkuAttributeCombinationService) Create(ctx context.Context, req *catalogV1.CreateSkuAttributeCombinationRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *SkuAttributeCombinationService) Update(ctx context.Context, req *catalogV1.UpdateSkuAttributeCombinationRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *SkuAttributeCombinationService) Delete(ctx context.Context, req *catalogV1.DeleteSkuAttributeCombinationRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
