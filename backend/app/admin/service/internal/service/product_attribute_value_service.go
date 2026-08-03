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

type ProductAttributeValueService struct {
	adminV1.ProductAttributeValueServiceHTTPServer

	log *log.Helper

	productAttributeValueServiceClient catalogV1.ProductAttributeValueServiceClient
}

func NewProductAttributeValueService(
	ctx *bootstrap.Context,
	productAttributeValueServiceClient catalogV1.ProductAttributeValueServiceClient,
) *ProductAttributeValueService {
	return &ProductAttributeValueService{
		log:                                ctx.NewLoggerHelper("product-attribute-value/service/admin-service"),
		productAttributeValueServiceClient: productAttributeValueServiceClient,
	}
}

func (s *ProductAttributeValueService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListProductAttributeValueResponse, error) {
	return s.productAttributeValueServiceClient.List(ctx, req)
}

func (s *ProductAttributeValueService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.CountProductAttributeValueResponse, error) {
	return s.productAttributeValueServiceClient.Count(ctx, req)
}

func (s *ProductAttributeValueService) Get(ctx context.Context, req *catalogV1.GetProductAttributeValueRequest) (*catalogV1.ProductAttributeValue, error) {
	return s.productAttributeValueServiceClient.Get(ctx, req)
}

func (s *ProductAttributeValueService) Create(ctx context.Context, req *catalogV1.CreateProductAttributeValueRequest) (*emptypb.Empty, error) {
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

	return s.productAttributeValueServiceClient.Create(ctx, req)
}

func (s *ProductAttributeValueService) BatchCreate(ctx context.Context, req *catalogV1.BatchCreateProductAttributeValuesRequest) (*emptypb.Empty, error) {
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

	return s.productAttributeValueServiceClient.BatchCreate(ctx, req)
}

func (s *ProductAttributeValueService) Update(ctx context.Context, req *catalogV1.UpdateProductAttributeValueRequest) (*emptypb.Empty, error) {
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

	return s.productAttributeValueServiceClient.Update(ctx, req)
}

func (s *ProductAttributeValueService) Delete(ctx context.Context, req *catalogV1.DeleteProductAttributeValueRequest) (*emptypb.Empty, error) {
	return s.productAttributeValueServiceClient.Delete(ctx, req)
}

func (s *ProductAttributeValueService) TranslationExists(ctx context.Context, req *catalogV1.ProductAttributeValueTranslationExistsRequest) (*catalogV1.ProductAttributeValueTranslationExistsResponse, error) {
	return s.productAttributeValueServiceClient.TranslationExists(ctx, req)
}

func (s *ProductAttributeValueService) GetTranslation(ctx context.Context, req *catalogV1.GetProductAttributeValueRequest) (*catalogV1.ProductAttributeValueTranslation, error) {
	return s.productAttributeValueServiceClient.GetTranslation(ctx, req)
}

func (s *ProductAttributeValueService) CreateTranslation(ctx context.Context, req *catalogV1.CreateProductAttributeValueTranslationRequest) (*catalogV1.ProductAttributeValueTranslation, error) {
	return s.productAttributeValueServiceClient.CreateTranslation(ctx, req)
}

func (s *ProductAttributeValueService) UpdateTranslation(ctx context.Context, req *catalogV1.UpdateProductAttributeValueTranslationRequest) (*catalogV1.ProductAttributeValueTranslation, error) {
	return s.productAttributeValueServiceClient.UpdateTranslation(ctx, req)
}

func (s *ProductAttributeValueService) DeleteTranslation(ctx context.Context, req *catalogV1.DeleteProductAttributeValueTranslationRequest) (*emptypb.Empty, error) {
	return s.productAttributeValueServiceClient.DeleteTranslation(ctx, req)
}
