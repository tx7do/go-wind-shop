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
	"go-wind-shop/app/core/service/internal/data/ent/taxrate"

	taxV1 "go-wind-shop/api/gen/go/tax/service/v1"
)

type TaxRateRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper         *mapper.CopierMapper[taxV1.TaxRate, ent.TaxRate]
	statusConverter *mapper.EnumTypeConverter[taxV1.TaxRate_Status, taxrate.Status]

	repository *entCrud.Repository[
		ent.TaxRateQuery, ent.TaxRateSelect,
		ent.TaxRateCreate, ent.TaxRateCreateBulk,
		ent.TaxRateUpdate, ent.TaxRateUpdateOne,
		ent.TaxRateDelete,
		predicate.TaxRate,
		taxV1.TaxRate, ent.TaxRate,
	]
}

func NewTaxRateRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *TaxRateRepo {
	repo := &TaxRateRepo{
		log:             ctx.NewLoggerHelper("tax-rate/repo/core-service"),
		entClient:       entClient,
		mapper:          mapper.NewCopierMapper[taxV1.TaxRate, ent.TaxRate](),
		statusConverter: mapper.NewEnumTypeConverter[taxV1.TaxRate_Status, taxrate.Status](taxV1.TaxRate_Status_name, taxV1.TaxRate_Status_value),
	}

	repo.init()

	return repo
}

func (r *TaxRateRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.TaxRateQuery, ent.TaxRateSelect,
		ent.TaxRateCreate, ent.TaxRateCreateBulk,
		ent.TaxRateUpdate, ent.TaxRateUpdateOne,
		ent.TaxRateDelete,
		predicate.TaxRate,
		taxV1.TaxRate, ent.TaxRate,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	r.mapper.AppendConverters(r.statusConverter.NewConverterPair())
}

func (r *TaxRateRepo) Count(ctx context.Context, req *paginationV1.PagingRequest) (int, error) {
	builder := r.entClient.Client().TaxRate.Query()

	whereSelectors, _, err := r.repository.BuildListSelectorWithPaging(builder, req)
	if len(whereSelectors) != 0 {
		builder.Modify(whereSelectors...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query tax rate count failed: %s", err.Error())
		return 0, taxV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *TaxRateRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*taxV1.ListTaxRateResponse, error) {
	if req == nil {
		return nil, taxV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().TaxRate.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &taxV1.ListTaxRateResponse{Total: 0, Items: nil}, nil
	}

	return &taxV1.ListTaxRateResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *TaxRateRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().TaxRate.Query().
		Where(taxrate.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, taxV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *TaxRateRepo) Get(ctx context.Context, req *taxV1.GetTaxRateRequest) (*taxV1.TaxRate, error) {
	if req == nil {
		return nil, taxV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().TaxRate.Query()

	var whereCond []func(s *sql.Selector)
	switch req.GetQueryBy().(type) {
	default:
	case *taxV1.GetTaxRateRequest_Id:
		whereCond = append(whereCond, taxrate.IDEQ(req.GetId()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *TaxRateRepo) Create(ctx context.Context, req *taxV1.CreateTaxRateRequest) error {
	if req == nil || req.Data == nil {
		return taxV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().TaxRate.Create().
		SetNillableRegion(req.Data.Region).
		SetNillableTaxRate(req.Data.TaxRate).
		SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
		SetNillableCurrency(req.Data.Currency).
		SetNillableTenantID(req.Data.TenantId).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert tax rate failed: %s", err.Error())
		return taxV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *TaxRateRepo) Update(ctx context.Context, req *taxV1.UpdateTaxRateRequest) error {
	if req == nil || req.Data == nil {
		return taxV1.ErrorBadRequest("invalid parameter")
	}
	if req.GetId() == 0 {
		return taxV1.ErrorBadRequest("id is required")
	}

	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &taxV1.CreateTaxRateRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	builder := r.entClient.Client().Debug().TaxRate.Update()

	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *taxV1.TaxRate) {
			builder.
				SetNillableRegion(req.Data.Region).
				SetNillableTaxRate(req.Data.TaxRate).
				SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
				SetNillableCurrency(req.Data.Currency).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(taxrate.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *TaxRateRepo) Delete(ctx context.Context, req *taxV1.DeleteTaxRateRequest) error {
	if req == nil {
		return taxV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Debug().TaxRate.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(taxrate.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete tax rate failed: %s", err.Error())
		return taxV1.ErrorInternalServerError("delete tax rate failed")
	}

	return nil
}
