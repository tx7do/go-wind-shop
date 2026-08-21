package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	taxV1 "go-wind-shop/api/gen/go/tax/service/v1"
)

type TaxRateService struct {
	taxV1.UnimplementedTaxRateServiceServer

	log *log.Helper

	repo *data.TaxRateRepo
}

func NewTaxRateService(ctx *bootstrap.Context, repo *data.TaxRateRepo) *TaxRateService {
	return &TaxRateService{
		log:  ctx.NewLoggerHelper("tax-rate/service/core-service"),
		repo: repo,
	}
}

func (s *TaxRateService) List(ctx context.Context, req *paginationV1.PagingRequest) (*taxV1.ListTaxRateResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *TaxRateService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*taxV1.CountTaxRateResponse, error) {
	count, err := s.repo.Count(ctx, req)
	if err != nil {
		return nil, err
	}

	return &taxV1.CountTaxRateResponse{
		Count: uint64(count),
	}, nil
}

func (s *TaxRateService) Get(ctx context.Context, req *taxV1.GetTaxRateRequest) (*taxV1.TaxRate, error) {
	return s.repo.Get(ctx, req)
}

func (s *TaxRateService) Create(ctx context.Context, req *taxV1.CreateTaxRateRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, taxV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *TaxRateService) Update(ctx context.Context, req *taxV1.UpdateTaxRateRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, taxV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *TaxRateService) Delete(ctx context.Context, req *taxV1.DeleteTaxRateRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
