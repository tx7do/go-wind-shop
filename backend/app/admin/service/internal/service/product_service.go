package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	adminV1 "go-wind-shop/api/gen/go/admin/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

type ProductService struct {
	adminV1.ProductServiceHTTPServer

	log *log.Helper

	productServiceClient catalogV1.ProductServiceClient
}

func NewProductService(
	ctx *bootstrap.Context,
	productServiceClient catalogV1.ProductServiceClient,
) *ProductService {
	return &ProductService{
		log:                  ctx.NewLoggerHelper("product/service/admin-service"),
		productServiceClient: productServiceClient,
	}
}

func (s *ProductService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListProductResponse, error) {
	return s.productServiceClient.List(ctx, req)
}

func (s *ProductService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.CountProductResponse, error) {
	return s.productServiceClient.Count(ctx, req)
}

func (s *ProductService) Get(ctx context.Context, req *catalogV1.GetProductRequest) (*catalogV1.Product, error) {
	return s.productServiceClient.Get(ctx, req)
}

func (s *ProductService) Create(ctx context.Context, req *catalogV1.CreateProductRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	for i := range req.Data.Translations {
		req.Data.Translations[i].CreatedBy = trans.Ptr(operator.UserId)
	}

	return s.productServiceClient.Create(ctx, req)
}

func (s *ProductService) BatchCreate(ctx context.Context, req *catalogV1.BatchCreateProductsRequest) (*emptypb.Empty, error) {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	for i := range req.Items {
		req.Items[i].CreatedBy = trans.Ptr(operator.UserId)
		for j := range req.Items[i].Translations {
			req.Items[i].Translations[j].CreatedBy = trans.Ptr(operator.UserId)
		}
	}

	return s.productServiceClient.BatchCreate(ctx, req)
}

func (s *ProductService) Update(ctx context.Context, req *catalogV1.UpdateProductRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.Id = trans.Ptr(req.GetId())

	req.Data.UpdatedBy = trans.Ptr(operator.GetUserId())
	if req.UpdateMask != nil {
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "updated_by")
	}

	for i := range req.Data.Translations {
		req.Data.Translations[i].UpdatedBy = trans.Ptr(operator.UserId)
	}

	return s.productServiceClient.Update(ctx, req)
}

func (s *ProductService) Delete(ctx context.Context, req *catalogV1.DeleteProductRequest) (*emptypb.Empty, error) {
	return s.productServiceClient.Delete(ctx, req)
}

func (s *ProductService) TranslationExists(ctx context.Context, req *catalogV1.ProductTranslationExistsRequest) (*catalogV1.ProductTranslationExistsResponse, error) {
	return s.productServiceClient.TranslationExists(ctx, req)
}

func (s *ProductService) GetTranslation(ctx context.Context, req *catalogV1.GetProductRequest) (*catalogV1.ProductTranslation, error) {
	return s.productServiceClient.GetTranslation(ctx, req)
}

func (s *ProductService) CreateTranslation(ctx context.Context, req *catalogV1.CreateProductTranslationRequest) (*catalogV1.ProductTranslation, error) {
	return s.productServiceClient.CreateTranslation(ctx, req)
}

func (s *ProductService) UpdateTranslation(ctx context.Context, req *catalogV1.UpdateProductTranslationRequest) (*catalogV1.ProductTranslation, error) {
	return s.productServiceClient.UpdateTranslation(ctx, req)
}

func (s *ProductService) DeleteTranslation(ctx context.Context, req *catalogV1.DeleteProductTranslationRequest) (*emptypb.Empty, error) {
	return s.productServiceClient.DeleteTranslation(ctx, req)
}
