package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	wishlistV1 "go-wind-shop/api/gen/go/wishlist/service/v1"
)

type WishlistService struct {
	wishlistV1.UnimplementedWishlistServiceServer

	log  *log.Helper
	repo *data.WishlistRepo
}

func NewWishlistService(ctx *bootstrap.Context, repo *data.WishlistRepo) *WishlistService {
	return &WishlistService{
		log:  ctx.NewLoggerHelper("wishlist/service/core-service"),
		repo: repo,
	}
}

func (s *WishlistService) List(ctx context.Context, req *paginationV1.PagingRequest) (*wishlistV1.ListWishlistResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *WishlistService) Create(ctx context.Context, req *wishlistV1.CreateWishlistRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, wishlistV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *WishlistService) Delete(ctx context.Context, req *wishlistV1.DeleteWishlistRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
