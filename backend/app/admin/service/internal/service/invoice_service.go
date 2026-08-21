package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	adminV1 "go-wind-shop/api/gen/go/admin/service/v1"
	invoiceV1 "go-wind-shop/api/gen/go/invoice/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

// InvoiceService 发票管理（admin BFF，REST → gRPC 转发）。
//
// 全 CRUD：admin 可从订单数据生成发票记录、审核开票、作废。
// 写操作注入 CreatedBy/UpdatedBy（取自运营 token），客户端传值被忽略。
// 发票关联的 order_id/user_id 由 admin 创建时指定（运营代开）。
type InvoiceService struct {
	adminV1.InvoiceServiceHTTPServer

	log *log.Helper

	invoiceServiceClient invoiceV1.InvoiceServiceClient
}

func NewInvoiceService(
	ctx *bootstrap.Context,
	invoiceServiceClient invoiceV1.InvoiceServiceClient,
) *InvoiceService {
	return &InvoiceService{
		log:                  ctx.NewLoggerHelper("invoice/service/admin-service"),
		invoiceServiceClient: invoiceServiceClient,
	}
}

func (s *InvoiceService) List(ctx context.Context, req *paginationV1.PagingRequest) (*invoiceV1.ListInvoiceResponse, error) {
	return s.invoiceServiceClient.List(ctx, req)
}

func (s *InvoiceService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*invoiceV1.CountInvoiceResponse, error) {
	return s.invoiceServiceClient.Count(ctx, req)
}

func (s *InvoiceService) Get(ctx context.Context, req *invoiceV1.GetInvoiceRequest) (*invoiceV1.Invoice, error) {
	return s.invoiceServiceClient.Get(ctx, req)
}

func (s *InvoiceService) Create(ctx context.Context, req *invoiceV1.CreateInvoiceRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	return s.invoiceServiceClient.Create(ctx, req)
}

func (s *InvoiceService) Update(ctx context.Context, req *invoiceV1.UpdateInvoiceRequest) (*emptypb.Empty, error) {
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

	return s.invoiceServiceClient.Update(ctx, req)
}

func (s *InvoiceService) Delete(ctx context.Context, req *invoiceV1.DeleteInvoiceRequest) (*emptypb.Empty, error) {
	return s.invoiceServiceClient.Delete(ctx, req)
}
