package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	orderV1 "go-wind-shop/api/gen/go/order/service/v1"
)

type OrderItemService struct {
	orderV1.UnimplementedOrderItemServiceServer

	log  *log.Helper
	repo *data.OrderItemRepo
}

func NewOrderItemService(ctx *bootstrap.Context, repo *data.OrderItemRepo) *OrderItemService {
	return &OrderItemService{
		log:  ctx.NewLoggerHelper("order-item/service/core-service"),
		repo: repo,
	}
}

func (s *OrderItemService) List(ctx context.Context, req *paginationV1.PagingRequest) (*orderV1.ListOrderItemResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *OrderItemService) Get(ctx context.Context, req *orderV1.GetOrderItemRequest) (*orderV1.OrderItem, error) {
	return s.repo.Get(ctx, req)
}

func (s *OrderItemService) Create(ctx context.Context, req *orderV1.CreateOrderItemRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, orderV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *OrderItemService) Update(ctx context.Context, req *orderV1.UpdateOrderItemRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, orderV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *OrderItemService) Delete(ctx context.Context, req *orderV1.DeleteOrderItemRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
