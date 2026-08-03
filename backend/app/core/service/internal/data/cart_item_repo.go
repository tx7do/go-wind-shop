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
	"go-wind-shop/app/core/service/internal/data/ent/cartitem"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"

	cartV1 "go-wind-shop/api/gen/go/cart/service/v1"
)

type CartItemRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[cartV1.CartItem, ent.CartItem]

	repository *entCrud.Repository[
		ent.CartItemQuery, ent.CartItemSelect,
		ent.CartItemCreate, ent.CartItemCreateBulk,
		ent.CartItemUpdate, ent.CartItemUpdateOne,
		ent.CartItemDelete,
		predicate.CartItem,
		cartV1.CartItem, ent.CartItem,
	]
}

func NewCartItemRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *CartItemRepo {
	repo := &CartItemRepo{
		log:       ctx.NewLoggerHelper("cart-item/repo/core-service"),
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[cartV1.CartItem, ent.CartItem](),
	}

	repo.init()

	return repo
}

func (r *CartItemRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.CartItemQuery, ent.CartItemSelect,
		ent.CartItemCreate, ent.CartItemCreateBulk,
		ent.CartItemUpdate, ent.CartItemUpdateOne,
		ent.CartItemDelete,
		predicate.CartItem,
		cartV1.CartItem, ent.CartItem,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
}

func (r *CartItemRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().CartItem.Query()
	if len(whereCond) != 0 {
		builder.Modify(whereCond...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, cartV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *CartItemRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*cartV1.ListCartItemResponse, error) {
	if req == nil {
		return nil, cartV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().CartItem.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &cartV1.ListCartItemResponse{Total: 0, Items: nil}, nil
	}

	return &cartV1.ListCartItemResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *CartItemRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().CartItem.Query().
		Where(cartitem.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, cartV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *CartItemRepo) Get(ctx context.Context, req *cartV1.GetCartItemRequest) (*cartV1.CartItem, error) {
	if req == nil {
		return nil, cartV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().CartItem.Query()

	var whereCond []func(s *sql.Selector)
	switch req.QueryBy.(type) {
	default:
	case *cartV1.GetCartItemRequest_Id:
		whereCond = append(whereCond, cartitem.IDEQ(req.GetId()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *CartItemRepo) Create(ctx context.Context, req *cartV1.CreateCartItemRequest) error {
	if req == nil || req.Data == nil {
		return cartV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().CartItem.Create().
		SetNillableCartID(req.Data.CartId).
		SetNillableSkuID(req.Data.SkuId).
		SetNillableQuantity(req.Data.Quantity).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert cart item failed: %s", err.Error())
		return cartV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *CartItemRepo) Update(ctx context.Context, req *cartV1.UpdateCartItemRequest) error {
	if req == nil || req.Data == nil {
		return cartV1.ErrorBadRequest("invalid parameter")
	}

	// 如果不存在则创建
	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &cartV1.CreateCartItemRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	builder := r.entClient.Client().Debug().CartItem.Update()
	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *cartV1.CartItem) {
			builder.
				SetNillableCartID(req.Data.CartId).
				SetNillableSkuID(req.Data.SkuId).
				SetNillableQuantity(req.Data.Quantity).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(cartitem.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *CartItemRepo) Delete(ctx context.Context, req *cartV1.DeleteCartItemRequest) error {
	if req == nil {
		return cartV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Debug().CartItem.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(cartitem.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete cart item failed: %s", err.Error())
		return cartV1.ErrorInternalServerError("delete cart item failed")
	}

	return nil
}
