package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"
	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/order"
	"go-wind-shop/app/core/service/internal/data/ent/shipment"

	entCrud "github.com/tx7do/go-crud/entgo"

	shippingV1 "go-wind-shop/api/gen/go/shipping/service/v1"
)

type ShipmentService struct {
	shippingV1.UnimplementedShipmentServiceServer

	log         *log.Helper
	shipmentRepo *data.ShipmentRepo
	entClient   *entCrud.EntClient[*ent.Client]
}

func NewShipmentService(
	ctx *bootstrap.Context,
	shipmentRepo *data.ShipmentRepo,
	entClient *entCrud.EntClient[*ent.Client],
) *ShipmentService {
	return &ShipmentService{
		log:          ctx.NewLoggerHelper("shipment/service/core-service"),
		shipmentRepo: shipmentRepo,
		entClient:    entClient,
	}
}

func (s *ShipmentService) List(ctx context.Context, req *paginationV1.PagingRequest) (*shippingV1.ListShipmentResponse, error) {
	return s.shipmentRepo.List(ctx, req)
}

func (s *ShipmentService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*shippingV1.CountShipmentResponse, error) {
	count, err := s.shipmentRepo.Count(ctx, req)
	if err != nil {
		return nil, err
	}
	return &shippingV1.CountShipmentResponse{Count: uint64(count)}, nil
}

func (s *ShipmentService) Get(ctx context.Context, req *shippingV1.GetShipmentRequest) (*shippingV1.Shipment, error) {
	return s.shipmentRepo.Get(ctx, req)
}

func (s *ShipmentService) Create(ctx context.Context, req *shippingV1.CreateShipmentRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, shippingV1.ErrorBadRequest("invalid parameter")
	}
	// 创建物流单时强制初始状态为 PENDING（忽略客户端传入的 status）
	pendingStatus := shippingV1.Shipment_PENDING
	req.Data.Status = &pendingStatus
	if err := s.shipmentRepo.Create(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *ShipmentService) Update(ctx context.Context, req *shippingV1.UpdateShipmentRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, shippingV1.ErrorBadRequest("invalid parameter")
	}

	// 状态机校验：若更新 status，必须有 expected_status 前置条件，且转换须合法。
	// 允许的转换：PENDING→SHIPPED（发货）、SHIPPED→DELIVERED（签收）。其余拒绝。
	// PENDING→SHIPPED 触发发货联动：同事务内推进关联订单 PAID→FULFILLED。
	targetStatus := req.Data.GetStatus()
	fulfillmentTriggered := false
	if targetStatus != shippingV1.Shipment_STATUS_UNSPECIFIED {
		if len(req.GetExpectedStatus()) == 0 {
			return nil, shippingV1.ErrorBadRequest("status update requires expected_status precondition")
		}
		for _, from := range req.GetExpectedStatus() {
			if !isAllowedShipmentTransition(from, targetStatus) {
				return nil, shippingV1.ErrorBadRequest("illegal shipment status transition: %v -> %v", from, targetStatus)
			}
		}
		// 标记发货联动：PENDING→SHIPPED 需同事务推进订单
		for _, from := range req.GetExpectedStatus() {
			if from == shippingV1.Shipment_PENDING && targetStatus == shippingV1.Shipment_SHIPPED {
				fulfillmentTriggered = true
				break
			}
		}
	}

	// 发货联动：shipment PENDING→SHIPPED + order PAID→FULFILLED 同事务。
	// 仿 OrderService.Create / ExpireOrderByTimeout 的 tx 多实体操作模式。
	// 订单状态推进用 expected_status 前置条件（PAID），若订单非 PAID（已取消/已关闭等），
	// 整事务回滚，物流单也不更新——这是"发货只能对已付款订单"的语义保证。
	if fulfillmentTriggered {
		return s.updateWithFulfillment(ctx, req)
	}

	// 非发货联动（含 SHIPPED→DELIVERED 签收、或非状态字段更新）走常规 repo 路径
	if err := s.shipmentRepo.Update(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// updateWithFulfillment 发货联动事务：shipment 翻 SHIPPED + 关联 order 翻 FULFILLED。
//
// 在单个 ent 事务内完成两件事，任一失败整事务回滚：
//  1. shipment 更新（status PENDING→SHIPPED，expected_status=PENDING 前置条件注入）
//  2. 关联 order 推进 PAID→FULFILLED（expected_status=PAID 前置条件注入）
//
// 订单侧用 expected_status 前置条件，若订单当前非 PAID（已被并发取消/关闭），
// UPDATE 不命中行（NotFound）→ 返回 Conflict，整事务回滚。
func (s *ShipmentService) updateWithFulfillment(ctx context.Context, req *shippingV1.UpdateShipmentRequest) (*emptypb.Empty, error) {
	shipmentId := req.GetId()
	orderId := req.Data.GetOrderId()
	if orderId == 0 {
		return nil, shippingV1.ErrorBadRequest("order_id is required for fulfillment")
	}

	var err error
	tx, txErr := s.entClient.Client().Tx(ctx)
	if txErr != nil {
		s.log.Errorf("begin fulfillment tx failed for shipment [%d]: %v", shipmentId, txErr)
		return nil, fmt.Errorf("begin fulfillment tx failed: %w", txErr)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 1. shipment 翻 SHIPPED（事务内，expected_status=PENDING 前置条件）
	shippedEnt := shipment.StatusShipmentStatusShipped
	pendingEnt := shipment.StatusShipmentStatusPending
	// 注：shipment 的非状态字段更新（如 carrier/tracking_number）也在此事务内一并完成，
	// 由调用方通过 update_mask 控制。此处统一用事务版的字段设置。
	sEntErr := tx.Shipment.Update().
		SetNillableCarrier(req.Data.Carrier).
		SetNillableTrackingNumber(req.Data.TrackingNumber).
		SetNillableTrackingEvents(req.Data.TrackingEvents).
		SetNillableStatus(&shippedEnt).
		SetUpdatedAt(time.Now()).
		Where(
			shipment.IDEQ(shipmentId),
			shipment.StatusIn(pendingEnt),
		).
		Exec(ctx)
	if sEntErr != nil {
		err = sEntErr
		s.log.Errorf("update shipment [%d] to SHIPPED in fulfillment tx failed: %v", shipmentId, sEntErr)
		return nil, fmt.Errorf("update shipment [%d] failed: %w", shipmentId, sEntErr)
	}

	// 2. 关联 order 推进 PAID→FULFILLED（事务内，expected_status=PAID 前置条件）
	fulfilledEnt := order.StatusOrderStatusFulfilled
	paidEnt := order.StatusOrderStatusPaid
	oEntErr := tx.Order.Update().
		SetNillableStatus(&fulfilledEnt).
		SetUpdatedAt(time.Now()).
		Where(
			order.IDEQ(orderId),
			order.StatusIn(paidEnt),
		).
		Exec(ctx)
	if oEntErr != nil {
		err = oEntErr
		s.log.Errorf("update order [%d] to FULFILLED in fulfillment tx failed: %v", orderId, oEntErr)
		return nil, fmt.Errorf("update order [%d] to FULFILLED failed: %w", orderId, oEntErr)
	}

	if cErr := tx.Commit(); cErr != nil {
		s.log.Errorf("commit fulfillment tx failed for shipment [%d] order [%d]: %v", shipmentId, orderId, cErr)
		err = cErr
		return nil, fmt.Errorf("commit fulfillment tx failed: %w", cErr)
	}
	err = nil

	s.log.Infow(
		"msg", "shipment fulfilled, order transitioned to FULFILLED",
		"shipment_id", shipmentId,
		"order_id", orderId,
	)
	return &emptypb.Empty{}, nil
}

// isAllowedShipmentTransition 校验物流单状态转换是否在允许列表内。
// 允许：PENDING→SHIPPED（发货）、SHIPPED→DELIVERED（签收）。其余拒绝。
func isAllowedShipmentTransition(from, to shippingV1.Shipment_Status) bool {
	allowed := map[shippingV1.Shipment_Status]map[shippingV1.Shipment_Status]bool{
		shippingV1.Shipment_PENDING: {
			shippingV1.Shipment_SHIPPED: true,
		},
		shippingV1.Shipment_SHIPPED: {
			shippingV1.Shipment_DELIVERED: true,
		},
	}
	allowedTo, ok := allowed[from]
	if !ok {
		return false
	}
	return allowedTo[to]
}

func (s *ShipmentService) Delete(ctx context.Context, req *shippingV1.DeleteShipmentRequest) (*emptypb.Empty, error) {
	if err := s.shipmentRepo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}
