package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"github.com/hibiken/asynq"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"
	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/cart"
	"go-wind-shop/app/core/service/internal/data/ent/cartitem"
	"go-wind-shop/app/core/service/internal/data/ent/order"
	"go-wind-shop/app/core/service/internal/data/ent/orderitem"
	"go-wind-shop/app/core/service/internal/data/ent/sku"
	"go-wind-shop/app/core/service/internal/data/ent/skuprice"

	entCrud "github.com/tx7do/go-crud/entgo"

	orderV1 "go-wind-shop/api/gen/go/order/service/v1"
	appViewer "go-wind-shop/pkg/entgo/viewer"
	"go-wind-shop/pkg/task"
)

// OrderPendingTTL 订单待支付超时时间，超时后自动取消并释放库存。
const OrderPendingTTL = 30 * time.Minute

type OrderService struct {
	orderV1.UnimplementedOrderServiceServer

	log *log.Helper

	orderRepo     *data.OrderRepo
	orderItemRepo *data.OrderItemRepo
	skuRepo       *data.SkuRepo
	cartItemRepo  *data.CartItemRepo
	entClient     *entCrud.EntClient[*ent.Client]

	taskScheduler TaskScheduler
}

func NewOrderService(
	ctx *bootstrap.Context,
	orderRepo *data.OrderRepo,
	orderItemRepo *data.OrderItemRepo,
	skuRepo *data.SkuRepo,
	cartItemRepo *data.CartItemRepo,
	entClient *entCrud.EntClient[*ent.Client],
) *OrderService {
	return &OrderService{
		log:           ctx.NewLoggerHelper("order/service/core-service"),
		orderRepo:     orderRepo,
		orderItemRepo: orderItemRepo,
		skuRepo:       skuRepo,
		cartItemRepo:  cartItemRepo,
		entClient:     entClient,
	}
}

// RegisterTaskScheduler 由 server/asynq_server.go 在启动期注入 asynq server。
func (s *OrderService) RegisterTaskScheduler(ts TaskScheduler) {
	s.taskScheduler = ts
}

func (s *OrderService) List(ctx context.Context, req *paginationV1.PagingRequest) (*orderV1.ListOrderResponse, error) {
	return s.orderRepo.List(ctx, req)
}

func (s *OrderService) Get(ctx context.Context, req *orderV1.GetOrderRequest) (*orderV1.Order, error) {
	orderEnt, err := s.orderRepo.Get(ctx, req)
	if err != nil {
		return nil, err
	}
	// 越权校验：普通用户只能查自己的订单。系统/平台视图（如 asynq 超时任务）放行。
	viewerUid, ok := viewerUserIDFromContext(ctx)
	if ok && orderEnt.GetUserId() != viewerUid {
		s.log.Warnf("user [%d] attempted to access order [%d] belonging to user [%d]", viewerUid, orderEnt.GetId(), orderEnt.GetUserId())
		return nil, orderV1.ErrorForbidden("order not found")
	}
	return orderEnt, nil
}

// Update 通用更新。
// 状态变更（status != UNSPECIFIED）必须携带 expected_status 前置条件，否则拒绝。
// 这保证了 MarkPaid（PENDING_PAYMENT→PAID，带 expected_status=[PENDING_PAYMENT]）与超时取消
// （PENDING_PAYMENT→CANCELLED，带 expected_status=[PENDING_PAYMENT]）都走乐观并发路径，
// 防止并发覆盖终态。
func (s *OrderService) Update(ctx context.Context, req *orderV1.UpdateOrderRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, orderV1.ErrorBadRequest("invalid parameter")
	}

	// 状态变更强制要求前置条件
	if req.Data.Status != nil && req.Data.GetStatus() != orderV1.Order_STATUS_UNSPECIFIED {
		if len(req.GetExpectedStatus()) == 0 {
			return nil, orderV1.ErrorBadRequest("status update requires expected_status precondition")
		}
	}

	if err := s.orderRepo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *OrderService) Delete(ctx context.Context, req *orderV1.DeleteOrderRequest) (*emptypb.Empty, error) {
	if err := s.orderRepo.Delete(ctx, req); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// Create 下单。
// 单 ent 事务内：
//  1. 插入 Order（状态 PENDING_PAYMENT，携带 idempotency_key/business_ref_id/currency 等支付 mixin 字段）
//  2. 遍历购物车项：插入 OrderItem（含 SKU 快照）+ 扣减对应 SKU 的 stock_qty
//  3. 清空该购物车的所有 CartItem
//  4. 提交事务
//  5. 注册 asynq 订单超时延时任务（N 分钟后未支付自动 CANCELLED + 释放库存）
//
// 任一步失败整事务回滚。库存扣减采用行级锁（ent sql/lock feature）防超卖。
func (s *OrderService) Create(ctx context.Context, req *orderV1.CreateOrderRequest) (*emptypb.Empty, error) {
	if req == nil || req.Data == nil {
		return nil, orderV1.ErrorBadRequest("invalid parameter")
	}
	if req.Data.GetIdempotencyKey() == "" {
		return nil, orderV1.ErrorBadRequest("idempotency_key is required")
	}
	tenantId := req.Data.GetTenantId()
	userId := req.Data.GetUserId()
	if tenantId == 0 {
		return nil, orderV1.ErrorBadRequest("tenant_id is required")
	}
	if userId == 0 {
		return nil, orderV1.ErrorBadRequest("user_id is required")
	}

	// 开启事务
	tx, err := s.entClient.Client().Tx(ctx)
	if err != nil {
		s.log.Errorf("begin tx failed: %v", err)
		return nil, orderV1.ErrorInternalServerError("begin tx failed")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// 0. 在事务内查找该用户的购物车（(tenant_id,user_id) 唯一）。
	//    购物车不存在视为空购物车 → 拒绝下单。
	cartEnt, cErr := tx.Cart.Query().
		Where(
			cart.TenantIDEQ(tenantId),
			cart.UserIDEQ(userId),
		).
		Only(ctx)
	if cErr != nil || cartEnt == nil {
		s.log.Warnf("cart not found for tenant [%d] user [%d]: %v", tenantId, userId, cErr)
		err = orderV1.ErrorBadRequest("cart is empty or not found")
		return nil, err
	}
	cartIdVal := cartEnt.ID

	// 1. 在事务内插入 Order。
	//    复用 orderRepo 的字段映射逻辑，但走事务客户端。
	orderEntBuilder := tx.Order.Create().
		SetNillableUserID(req.Data.UserId).
		SetNillableTotalAmount(req.Data.TotalAmount).
		SetNillableCurrency(req.Data.Currency).
		SetNillableBusinessRefID(req.Data.BusinessRefId).
		SetNillableIdempotencyKey(req.Data.IdempotencyKey).
		SetNillableRecipientName(req.Data.RecipientName).
		SetNillableRecipientPhone(req.Data.RecipientPhone).
		SetNillableShippingAddress(req.Data.ShippingAddress).
		SetNillableTenantID(req.Data.TenantId).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	// 状态固定为 PENDING_PAYMENT（下单初始态，忽略客户端传入）
	pendingStatus := order.StatusOrderStatusPendingPayment
	orderEntBuilder.SetStatus(pendingStatus)

	if err = orderEntBuilder.Exec(ctx); err != nil {
		s.log.Errorf("insert order failed: %v", err)
		err = orderV1.ErrorInternalServerError("create order failed")
		return nil, err
	}

	// 取回刚创建的订单ID（用于关联 OrderItem 与超时任务）。
	createdOrder, qErr := tx.Order.Query().
		Where(
			order.IdempotencyKeyEQ(req.Data.GetIdempotencyKey()),
			order.TenantIDEQ(tenantId),
		).
		Only(ctx)
	if qErr != nil || createdOrder == nil {
		s.log.Errorf("query back created order failed: %v", qErr)
		err = orderV1.ErrorInternalServerError("create order failed")
		return nil, err
	}
	orderId := createdOrder.ID

	// 2. 查询购物车项（事务内），为每个项插入 OrderItem + 扣减 SKU 库存。
	cartItems, qErr := tx.CartItem.Query().
		Where(cartitem.CartIDEQ(cartIdVal)).
		All(ctx)
	if qErr != nil {
		s.log.Errorf("query cart items failed: %v", qErr)
		err = orderV1.ErrorInternalServerError("create order failed")
		return nil, err
	}

	var totalAmount int64 = 0
	for _, ci := range cartItems {
		skuId := ci.SkuID
		if skuId == nil || *skuId == 0 {
			continue
		}
		quantity := ci.Quantity
		if quantity == nil || *quantity <= 0 {
			continue
		}
		qty := *quantity

		// 2a. 查询 SKU（带行锁防超卖），取当前价格与库存。
		skuEnt, skuErr := tx.Sku.Query().
			Where(sku.IDEQ(*skuId)).
			ForUpdate().
			Only(ctx)
		if skuErr != nil {
			s.log.Errorf("query sku [%d] for update failed: %v", *skuId, skuErr)
			err = orderV1.ErrorInternalServerError("create order failed")
			return nil, err
		}
		if skuEnt.StockQty == nil || *skuEnt.StockQty < qty {
			s.log.Warnf("insufficient stock for sku [%d]: have %d want %d", *skuId, *skuEnt.StockQty, qty)
			err = orderV1.ErrorBadRequest("insufficient stock for sku [%d]", *skuId)
			return nil, err
		}

		// 2b-1. 查询该 SKU 在结算币下的价格行（SkuPrice，sku_id + currency 唯一）。
		//       订单按固定结算币（req.Data.currency，由 Currency mixin 注入，默认 CNY）结算。
		//       无价格行或金额非法 → 不可售，整事务回滚。
		currency := req.Data.GetCurrency()
		priceEnt, priceErr := tx.SkuPrice.Query().
			Where(
				skuprice.SkuIDEQ(*skuId),
				skuprice.CurrencyEQ(currency),
			).
			Only(ctx)
		if priceErr != nil || priceEnt == nil || priceEnt.Amount == nil || *priceEnt.Amount == "" {
			s.log.Warnf("no price row for sku [%d] currency [%s]: %v", *skuId, currency, priceErr)
			err = orderV1.ErrorBadRequest("sku [%d] has no price for settlement currency", *skuId)
			return nil, err
		}
		unitPrice, parseErr := strconv.ParseInt(*priceEnt.Amount, 10, 64)
		if parseErr != nil || unitPrice <= 0 {
			s.log.Warnf("invalid price [%s] for sku [%d] currency [%s]: %v", *priceEnt.Amount, *skuId, currency, parseErr)
			err = orderV1.ErrorBadRequest("sku [%d] has invalid price", *skuId)
			return nil, err
		}
		subtotal := unitPrice * int64(qty)
		totalAmount += subtotal

		// 2b. 构造 SKU 快照 JSON（下单时名称/价格/属性固化，便于售后与展示）。
		snapshot := map[string]any{
			"sku_id":   *skuId,
			"quantity": qty,
			"currency": currency,
			"unit_price": unitPrice,
		}
		snapshotJSON, mErr := json.Marshal(snapshot)
		if mErr != nil {
			s.log.Errorf("marshal sku snapshot failed: %v", mErr)
			err = orderV1.ErrorInternalServerError("create order failed")
			return nil, err
		}

		// 2c. 插入 OrderItem。
		if oErr := tx.OrderItem.Create().
			SetNillableOrderID(&orderId).
			SetNillableSkuID(skuId).
			SetSkuSnapshot(string(snapshotJSON)).
			SetNillableQuantity(&qty).
			SetNillableUnitPrice(&unitPrice).
			SetNillableSubtotal(&subtotal).
			SetNillableCreatedBy(req.Data.CreatedBy).
			SetCreatedAt(time.Now()).
			Exec(ctx); oErr != nil {
			s.log.Errorf("insert order item failed: %v", oErr)
			err = orderV1.ErrorInternalServerError("create order failed")
			return nil, err
		}

		// 2d. 扣减 SKU 库存（行锁内）。
		newStock := *skuEnt.StockQty - qty
		if uErr := tx.Sku.UpdateOneID(*skuId).
			SetStockQty(newStock).
			Exec(ctx); uErr != nil {
			s.log.Errorf("deduct sku [%d] stock failed: %v", *skuId, uErr)
			err = orderV1.ErrorInternalServerError("create order failed")
			return nil, err
		}
	}

	// 3. 清空该购物车的所有 CartItem（事务内）。
	if _, dErr := tx.CartItem.Delete().
		Where(cartitem.CartIDEQ(cartIdVal)).
		Exec(ctx); dErr != nil {
		s.log.Errorf("clear cart [%d] items failed: %v", cartIdVal, dErr)
		err = orderV1.ErrorInternalServerError("create order failed")
		return nil, err
	}

	// 回写订单总金额（事务内）。
	if uErr := tx.Order.UpdateOneID(orderId).
		SetTotalAmount(totalAmount).
		Exec(ctx); uErr != nil {
		s.log.Errorf("update order total amount failed: %v", uErr)
		err = orderV1.ErrorInternalServerError("create order failed")
		return nil, err
	}

	// 4. 提交事务。
	if cErr := tx.Commit(); cErr != nil {
		s.log.Errorf("commit tx failed: %v", cErr)
		err = orderV1.ErrorInternalServerError("create order failed")
		return nil, err
	}
	err = nil

	// 5. 注册订单超时延时任务（事务提交后，N 分钟后触发 ExpireOrderByTimeout）。
	s.scheduleOrderTimeout(orderId)

	s.log.Infof("order [%d] created for user [%d] tenant [%d]", orderId, userId, tenantId)
	return &emptypb.Empty{}, nil
}

// scheduleOrderTimeout 投递订单超时延时任务。
// 由本地 asynq worker 在 TTL 后触发 HandleOrderTimeout → ExpireOrderByTimeout。
// task_id 用于全局幂等去重（同订单 ID 重复投递会被 asynq.Unique 去重）。
func (s *OrderService) scheduleOrderTimeout(orderId uint32) {
	if s.taskScheduler == nil {
		s.log.Warnf("task scheduler not available, skip scheduling order timeout for [%d]", orderId)
		return
	}

	taskId := task.CreateOrderTimeoutTaskID(orderId)

	taskPayload := task.OrderTimeoutTaskData{
		OrderId: orderId,
		TaskId:  taskId,
		TraceId: fmt.Sprintf("order_timeout:%d:%d", orderId, time.Now().UnixNano()),
	}

	if tErr := s.taskScheduler.NewTask(
		task.OrderTimeoutTaskType,
		taskPayload,
		asynq.ProcessIn(OrderPendingTTL),
		asynq.TaskID(taskId),
		asynq.Unique(OrderPendingTTL),
	); tErr != nil {
		s.log.Errorf("schedule order timeout task failed for [%d]: %v", orderId, tErr)
	} else {
		s.log.Infof("scheduled order timeout task for order [%d] (ttl=%s)", orderId, OrderPendingTTL)
	}
}

// ExpireOrderByTimeout 超时过期订单（由 asynq 延时任务触发）。
// 幂等检查：仅 PENDING_PAYMENT 状态才过期；否则跳过。
// 过期动作：状态 PENDING_PAYMENT→CANCELLED（带 expected_status 前置条件，乐观并发）+ 释放对应订单项占用的库存。
// MVP 支付为 STUB（不真实扣款），故无需退款；真实网关接入后此处需补充退款逻辑。
func (s *OrderService) ExpireOrderByTimeout(ctx context.Context, req *orderV1.ExpireOrderByTimeoutRequest) (*orderV1.ExpireOrderByTimeoutResponse, error) {
	s.log.Infow(
		"msg", "ExpireOrderByTimeout started",
		"task_id", req.GetTaskId(),
		"trace_id", req.GetTraceId(),
		"order_id", req.GetOrderId(),
	)

	orderId := req.GetOrderId()
	if orderId == 0 {
		return nil, orderV1.ErrorBadRequest("order_id is required")
	}

	// 查询订单（带 tenant 行级过滤由 viewer 注入）。
	orderEnt, err := s.orderRepo.Get(ctx, &orderV1.GetOrderRequest{
		QueryBy: &orderV1.GetOrderRequest_Id{Id: orderId},
	})
	if err != nil {
		return nil, fmt.Errorf("get order [%d] failed: %w", orderId, err)
	}

	// 幂等检查：仅 PENDING_PAYMENT 才需要超时处理。
	if orderEnt.GetStatus() != orderV1.Order_PENDING_PAYMENT {
		s.log.Infow(
			"msg", "order not PENDING_PAYMENT, skip timeout",
			"task_id", req.GetTaskId(),
			"trace_id", req.GetTraceId(),
			"order_id", orderId,
			"status", orderEnt.GetStatus().String(),
		)
		return &orderV1.ExpireOrderByTimeoutResponse{Expired: boolPtr(false)}, nil
	}

	// 释放库存 + 状态推进 CANCELLED 在同一 ent 事务内完成，保证原子性。
	// 任一步失败整事务回滚，避免库存已加回但订单未取消（或反之）的不一致。
	tx, err := s.entClient.Client().Tx(ctx)
	if err != nil {
		s.log.Errorf("begin expire tx failed for order [%d]: %v", orderId, err)
		return nil, fmt.Errorf("begin expire tx failed: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	items, qErr := tx.OrderItem.Query().
		Where(orderitem.OrderIDEQ(orderId)).
		All(ctx)
	if qErr != nil {
		s.log.Errorf("list order [%d] items for stock release failed: %v", orderId, qErr)
		err = qErr
		return nil, fmt.Errorf("list order items failed: %w", qErr)
	}
	if err != nil {
		s.log.Errorf("list order [%d] items for stock release failed: %v", orderId, err)
		_ = tx.Rollback()
		return nil, fmt.Errorf("list order items failed: %w", err)
	}

	for _, it := range items {
		skuId := it.SkuID
		if skuId == nil || *skuId == 0 {
			continue
		}
		qty := it.Quantity
		if qty == nil || *qty <= 0 {
			continue
		}
		// 释放库存（事务内）。
		if rErr := tx.Sku.UpdateOneID(*skuId).
			AddStockQty(*qty).
			Exec(ctx); rErr != nil {
			s.log.Errorf("release stock for sku [%d] during order [%d] expire failed: %v", *skuId, orderId, rErr)
			_ = tx.Rollback()
			return nil, fmt.Errorf("release stock for sku [%d] failed: %w", *skuId, rErr)
		}
	}

	// 状态推进 PENDING_PAYMENT→CANCELLED（事务内，batch Update + expected_status selector，
	// 与 orderRepo.Update 同模式）。
	// 目标状态 CANCELLED；expected_status 前置条件 = 当前须为 PENDING_PAYMENT。
	cancelledEnt := order.StatusOrderStatusCancelled
	pendingEnt := order.StatusOrderStatusPendingPayment
	uErr := tx.Order.Update().
		SetNillableStatus(&cancelledEnt).
		SetUpdatedAt(time.Now()).
		Where(
			order.IDEQ(orderId),
			order.StatusIn(pendingEnt),
		).
		Exec(ctx)
	if uErr != nil {
		err = uErr
		_ = tx.Rollback()
		// NotFound = 状态已被并发推进（expected_status 前置条件未满足），跳过。
		if ent.IsNotFound(uErr) {
			s.log.Infow(
				"msg", "expire order skipped: status changed concurrently",
				"order_id", orderId,
			)
			return &orderV1.ExpireOrderByTimeoutResponse{Expired: boolPtr(false)}, nil
		}
		return nil, fmt.Errorf("expire order [%d] failed: %w", orderId, uErr)
	}

	if cErr := tx.Commit(); cErr != nil {
		s.log.Errorf("commit expire tx failed for order [%d]: %v", orderId, cErr)
		return nil, fmt.Errorf("commit expire tx failed: %w", cErr)
	}
	err = nil

	s.log.Infow(
		"msg", "order expired due to timeout",
		"task_id", req.GetTaskId(),
		"trace_id", req.GetTraceId(),
		"order_id", orderId,
	)

	return &orderV1.ExpireOrderByTimeoutResponse{Expired: boolPtr(true)}, nil
}

// HandleOrderTimeout asynq 任务处理器（薄委托）。
// 由 server/asynq_server.go 注册，asynq worker 在 TTL 后触发。
// 所有业务逻辑内聚在 ExpireOrderByTimeout RPC 中。
func (s *OrderService) HandleOrderTimeout(taskType string, taskData *task.OrderTimeoutTaskData) error {
	s.log.Infow(
		"msg", "HandleOrderTimeout started",
		"task_type", taskType,
		"task_id", taskData.TaskId,
		"order_id", taskData.OrderId,
	)

	// 调用本服务 ExpireOrderByTimeout。
	// asynq 任务无 HTTP/gRPC 上下文，注入 SystemViewer context 以通过 ent privacy
	// 的 viewer 校验（SystemViewer.IsSystemContext() 放行，不做 tenant 过滤）。
	// 与 asynq_server.go 的 StartAllTask 使用同一 SystemViewer 注入模式。
	ctx := appViewer.NewSystemViewerContext(context.Background())
	_, err := s.ExpireOrderByTimeout(ctx, &orderV1.ExpireOrderByTimeoutRequest{
		OrderId: &taskData.OrderId,
		TaskId:  &taskData.TaskId,
		TraceId: &taskData.TraceId,
	})
	if err != nil {
		return fmt.Errorf("ExpireOrderByTimeout RPC failed for order [%d]: %w", taskData.OrderId, err)
	}
	return nil
}

func boolPtr(v bool) *bool { return &v }
