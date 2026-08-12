package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	couponV1 "go-wind-shop/api/gen/go/coupon/service/v1"
)

type CouponTemplateService struct {
	couponV1.UnimplementedCouponTemplateServiceServer

	log *log.Helper

	repo *data.CouponTemplateRepo
}

func NewCouponTemplateService(ctx *bootstrap.Context, repo *data.CouponTemplateRepo) *CouponTemplateService {
	return &CouponTemplateService{
		log:  ctx.NewLoggerHelper("coupon-template/service/core-service"),
		repo: repo,
	}
}

func (s *CouponTemplateService) List(ctx context.Context, req *paginationV1.PagingRequest) (*couponV1.ListCouponTemplateResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *CouponTemplateService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*couponV1.CountCouponTemplateResponse, error) {
	count, err := s.repo.Count(ctx, req)
	if err != nil {
		return nil, err
	}

	return &couponV1.CountCouponTemplateResponse{
		Count: uint64(count),
	}, nil
}

func (s *CouponTemplateService) Get(ctx context.Context, req *couponV1.GetCouponTemplateRequest) (*couponV1.CouponTemplate, error) {
	return s.repo.Get(ctx, req)
}

func (s *CouponTemplateService) Create(ctx context.Context, req *couponV1.CreateCouponTemplateRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, couponV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *CouponTemplateService) Update(ctx context.Context, req *couponV1.UpdateCouponTemplateRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, couponV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *CouponTemplateService) Delete(ctx context.Context, req *couponV1.DeleteCouponTemplateRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
