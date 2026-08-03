package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type SkuPriceService struct {
	appV1.SkuPriceServiceHTTPServer

	log *log.Helper

	skuPriceServiceClient catalogV1.SkuPriceServiceClient
}

func NewSkuPriceService(
	ctx *bootstrap.Context,
	skuPriceServiceClient catalogV1.SkuPriceServiceClient,
) *SkuPriceService {
	return &SkuPriceService{
		log:                   ctx.NewLoggerHelper("sku-price/service/app-service"),
		skuPriceServiceClient: skuPriceServiceClient,
	}
}

func (s *SkuPriceService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListSkuPriceResponse, error) {
	return s.skuPriceServiceClient.List(ctx, req)
}

func (s *SkuPriceService) Get(ctx context.Context, req *catalogV1.GetSkuPriceRequest) (*catalogV1.SkuPrice, error) {
	return s.skuPriceServiceClient.Get(ctx, req)
}
