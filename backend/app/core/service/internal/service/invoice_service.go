package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	invoiceV1 "go-wind-shop/api/gen/go/invoice/service/v1"
)

type InvoiceService struct {
	invoiceV1.UnimplementedInvoiceServiceServer

	log *log.Helper

	repo *data.InvoiceRepo
}

func NewInvoiceService(ctx *bootstrap.Context, repo *data.InvoiceRepo) *InvoiceService {
	return &InvoiceService{
		log:  ctx.NewLoggerHelper("invoice/service/core-service"),
		repo: repo,
	}
}

func (s *InvoiceService) List(ctx context.Context, req *paginationV1.PagingRequest) (*invoiceV1.ListInvoiceResponse, error) {
	return s.repo.List(ctx, req)
}

func (s *InvoiceService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*invoiceV1.CountInvoiceResponse, error) {
	count, err := s.repo.Count(ctx, req)
	if err != nil {
		return nil, err
	}

	return &invoiceV1.CountInvoiceResponse{
		Count: uint64(count),
	}, nil
}

func (s *InvoiceService) Get(ctx context.Context, req *invoiceV1.GetInvoiceRequest) (*invoiceV1.Invoice, error) {
	return s.repo.Get(ctx, req)
}

func (s *InvoiceService) Create(ctx context.Context, req *invoiceV1.CreateInvoiceRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, invoiceV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Create(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *InvoiceService) Update(ctx context.Context, req *invoiceV1.UpdateInvoiceRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, invoiceV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.repo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *InvoiceService) Delete(ctx context.Context, req *invoiceV1.DeleteInvoiceRequest) (*emptypb.Empty, error) {
	if err := s.repo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
