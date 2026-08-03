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

func (s *OrderItemService) Create(ctx context.Context, req *orderV1.CreateOrderItemRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, appV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	return s.orderItemServiceClient.Create(ctx, req)
}

func (s *OrderItemService) BatchCreate(ctx context.Context, req *orderV1.BatchCreateOrderItemsRequest) (*emptypb.Empty, error) {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	for i := range req.Items {
		req.Items[i].CreatedBy = trans.Ptr(operator.UserId)
	}

	return s.orderItemServiceClient.BatchCreate(ctx, req)
}

func (s *OrderItemService) Update(ctx context.Context, req *orderV1.UpdateOrderItemRequest) (*emptypb.Empty, error) {
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

	return s.orderItemServiceClient.Update(ctx, req)
}

func (s *OrderItemService) Delete(ctx context.Context, req *orderV1.DeleteOrderItemRequest) (*emptypb.Empty, error) {
	return s.orderItemServiceClient.Delete(ctx, req)
}
