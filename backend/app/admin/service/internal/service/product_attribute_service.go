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

type ProductAttributeService struct {
	adminV1.ProductAttributeServiceHTTPServer

	log *log.Helper

	productAttributeServiceClient catalogV1.ProductAttributeServiceClient
}

func NewProductAttributeService(
	ctx *bootstrap.Context,
	productAttributeServiceClient catalogV1.ProductAttributeServiceClient,
) *ProductAttributeService {
	return &ProductAttributeService{
		log:                           ctx.NewLoggerHelper("product-attribute/service/admin-service"),
		productAttributeServiceClient: productAttributeServiceClient,
	}
}

func (s *ProductAttributeService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListProductAttributeResponse, error) {
	return s.productAttributeServiceClient.List(ctx, req)
}

func (s *ProductAttributeService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.CountProductAttributeResponse, error) {
	return s.productAttributeServiceClient.Count(ctx, req)
}

func (s *ProductAttributeService) Get(ctx context.Context, req *catalogV1.GetProductAttributeRequest) (*catalogV1.ProductAttribute, error) {
	return s.productAttributeServiceClient.Get(ctx, req)
}

func (s *ProductAttributeService) Create(ctx context.Context, req *catalogV1.CreateProductAttributeRequest) (*emptypb.Empty, error) {
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

	return s.productAttributeServiceClient.Create(ctx, req)
}

func (s *ProductAttributeService) BatchCreate(ctx context.Context, req *catalogV1.BatchCreateProductAttributesRequest) (*emptypb.Empty, error) {
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

	return s.productAttributeServiceClient.BatchCreate(ctx, req)
}

func (s *ProductAttributeService) Update(ctx context.Context, req *catalogV1.UpdateProductAttributeRequest) (*emptypb.Empty, error) {
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

	return s.productAttributeServiceClient.Update(ctx, req)
}

func (s *ProductAttributeService) Delete(ctx context.Context, req *catalogV1.DeleteProductAttributeRequest) (*emptypb.Empty, error) {
	return s.productAttributeServiceClient.Delete(ctx, req)
}

func (s *ProductAttributeService) TranslationExists(ctx context.Context, req *catalogV1.ProductAttributeTranslationExistsRequest) (*catalogV1.ProductAttributeTranslationExistsResponse, error) {
	return s.productAttributeServiceClient.TranslationExists(ctx, req)
}

func (s *ProductAttributeService) GetTranslation(ctx context.Context, req *catalogV1.GetProductAttributeRequest) (*catalogV1.ProductAttributeTranslation, error) {
	return s.productAttributeServiceClient.GetTranslation(ctx, req)
}

func (s *ProductAttributeService) CreateTranslation(ctx context.Context, req *catalogV1.CreateProductAttributeTranslationRequest) (*catalogV1.ProductAttributeTranslation, error) {
	return s.productAttributeServiceClient.CreateTranslation(ctx, req)
}

func (s *ProductAttributeService) UpdateTranslation(ctx context.Context, req *catalogV1.UpdateProductAttributeTranslationRequest) (*catalogV1.ProductAttributeTranslation, error) {
	return s.productAttributeServiceClient.UpdateTranslation(ctx, req)
}

func (s *ProductAttributeService) DeleteTranslation(ctx context.Context, req *catalogV1.DeleteProductAttributeTranslationRequest) (*emptypb.Empty, error) {
	return s.productAttributeServiceClient.DeleteTranslation(ctx, req)
}
