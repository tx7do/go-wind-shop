package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	adminV1 "go-wind-shop/api/gen/go/admin/service/v1"
	shippingV1 "go-wind-shop/api/gen/go/shipping/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

// ShippingRateService 运费模板管理（admin BFF，REST → gRPC 转发）。
type ShippingRateService struct {
	adminV1.ShippingRateServiceHTTPServer

	log *log.Helper

	shippingRateServiceClient shippingV1.ShippingRateServiceClient
}

func NewShippingRateService(
	ctx *bootstrap.Context,
	shippingRateServiceClient shippingV1.ShippingRateServiceClient,
) *ShippingRateService {
	return &ShippingRateService{
		log:                       ctx.NewLoggerHelper("shipping-rate/service/admin-service"),
		shippingRateServiceClient: shippingRateServiceClient,
	}
}

func (s *ShippingRateService) List(ctx context.Context, req *paginationV1.PagingRequest) (*shippingV1.ListShippingRateResponse, error) {
	return s.shippingRateServiceClient.List(ctx, req)
}

func (s *ShippingRateService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*shippingV1.CountShippingRateResponse, error) {
	return s.shippingRateServiceClient.Count(ctx, req)
}

func (s *ShippingRateService) Get(ctx context.Context, req *shippingV1.GetShippingRateRequest) (*shippingV1.ShippingRate, error) {
	return s.shippingRateServiceClient.Get(ctx, req)
}

func (s *ShippingRateService) Create(ctx context.Context, req *shippingV1.CreateShippingRateRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	return s.shippingRateServiceClient.Create(ctx, req)
}

func (s *ShippingRateService) Update(ctx context.Context, req *shippingV1.UpdateShippingRateRequest) (*emptypb.Empty, error) {
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

	return s.shippingRateServiceClient.Update(ctx, req)
}

func (s *ShippingRateService) Delete(ctx context.Context, req *shippingV1.DeleteShippingRateRequest) (*emptypb.Empty, error) {
	return s.shippingRateServiceClient.Delete(ctx, req)
}
