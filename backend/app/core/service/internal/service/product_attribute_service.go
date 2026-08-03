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

type ProductAttributeService struct {
	catalogV1.UnimplementedProductAttributeServiceServer

	log *log.Helper

	repo *data.ProductAttributeRepo
}

func NewProductAttributeService(ctx *bootstrap.Context, repo *data.ProductAttributeRepo) *ProductAttributeService {
	return &ProductAttributeService{
		log:  ctx.NewLoggerHelper("product-attribute/service/core-service"),
		repo: repo,
	}
}

func (s *ProductAttributeService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListProductAttributeResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *ProductAttributeService) Get(ctx context.Context, req *catalogV1.GetProductAttributeRequest) (*catalogV1.ProductAttribute, error) {
	return s.repo.Get(ctx, req)
}

func (s *ProductAttributeService) Create(ctx context.Context, req *catalogV1.CreateProductAttributeRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if _, err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *ProductAttributeService) Update(ctx context.Context, req *catalogV1.UpdateProductAttributeRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if _, err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *ProductAttributeService) Delete(ctx context.Context, req *catalogV1.DeleteProductAttributeRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *ProductAttributeService) TranslationExists(ctx context.Context, req *catalogV1.ProductAttributeTranslationExistsRequest) (*catalogV1.ProductAttributeTranslationExistsResponse, error) {
	exists, err := s.repo.TranslationExists(ctx, req.GetAttributeId(), req.GetLanguageCode())
	if err != nil {
		return nil, err
	}

	return &catalogV1.ProductAttributeTranslationExistsResponse{
		Exists: exists,
	}, nil
}

func (s *ProductAttributeService) GetTranslation(ctx context.Context, req *catalogV1.GetProductAttributeRequest) (*catalogV1.ProductAttributeTranslation, error) {
	return s.repo.GetTranslation(ctx, req)
}

func (s *ProductAttributeService) CreateTranslation(ctx context.Context, req *catalogV1.CreateProductAttributeTranslationRequest) (*catalogV1.ProductAttributeTranslation, error) {
	return s.repo.CreateTranslation(ctx, req)
}

func (s *ProductAttributeService) UpdateTranslation(ctx context.Context, req *catalogV1.UpdateProductAttributeTranslationRequest) (*catalogV1.ProductAttributeTranslation, error) {
	return s.repo.UpdateTranslation(ctx, req)
}

func (s *ProductAttributeService) DeleteTranslation(ctx context.Context, req *catalogV1.DeleteProductAttributeTranslationRequest) (*emptypb.Empty, error) {
	err := s.repo.DeleteTranslation(ctx, req)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
