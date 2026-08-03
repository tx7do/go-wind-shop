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

type SkuAttributeCombinationService struct {
	adminV1.SkuAttributeCombinationServiceHTTPServer

	log *log.Helper

	skuAttributeCombinationServiceClient catalogV1.SkuAttributeCombinationServiceClient
}

func NewSkuAttributeCombinationService(
	ctx *bootstrap.Context,
	skuAttributeCombinationServiceClient catalogV1.SkuAttributeCombinationServiceClient,
) *SkuAttributeCombinationService {
	return &SkuAttributeCombinationService{
		log:                                  ctx.NewLoggerHelper("sku-attribute-combination/service/admin-service"),
		skuAttributeCombinationServiceClient: skuAttributeCombinationServiceClient,
	}
}

func (s *SkuAttributeCombinationService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListSkuAttributeCombinationResponse, error) {
	return s.skuAttributeCombinationServiceClient.List(ctx, req)
}

func (s *SkuAttributeCombinationService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.CountSkuAttributeCombinationResponse, error) {
	return s.skuAttributeCombinationServiceClient.Count(ctx, req)
}

func (s *SkuAttributeCombinationService) Get(ctx context.Context, req *catalogV1.GetSkuAttributeCombinationRequest) (*catalogV1.SkuAttributeCombination, error) {
	return s.skuAttributeCombinationServiceClient.Get(ctx, req)
}

func (s *SkuAttributeCombinationService) Create(ctx context.Context, req *catalogV1.CreateSkuAttributeCombinationRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	return s.skuAttributeCombinationServiceClient.Create(ctx, req)
}

func (s *SkuAttributeCombinationService) BatchCreate(ctx context.Context, req *catalogV1.BatchCreateSkuAttributeCombinationsRequest) (*emptypb.Empty, error) {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	for i := range req.Items {
		req.Items[i].CreatedBy = trans.Ptr(operator.UserId)
	}

	return s.skuAttributeCombinationServiceClient.BatchCreate(ctx, req)
}

func (s *SkuAttributeCombinationService) Update(ctx context.Context, req *catalogV1.UpdateSkuAttributeCombinationRequest) (*emptypb.Empty, error) {
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

	return s.skuAttributeCombinationServiceClient.Update(ctx, req)
}

func (s *SkuAttributeCombinationService) Delete(ctx context.Context, req *catalogV1.DeleteSkuAttributeCombinationRequest) (*emptypb.Empty, error) {
	return s.skuAttributeCombinationServiceClient.Delete(ctx, req)
}
