package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	adminV1 "go-wind-shop/api/gen/go/admin/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

// StockAlertService 库存预警记录管理（admin BFF，REST → gRPC 转发）。
// 裁剪版：仅 List/Get/Update。Update 唯一允许的转移 OPEN→RESOLVED 由 core
// service 强制校验；BFF 此处照搬 shipping_rate 的 Update 注入范式（id 从 URL
// path 绑定、updated_by 从 operator 注入、update_mask 追加 updated_by）。
type StockAlertService struct {
	adminV1.StockAlertServiceHTTPServer

	log *log.Helper

	stockAlertServiceClient catalogV1.StockAlertServiceClient
}

func NewStockAlertService(
	ctx *bootstrap.Context,
	stockAlertServiceClient catalogV1.StockAlertServiceClient,
) *StockAlertService {
	return &StockAlertService{
		log:                     ctx.NewLoggerHelper("stock-alert/service/admin-service"),
		stockAlertServiceClient: stockAlertServiceClient,
	}
}

func (s *StockAlertService) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListStockAlertResponse, error) {
	return s.stockAlertServiceClient.List(ctx, req)
}

func (s *StockAlertService) Get(ctx context.Context, req *catalogV1.GetStockAlertRequest) (*catalogV1.StockAlert, error) {
	return s.stockAlertServiceClient.Get(ctx, req)
}

func (s *StockAlertService) Update(ctx context.Context, req *catalogV1.UpdateStockAlertRequest) (*emptypb.Empty, error) {
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

	return s.stockAlertServiceClient.Update(ctx, req)
}
