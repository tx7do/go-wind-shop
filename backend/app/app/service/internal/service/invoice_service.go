package service

import (
	"context"
	"encoding/json"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	invoiceV1 "go-wind-shop/api/gen/go/invoice/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

// InvoiceService 发票前台服务（app BFF，REST → gRPC 转发）。
//
// 裁剪写 RPC：仅 List/Get，无 Create/Update/Delete——用户不能自己开发票。
// 用户隔离双重保障：
//   - BFF fail-closed：List 强制注入 userId=当前登录用户，JSON 解析失败即拒（仿
//     user_coupon_service.go 的 List）。
//   - core UserPrivacy：invoice 表注入 UserPrivacy，按 viewer.user_id 自动注入
//     WHERE 条件。Get 的隔离完全依赖 core UserPrivacy（请求体无 user_id 字段
//     可注入，但 ent privacy 层强制兜底，跨用户 invoice_id 查不到）。
type InvoiceService struct {
	appV1.InvoiceServiceHTTPServer

	log *log.Helper

	invoiceServiceClient invoiceV1.InvoiceServiceClient
}

func NewInvoiceService(
	ctx *bootstrap.Context,
	invoiceServiceClient invoiceV1.InvoiceServiceClient,
) *InvoiceService {
	return &InvoiceService{
		log:                  ctx.NewLoggerHelper("invoice/service/app-service"),
		invoiceServiceClient: invoiceServiceClient,
	}
}

// List 强制把 userId=当前用户 注入到分页 query，确保只返回本人的发票。
// JSON 解析/序列化任一失败均 fail-closed 返回错误，避免保留客户端原 query 越权。
func (s *InvoiceService) List(ctx context.Context, req *paginationV1.PagingRequest) (*invoiceV1.ListInvoiceResponse, error) {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	userId := operator.GetUserId()

	queryMap := map[string]any{}
	if raw := req.GetQuery(); raw != "" {
		if jErr := json.Unmarshal([]byte(raw), &queryMap); jErr != nil {
			return nil, appV1.ErrorInternalServerError("internal error")
		}
	}
	queryMap["userId"] = userId
	newJSON, mErr := json.Marshal(queryMap)
	if mErr != nil {
		return nil, appV1.ErrorInternalServerError("internal error")
	}
	req.FilteringType = &paginationV1.PagingRequest_Query{Query: string(newJSON)}

	return s.invoiceServiceClient.List(ctx, req)
}

// Get 透传。隔离靠 core UserPrivacy（按 viewer.user_id 自动注入 WHERE，跨用户
// invoice_id 查不到→NotFound）。BFF 侧请求体无 user_id 字段可注入。
func (s *InvoiceService) Get(ctx context.Context, req *invoiceV1.GetInvoiceRequest) (*invoiceV1.Invoice, error) {
	return s.invoiceServiceClient.Get(ctx, req)
}
