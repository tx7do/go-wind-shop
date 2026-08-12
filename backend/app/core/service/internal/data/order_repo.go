package data

import (
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	entCrud "github.com/tx7do/go-crud/entgo"

	"github.com/tx7do/go-utils/copierutil"
	"github.com/tx7do/go-utils/mapper"

	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/order"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"

	orderV1 "go-wind-shop/api/gen/go/order/service/v1"
)

type OrderRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[orderV1.Order, ent.Order]

	statusConverter *mapper.EnumTypeConverter[orderV1.Order_Status, order.Status]

	repository *entCrud.Repository[
		ent.OrderQuery, ent.OrderSelect,
		ent.OrderCreate, ent.OrderCreateBulk,
		ent.OrderUpdate, ent.OrderUpdateOne,
		ent.OrderDelete,
		predicate.Order,
		orderV1.Order, ent.Order,
	]
}

func NewOrderRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *OrderRepo {
	repo := &OrderRepo{
		log:             ctx.NewLoggerHelper("order/repo/core-service"),
		entClient:       entClient,
		mapper:          mapper.NewCopierMapper[orderV1.Order, ent.Order](),
		statusConverter: mapper.NewEnumTypeConverter[orderV1.Order_Status, order.Status](orderV1.Order_Status_name, orderV1.Order_Status_value),
	}

	repo.init()

	return repo
}

func (r *OrderRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.OrderQuery, ent.OrderSelect,
		ent.OrderCreate, ent.OrderCreateBulk,
		ent.OrderUpdate, ent.OrderUpdateOne,
		ent.OrderDelete,
		predicate.Order,
		orderV1.Order, ent.Order,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	r.mapper.AppendConverters(r.statusConverter.NewConverterPair())
}

func (r *OrderRepo) Count(ctx context.Context, req *paginationV1.PagingRequest) (int, error) {
	builder := r.entClient.Client().Order.Query()

	whereSelectors, _, err := r.repository.BuildListSelectorWithPaging(builder, req)
	if len(whereSelectors) != 0 {
		builder.Modify(whereSelectors...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query order count failed: %s", err.Error())
		return 0, orderV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *OrderRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*orderV1.ListOrderResponse, error) {
	if req == nil {
		return nil, orderV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Order.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &orderV1.ListOrderResponse{Total: 0, Items: nil}, nil
	}

	return &orderV1.ListOrderResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *OrderRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().Order.Query().
		Where(order.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, orderV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *OrderRepo) Get(ctx context.Context, req *orderV1.GetOrderRequest) (*orderV1.Order, error) {
	if req == nil {
		return nil, orderV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Order.Query()

	var whereCond []func(s *sql.Selector)
	switch req.GetQueryBy().(type) {
	default:
	case *orderV1.GetOrderRequest_Id:
		whereCond = append(whereCond, order.IDEQ(req.GetId()))

	case *orderV1.GetOrderRequest_IdempotencyKey:
		// 幂等键在租户内唯一，必须配合 tenant_id 过滤，否则跨租户可能命中他人订单。
		tenantId := req.GetTenantId()
		whereCond = append(whereCond, func(s *sql.Selector) {
			s.Where(
				sql.And(
					sql.EQ(order.FieldIdempotencyKey, req.GetIdempotencyKey()),
					sql.EQ(order.FieldTenantID, tenantId),
				),
			)
		})
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *OrderRepo) Create(ctx context.Context, req *orderV1.CreateOrderRequest) error {
	if req == nil || req.Data == nil {
		return orderV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Order.Create().
		SetNillableUserID(req.Data.UserId).
		SetNillableTotalAmount(req.Data.TotalAmount).
		SetNillableOriginalAmount(req.Data.OriginalAmount).
		SetNillableDiscountAmount(req.Data.DiscountAmount).
		SetNillableCurrency(req.Data.Currency).
		SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
		SetNillableBusinessRefID(req.Data.BusinessRefId).
		SetNillableIdempotencyKey(req.Data.IdempotencyKey).
		SetNillableRecipientName(req.Data.RecipientName).
		SetNillableRecipientPhone(req.Data.RecipientPhone).
		SetNillableShippingAddress(req.Data.ShippingAddress).
		SetNillableTenantID(req.Data.TenantId).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert order failed: %s", err.Error())
		return orderV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *OrderRepo) Update(ctx context.Context, req *orderV1.UpdateOrderRequest) error {
	if req == nil || req.Data == nil {
		return orderV1.ErrorBadRequest("invalid parameter")
	}
	if req.GetId() == 0 {
		return orderV1.ErrorBadRequest("id is required")
	}

	// 如果不存在则创建
	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &orderV1.CreateOrderRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	builder := r.entClient.Client().Debug().Order.Update()

	// 状态前置条件：expected_status 非空时，把条件注入 SQL selector，
	// 仅当订单当前 status ∈ expected_status 时 UPDATE 才命中（affected_rows=0 表示状态已被并发推进）。
	expectedStatusVals := make([]any, 0, len(req.GetExpectedStatus()))
	for _, st := range req.GetExpectedStatus() {
		st := st // 取地址前避免 range 变量复用
		if v := r.statusConverter.ToEntity(&st); v != nil {
			expectedStatusVals = append(expectedStatusVals, *v)
		}
	}

	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *orderV1.Order) {
			builder.
				SetNillableUserID(req.Data.UserId).
				SetNillableTotalAmount(req.Data.TotalAmount).
				SetNillableOriginalAmount(req.Data.OriginalAmount).
				SetNillableDiscountAmount(req.Data.DiscountAmount).
				SetNillableCurrency(req.Data.Currency).
				SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
				SetNillableBusinessRefID(req.Data.BusinessRefId).
				SetNillableIdempotencyKey(req.Data.IdempotencyKey).
				SetNillableRecipientName(req.Data.RecipientName).
				SetNillableRecipientPhone(req.Data.RecipientPhone).
				SetNillableShippingAddress(req.Data.ShippingAddress).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(order.FieldID, req.GetId()))
			if len(expectedStatusVals) > 0 {
				s.Where(sql.In(order.FieldStatus, expectedStatusVals...))
			}
		},
	)
	if err != nil {
		// UpdateX 在 WHERE 条件不匹配（状态已被并发推进）时返回 NotFound。
		// 此时若传入了 expected_status，说明是状态前置条件失败，返回 Conflict 供调用方识别。
		if ent.IsNotFound(err) && len(expectedStatusVals) > 0 {
			r.log.Infof("update order [%d] precondition failed: status not in expected set", req.GetId())
			return orderV1.ErrorConflict("order status precondition failed")
		}
		return err
	}

	return nil
}

func (r *OrderRepo) Delete(ctx context.Context, req *orderV1.DeleteOrderRequest) error {
	if req == nil {
		return orderV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Debug().Order.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(order.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete order failed: %s", err.Error())
		return orderV1.ErrorInternalServerError("delete order failed")
	}

	return nil
}
