package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	adminV1 "go-wind-shop/api/gen/go/admin/service/v1"
	shippingV1 "go-wind-shop/api/gen/go/shipping/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

type ShipmentService struct {
	adminV1.ShipmentServiceHTTPServer

	log *log.Helper

	shipmentServiceClient shippingV1.ShipmentServiceClient
}

func NewShipmentService(
	ctx *bootstrap.Context,
	shipmentServiceClient shippingV1.ShipmentServiceClient,
) *ShipmentService {
	return &ShipmentService{
		log:                   ctx.NewLoggerHelper("shipment/service/admin-service"),
		shipmentServiceClient: shipmentServiceClient,
	}
}

func (s *ShipmentService) List(ctx context.Context, req *paginationV1.PagingRequest) (*shippingV1.ListShipmentResponse, error) {
	return s.shipmentServiceClient.List(ctx, req)
}

func (s *ShipmentService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*shippingV1.CountShipmentResponse, error) {
	return s.shipmentServiceClient.Count(ctx, req)
}

func (s *ShipmentService) Get(ctx context.Context, req *shippingV1.GetShipmentRequest) (*shippingV1.Shipment, error) {
	return s.shipmentServiceClient.Get(ctx, req)
}

func (s *ShipmentService) Create(ctx context.Context, req *shippingV1.CreateShipmentRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.CreatedBy = trans.Ptr(operator.GetUserId())
	req.Data.UpdatedBy = nil

	// 创建物流单时强制初始状态 PENDING（core ShipmentService.Create 会再次覆盖保证）
	pendingStatus := shippingV1.Shipment_PENDING
	req.Data.Status = &pendingStatus

	return s.shipmentServiceClient.Create(ctx, req)
}

func (s *ShipmentService) Update(ctx context.Context, req *shippingV1.UpdateShipmentRequest) (*emptypb.Empty, error) {
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

	// 状态变更补全 expected_status（core ShipmentService.Update 强制要求）：
	//   - 目标 SHIPPED（发货）：当前须为 PENDING
	//   - 目标 DELIVERED（签收）：当前须为 SHIPPED
	// 其他目标态留空，由 core 拒绝。
	targetStatus := req.Data.GetStatus()
	switch targetStatus {
	case shippingV1.Shipment_SHIPPED:
		req.ExpectedStatus = []shippingV1.Shipment_Status{shippingV1.Shipment_PENDING}
	case shippingV1.Shipment_DELIVERED:
		req.ExpectedStatus = []shippingV1.Shipment_Status{shippingV1.Shipment_SHIPPED}
	}

	return s.shipmentServiceClient.Update(ctx, req)
}

func (s *ShipmentService) Delete(ctx context.Context, req *shippingV1.DeleteShipmentRequest) (*emptypb.Empty, error) {
	return s.shipmentServiceClient.Delete(ctx, req)
}
