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

type SkuService struct {
	adminV1.SkuServiceHTTPServer

	log *log.Helper

	skuServiceClient catalogV1.SkuServiceClient
}

func NewSkuService(
	ctx *bootstrap.Context,
	skuServiceClient catalogV1.SkuServiceClient,
) *SkuService {
	return &SkuService{
		log:              ctx.NewLoggerHelper("sku/service/admin-service"),
		skuServiceClient: skuServiceClient,
	}
}

func (s *SkuService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListSkuResponse, error) {
	return s.skuServiceClient.List(ctx, req)
}

func (s *SkuService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.CountSkuResponse, error) {
	return s.skuServiceClient.Count(ctx, req)
}

func (s *SkuService) Get(ctx context.Context, req *catalogV1.GetSkuRequest) (*catalogV1.Sku, error) {
	return s.skuServiceClient.Get(ctx, req)
}

func (s *SkuService) Create(ctx context.Context, req *catalogV1.CreateSkuRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	return s.skuServiceClient.Create(ctx, req)
}

func (s *SkuService) BatchCreate(ctx context.Context, req *catalogV1.BatchCreateSkusRequest) (*emptypb.Empty, error) {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	for i := range req.Items {
		req.Items[i].CreatedBy = trans.Ptr(operator.UserId)
	}

	return s.skuServiceClient.BatchCreate(ctx, req)
}

func (s *SkuService) Update(ctx context.Context, req *catalogV1.UpdateSkuRequest) (*emptypb.Empty, error) {
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

	return s.skuServiceClient.Update(ctx, req)
}

func (s *SkuService) Delete(ctx context.Context, req *catalogV1.DeleteSkuRequest) (*emptypb.Empty, error) {
	return s.skuServiceClient.Delete(ctx, req)
}
