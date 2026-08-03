package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type ProductService struct {
	appV1.ProductServiceHTTPServer

	log *log.Helper

	productServiceClient catalogV1.ProductServiceClient
}

func NewProductService(
	ctx *bootstrap.Context,
	productServiceClient catalogV1.ProductServiceClient,
) *ProductService {
	return &ProductService{
		log:                  ctx.NewLoggerHelper("product/service/app-service"),
		productServiceClient: productServiceClient,
	}
}

func (s *ProductService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListProductResponse, error) {
	return s.productServiceClient.List(ctx, req)
}

func (s *ProductService) Get(ctx context.Context, req *catalogV1.GetProductRequest) (*catalogV1.Product, error) {
	return s.productServiceClient.Get(ctx, req)
}
