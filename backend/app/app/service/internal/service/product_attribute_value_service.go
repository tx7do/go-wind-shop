package service

import (
	"context"

	"go-wind-shop/pkg/logx"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type ProductAttributeValueService struct {
	appV1.ProductAttributeValueServiceHTTPServer

	log *logx.Helper

	productAttributeValueServiceClient catalogV1.ProductAttributeValueServiceClient
}

func NewProductAttributeValueService(
	ctx *bootstrap.Context,
	productAttributeValueServiceClient catalogV1.ProductAttributeValueServiceClient,
) *ProductAttributeValueService {
	return &ProductAttributeValueService{
		log:                                logx.New(ctx.NewLoggerHelper("product-attribute-value/service/app-service")),
		productAttributeValueServiceClient: productAttributeValueServiceClient,
	}
}

func (s *ProductAttributeValueService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListProductAttributeValueResponse, error) {
	return s.productAttributeValueServiceClient.List(ctx, req)
}

func (s *ProductAttributeValueService) Get(ctx context.Context, req *catalogV1.GetProductAttributeValueRequest) (*catalogV1.ProductAttributeValue, error) {
	return s.productAttributeValueServiceClient.Get(ctx, req)
}
