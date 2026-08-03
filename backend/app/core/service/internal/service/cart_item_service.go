package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	cartV1 "go-wind-shop/api/gen/go/cart/service/v1"
)

type CartItemService struct {
	cartV1.UnimplementedCartItemServiceServer

	log  *log.Helper
	repo *data.CartItemRepo
}

func NewCartItemService(ctx *bootstrap.Context, repo *data.CartItemRepo) *CartItemService {
	return &CartItemService{
		log:  ctx.NewLoggerHelper("cart-item/service/core-service"),
		repo: repo,
	}
}

func (s *CartItemService) List(ctx context.Context, req *paginationV1.PagingRequest) (*cartV1.ListCartItemResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *CartItemService) Get(ctx context.Context, req *cartV1.GetCartItemRequest) (*cartV1.CartItem, error) {
	return s.repo.Get(ctx, req)
}

func (s *CartItemService) Create(ctx context.Context, req *cartV1.CreateCartItemRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, cartV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *CartItemService) Update(ctx context.Context, req *cartV1.UpdateCartItemRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, cartV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *CartItemService) Delete(ctx context.Context, req *cartV1.DeleteCartItemRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
