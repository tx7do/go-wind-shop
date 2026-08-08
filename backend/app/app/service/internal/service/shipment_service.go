package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	shippingV1 "go-wind-shop/api/gen/go/shipping/service/v1"
)

// ShipmentService 物流单前台服务（BFF 网关，REST → gRPC 转发）。
// 仅暴露只读 List/Get：买家查询自己订单的物流轨迹。无 Create/Update/Delete。
// 行级隔离由 core 层 UserPrivacy 策略强制（user_id = 当前登录用户）。
type ShipmentService struct {
	appV1.ShipmentServiceHTTPServer

	log *log.Helper

	shipmentServiceClient shippingV1.ShipmentServiceClient
}

func NewShipmentService(
	ctx *bootstrap.Context,
	shipmentServiceClient shippingV1.ShipmentServiceClient,
) *ShipmentService {
	return &ShipmentService{
		log:                   ctx.NewLoggerHelper("shipment/service/app-service"),
		shipmentServiceClient: shipmentServiceClient,
	}
}

func (s *ShipmentService) List(ctx context.Context, req *paginationV1.PagingRequest) (*shippingV1.ListShipmentResponse, error) {
	return s.shipmentServiceClient.List(ctx, req)
}

func (s *ShipmentService) Get(ctx context.Context, req *shippingV1.GetShipmentRequest) (*shippingV1.Shipment, error) {
	return s.shipmentServiceClient.Get(ctx, req)
}
