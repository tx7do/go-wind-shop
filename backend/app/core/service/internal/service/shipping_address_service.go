package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	addressV1 "go-wind-shop/api/gen/go/address/service/v1"
)

// ShippingAddressService 收货地址核心服务（gRPC）。
// 纯 CRUD 委托到 repo；user_id 的可见性隔离由 UserPrivacy 策略 + app 网关注入保证。
type ShippingAddressService struct {
	addressV1.UnimplementedShippingAddressServiceServer

	log  *log.Helper
	repo *data.ShippingAddressRepo
}

func NewShippingAddressService(
	ctx *bootstrap.Context,
	repo *data.ShippingAddressRepo,
) *ShippingAddressService {
	return &ShippingAddressService{
		log:  ctx.NewLoggerHelper("shipping-address/service/core-service"),
		repo: repo,
	}
}

func (s *ShippingAddressService) List(ctx context.Context, req *paginationV1.PagingRequest) (*addressV1.ListShippingAddressResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *ShippingAddressService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*addressV1.CountShippingAddressResponse, error) {
	count, err := s.repo.Count(ctx, req)
	if err != nil {
		return nil, err
	}
	return &addressV1.CountShippingAddressResponse{Count: uint64(count)}, nil
}

func (s *ShippingAddressService) Get(ctx context.Context, req *addressV1.GetShippingAddressRequest) (*addressV1.ShippingAddress, error) {
	return s.repo.Get(ctx, req)
}

func (s *ShippingAddressService) Create(ctx context.Context, req *addressV1.CreateShippingAddressRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, addressV1.ErrorBadRequest("invalid parameter")
	}
	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *ShippingAddressService) Update(ctx context.Context, req *addressV1.UpdateShippingAddressRequest) (*emptypb.Empty, error) {
	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *ShippingAddressService) Delete(ctx context.Context, req *addressV1.DeleteShippingAddressRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
