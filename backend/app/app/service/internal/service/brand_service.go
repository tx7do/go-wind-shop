package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type BrandService struct {
	appV1.BrandServiceHTTPServer

	log *log.Helper

	brandServiceClient catalogV1.BrandServiceClient
}

func NewBrandService(
	ctx *bootstrap.Context,
	brandServiceClient catalogV1.BrandServiceClient,
) *BrandService {
	return &BrandService{
		log:                ctx.NewLoggerHelper("brand/service/app-service"),
		brandServiceClient: brandServiceClient,
	}
}

func (s *BrandService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListBrandResponse, error) {
	return s.brandServiceClient.List(ctx, req)
}

func (s *BrandService) Get(ctx context.Context, req *catalogV1.GetBrandRequest) (*catalogV1.Brand, error) {
	return s.brandServiceClient.Get(ctx, req)
}
