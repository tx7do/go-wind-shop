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

type CartService struct {
	appV1.CartServiceHTTPServer

	log *log.Helper

	cartServiceClient cartV1.CartServiceClient
}

func NewCartService(
	ctx *bootstrap.Context,
	cartServiceClient cartV1.CartServiceClient,
) *CartService {
	return &CartService{
		log:               ctx.NewLoggerHelper("cart/service/app-service"),
		cartServiceClient: cartServiceClient,
	}
}

func (s *CartService) List(ctx context.Context, req *paginationV1.PagingRequest) (*cartV1.ListCartResponse, error) {
	return s.cartServiceClient.List(ctx, req)
}

func (s *CartService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*cartV1.CountCartResponse, error) {
	return s.cartServiceClient.Count(ctx, req)
}

func (s *CartService) Get(ctx context.Context, req *cartV1.GetCartRequest) (*cartV1.Cart, error) {
	return s.cartServiceClient.Get(ctx, req)
}

func (s *CartService) Create(ctx context.Context, req *cartV1.CreateCartRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, appV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	return s.cartServiceClient.Create(ctx, req)
}

func (s *CartService) BatchCreate(ctx context.Context, req *cartV1.BatchCreateCartsRequest) (*emptypb.Empty, error) {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	for i := range req.Items {
		req.Items[i].CreatedBy = trans.Ptr(operator.UserId)
	}

	return s.cartServiceClient.BatchCreate(ctx, req)
}

func (s *CartService) Update(ctx context.Context, req *cartV1.UpdateCartRequest) (*emptypb.Empty, error) {
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

	return s.cartServiceClient.Update(ctx, req)
}

func (s *CartService) Delete(ctx context.Context, req *cartV1.DeleteCartRequest) (*emptypb.Empty, error) {
	return s.cartServiceClient.Delete(ctx, req)
}
