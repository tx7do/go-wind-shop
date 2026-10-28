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
	"go-wind-shop/app/core/service/internal/data/ent/promotion"

	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

type PromotionRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *logx.Helper

	mapper *mapper.CopierMapper[catalogV1.Promotion, ent.Promotion]

	promoTypeConverter *mapper.EnumTypeConverter[catalogV1.Promotion_PromoType, promotion.PromoType]
	statusConverter    *mapper.EnumTypeConverter[catalogV1.Promotion_Status, promotion.Status]

	repository *entCrud.Repository[
		ent.PromotionQuery, ent.PromotionSelect,
		ent.PromotionCreate, ent.PromotionCreateBulk,
		ent.PromotionUpdate, ent.PromotionUpdateOne,
		ent.PromotionDelete,
		predicate.Promotion,
		catalogV1.Promotion, ent.Promotion,
	]
}

func NewPromotionRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *PromotionRepo {
	repo := &PromotionRepo{
		log:                logx.New(ctx.NewLoggerHelper("promotion/repo/core-service")),
		entClient:          entClient,
		mapper:             mapper.NewCopierMapper[catalogV1.Promotion, ent.Promotion](),
		promoTypeConverter: mapper.NewEnumTypeConverter[catalogV1.Promotion_PromoType, promotion.PromoType](catalogV1.Promotion_PromoType_name, catalogV1.Promotion_PromoType_value),
		statusConverter:    mapper.NewEnumTypeConverter[catalogV1.Promotion_Status, promotion.Status](catalogV1.Promotion_Status_name, catalogV1.Promotion_Status_value),
	}

	repo.init()

	return repo
}

func (r *PromotionRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.PromotionQuery, ent.PromotionSelect,
		ent.PromotionCreate, ent.PromotionCreateBulk,
		ent.PromotionUpdate, ent.PromotionUpdateOne,
		ent.PromotionDelete,
		predicate.Promotion,
		catalogV1.Promotion, ent.Promotion,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	r.mapper.AppendConverters(r.promoTypeConverter.NewConverterPair())
	r.mapper.AppendConverters(r.statusConverter.NewConverterPair())
}

func (r *PromotionRepo) Count(ctx context.Context, req *paginationV1.PagingRequest) (int, error) {
	builder := r.entClient.Client().Promotion.Query()

	whereSelectors, _, err := r.repository.BuildListSelectorWithPaging(builder, req)
	if len(whereSelectors) != 0 {
		builder.Modify(whereSelectors...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query promotion count failed: %s", err.Error())
		return 0, catalogV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *PromotionRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListPromotionResponse, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Promotion.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &catalogV1.ListPromotionResponse{Total: 0, Items: nil}, nil
	}

	return &catalogV1.ListPromotionResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

// ListActive 列出当前生效中的活动（status=ACTIVE 且处于有效窗口）。
// 供 C端商品页角标/活动列表调用。租户隔离由 TenantPrivacy 按 viewer 自动注入。
func (r *PromotionRepo) ListActive(ctx context.Context) (*catalogV1.ListPromotionResponse, error) {
	now := time.Now()
	builder := r.entClient.Client().Promotion.Query().
		Where(
			promotion.StatusEQ(promotion.StatusPromotionStatusActive),
			promotion.ValidFromLTE(now),
			promotion.ValidUntilGTE(now),
		)

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), &paginationV1.PagingRequest{
		NoPaging: trans.Ptr(true),
	})
	if err != nil {
		r.log.Errorf("query active promotions failed: %s", err.Error())
		return nil, catalogV1.ErrorInternalServerError("query failed")
	}
	if ret == nil {
		return &catalogV1.ListPromotionResponse{Total: 0, Items: nil}, nil
	}

	return &catalogV1.ListPromotionResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *PromotionRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().Promotion.Query().
		Where(promotion.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, catalogV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *PromotionRepo) Get(ctx context.Context, req *catalogV1.GetPromotionRequest) (*catalogV1.Promotion, error) {
	if req == nil {
		return nil, catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Promotion.Query()

	var whereCond []func(s *sql.Selector)
	switch req.GetQueryBy().(type) {
	default:
	case *catalogV1.GetPromotionRequest_Id:
		whereCond = append(whereCond, promotion.IDEQ(req.GetId()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *PromotionRepo) Create(ctx context.Context, req *catalogV1.CreatePromotionRequest) error {
	if req == nil || req.Data == nil {
		return catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Promotion.Create().
		SetNillableName(req.Data.Name).
		SetNillablePromoType(r.promoTypeConverter.ToEntity(req.Data.PromoType)).
		SetNillableProductID(req.Data.ProductId).
		SetNillableSpecialPrice(req.Data.SpecialPrice).
		SetNillableThresholdAmount(req.Data.ThresholdAmount).
		SetNillableReductionAmount(req.Data.ReductionAmount).
		SetNillablePriority(req.Data.Priority).
		SetNillableValidFrom(timeutil.TimestamppbToTime(req.Data.ValidFrom)).
		SetNillableValidUntil(timeutil.TimestamppbToTime(req.Data.ValidUntil)).
		SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
		SetNillableCurrency(req.Data.Currency).
		SetNillableTenantID(req.Data.TenantId).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert promotion failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *PromotionRepo) Update(ctx context.Context, req *catalogV1.UpdatePromotionRequest) error {
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
			createReq := &catalogV1.CreatePromotionRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	builder := r.entClient.Client().Promotion.Update()

	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *catalogV1.Promotion) {
			builder.
				SetNillableName(req.Data.Name).
				SetNillablePromoType(r.promoTypeConverter.ToEntity(req.Data.PromoType)).
				SetNillableProductID(req.Data.ProductId).
				SetNillableSpecialPrice(req.Data.SpecialPrice).
				SetNillableThresholdAmount(req.Data.ThresholdAmount).
				SetNillableReductionAmount(req.Data.ReductionAmount).
				SetNillablePriority(req.Data.Priority).
				SetNillableValidFrom(timeutil.TimestamppbToTime(req.Data.ValidFrom)).
				SetNillableValidUntil(timeutil.TimestamppbToTime(req.Data.ValidUntil)).
				SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
				SetNillableCurrency(req.Data.Currency).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(promotion.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *PromotionRepo) Delete(ctx context.Context, req *catalogV1.DeletePromotionRequest) error {
	if req == nil {
		return catalogV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Promotion.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(promotion.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete promotion failed: %s", err.Error())
		return catalogV1.ErrorInternalServerError("delete data failed")
	}

	return nil
}
