package data

import (
	"go-wind-shop/pkg/logx"
	"context"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"github.com/tx7do/go-utils/trans"

	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	entCrud "github.com/tx7do/go-crud/entgo"

	"github.com/tx7do/go-utils/copierutil"
	"github.com/tx7do/go-utils/mapper"
	"github.com/tx7do/go-utils/timeutil"

	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"
	"go-wind-shop/app/core/service/internal/data/ent/seckillactivity"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type SeckillActivityRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *logx.Helper

	mapper *mapper.CopierMapper[catalogV1.SeckillActivity, ent.SeckillActivity]

	statusConverter *mapper.EnumTypeConverter[catalogV1.SeckillActivity_Status, seckillactivity.Status]

	repository *entCrud.Repository[
		ent.SeckillActivityQuery, ent.SeckillActivitySelect,
		ent.SeckillActivityCreate, ent.SeckillActivityCreateBulk,
		ent.SeckillActivityUpdate, ent.SeckillActivityUpdateOne,
		ent.SeckillActivityDelete,
		predicate.SeckillActivity,
		catalogV1.SeckillActivity, ent.SeckillActivity,
	]
}

func NewSeckillActivityRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *SeckillActivityRepo {
	repo := &SeckillActivityRepo{
		log:             logx.New(ctx.NewLoggerHelper("seckill-activity/repo/core-service")),
		entClient:       entClient,
		mapper:          mapper.NewCopierMapper[catalogV1.SeckillActivity, ent.SeckillActivity](),
		statusConverter: mapper.NewEnumTypeConverter[catalogV1.SeckillActivity_Status, seckillactivity.Status](catalogV1.SeckillActivity_Status_name, catalogV1.SeckillActivity_Status_value),
	}

	repo.init()

	return repo
}

func (r *SeckillActivityRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.SeckillActivityQuery, ent.SeckillActivitySelect,
		ent.SeckillActivityCreate, ent.SeckillActivityCreateBulk,
		ent.SeckillActivityUpdate, ent.SeckillActivityUpdateOne,
		ent.SeckillActivityDelete,
		predicate.SeckillActivity,
		catalogV1.SeckillActivity, ent.SeckillActivity,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	r.mapper.AppendConverters(r.statusConverter.NewConverterPair())
}

func (r *SeckillActivityRepo) Count(ctx context.Context, req *paginationV1.PagingRequest) (int, error) {
	builder := r.entClient.Client().SeckillActivity.Query()

	whereSelectors, _, err := r.repository.BuildListSelectorWithPaging(builder, req)
	if len(whereSelectors) != 0 {
		builder.Modify(whereSelectors...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query seckill activity count failed: %s", err.Error())
		return 0, catalogV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *SeckillActivityRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListSeckillActivityResponse, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().SeckillActivity.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &catalogV1.ListSeckillActivityResponse{Total: 0, Items: nil}, nil
	}

	return &catalogV1.ListSeckillActivityResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

// ListActive 列出当前生效中的秒杀（status=ACTIVE、处于有效窗口且未售罄）。
// 供 C 端商品页秒杀价角标/倒计时调用。租户隔离由 TenantPrivacy 按 viewer 注入。
func (r *SeckillActivityRepo) ListActive(ctx context.Context) (*catalogV1.ListSeckillActivityResponse, error) {
	now := time.Now()
	builder := r.entClient.Client().SeckillActivity.Query().
		Where(
			seckillactivity.StatusEQ(seckillactivity.StatusSeckillStatusActive),
			seckillactivity.ValidFromLTE(now),
			seckillactivity.ValidUntilGTE(now),
		)

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), &paginationV1.PagingRequest{
		NoPaging: trans.Ptr(true),
	})
	if err != nil {
		r.log.Errorf("query active seckills failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query failed")
	}

	resp := &catalogV1.ListSeckillActivityResponse{Items: []*catalogV1.SeckillActivity{}}
	if ret == nil {
		return resp, nil
	}
	// 未售罄过滤（sold_count < total_stock）
	for _, it := range ret.Items {
		if it.TotalStock != nil && it.SoldCount != nil && *it.SoldCount >= *it.TotalStock {
			continue
		}
		resp.Items = append(resp.Items, it)
	}
	resp.Total = uint64(len(resp.Items))

	return resp, nil
}

func (r *SeckillActivityRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().SeckillActivity.Query().
		Where(seckillactivity.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, catalogV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *SeckillActivityRepo) Get(ctx context.Context, req *catalogV1.GetSeckillActivityRequest) (*catalogV1.SeckillActivity, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().SeckillActivity.Query()

	var whereCond []func(s *sql.Selector)
	whereCond = append(whereCond, seckillactivity.IDEQ(req.GetId()))

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *SeckillActivityRepo) Create(ctx context.Context, req *catalogV1.CreateSeckillActivityRequest) error {
	if req == nil || req.Data == nil {
		return catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().SeckillActivity.Create().
		SetName(req.Data.GetName()).
		SetNillableSkuID(req.Data.SkuId).
		SetNillableSeckillPrice(req.Data.SeckillPrice).
		SetNillableTotalStock(req.Data.TotalStock).
		SetSoldCount(req.Data.GetSoldCount()).
		SetPerUserLimit(req.Data.GetPerUserLimit()).
		SetPriority(req.Data.GetPriority()).
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
		r.log.Errorf("insert seckill activity failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *SeckillActivityRepo) Update(ctx context.Context, req *catalogV1.UpdateSeckillActivityRequest) error {
	if req == nil || req.Data == nil {
		return catalogV1.ErrorBadRequest("invalid parameter")
	}
	if req.GetId() == 0 {
		return catalogV1.ErrorBadRequest("id is required")
	}

	// 如果不存在则创建
	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &catalogV1.CreateSeckillActivityRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	builder := r.entClient.Client().SeckillActivity.Update()

	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *catalogV1.SeckillActivity) {
			builder.
				SetNillableName(req.Data.Name).
				SetNillableSkuID(req.Data.SkuId).
				SetNillableSeckillPrice(req.Data.SeckillPrice).
				SetNillableTotalStock(req.Data.TotalStock).
				SetSoldCount(req.Data.GetSoldCount()).
				SetPerUserLimit(req.Data.GetPerUserLimit()).
				SetPriority(req.Data.GetPriority()).
				SetNillableValidFrom(timeutil.TimestamppbToTime(req.Data.ValidFrom)).
				SetNillableValidUntil(timeutil.TimestamppbToTime(req.Data.ValidUntil)).
				SetNillableCurrency(req.Data.Currency).
				SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(seckillactivity.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *SeckillActivityRepo) Delete(ctx context.Context, req *catalogV1.DeleteSeckillActivityRequest) error {
	if req == nil {
		return catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().SeckillActivity.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(seckillactivity.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete seckill activity failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("delete data failed")
	}

	return nil
}
