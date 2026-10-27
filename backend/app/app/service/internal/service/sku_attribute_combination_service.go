package service

import (
	"context"

	"go-wind-shop/pkg/logx"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type SkuAttributeCombinationService struct {
	appV1.SkuAttributeCombinationServiceHTTPServer

	log *logx.Helper

	skuAttributeCombinationServiceClient catalogV1.SkuAttributeCombinationServiceClient
}

func NewSkuAttributeCombinationService(
	ctx *bootstrap.Context,
	skuAttributeCombinationServiceClient catalogV1.SkuAttributeCombinationServiceClient,
) *SkuAttributeCombinationService {
	return &SkuAttributeCombinationService{
		log:                                  logx.New(ctx.NewLoggerHelper("sku-attribute-combination/service/app-service")),
		skuAttributeCombinationServiceClient: skuAttributeCombinationServiceClient,
	}
}

func (s *SkuAttributeCombinationService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListSkuAttributeCombinationResponse, error) {
	return s.skuAttributeCombinationServiceClient.List(ctx, req)
}

func (s *SkuAttributeCombinationService) Get(ctx context.Context, req *catalogV1.GetSkuAttributeCombinationRequest) (*catalogV1.SkuAttributeCombination, error) {
	return s.skuAttributeCombinationServiceClient.Get(ctx, req)
}
