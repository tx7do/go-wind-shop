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
	"go-wind-shop/app/core/service/internal/data/ent/coupontemplate"
	"go-wind-shop/app/core/service/internal/data/ent/order"
	"go-wind-shop/app/core/service/internal/data/ent/orderitem"
	"go-wind-shop/app/core/service/internal/data/ent/paymenttransaction"
	"go-wind-shop/app/core/service/internal/data/ent/sku"
	"go-wind-shop/app/core/service/internal/data/ent/skuprice"
	"go-wind-shop/app/core/service/internal/data/ent/shippingrate"
	"go-wind-shop/app/core/service/internal/data/ent/taxrate"
	"go-wind-shop/app/core/service/internal/data/ent/usercoupon"

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

func (s *OrderService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*orderV1.CountOrderResponse, error) {
	count, err := s.orderRepo.Count(ctx, req)
	if err != nil {
		return nil, err
	}

	return &orderV1.CountOrderResponse{
		Count: uint64(count),
	}, nil
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
		// 校验状态转换合法性。expected_status 是调用方声明的当前状态，
		// target 是要求变更到的目标状态。仅允许下列转换，其余拒绝：
		//   PENDING_PAYMENT → PAID        （MarkPaid，支付 stub 调用）
		//   PENDING_PAYMENT → CANCELLED   （超时自动取消）
		//   PAID → FULFILLED               （发货）
		//   PAID → CLOSED                  （关闭）
		//   FULFILLED → CLOSED            （关闭）
		// 终态（CANCELLED/CLOSED）为吸收态，不可再翻转。
		target := req.Data.GetStatus()
		for _, from := range req.GetExpectedStatus() {
			if !isAllowedTransition(from, target) {
				return nil, orderV1.ErrorBadRequest("illegal order status transition: %v -> %v", from, target)
			}
		}
	}

	if err := s.orderRepo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

// isAllowedTransition 校验订单状态转换是否在允许列表内。
func isAllowedTransition(from, to orderV1.Order_Status) bool {
	allowed := map[orderV1.Order_Status]map[orderV1.Order_Status]bool{
		orderV1.Order_PENDING_PAYMENT: {
			orderV1.Order_PAID:      true,
			orderV1.Order_CANCELLED: true,
		},
		orderV1.Order_PAID: {
			orderV1.Order_FULFILLED: true,
			orderV1.Order_CLOSED:    true,
		},
		orderV1.Order_FULFILLED: {
			orderV1.Order_CLOSED: true,
		},
	}
	allowedTo, ok := allowed[from]
	if !ok {
		return false
	}
	return allowedTo[to]
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

	// ===== 优惠券核销（事务内）=====
	//
	// 至此 totalAmount 为折前总额（各 subtotal 之和）。若下单请求携带 coupon_id，
	// 在同一事务内 ForUpdate 锁定 user_coupon 与 coupon_template 行，校验：
	//   - 券归属（UserPrivacy 按 caller user_id 自动注入 WHERE，跨用户查不到→NotFound→回滚）
	//   - 券状态 UNUSED（手动前置条件，仿库存扣减模式）
	//   - 模板状态 ACTIVE、币种一致、有效窗口、未超 max_redemptions
	// 任一校验失败 → err 非空 → defer tx.Rollback() → 不下单不核销（fail-closed）。
	//
	// 核销成功后：
	//   - 模板 redeemed_count++（行锁内）
	//   - user_coupon 翻 USED + 写 redeemed_at/redeemed_order_id/applied_discount_amount
	//   - discount = computeDiscount(...)（与 Quote 共用，单一真相源）
	//
	// 无券时 discount=0。最终：
	//   - order.original_amount = 折前总额（审计）
	//   - order.discount_amount = discount（审计）
	//   - order.total_amount = max(0, 折前总额 - discount)（折后应付额，支付/退款取此值）
	originalAmount := totalAmount
	var discount int64 = 0
	couponId := req.GetCouponId()
	if couponId != 0 {
		discount, err = s.redeemCouponInTx(ctx, tx, couponId, orderId, originalAmount, req.Data.GetCurrency())
		if err != nil {
			return nil, err
		}
	}

	// 运费计算：按 (viewer tenant_id, shipping_region, currency) 查询运费模板。
	// shipping_region 由 BFF 从收货地址强制注入（core 不信任请求方传值）。
	// 运费 = base_fee + per_unit_fee * item_count。无规则时 shipping_fee = 0，不阻塞下单。
	shippingRegion := req.Data.GetShippingRegion()
	currency := req.Data.GetCurrency()
	var shippingFee int64 = 0
	var itemCount int64 = 0
	if shippingRegion != "" && tenantId != 0 {
		rateEnt, rateErr := tx.ShippingRate.Query().
			Where(
				shippingrate.TenantIDEQ(tenantId),
				shippingrate.RegionEQ(shippingRegion),
				shippingrate.CurrencyEQ(currency),
				shippingrate.StatusEQ(shippingrate.StatusShippingRateStatusActive),
			).
			Only(ctx)
		if rateErr == nil && rateEnt != nil {
			itemCount = int64(len(cartItems))
			baseFee := int64(0)
			perUnit := int64(0)
			if rateEnt.BaseFee != nil {
				baseFee = *rateEnt.BaseFee
			}
			if rateEnt.PerUnitFee != nil {
				perUnit = *rateEnt.PerUnitFee
			}
			shippingFee = baseFee + perUnit*itemCount
			if shippingFee < 0 {
				shippingFee = 0
			}
		}
	}

	// 税费计算：按 (viewer tenant_id, shipping_region, currency) 查询税率规则。
	// tax_amount = (originalAmount - discount) * tax_rate / 100（整数 floor）。
	// 无规则时 tax_amount = 0，不阻塞下单。
	var taxAmount int64 = 0
	if shippingRegion != "" && tenantId != 0 {
		taxEnt, taxErr := tx.TaxRate.Query().
			Where(
				taxrate.TenantIDEQ(tenantId),
				taxrate.RegionEQ(shippingRegion),
				taxrate.CurrencyEQ(currency),
				taxrate.StatusEQ(taxrate.StatusTaxRateStatusActive),
			).
			Only(ctx)
		if taxErr == nil && taxEnt != nil && taxEnt.TaxRate != nil {
			tr := int64(*taxEnt.TaxRate)
			if tr > 0 {
				taxable := originalAmount - discount
				if taxable > 0 {
					taxAmount = (taxable * tr) / 100
				}
			}
		}
	}

	// 折后应付额 = 折前总额 - 抵扣 + 运费 + 税费。
	finalTotal := originalAmount - discount + shippingFee + taxAmount
	if finalTotal < 0 {
		finalTotal = 0
	}

	// 回写订单金额（事务内）：折后应付额、折前总额、抵扣额、运费、税费、收货区域。
	if uErr := tx.Order.UpdateOneID(orderId).
		SetTotalAmount(finalTotal).
		SetOriginalAmount(originalAmount).
		SetDiscountAmount(discount).
		SetShippingFee(shippingFee).
		SetTaxAmount(taxAmount).
		SetShippingRegion(shippingRegion).
		Exec(ctx); uErr != nil {
		s.log.Errorf("update order amount failed: %v", uErr)
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

	// 孤儿支付清理：若该订单存在因 MarkPaid 竞态失败而遗留的 SUCCEEDED
	// 支付记录（支付"成功"但订单未推进到 PAID，现已被取消），在同一事务
	// 内将其翻为 FAILED，消除"已成功支付但订单已取消"的不一致。
	failedStatus := paymenttransaction.StatusPaymentTransactionStatusFailed
	succeededStatus := paymenttransaction.StatusPaymentTransactionStatusSucceeded
	if cErr := tx.PaymentTransaction.Update().
		SetStatus(failedStatus).
		SetUpdatedAt(time.Now()).
		Where(
			paymenttransaction.OrderIDEQ(orderId),
			paymenttransaction.StatusEQ(succeededStatus),
		).
		Exec(ctx); cErr != nil {
		err = cErr
		_ = tx.Rollback()
		s.log.Errorf("cleanup orphan succeeded payments for order [%d] failed: %v", orderId, cErr)
		return nil, fmt.Errorf("cleanup orphan payments for order [%d] failed: %w", orderId, cErr)
	}

	// 优惠券返还：若该订单核销过券，在同一事务内把关联 user_coupon 翻回 UNUSED、
	// 模板 redeemed_count--。与库存释放同事务同原子性，避免"券已用但订单已取消"
	// 的不一致。系统 viewer 上下文（本函数由 SystemViewer 调用）绕过 privacy，
	// 可跨用户反查 redeemed_order_id。
	s.restoreCouponInTx(ctx, tx, orderId)

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

// RestoreStockForRefund 退款成功后回补库存。
//
// 退款（PaymentRefundService.Update 翻 SUCCEEDED）应把下单时扣减的 SKU 库存加回，
// 否则退款后 stock_qty 永久丢失。本方法在单事务内遍历关联订单的 OrderItem，
// 按 quantity 调 AddStockQty 回补——与 ExpireOrderByTimeout 的释放模式一致，
// 但不改订单状态（退款不改变订单状态，这是现有设计；仅恢复库存）。
//
// 调用方：PaymentRefundService.Update（经 PaymentTransactionService 转发）。
// 入参 orderId 取自关联支付流水 payment_transaction.order_id。
//
// 注意：本方法专为退款回补设计，区别于超时取消（后者还推进 PENDING_PAYMENT→CANCELLED）。
func (s *OrderService) RestoreStockForRefund(ctx context.Context, orderId uint32) error {
	if orderId == 0 {
		return fmt.Errorf("invalid order id for stock restore")
	}

	tx, err := s.entClient.Client().Tx(ctx)
	if err != nil {
		s.log.Errorf("begin stock-restore tx failed for order [%d]: %v", orderId, err)
		return fmt.Errorf("begin stock-restore tx failed: %w", err)
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
		s.log.Errorf("list order [%d] items for stock restore failed: %v", orderId, qErr)
		err = qErr
		return fmt.Errorf("list order items failed: %w", qErr)
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
		// 回补库存（事务内）。
		if rErr := tx.Sku.UpdateOneID(*skuId).
			AddStockQty(*qty).
			Exec(ctx); rErr != nil {
			s.log.Errorf("restore stock for sku [%d] during refund of order [%d] failed: %v", *skuId, orderId, rErr)
			err = rErr
			return fmt.Errorf("restore stock for sku [%d] failed: %w", *skuId, rErr)
		}
	}

	// 优惠券返还：若该订单核销过券，在同一事务内把关联 user_coupon 翻回 UNUSED、
	// 模板 redeemed_count--。与库存回补同事务同原子性，避免"券已用但已退款"
	// 的不一致。本函数由 admin 平台 viewer（tid=0）调用，绕过 privacy，
	// 可跨用户反查 redeemed_order_id。
	s.restoreCouponInTx(ctx, tx, orderId)

	if cErr := tx.Commit(); cErr != nil {
		s.log.Errorf("commit stock-restore tx failed for order [%d]: %v", orderId, cErr)
		err = cErr
		return fmt.Errorf("commit stock-restore tx failed: %w", cErr)
	}
	err = nil

	s.log.Infow(
		"msg", "stock restored for refund",
		"order_id", orderId,
	)
	return nil
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

// redeemCouponInTx 在下单事务内核销优惠券。
//
// 该函数在 OrderService.Create 的事务内调用，与库存扣减同事务，保证"下单 + 核销"
// 原子性。任一步失败 → 返回 error → 调用方 defer tx.Rollback() → 不下单不核销（fail-closed）。
//
// 步骤：
//  1. ForUpdate 锁定 user_coupon 行。UserPrivacy 按 caller user_id 自动注入 WHERE，
//     跨用户 coupon_id 查不到 → NotFound → 返回 error → 整事务回滚。
//  2. 校验券状态 == UNUSED（手动前置条件，仿库存扣减模式）。
//  3. ForUpdate 锁定 coupon_template 行（TenantPrivacy 按 caller tenant_id 自动注入 WHERE）。
//  4. 校验模板状态 ACTIVE、币种一致、有效窗口、未超 max_redemptions。
//  5. discount = computeDiscount(originalAmount, discountParamsFromEntity(tmpl))。
//  6. 模板 redeemed_count++（行锁内）。
//  7. user_coupon 翻 USED + 写 redeemed_at/redeemed_order_id/applied_discount_amount。
//
// 返回 discount（≥0）。调用方据此算 finalTotal = max(0, originalAmount - discount)。
func (s *OrderService) redeemCouponInTx(
	ctx context.Context,
	tx *ent.Tx,
	couponId uint32,
	orderId uint32,
	originalAmount int64,
	currency string,
) (int64, error) {
	// 1. ForUpdate 锁定 user_coupon 行。
	//    UserPrivacy 按 caller user_id 自动注入 WHERE——跨用户查不到会返回 NotFound。
	uc, qErr := tx.UserCoupon.Query().
		Where(usercoupon.IDEQ(couponId)).
		ForUpdate().
		Only(ctx)
	if qErr != nil || uc == nil {
		s.log.Warnf("coupon [%d] not found or not owned by caller: %v", couponId, qErr)
		return 0, orderV1.ErrorBadRequest("coupon not found or not owned by caller")
	}

	// 2. 校验券状态 == UNUSED。
	if uc.Status == nil || *uc.Status != usercoupon.StatusUserCouponStatusUnused {
		s.log.Warnf("coupon [%d] not in UNUSED state", couponId)
		return 0, orderV1.ErrorBadRequest("coupon is not available")
	}

	// 3. ForUpdate 锁定 coupon_template 行。
	tmplId := uc.CouponTemplateID
	if tmplId == nil || *tmplId == 0 {
		return 0, orderV1.ErrorBadRequest("coupon has no associated template")
	}
	tmpl, tErr := tx.CouponTemplate.Query().
		Where(coupontemplate.IDEQ(*tmplId)).
		ForUpdate().
		Only(ctx)
	if tErr != nil || tmpl == nil {
		s.log.Warnf("coupon template [%d] not found: %v", *tmplId, tErr)
		return 0, orderV1.ErrorBadRequest("coupon template not found")
	}

	// 4. 校验模板。
	if tmpl.Status == nil || *tmpl.Status != coupontemplate.StatusCouponTemplateStatusActive {
		return 0, orderV1.ErrorBadRequest("coupon template is inactive")
	}
	// 币种一致：模板 currency 须与订单结算币一致。
	if tmpl.Currency == nil || *tmpl.Currency != currency {
		return 0, orderV1.ErrorBadRequest("coupon currency mismatch")
	}
	// 有效窗口。
	if !couponApplicableNowEntity(tmpl, time.Now()) {
		return 0, orderV1.ErrorBadRequest("coupon is not within its valid window")
	}
	// 限量校验。
	if tmpl.MaxRedemptions != nil && *tmpl.MaxRedemptions > 0 {
		if tmpl.RedeemedCount == nil || *tmpl.RedeemedCount >= *tmpl.MaxRedemptions {
			return 0, orderV1.ErrorBadRequest("coupon redemption limit reached")
		}
	}

	// per-user 限用校验：该用户该模板已 USED 的券数 >= max_redemptions_per_user 则拒。
	// 在 ForUpdate 锁定的模板行内读取 MaxRedemptionsPerUser（防 admin 并发改上限）。
	// Count 查询在事务内，UserPrivacy 按 caller user_id 自动注入 WHERE（双重保护）。
	// uc.UserID 取自 ForUpdate 锁定的券行（非请求参数，防伪造归属）。
	if tmpl.MaxRedemptionsPerUser != nil && *tmpl.MaxRedemptionsPerUser > 0 {
		if uc.UserID == nil {
			return 0, orderV1.ErrorBadRequest("coupon has no owner")
		}
		usedCount, cErr := tx.UserCoupon.Query().
			Where(
				usercoupon.And(
					usercoupon.UserIDEQ(*uc.UserID),
					usercoupon.CouponTemplateIDEQ(*tmplId),
					usercoupon.StatusEQ(usercoupon.StatusUserCouponStatusUsed),
				),
			).
			Count(ctx)
		if cErr != nil {
			s.log.Errorf("per-user usage count failed for user [%d] template [%d]: %v", *uc.UserID, *tmplId, cErr)
			return 0, orderV1.ErrorInternalServerError("per-user usage check failed")
		}
		if usedCount >= int(*tmpl.MaxRedemptionsPerUser) {
			s.log.Warnf("per-user usage limit reached for user [%d] template [%d]: %d >= %d", *uc.UserID, *tmplId, usedCount, *tmpl.MaxRedemptionsPerUser)
			return 0, orderV1.ErrorBadRequest("coupon per-user usage limit reached")
		}
	}

	// 5. 计算抵扣。
	params := discountParamsFromEntity(tmpl)
	discount, applicable := computeDiscount(originalAmount, params)
	if !applicable || discount <= 0 {
		return 0, orderV1.ErrorBadRequest("coupon produces no discount")
	}

	// 6. 模板 redeemed_count++（行锁内）。
	if uErr := tx.CouponTemplate.UpdateOneID(*tmplId).
		AddRedeemedCount(1).
		Exec(ctx); uErr != nil {
		s.log.Errorf("increment template [%d] redeemed_count failed: %v", *tmplId, uErr)
		return 0, orderV1.ErrorInternalServerError("coupon redemption failed")
	}

	// 7. user_coupon 翻 USED + 写审计字段。
	usedStatus := usercoupon.StatusUserCouponStatusUsed
	now := time.Now()
	if uErr := tx.UserCoupon.UpdateOneID(couponId).
		SetStatus(usedStatus).
		SetRedeemedAt(now).
		SetRedeemedOrderID(orderId).
		SetAppliedDiscountAmount(discount).
		Exec(ctx); uErr != nil {
		s.log.Errorf("mark coupon [%d] as USED failed: %v", couponId, uErr)
		return 0, orderV1.ErrorInternalServerError("coupon redemption failed")
	}

	s.log.Infof("coupon [%d] redeemed for order [%d], discount=%d", couponId, orderId, discount)
	return discount, nil
}

// restoreCouponInTx 在订单取消/退款事务内返还优惠券。
//
// 与库存回补同事务、同原子性。按 redeemed_order_id=orderId 反查关联的 user_coupon
// （SystemViewer/平台 viewer 绕过 UserPrivacy/TenantPrivacy，可跨用户/租户反查）。
// 若命中且 status==USED → 翻 UNUSED、清审计字段、模板 redeemed_count--。
// 无关联券或券非 USED 时静默跳过（幂等）。
//
// 调用点：
//   - ExpireOrderByTimeout（超时取消，SystemViewer 上下文）
//   - RestoreStockForRefund（退款回补，平台 admin viewer 上下文）
func (s *OrderService) restoreCouponInTx(
	ctx context.Context,
	tx *ent.Tx,
	orderId uint32,
) {
	// 反查关联券。返还钩子在 system/platform viewer 下运行，privacy 放行。
	ucs, qErr := tx.UserCoupon.Query().
		Where(usercoupon.RedeemedOrderIDEQ(orderId)).
		All(ctx)
	if qErr != nil {
		s.log.Errorf("query coupons for order [%d] restore failed: %v", orderId, qErr)
		return
	}
	for _, uc := range ucs {
		if uc.Status == nil || *uc.Status != usercoupon.StatusUserCouponStatusUsed {
			continue
		}
		// 翻 UNUSED + 清审计字段。
		unusedStatus := usercoupon.StatusUserCouponStatusUnused
		if uErr := tx.UserCoupon.UpdateOneID(uc.ID).
			SetStatus(unusedStatus).
			ClearRedeemedAt().
			ClearRedeemedOrderID().
			ClearAppliedDiscountAmount().
			Exec(ctx); uErr != nil {
			s.log.Errorf("restore coupon [%d] to UNUSED failed: %v", uc.ID, uErr)
			continue
		}
		// 模板 redeemed_count--。
		tmplId := uc.CouponTemplateID
		if tmplId != nil && *tmplId != 0 {
			if uErr := tx.CouponTemplate.UpdateOneID(*tmplId).
				AddRedeemedCount(-1).
				Exec(ctx); uErr != nil {
				s.log.Errorf("decrement template [%d] redeemed_count during restore failed: %v", *tmplId, uErr)
			}
		}
		s.log.Infof("coupon [%d] restored to UNUSED for order [%d]", uc.ID, orderId)
	}
}

