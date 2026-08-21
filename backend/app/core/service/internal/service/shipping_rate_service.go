package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	shippingV1 "go-wind-shop/api/gen/go/shipping/service/v1"
)

type ShippingRateService struct {
	shippingV1.UnimplementedShippingRateServiceServer

	log *log.Helper

	repo *data.ShippingRateRepo
}

func NewShippingRateService(ctx *bootstrap.Context, repo *data.ShippingRateRepo) *ShippingRateService {
	return &ShippingRateService{
		log:  ctx.NewLoggerHelper("shipping-rate/service/core-service"),
		repo: repo,
	}
}

func (s *ShippingRateService) List(ctx context.Context, req *paginationV1.PagingRequest) (*shippingV1.ListShippingRateResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *ShippingRateService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*shippingV1.CountShippingRateResponse, error) {
	count, err := s.repo.Count(ctx, req)
	if err != nil {
		return nil, err
	}

	return &shippingV1.CountShippingRateResponse{
		Count: uint64(count),
	}, nil
}

func (s *ShippingRateService) Get(ctx context.Context, req *shippingV1.GetShippingRateRequest) (*shippingV1.ShippingRate, error) {
	return s.repo.Get(ctx, req)
}

func (s *ShippingRateService) Create(ctx context.Context, req *shippingV1.CreateShippingRateRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, shippingV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *ShippingRateService) Update(ctx context.Context, req *shippingV1.UpdateShippingRateRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, shippingV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *ShippingRateService) Delete(ctx context.Context, req *shippingV1.DeleteShippingRateRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
