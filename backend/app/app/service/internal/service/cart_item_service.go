package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	cartV1 "go-wind-shop/api/gen/go/cart/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

type CartItemService struct {
	appV1.CartItemServiceHTTPServer

	log *log.Helper

	cartItemServiceClient cartV1.CartItemServiceClient
}

func NewCartItemService(
	ctx *bootstrap.Context,
	cartItemServiceClient cartV1.CartItemServiceClient,
) *CartItemService {
	return &CartItemService{
		log:                   ctx.NewLoggerHelper("cart-item/service/app-service"),
		cartItemServiceClient: cartItemServiceClient,
	}
}

func (s *CartItemService) List(ctx context.Context, req *paginationV1.PagingRequest) (*cartV1.ListCartItemResponse, error) {
	return s.cartItemServiceClient.List(ctx, req)
}

func (s *CartItemService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*cartV1.CountCartItemResponse, error) {
	return s.cartItemServiceClient.Count(ctx, req)
}

func (s *CartItemService) Get(ctx context.Context, req *cartV1.GetCartItemRequest) (*cartV1.CartItem, error) {
	return s.cartItemServiceClient.Get(ctx, req)
}

func (s *CartItemService) Create(ctx context.Context, req *cartV1.CreateCartItemRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, appV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	return s.cartItemServiceClient.Create(ctx, req)
}

func (s *CartItemService) BatchCreate(ctx context.Context, req *cartV1.BatchCreateCartItemsRequest) (*emptypb.Empty, error) {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	for i := range req.Items {
		req.Items[i].CreatedBy = trans.Ptr(operator.UserId)
	}

	return s.cartItemServiceClient.BatchCreate(ctx, req)
}

func (s *CartItemService) Update(ctx context.Context, req *cartV1.UpdateCartItemRequest) (*emptypb.Empty, error) {
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

	return s.cartItemServiceClient.Update(ctx, req)
}

func (s *CartItemService) Delete(ctx context.Context, req *cartV1.DeleteCartItemRequest) (*emptypb.Empty, error) {
	return s.cartItemServiceClient.Delete(ctx, req)
}
