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

type CategoryService struct {
	catalogV1.UnimplementedCategoryServiceServer

	log *log.Helper

	repo *data.CategoryRepo
}

func NewCategoryService(ctx *bootstrap.Context, repo *data.CategoryRepo) *CategoryService {
	return &CategoryService{
		log:  ctx.NewLoggerHelper("category/service/core-service"),
		repo: repo,
	}
}

func (s *CategoryService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListCategoryResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *CategoryService) Get(ctx context.Context, req *catalogV1.GetCategoryRequest) (*catalogV1.Category, error) {
	return s.repo.Get(ctx, req)
}

func (s *CategoryService) Create(ctx context.Context, req *catalogV1.CreateCategoryRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if _, err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *CategoryService) Update(ctx context.Context, req *catalogV1.UpdateCategoryRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if _, err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *CategoryService) Delete(ctx context.Context, req *catalogV1.DeleteCategoryRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *CategoryService) TranslationExists(ctx context.Context, req *catalogV1.CategoryTranslationExistsRequest) (*catalogV1.CategoryTranslationExistsResponse, error) {
	exists, err := s.repo.TranslationExists(ctx, req.GetCategoryId(), req.GetLanguageCode())
	if err != nil {
		return nil, err
	}

	return &catalogV1.CategoryTranslationExistsResponse{
		Exists: exists,
	}, nil
}

func (s *CategoryService) GetTranslation(ctx context.Context, req *catalogV1.GetCategoryRequest) (*catalogV1.CategoryTranslation, error) {
	return s.repo.GetTranslation(ctx, req)
}

func (s *CategoryService) CreateTranslation(ctx context.Context, req *catalogV1.CreateCategoryTranslationRequest) (*catalogV1.CategoryTranslation, error) {
	return s.repo.CreateTranslation(ctx, req)
}

func (s *CategoryService) UpdateTranslation(ctx context.Context, req *catalogV1.UpdateCategoryTranslationRequest) (*catalogV1.CategoryTranslation, error) {
	return s.repo.UpdateTranslation(ctx, req)
}

func (s *CategoryService) DeleteTranslation(ctx context.Context, req *catalogV1.DeleteCategoryTranslationRequest) (*emptypb.Empty, error) {
	err := s.repo.DeleteTranslation(ctx, req)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
