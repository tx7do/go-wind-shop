package service

import (
	"context"

	"go-wind-shop/pkg/logx"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type ProductAttributeService struct {
	appV1.ProductAttributeServiceHTTPServer

	log *logx.Helper

	productAttributeServiceClient catalogV1.ProductAttributeServiceClient
}

func NewProductAttributeService(
	ctx *bootstrap.Context,
	productAttributeServiceClient catalogV1.ProductAttributeServiceClient,
) *ProductAttributeService {
	return &ProductAttributeService{
		log:                           logx.New(ctx.NewLoggerHelper("product-attribute/service/app-service")),
		productAttributeServiceClient: productAttributeServiceClient,
	}
}

func (s *ProductAttributeService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListProductAttributeResponse, error) {
	return s.productAttributeServiceClient.List(ctx, req)
}

func (s *ProductAttributeService) Get(ctx context.Context, req *catalogV1.GetProductAttributeRequest) (*catalogV1.ProductAttribute, error) {
	return s.productAttributeServiceClient.Get(ctx, req)
}
