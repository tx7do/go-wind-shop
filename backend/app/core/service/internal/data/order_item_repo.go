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
	"go-wind-shop/app/core/service/internal/data/ent/orderitem"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"

	orderV1 "go-wind-shop/api/gen/go/order/service/v1"
)

type OrderItemRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[orderV1.OrderItem, ent.OrderItem]

	repository *entCrud.Repository[
		ent.OrderItemQuery, ent.OrderItemSelect,
		ent.OrderItemCreate, ent.OrderItemCreateBulk,
		ent.OrderItemUpdate, ent.OrderItemUpdateOne,
		ent.OrderItemDelete,
		predicate.OrderItem,
		orderV1.OrderItem, ent.OrderItem,
	]
}

func NewOrderItemRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *OrderItemRepo {
	repo := &OrderItemRepo{
		log:       ctx.NewLoggerHelper("order-item/repo/core-service"),
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[orderV1.OrderItem, ent.OrderItem](),
	}

	repo.init()

	return repo
}

func (r *OrderItemRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.OrderItemQuery, ent.OrderItemSelect,
		ent.OrderItemCreate, ent.OrderItemCreateBulk,
		ent.OrderItemUpdate, ent.OrderItemUpdateOne,
		ent.OrderItemDelete,
		predicate.OrderItem,
		orderV1.OrderItem, ent.OrderItem,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
}

func (r *OrderItemRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().OrderItem.Query()
	if len(whereCond) != 0 {
		builder.Modify(whereCond...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, orderV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *OrderItemRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*orderV1.ListOrderItemResponse, error) {
	if req == nil {
		return nil, orderV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().OrderItem.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &orderV1.ListOrderItemResponse{Total: 0, Items: nil}, nil
	}

	return &orderV1.ListOrderItemResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *OrderItemRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().OrderItem.Query().
		Where(orderitem.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, orderV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

// ListByOrderId 列出指定订单下的所有订单项（用于超时过期时的库存释放）。
// 直接走 ent 客户端（非分页），返回 ent 实体切片。
func (r *OrderItemRepo) ListByOrderId(ctx context.Context, orderId uint32) ([]*ent.OrderItem, error) {
	items, err := r.entClient.Client().OrderItem.Query().
		Where(orderitem.OrderIDEQ(orderId)).
		All(ctx)
	if err != nil {
		r.log.Errorf("list order items by order [%d] failed: %s", orderId, err.Error())
		return nil, orderV1.ErrorInternalServerError("list order items failed")
	}
	return items, nil
}

func (r *OrderItemRepo) Get(ctx context.Context, req *orderV1.GetOrderItemRequest) (*orderV1.OrderItem, error) {
	if req == nil {
		return nil, orderV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().OrderItem.Query()

	var whereCond []func(s *sql.Selector)
	switch req.QueryBy.(type) {
	default:
	case *orderV1.GetOrderItemRequest_Id:
		whereCond = append(whereCond, orderitem.IDEQ(req.GetId()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *OrderItemRepo) Create(ctx context.Context, req *orderV1.CreateOrderItemRequest) error {
	if req == nil || req.Data == nil {
		return orderV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().OrderItem.Create().
		SetNillableOrderID(req.Data.OrderId).
		SetNillableSkuID(req.Data.SkuId).
		SetNillableSkuSnapshot(req.Data.SkuSnapshot).
		SetNillableQuantity(req.Data.Quantity).
		SetNillableUnitPrice(req.Data.UnitPrice).
		SetNillableSubtotal(req.Data.Subtotal).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert order item failed: %s", err.Error())
		return orderV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *OrderItemRepo) Update(ctx context.Context, req *orderV1.UpdateOrderItemRequest) error {
	if req == nil || req.Data == nil {
		return orderV1.ErrorBadRequest("invalid parameter")
	}

	// 如果不存在则创建
	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &orderV1.CreateOrderItemRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	builder := r.entClient.Client().Debug().OrderItem.Update()
	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *orderV1.OrderItem) {
			builder.
				SetNillableOrderID(req.Data.OrderId).
				SetNillableSkuID(req.Data.SkuId).
				SetNillableSkuSnapshot(req.Data.SkuSnapshot).
				SetNillableQuantity(req.Data.Quantity).
				SetNillableUnitPrice(req.Data.UnitPrice).
				SetNillableSubtotal(req.Data.Subtotal).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(orderitem.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *OrderItemRepo) Delete(ctx context.Context, req *orderV1.DeleteOrderItemRequest) error {
	if req == nil {
		return orderV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Debug().OrderItem.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(orderitem.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete order item failed: %s", err.Error())
		return orderV1.ErrorInternalServerError("delete order item failed")
	}

	return nil
}
