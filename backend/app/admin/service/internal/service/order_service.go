package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	adminV1 "go-wind-shop/api/gen/go/admin/service/v1"
	orderV1 "go-wind-shop/api/gen/go/order/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

type OrderService struct {
	adminV1.OrderServiceHTTPServer

	log *log.Helper

	orderServiceClient orderV1.OrderServiceClient
}

func NewOrderService(
	ctx *bootstrap.Context,
	orderServiceClient orderV1.OrderServiceClient,
) *OrderService {
	return &OrderService{
		log:                ctx.NewLoggerHelper("order/service/admin-service"),
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

func (s *OrderService) Update(ctx context.Context, req *orderV1.UpdateOrderRequest) (*emptypb.Empty, error) {
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

	// 状态变更必须携带 expected_status 前置条件（core OrderService.Update 强制要求）。
	// 发货（PAID→FULFILLED）、关闭（PAID→CLOSED 或 FULFILLED→CLOSED）均属状态机白名单内转换，
	// 这里按目标状态补全 expected_status，让 core 的乐观状态机校验通过：
	//   - 目标 FULFILLED：当前须为 PAID
	//   - 目标 CLOSED：当前须为 PAID 或 FULFILLED
	// 目标 PENDING_PAYMENT/CANCELLED 不由 admin 发起，无需补；其他目标态留空由 core 拒绝。
	targetStatus := req.Data.GetStatus()
	switch targetStatus {
	case orderV1.Order_FULFILLED:
		req.ExpectedStatus = []orderV1.Order_Status{orderV1.Order_PAID}
	case orderV1.Order_CLOSED:
		req.ExpectedStatus = []orderV1.Order_Status{orderV1.Order_PAID, orderV1.Order_FULFILLED}
	}

	return s.orderServiceClient.Update(ctx, req)
}

func (s *OrderService) Delete(ctx context.Context, req *orderV1.DeleteOrderRequest) (*emptypb.Empty, error) {
	return s.orderServiceClient.Delete(ctx, req)
}
