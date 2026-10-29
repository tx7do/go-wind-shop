package data

import (
	"go-wind-shop/pkg/logx"
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	entCrud "github.com/tx7do/go-crud/entgo"

	"github.com/tx7do/go-utils/copierutil"
	"github.com/tx7do/go-utils/mapper"
	"github.com/tx7do/go-utils/timeutil"

	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"
	"go-wind-shop/app/core/service/internal/data/ent/grouponactivity"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type GrouponActivityRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *logx.Helper

	mapper *mapper.CopierMapper[catalogV1.GrouponActivity, ent.GrouponActivity]

	statusConverter *mapper.EnumTypeConverter[catalogV1.GrouponActivity_Status, grouponactivity.Status]

	repository *entCrud.Repository[
		ent.GrouponActivityQuery, ent.GrouponActivitySelect,
		ent.GrouponActivityCreate, ent.GrouponActivityCreateBulk,
		ent.GrouponActivityUpdate, ent.GrouponActivityUpdateOne,
		ent.GrouponActivityDelete,
		predicate.GrouponActivity,
		catalogV1.GrouponActivity, ent.GrouponActivity,
	]
}

func NewGrouponActivityRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *GrouponActivityRepo {
	repo := &GrouponActivityRepo{
		log:             logx.New(ctx.NewLoggerHelper("groupon-activity/repo/core-service")),
		entClient:       entClient,
		mapper:          mapper.NewCopierMapper[catalogV1.GrouponActivity, ent.GrouponActivity](),
		statusConverter: mapper.NewEnumTypeConverter[catalogV1.GrouponActivity_Status, grouponactivity.Status](catalogV1.GrouponActivity_Status_name, catalogV1.GrouponActivity_Status_value),
	}

	repo.init()

	return repo
}

func (r *GrouponActivityRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.GrouponActivityQuery, ent.GrouponActivitySelect,
		ent.GrouponActivityCreate, ent.GrouponActivityCreateBulk,
		ent.GrouponActivityUpdate, ent.GrouponActivityUpdateOne,
		ent.GrouponActivityDelete,
		predicate.GrouponActivity,
		catalogV1.GrouponActivity, ent.GrouponActivity,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	r.mapper.AppendConverters(r.statusConverter.NewConverterPair())
}

func (r *GrouponActivityRepo) Count(ctx context.Context, req *paginationV1.PagingRequest) (int, error) {
	builder := r.entClient.Client().GrouponActivity.Query()

	whereSelectors, _, err := r.repository.BuildListSelectorWithPaging(builder, req)
	if len(whereSelectors) != 0 {
		builder.Modify(whereSelectors...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query groupon activity count failed: %s", err.Error())
		return 0, catalogV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *GrouponActivityRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListGrouponActivityResponse, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().GrouponActivity.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &catalogV1.ListGrouponActivityResponse{Total: 0, Items: nil}, nil
	}

	return &catalogV1.ListGrouponActivityResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *GrouponActivityRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().GrouponActivity.Query().
		Where(grouponactivity.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, catalogV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *GrouponActivityRepo) Get(ctx context.Context, req *catalogV1.GetGrouponActivityRequest) (*catalogV1.GrouponActivity, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().GrouponActivity.Query()

	var whereCond []func(s *sql.Selector)
	whereCond = append(whereCond, grouponactivity.IDEQ(req.GetId()))

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *GrouponActivityRepo) Create(ctx context.Context, req *catalogV1.CreateGrouponActivityRequest) error {
	if req == nil || req.Data == nil {
		return catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().GrouponActivity.Create().
		SetName(req.Data.GetName()).
		SetNillableSkuID(req.Data.SkuId).
		SetNillableGroupPrice(req.Data.GroupPrice).
		SetNillableGroupSize(req.Data.GroupSize).
		SetNillableGroupValidHours(req.Data.GroupValidHours).
		SetNillablePerUserLimit(req.Data.PerUserLimit).
		SetNillableValidFrom(timeutil.TimestamppbToTime(req.Data.ValidFrom)).
		SetNillableValidUntil(timeutil.TimestamppbToTime(req.Data.ValidUntil)).
		SetNillableCurrency(req.Data.Currency).
		SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
		SetNillableTenantID(req.Data.TenantId).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert groupon activity failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *GrouponActivityRepo) Update(ctx context.Context, req *catalogV1.UpdateGrouponActivityRequest) error {
	if req == nil || req.Data == nil {
		return catalogV1.ErrorBadRequest("invalid parameter")
	}
	if req.GetId() == 0 {
		return catalogV1.ErrorBadRequest("id is required")
	}

	builder := r.entClient.Client().GrouponActivity.Update()

	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *catalogV1.GrouponActivity) {
			builder.
				SetNillableName(req.Data.Name).
				SetNillableSkuID(req.Data.SkuId).
				SetNillableGroupPrice(req.Data.GroupPrice).
				SetNillableGroupSize(req.Data.GroupSize).
				SetNillableGroupValidHours(req.Data.GroupValidHours).
				SetNillablePerUserLimit(req.Data.PerUserLimit).
				SetNillableValidFrom(timeutil.TimestamppbToTime(req.Data.ValidFrom)).
				SetNillableValidUntil(timeutil.TimestamppbToTime(req.Data.ValidUntil)).
				SetNillableCurrency(req.Data.Currency).
				SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(grouponactivity.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *GrouponActivityRepo) Delete(ctx context.Context, req *catalogV1.DeleteGrouponActivityRequest) error {
	if req == nil {
		return catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().GrouponActivity.Delete()

	var err error
	if _, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(grouponactivity.FieldID, req.GetId()))
	}); err != nil {
		r.log.Errorf("delete groupon activity failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("delete data failed")
	}

	return nil
}
