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

	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"
	"go-wind-shop/app/core/service/internal/data/ent/membertier"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type MemberTierRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *logx.Helper

	mapper *mapper.CopierMapper[catalogV1.MemberTier, ent.MemberTier]

	statusConverter *mapper.EnumTypeConverter[catalogV1.MemberTier_Status, membertier.Status]

	repository *entCrud.Repository[
		ent.MemberTierQuery, ent.MemberTierSelect,
		ent.MemberTierCreate, ent.MemberTierCreateBulk,
		ent.MemberTierUpdate, ent.MemberTierUpdateOne,
		ent.MemberTierDelete,
		predicate.MemberTier,
		catalogV1.MemberTier, ent.MemberTier,
	]
}

func NewMemberTierRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *MemberTierRepo {
	repo := &MemberTierRepo{
		log:             logx.New(ctx.NewLoggerHelper("member-tier/repo/core-service")),
		entClient:       entClient,
		mapper:          mapper.NewCopierMapper[catalogV1.MemberTier, ent.MemberTier](),
		statusConverter: mapper.NewEnumTypeConverter[catalogV1.MemberTier_Status, membertier.Status](catalogV1.MemberTier_Status_name, catalogV1.MemberTier_Status_value),
	}

	repo.init()

	return repo
}

func (r *MemberTierRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.MemberTierQuery, ent.MemberTierSelect,
		ent.MemberTierCreate, ent.MemberTierCreateBulk,
		ent.MemberTierUpdate, ent.MemberTierUpdateOne,
		ent.MemberTierDelete,
		predicate.MemberTier,
		catalogV1.MemberTier, ent.MemberTier,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	r.mapper.AppendConverters(r.statusConverter.NewConverterPair())
}

func (r *MemberTierRepo) Count(ctx context.Context, req *paginationV1.PagingRequest) (int, error) {
	builder := r.entClient.Client().MemberTier.Query()

	whereSelectors, _, err := r.repository.BuildListSelectorWithPaging(builder, req)
	if len(whereSelectors) != 0 {
		builder.Modify(whereSelectors...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query member tier count failed: %s", err.Error())
		return 0, catalogV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *MemberTierRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListMemberTierResponse, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().MemberTier.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &catalogV1.ListMemberTierResponse{Total: 0, Items: nil}, nil
	}

	return &catalogV1.ListMemberTierResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *MemberTierRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().MemberTier.Query().
		Where(membertier.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, catalogV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *MemberTierRepo) Get(ctx context.Context, req *catalogV1.GetMemberTierRequest) (*catalogV1.MemberTier, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().MemberTier.Query()

	var whereCond []func(s *sql.Selector)
	whereCond = append(whereCond, membertier.IDEQ(req.GetId()))

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *MemberTierRepo) Create(ctx context.Context, req *catalogV1.CreateMemberTierRequest) error {
	if req == nil || req.Data == nil {
		return catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().MemberTier.Create().
		SetName(req.Data.GetName()).
		SetLevel(req.Data.GetLevel()).
		SetMinSpend(req.Data.GetMinSpend()).
		SetPointsMultiplier(req.Data.GetPointsMultiplier()).
		SetNillableBadgeIcon(req.Data.BadgeIcon).
		SetCurrency(req.Data.GetCurrency()).
		SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
		SetNillableTenantID(req.Data.TenantId).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert member tier failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *MemberTierRepo) Update(ctx context.Context, req *catalogV1.UpdateMemberTierRequest) error {
	if req == nil || req.Data == nil {
		return catalogV1.ErrorBadRequest("invalid parameter")
	}
	if req.GetId() == 0 {
		return catalogV1.ErrorBadRequest("id is required")
	}

	builder := r.entClient.Client().MemberTier.Update()

	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *catalogV1.MemberTier) {
			builder.
				SetNillableName(req.Data.Name).
				SetLevel(req.Data.GetLevel()).
				SetMinSpend(req.Data.GetMinSpend()).
				SetPointsMultiplier(req.Data.GetPointsMultiplier()).
				SetNillableBadgeIcon(req.Data.BadgeIcon).
				SetCurrency(req.Data.GetCurrency()).
				SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(membertier.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *MemberTierRepo) Delete(ctx context.Context, req *catalogV1.DeleteMemberTierRequest) error {
	if req == nil {
		return catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().MemberTier.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(membertier.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete member tier failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("delete data failed")
	}

	return nil
}
