package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	adminV1 "go-wind-shop/api/gen/go/admin/service/v1"
	taxV1 "go-wind-shop/api/gen/go/tax/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

// TaxRateService 税率规则管理（admin BFF，REST → gRPC 转发）。
type TaxRateService struct {
	adminV1.TaxRateServiceHTTPServer

	log *log.Helper

	taxRateServiceClient taxV1.TaxRateServiceClient
}

func NewTaxRateService(
	ctx *bootstrap.Context,
	taxRateServiceClient taxV1.TaxRateServiceClient,
) *TaxRateService {
	return &TaxRateService{
		log:                  ctx.NewLoggerHelper("tax-rate/service/admin-service"),
		taxRateServiceClient: taxRateServiceClient,
	}
}

func (s *TaxRateService) List(ctx context.Context, req *paginationV1.PagingRequest) (*taxV1.ListTaxRateResponse, error) {
	return s.taxRateServiceClient.List(ctx, req)
}

func (s *TaxRateService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*taxV1.CountTaxRateResponse, error) {
	return s.taxRateServiceClient.Count(ctx, req)
}

func (s *TaxRateService) Get(ctx context.Context, req *taxV1.GetTaxRateRequest) (*taxV1.TaxRate, error) {
	return s.taxRateServiceClient.Get(ctx, req)
}

func (s *TaxRateService) Create(ctx context.Context, req *taxV1.CreateTaxRateRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	return s.taxRateServiceClient.Create(ctx, req)
}

func (s *TaxRateService) Update(ctx context.Context, req *taxV1.UpdateTaxRateRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
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

	return s.taxRateServiceClient.Update(ctx, req)
}

func (s *TaxRateService) Delete(ctx context.Context, req *taxV1.DeleteTaxRateRequest) (*emptypb.Empty, error) {
	return s.taxRateServiceClient.Delete(ctx, req)
}
