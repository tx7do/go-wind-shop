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

type BrandService struct {
	catalogV1.UnimplementedBrandServiceServer

	log *log.Helper

	repo *data.BrandRepo
}

func NewBrandService(ctx *bootstrap.Context, repo *data.BrandRepo) *BrandService {
	return &BrandService{
		log:  ctx.NewLoggerHelper("brand/service/core-service"),
		repo: repo,
	}
}

func (s *BrandService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListBrandResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *BrandService) Get(ctx context.Context, req *catalogV1.GetBrandRequest) (*catalogV1.Brand, error) {
	return s.repo.Get(ctx, req)
}

func (s *BrandService) Create(ctx context.Context, req *catalogV1.CreateBrandRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if _, err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *BrandService) Update(ctx context.Context, req *catalogV1.UpdateBrandRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if _, err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *BrandService) Delete(ctx context.Context, req *catalogV1.DeleteBrandRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *BrandService) TranslationExists(ctx context.Context, req *catalogV1.BrandTranslationExistsRequest) (*catalogV1.BrandTranslationExistsResponse, error) {
	exists, err := s.repo.TranslationExists(ctx, req.GetBrandId(), req.GetLanguageCode())
	if err != nil {
		return nil, err
	}

	return &catalogV1.BrandTranslationExistsResponse{
		Exists: exists,
	}, nil
}

func (s *BrandService) GetTranslation(ctx context.Context, req *catalogV1.GetBrandRequest) (*catalogV1.BrandTranslation, error) {
	return s.repo.GetTranslation(ctx, req)
}

func (s *BrandService) CreateTranslation(ctx context.Context, req *catalogV1.CreateBrandTranslationRequest) (*catalogV1.BrandTranslation, error) {
	return s.repo.CreateTranslation(ctx, req)
}

func (s *BrandService) UpdateTranslation(ctx context.Context, req *catalogV1.UpdateBrandTranslationRequest) (*catalogV1.BrandTranslation, error) {
	return s.repo.UpdateTranslation(ctx, req)
}

func (s *BrandService) DeleteTranslation(ctx context.Context, req *catalogV1.DeleteBrandTranslationRequest) (*emptypb.Empty, error) {
	err := s.repo.DeleteTranslation(ctx, req)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
