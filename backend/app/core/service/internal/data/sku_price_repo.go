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
	"go-wind-shop/app/core/service/internal/data/ent/skuprice"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type SkuPriceRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[catalogV1.SkuPrice, ent.SkuPrice]

	repository *entCrud.Repository[
		ent.SkuPriceQuery, ent.SkuPriceSelect,
		ent.SkuPriceCreate, ent.SkuPriceCreateBulk,
		ent.SkuPriceUpdate, ent.SkuPriceUpdateOne,
		ent.SkuPriceDelete,
		predicate.SkuPrice,
		catalogV1.SkuPrice, ent.SkuPrice,
	]
}

func NewSkuPriceRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *SkuPriceRepo {
	repo := &SkuPriceRepo{
		log:       ctx.NewLoggerHelper("sku-price/repo/core-service"),
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[catalogV1.SkuPrice, ent.SkuPrice](),
	}

	repo.init()

	return repo
}

func (r *SkuPriceRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.SkuPriceQuery, ent.SkuPriceSelect,
		ent.SkuPriceCreate, ent.SkuPriceCreateBulk,
		ent.SkuPriceUpdate, ent.SkuPriceUpdateOne,
		ent.SkuPriceDelete,
		predicate.SkuPrice,
		catalogV1.SkuPrice, ent.SkuPrice,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
}

func (r *SkuPriceRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().SkuPrice.Query()
	if len(whereCond) != 0 {
		builder.Modify(whereCond...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query count failed: %s", err.Error())
		return 0, catalogV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *SkuPriceRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListSkuPriceResponse, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().SkuPrice.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &catalogV1.ListSkuPriceResponse{Total: 0, Items: nil}, nil
	}

	return &catalogV1.ListSkuPriceResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *SkuPriceRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().SkuPrice.Query().
		Where(skuprice.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, catalogV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *SkuPriceRepo) Get(ctx context.Context, req *catalogV1.GetSkuPriceRequest) (*catalogV1.SkuPrice, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().SkuPrice.Query()

	var whereCond []func(s *sql.Selector)
	switch req.QueryBy.(type) {
	default:
	case *catalogV1.GetSkuPriceRequest_Id:
		whereCond = append(whereCond, skuprice.IDEQ(req.GetId()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *SkuPriceRepo) Create(ctx context.Context, req *catalogV1.CreateSkuPriceRequest) error {
	if req == nil || req.Data == nil {
		return catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().SkuPrice.Create().
		SetNillableSkuID(req.Data.SkuId).
		SetNillableCurrency(req.Data.Currency).
		SetNillableAmount(req.Data.Amount).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert sku price failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *SkuPriceRepo) Update(ctx context.Context, req *catalogV1.UpdateSkuPriceRequest) error {
	if req == nil || req.Data == nil {
		return catalogV1.ErrorBadRequest("invalid parameter")
	}

	// 如果不存在则创建
	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &catalogV1.CreateSkuPriceRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	builder := r.entClient.Client().Debug().SkuPrice.Update()
	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *catalogV1.SkuPrice) {
			builder.
				SetNillableSkuID(req.Data.SkuId).
				SetNillableCurrency(req.Data.Currency).
				SetNillableAmount(req.Data.Amount).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(skuprice.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *SkuPriceRepo) Delete(ctx context.Context, req *catalogV1.DeleteSkuPriceRequest) error {
	if req == nil {
		return catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Debug().SkuPrice.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(skuprice.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete sku price failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("delete sku price failed")
	}

	return nil
}
