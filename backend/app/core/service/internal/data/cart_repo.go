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
	"go-wind-shop/app/core/service/internal/data/ent/cart"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"

	cartV1 "go-wind-shop/api/gen/go/cart/service/v1"
)

type CartRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[cartV1.Cart, ent.Cart]

	repository *entCrud.Repository[
		ent.CartQuery, ent.CartSelect,
		ent.CartCreate, ent.CartCreateBulk,
		ent.CartUpdate, ent.CartUpdateOne,
		ent.CartDelete,
		predicate.Cart,
		cartV1.Cart, ent.Cart,
	]
}

func NewCartRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *CartRepo {
	repo := &CartRepo{
		log:       ctx.NewLoggerHelper("cart/repo/core-service"),
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[cartV1.Cart, ent.Cart](),
	}

	repo.init()

	return repo
}

func (r *CartRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.CartQuery, ent.CartSelect,
		ent.CartCreate, ent.CartCreateBulk,
		ent.CartUpdate, ent.CartUpdateOne,
		ent.CartDelete,
		predicate.Cart,
		cartV1.Cart, ent.Cart,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
}

func (r *CartRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().Cart.Query()
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

func (r *CartRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*cartV1.ListCartResponse, error) {
	if req == nil {
		return nil, cartV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Cart.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &cartV1.ListCartResponse{Total: 0, Items: nil}, nil
	}

	return &cartV1.ListCartResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *CartRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().Cart.Query().
		Where(cart.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, cartV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *CartRepo) Get(ctx context.Context, req *cartV1.GetCartRequest) (*cartV1.Cart, error) {
	if req == nil {
		return nil, cartV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Cart.Query()

	var whereCond []func(s *sql.Selector)
	switch req.QueryBy.(type) {
	default:
	case *cartV1.GetCartRequest_Id:
		whereCond = append(whereCond, cart.IDEQ(req.GetId()))
	case *cartV1.GetCartRequest_UserId:
		whereCond = append(whereCond, cart.UserIDEQ(req.GetUserId()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *CartRepo) Create(ctx context.Context, req *cartV1.CreateCartRequest) error {
	if req == nil || req.Data == nil {
		return cartV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Cart.Create().
		SetNillableUserID(req.Data.UserId).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert cart failed: %s", err.Error())
		return cartV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *CartRepo) Update(ctx context.Context, req *cartV1.UpdateCartRequest) error {
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
			createReq := &cartV1.CreateCartRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	builder := r.entClient.Client().Debug().Cart.Update()
	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *cartV1.Cart) {
			builder.
				SetNillableUserID(req.Data.UserId).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(cart.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *CartRepo) Delete(ctx context.Context, req *cartV1.DeleteCartRequest) error {
	if req == nil {
		return cartV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Debug().Cart.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(cart.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete cart failed: %s", err.Error())
		return cartV1.ErrorInternalServerError("delete cart failed")
	}

	return nil
}
