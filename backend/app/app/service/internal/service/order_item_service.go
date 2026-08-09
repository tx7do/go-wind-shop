package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	orderV1 "go-wind-shop/api/gen/go/order/service/v1"
)

type OrderItemService struct {
	appV1.OrderItemServiceHTTPServer

	log *log.Helper

	orderItemServiceClient orderV1.OrderItemServiceClient
}

func NewOrderItemService(
	ctx *bootstrap.Context,
	orderItemServiceClient orderV1.OrderItemServiceClient,
) *OrderItemService {
	return &OrderItemService{
		log:                    ctx.NewLoggerHelper("order-item/service/app-service"),
		orderItemServiceClient: orderItemServiceClient,
	}
}

func (s *OrderItemService) List(ctx context.Context, req *paginationV1.PagingRequest) (*orderV1.ListOrderItemResponse, error) {
	return s.orderItemServiceClient.List(ctx, req)
}

func (s *OrderItemService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*orderV1.CountOrderItemResponse, error) {
	return s.orderItemServiceClient.Count(ctx, req)
}

func (s *OrderItemService) Get(ctx context.Context, req *orderV1.GetOrderItemRequest) (*orderV1.OrderItem, error) {
	return s.orderItemServiceClient.Get(ctx, req)
}

