package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	orderV1 "go-wind-shop/api/gen/go/order/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

type OrderService struct {
	appV1.OrderServiceHTTPServer

	log *log.Helper

	orderServiceClient orderV1.OrderServiceClient
}

func NewOrderService(
	ctx *bootstrap.Context,
	orderServiceClient orderV1.OrderServiceClient,
) *OrderService {
	return &OrderService{
		log:                ctx.NewLoggerHelper("order/service/app-service"),
		orderServiceClient: orderServiceClient,
	}
}

func (s *OrderService) List(ctx context.Context, req *paginationV1.PagingRequest) (*orderV1.ListOrderResponse, error) {
	return s.orderServiceClient.List(ctx, req)
}

func (s *OrderService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*orderV1.CountOrderResponse, error) {
	return s.orderServiceClient.Count(ctx, req)
}

func (s *OrderService) Get(ctx context.Context, req *orderV1.GetOrderRequest) (*orderV1.Order, error) {
	return s.orderServiceClient.Get(ctx, req)
}

func (s *OrderService) Create(ctx context.Context, req *orderV1.CreateOrderRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, appV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 强制覆盖为当前登录用户，防止客户端伪造 userId/tenantId 越权
	// 下单他人购物车或将订单挂到他人名下。
	req.Data.UserId = trans.Ptr(operator.GetUserId())
	req.Data.TenantId = trans.Ptr(operator.GetTenantId())
	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	return s.orderServiceClient.Create(ctx, req)
}

func (s *OrderService) Update(ctx context.Context, req *orderV1.UpdateOrderRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, appV1.ErrorBadRequest("invalid parameter")
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

	return s.orderServiceClient.Update(ctx, req)
}

func (s *OrderService) Delete(ctx context.Context, req *orderV1.DeleteOrderRequest) (*emptypb.Empty, error) {
	return s.orderServiceClient.Delete(ctx, req)
}
