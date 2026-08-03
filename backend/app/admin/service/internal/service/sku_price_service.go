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

type SkuPriceService struct {
	adminV1.SkuPriceServiceHTTPServer

	log *log.Helper

	skuPriceServiceClient catalogV1.SkuPriceServiceClient
}

func NewSkuPriceService(
	ctx *bootstrap.Context,
	skuPriceServiceClient catalogV1.SkuPriceServiceClient,
) *SkuPriceService {
	return &SkuPriceService{
		log:                   ctx.NewLoggerHelper("sku-price/service/admin-service"),
		skuPriceServiceClient: skuPriceServiceClient,
	}
}

func (s *SkuPriceService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListSkuPriceResponse, error) {
	return s.skuPriceServiceClient.List(ctx, req)
}

func (s *SkuPriceService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.CountSkuResponse, error) {
	return s.skuPriceServiceClient.Count(ctx, req)
}

func (s *SkuPriceService) Get(ctx context.Context, req *catalogV1.GetSkuPriceRequest) (*catalogV1.SkuPrice, error) {
	return s.skuPriceServiceClient.Get(ctx, req)
}

func (s *SkuPriceService) Create(ctx context.Context, req *catalogV1.CreateSkuPriceRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	return s.skuPriceServiceClient.Create(ctx, req)
}

func (s *SkuPriceService) BatchCreate(ctx context.Context, req *catalogV1.BatchCreateSkuPricesRequest) (*emptypb.Empty, error) {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	for i := range req.Items {
		req.Items[i].CreatedBy = trans.Ptr(operator.UserId)
	}

	return s.skuPriceServiceClient.BatchCreate(ctx, req)
}

func (s *SkuPriceService) Update(ctx context.Context, req *catalogV1.UpdateSkuPriceRequest) (*emptypb.Empty, error) {
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

	return s.skuPriceServiceClient.Update(ctx, req)
}

func (s *SkuPriceService) Delete(ctx context.Context, req *catalogV1.DeleteSkuPriceRequest) (*emptypb.Empty, error) {
	return s.skuPriceServiceClient.Delete(ctx, req)
}
