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
	"go-wind-shop/app/core/service/internal/data/ent/predicate"
	"go-wind-shop/app/core/service/internal/data/ent/shippingrate"

	shippingV1 "go-wind-shop/api/gen/go/shipping/service/v1"
)

type ShippingRateRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper         *mapper.CopierMapper[shippingV1.ShippingRate, ent.ShippingRate]
	statusConverter *mapper.EnumTypeConverter[shippingV1.ShippingRate_Status, shippingrate.Status]

	repository *entCrud.Repository[
		ent.ShippingRateQuery, ent.ShippingRateSelect,
		ent.ShippingRateCreate, ent.ShippingRateCreateBulk,
		ent.ShippingRateUpdate, ent.ShippingRateUpdateOne,
		ent.ShippingRateDelete,
		predicate.ShippingRate,
		shippingV1.ShippingRate, ent.ShippingRate,
	]
}

func NewShippingRateRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *ShippingRateRepo {
	repo := &ShippingRateRepo{
		log:             ctx.NewLoggerHelper("shipping-rate/repo/core-service"),
		entClient:       entClient,
		mapper:          mapper.NewCopierMapper[shippingV1.ShippingRate, ent.ShippingRate](),
		statusConverter: mapper.NewEnumTypeConverter[shippingV1.ShippingRate_Status, shippingrate.Status](shippingV1.ShippingRate_Status_name, shippingV1.ShippingRate_Status_value),
	}

	repo.init()

	return repo
}

func (r *ShippingRateRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.ShippingRateQuery, ent.ShippingRateSelect,
		ent.ShippingRateCreate, ent.ShippingRateCreateBulk,
		ent.ShippingRateUpdate, ent.ShippingRateUpdateOne,
		ent.ShippingRateDelete,
		predicate.ShippingRate,
		shippingV1.ShippingRate, ent.ShippingRate,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	r.mapper.AppendConverters(r.statusConverter.NewConverterPair())
}

func (r *ShippingRateRepo) Count(ctx context.Context, req *paginationV1.PagingRequest) (int, error) {
	builder := r.entClient.Client().ShippingRate.Query()

	whereSelectors, _, err := r.repository.BuildListSelectorWithPaging(builder, req)
	if len(whereSelectors) != 0 {
		builder.Modify(whereSelectors...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query shipping rate count failed: %s", err.Error())
		return 0, shippingV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *ShippingRateRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*shippingV1.ListShippingRateResponse, error) {
	if req == nil {
		return nil, shippingV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().ShippingRate.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &shippingV1.ListShippingRateResponse{Total: 0, Items: nil}, nil
	}

	return &shippingV1.ListShippingRateResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *ShippingRateRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().ShippingRate.Query().
		Where(shippingrate.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, shippingV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *ShippingRateRepo) Get(ctx context.Context, req *shippingV1.GetShippingRateRequest) (*shippingV1.ShippingRate, error) {
	if req == nil {
		return nil, shippingV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().ShippingRate.Query()

	var whereCond []func(s *sql.Selector)
	switch req.GetQueryBy().(type) {
	default:
	case *shippingV1.GetShippingRateRequest_Id:
		whereCond = append(whereCond, shippingrate.IDEQ(req.GetId()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *ShippingRateRepo) Create(ctx context.Context, req *shippingV1.CreateShippingRateRequest) error {
	if req == nil || req.Data == nil {
		return shippingV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().ShippingRate.Create().
		SetNillableRegion(req.Data.Region).
		SetNillableBaseFee(req.Data.BaseFee).
		SetNillablePerUnitFee(req.Data.PerUnitFee).
		SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
		SetNillableCurrency(req.Data.Currency).
		SetNillableTenantID(req.Data.TenantId).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert shipping rate failed: %s", err.Error())
		return shippingV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *ShippingRateRepo) Update(ctx context.Context, req *shippingV1.UpdateShippingRateRequest) error {
	if req == nil || req.Data == nil {
		return shippingV1.ErrorBadRequest("invalid parameter")
	}
	if req.GetId() == 0 {
		return shippingV1.ErrorBadRequest("id is required")
	}

	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &shippingV1.CreateShippingRateRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	builder := r.entClient.Client().Debug().ShippingRate.Update()

	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *shippingV1.ShippingRate) {
			builder.
				SetNillableRegion(req.Data.Region).
				SetNillableBaseFee(req.Data.BaseFee).
				SetNillablePerUnitFee(req.Data.PerUnitFee).
				SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
				SetNillableCurrency(req.Data.Currency).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(shippingrate.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *ShippingRateRepo) Delete(ctx context.Context, req *shippingV1.DeleteShippingRateRequest) error {
	if req == nil {
		return shippingV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Debug().ShippingRate.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(shippingrate.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete shipping rate failed: %s", err.Error())
		return shippingV1.ErrorInternalServerError("delete shipping rate failed")
	}

	return nil
}
