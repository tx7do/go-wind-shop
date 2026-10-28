package service

import (
	"context"

	"go-wind-shop/pkg/logx"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	orderV1 "go-wind-shop/api/gen/go/order/service/v1"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"go-wind-shop/pkg/middleware/auth"
)

type OrderService struct {
	appV1.OrderServiceHTTPServer

	log *logx.Helper

	orderServiceClient orderV1.OrderServiceClient
}

func NewOrderService(
	ctx *bootstrap.Context,
	orderServiceClient orderV1.OrderServiceClient,
) *OrderService {
	return &OrderService{
		log:                logx.New(ctx.NewLoggerHelper("order/service/app-service")),
		orderServiceClient: orderServiceClient,
	}
}

func (s *OrderService) List(ctx context.Context, req *paginationV1.PagingRequest) (*orderV1.ListOrderResponse, error) {
	return s.orderServiceClient.List(ctx, req)
}

func (s *OrderService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*orderV1.CountOrderResponse, error) {
	return s.orderServiceClient.Count(ctx, req)
}

func (s *OrderService) Get(ctx context.Context, req *orderV1.GetOrderRequest) (*orderV1.Order, error) {
	return s.orderServiceClient.Get(ctx, req)
}

func (s *OrderService) Create(ctx context.Context, req *orderV1.CreateOrderRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, appV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 强制覆盖为当前登录用户，防止客户端伪造 userId/tenantId 越权
	// 下单他人购物车或将订单挂到他人名下。
	req.Data.UserId = trans.Ptr(operator.GetUserId())
	req.Data.TenantId = trans.Ptr(operator.GetTenantId())
	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	return s.orderServiceClient.Create(ctx, req)
}

func (s *OrderService) Update(ctx context.Context, req *orderV1.UpdateOrderRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, appV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 前台买家不允许通过通用 PATCH 推进订单状态：取消/确认收货走专用 RPC
	// （CancelOrder 会释放库存/清理孤儿支付/返还优惠券，ConfirmReceipt 有专属
	// 通知），通用路径推进状态会绕过这些联动。
	if req.Data.Status != nil && req.Data.GetStatus() != orderV1.Order_STATUS_UNSPECIFIED {
		return nil, appV1.ErrorBadRequest("order status change via generic update is not allowed; use cancel / confirm-receipt endpoints")
	}

	// 金额/地址等持久字段是下单事务计算的单一真相源，不接受客户端改写——
	// 仅放行 updated_by（审计），update_mask 同步收紧，防"下单后 PATCH
	// total_amount=0"类篡改。
	req.Data.Id = trans.Ptr(req.GetId())
	req.Data.UpdatedBy = trans.Ptr(operator.GetUserId())
	req.UpdateMask = &fieldmaskpb.FieldMask{Paths: []string{"updated_by"}}

	return s.orderServiceClient.Update(ctx, req)
}

// CancelOrder 买家取消本人 PENDING_PAYMENT 订单。
// 归属校验与状态机由 core 强制（UserPrivacy + viewer 比对），BFF 仅要求登录。
func (s *OrderService) CancelOrder(ctx context.Context, req *orderV1.CancelOrderRequest) (*emptypb.Empty, error) {
	if _, err := auth.FromContext(ctx); err != nil {
		return nil, err
	}
	return s.orderServiceClient.CancelOrder(ctx, req)
}

// ConfirmReceipt 买家确认收货（本人 FULFILLED 订单 → CLOSED）。
func (s *OrderService) ConfirmReceipt(ctx context.Context, req *orderV1.ConfirmReceiptRequest) (*emptypb.Empty, error) {
	if _, err := auth.FromContext(ctx); err != nil {
		return nil, err
	}
	return s.orderServiceClient.ConfirmReceipt(ctx, req)
}
