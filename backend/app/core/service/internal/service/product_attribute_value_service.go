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

type ProductAttributeValueService struct {
	catalogV1.UnimplementedProductAttributeValueServiceServer

	log *log.Helper

	repo *data.ProductAttributeValueRepo
}

func NewProductAttributeValueService(ctx *bootstrap.Context, repo *data.ProductAttributeValueRepo) *ProductAttributeValueService {
	return &ProductAttributeValueService{
		log:  ctx.NewLoggerHelper("product-attribute-value/service/core-service"),
		repo: repo,
	}
}

func (s *ProductAttributeValueService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListProductAttributeValueResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *ProductAttributeValueService) Get(ctx context.Context, req *catalogV1.GetProductAttributeValueRequest) (*catalogV1.ProductAttributeValue, error) {
	return s.repo.Get(ctx, req)
}

func (s *ProductAttributeValueService) Create(ctx context.Context, req *catalogV1.CreateProductAttributeValueRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if _, err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *ProductAttributeValueService) Update(ctx context.Context, req *catalogV1.UpdateProductAttributeValueRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if _, err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *ProductAttributeValueService) Delete(ctx context.Context, req *catalogV1.DeleteProductAttributeValueRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *ProductAttributeValueService) TranslationExists(ctx context.Context, req *catalogV1.ProductAttributeValueTranslationExistsRequest) (*catalogV1.ProductAttributeValueTranslationExistsResponse, error) {
	exists, err := s.repo.TranslationExists(ctx, req.GetValueId(), req.GetLanguageCode())
	if err != nil {
		return nil, err
	}

	return &catalogV1.ProductAttributeValueTranslationExistsResponse{
		Exists: exists,
	}, nil
}

func (s *ProductAttributeValueService) GetTranslation(ctx context.Context, req *catalogV1.GetProductAttributeValueRequest) (*catalogV1.ProductAttributeValueTranslation, error) {
	return s.repo.GetTranslation(ctx, req)
}

func (s *ProductAttributeValueService) CreateTranslation(ctx context.Context, req *catalogV1.CreateProductAttributeValueTranslationRequest) (*catalogV1.ProductAttributeValueTranslation, error) {
	return s.repo.CreateTranslation(ctx, req)
}

func (s *ProductAttributeValueService) UpdateTranslation(ctx context.Context, req *catalogV1.UpdateProductAttributeValueTranslationRequest) (*catalogV1.ProductAttributeValueTranslation, error) {
	return s.repo.UpdateTranslation(ctx, req)
}

func (s *ProductAttributeValueService) DeleteTranslation(ctx context.Context, req *catalogV1.DeleteProductAttributeValueTranslationRequest) (*emptypb.Empty, error) {
	err := s.repo.DeleteTranslation(ctx, req)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
