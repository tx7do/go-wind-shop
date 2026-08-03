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
	"go-wind-shop/app/core/service/internal/data/ent/skuattributecombination"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type SkuAttributeCombinationRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[catalogV1.SkuAttributeCombination, ent.SkuAttributeCombination]

	repository *entCrud.Repository[
		ent.SkuAttributeCombinationQuery, ent.SkuAttributeCombinationSelect,
		ent.SkuAttributeCombinationCreate, ent.SkuAttributeCombinationCreateBulk,
		ent.SkuAttributeCombinationUpdate, ent.SkuAttributeCombinationUpdateOne,
		ent.SkuAttributeCombinationDelete,
		predicate.SkuAttributeCombination,
		catalogV1.SkuAttributeCombination, ent.SkuAttributeCombination,
	]
}

func NewSkuAttributeCombinationRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *SkuAttributeCombinationRepo {
	repo := &SkuAttributeCombinationRepo{
		log:       ctx.NewLoggerHelper("sku-attribute-combination/repo/core-service"),
		entClient: entClient,
		mapper:    mapper.NewCopierMapper[catalogV1.SkuAttributeCombination, ent.SkuAttributeCombination](),
	}

	repo.init()

	return repo
}

func (r *SkuAttributeCombinationRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.SkuAttributeCombinationQuery, ent.SkuAttributeCombinationSelect,
		ent.SkuAttributeCombinationCreate, ent.SkuAttributeCombinationCreateBulk,
		ent.SkuAttributeCombinationUpdate, ent.SkuAttributeCombinationUpdateOne,
		ent.SkuAttributeCombinationDelete,
		predicate.SkuAttributeCombination,
		catalogV1.SkuAttributeCombination, ent.SkuAttributeCombination,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())
}

func (r *SkuAttributeCombinationRepo) Count(ctx context.Context, whereCond []func(s *sql.Selector)) (int, error) {
	builder := r.entClient.Client().SkuAttributeCombination.Query()
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

func (r *SkuAttributeCombinationRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListSkuAttributeCombinationResponse, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().SkuAttributeCombination.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &catalogV1.ListSkuAttributeCombinationResponse{Total: 0, Items: nil}, nil
	}

	return &catalogV1.ListSkuAttributeCombinationResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *SkuAttributeCombinationRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().SkuAttributeCombination.Query().
		Where(skuattributecombination.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, catalogV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *SkuAttributeCombinationRepo) Get(ctx context.Context, req *catalogV1.GetSkuAttributeCombinationRequest) (*catalogV1.SkuAttributeCombination, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().SkuAttributeCombination.Query()

	var whereCond []func(s *sql.Selector)
	switch req.QueryBy.(type) {
	default:
	case *catalogV1.GetSkuAttributeCombinationRequest_Id:
		whereCond = append(whereCond, skuattributecombination.IDEQ(req.GetId()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *SkuAttributeCombinationRepo) Create(ctx context.Context, req *catalogV1.CreateSkuAttributeCombinationRequest) error {
	if req == nil || req.Data == nil {
		return catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().SkuAttributeCombination.Create().
		SetNillableSkuID(req.Data.SkuId).
		SetNillableAttributeID(req.Data.AttributeId).
		SetNillableAttributeValueID(req.Data.AttributeValueId).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert sku attribute combination failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *SkuAttributeCombinationRepo) Update(ctx context.Context, req *catalogV1.UpdateSkuAttributeCombinationRequest) error {
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
			createReq := &catalogV1.CreateSkuAttributeCombinationRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	builder := r.entClient.Client().Debug().SkuAttributeCombination.Update()
	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *catalogV1.SkuAttributeCombination) {
			builder.
				SetNillableSkuID(req.Data.SkuId).
				SetNillableAttributeID(req.Data.AttributeId).
				SetNillableAttributeValueID(req.Data.AttributeValueId).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(skuattributecombination.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *SkuAttributeCombinationRepo) Delete(ctx context.Context, req *catalogV1.DeleteSkuAttributeCombinationRequest) error {
	if req == nil {
		return catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Debug().SkuAttributeCombination.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(skuattributecombination.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete sku attribute combination failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("delete sku attribute combination failed")
	}

	return nil
}
