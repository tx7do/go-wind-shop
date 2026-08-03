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

type ProductService struct {
	catalogV1.UnimplementedProductServiceServer

	log *log.Helper

	repo *data.ProductRepo
}

func NewProductService(ctx *bootstrap.Context, repo *data.ProductRepo) *ProductService {
	return &ProductService{
		log:  ctx.NewLoggerHelper("product/service/core-service"),
		repo: repo,
	}
}

func (s *ProductService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListProductResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *ProductService) Get(ctx context.Context, req *catalogV1.GetProductRequest) (*catalogV1.Product, error) {
	return s.repo.Get(ctx, req)
}

func (s *ProductService) Create(ctx context.Context, req *catalogV1.CreateProductRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if _, err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *ProductService) Update(ctx context.Context, req *catalogV1.UpdateProductRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	if _, err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *ProductService) Delete(ctx context.Context, req *catalogV1.DeleteProductRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *ProductService) TranslationExists(ctx context.Context, req *catalogV1.ProductTranslationExistsRequest) (*catalogV1.ProductTranslationExistsResponse, error) {
	exists, err := s.repo.TranslationExists(ctx, req.GetProductId(), req.GetLanguageCode())
	if err != nil {
		return nil, err
	}

	return &catalogV1.ProductTranslationExistsResponse{
		Exists: exists,
	}, nil
}

func (s *ProductService) GetTranslation(ctx context.Context, req *catalogV1.GetProductRequest) (*catalogV1.ProductTranslation, error) {
	return s.repo.GetTranslation(ctx, req)
}

func (s *ProductService) CreateTranslation(ctx context.Context, req *catalogV1.CreateProductTranslationRequest) (*catalogV1.ProductTranslation, error) {
	return s.repo.CreateTranslation(ctx, req)
}

func (s *ProductService) UpdateTranslation(ctx context.Context, req *catalogV1.UpdateProductTranslationRequest) (*catalogV1.ProductTranslation, error) {
	return s.repo.UpdateTranslation(ctx, req)
}

func (s *ProductService) DeleteTranslation(ctx context.Context, req *catalogV1.DeleteProductTranslationRequest) (*emptypb.Empty, error) {
	err := s.repo.DeleteTranslation(ctx, req)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
