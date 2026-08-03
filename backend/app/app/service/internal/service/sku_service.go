package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type SkuService struct {
	appV1.SkuServiceHTTPServer

	log *log.Helper

	skuServiceClient catalogV1.SkuServiceClient
}

func NewSkuService(
	ctx *bootstrap.Context,
	skuServiceClient catalogV1.SkuServiceClient,
) *SkuService {
	return &SkuService{
		log:              ctx.NewLoggerHelper("sku/service/app-service"),
		skuServiceClient: skuServiceClient,
	}
}

func (s *SkuService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListSkuResponse, error) {
	return s.skuServiceClient.List(ctx, req)
}

func (s *SkuService) Get(ctx context.Context, req *catalogV1.GetSkuRequest) (*catalogV1.Sku, error) {
	return s.skuServiceClient.Get(ctx, req)
}
