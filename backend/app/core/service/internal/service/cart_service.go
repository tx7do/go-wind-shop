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

type CartService struct {
	cartV1.UnimplementedCartServiceServer

	log  *log.Helper
	repo *data.CartRepo
}

func NewCartService(ctx *bootstrap.Context, repo *data.CartRepo) *CartService {
	return &CartService{
		log:  ctx.NewLoggerHelper("cart/service/core-service"),
		repo: repo,
	}
}

func (s *CartService) List(ctx context.Context, req *paginationV1.PagingRequest) (*cartV1.ListCartResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *CartService) Get(ctx context.Context, req *cartV1.GetCartRequest) (*cartV1.Cart, error) {
	return s.repo.Get(ctx, req)
}

func (s *CartService) Create(ctx context.Context, req *cartV1.CreateCartRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, cartV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *CartService) Update(ctx context.Context, req *cartV1.UpdateCartRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, cartV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *CartService) Delete(ctx context.Context, req *cartV1.DeleteCartRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
