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
	"github.com/tx7do/go-utils/timeutil"

	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/coupontemplate"
	"go-wind-shop/app/core/service/internal/data/ent/predicate"

	couponV1 "go-wind-shop/api/gen/go/coupon/service/v1"
)

type CouponTemplateRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[couponV1.CouponTemplate, ent.CouponTemplate]

	discountTypeConverter *mapper.EnumTypeConverter[couponV1.CouponTemplate_DiscountType, coupontemplate.DiscountType]
	statusConverter       *mapper.EnumTypeConverter[couponV1.CouponTemplate_Status, coupontemplate.Status]

	repository *entCrud.Repository[
		ent.CouponTemplateQuery, ent.CouponTemplateSelect,
		ent.CouponTemplateCreate, ent.CouponTemplateCreateBulk,
		ent.CouponTemplateUpdate, ent.CouponTemplateUpdateOne,
		ent.CouponTemplateDelete,
		predicate.CouponTemplate,
		couponV1.CouponTemplate, ent.CouponTemplate,
	]
}

func NewCouponTemplateRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *CouponTemplateRepo {
	repo := &CouponTemplateRepo{
		log:                   ctx.NewLoggerHelper("coupon-template/repo/core-service"),
		entClient:             entClient,
		mapper:                mapper.NewCopierMapper[couponV1.CouponTemplate, ent.CouponTemplate](),
		discountTypeConverter: mapper.NewEnumTypeConverter[couponV1.CouponTemplate_DiscountType, coupontemplate.DiscountType](couponV1.CouponTemplate_DiscountType_name, couponV1.CouponTemplate_DiscountType_value),
		statusConverter:       mapper.NewEnumTypeConverter[couponV1.CouponTemplate_Status, coupontemplate.Status](couponV1.CouponTemplate_Status_name, couponV1.CouponTemplate_Status_value),
	}

	repo.init()

	return repo
}

func (r *CouponTemplateRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.CouponTemplateQuery, ent.CouponTemplateSelect,
		ent.CouponTemplateCreate, ent.CouponTemplateCreateBulk,
		ent.CouponTemplateUpdate, ent.CouponTemplateUpdateOne,
		ent.CouponTemplateDelete,
		predicate.CouponTemplate,
		couponV1.CouponTemplate, ent.CouponTemplate,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	r.mapper.AppendConverters(r.discountTypeConverter.NewConverterPair())
	r.mapper.AppendConverters(r.statusConverter.NewConverterPair())
}

func (r *CouponTemplateRepo) Count(ctx context.Context, req *paginationV1.PagingRequest) (int, error) {
	builder := r.entClient.Client().CouponTemplate.Query()

	whereSelectors, _, err := r.repository.BuildListSelectorWithPaging(builder, req)
	if len(whereSelectors) != 0 {
		builder.Modify(whereSelectors...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query coupon template count failed: %s", err.Error())
		return 0, couponV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *CouponTemplateRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*couponV1.ListCouponTemplateResponse, error) {
	if req == nil {
		return nil, couponV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().CouponTemplate.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &couponV1.ListCouponTemplateResponse{Total: 0, Items: nil}, nil
	}

	return &couponV1.ListCouponTemplateResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

// ListClaimable 列出公开可领模板（claimable=true AND status=ACTIVE）。
// 供 app-scope 领券中心调用。有效窗口由 service 层后过滤（时间判断非 ent 谓词）。
// TenantPrivacy 按 viewer tenant_id 自动注入 WHERE。
func (r *CouponTemplateRepo) ListClaimable(ctx context.Context, req *paginationV1.PagingRequest) (*couponV1.ListCouponTemplateResponse, error) {
	if req == nil {
		return nil, couponV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().CouponTemplate.Query().
		Where(
			coupontemplate.ClaimableEQ(true),
			coupontemplate.StatusEQ(coupontemplate.StatusCouponTemplateStatusActive),
		)

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &couponV1.ListCouponTemplateResponse{Total: 0, Items: nil}, nil
	}

	return &couponV1.ListCouponTemplateResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *CouponTemplateRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().CouponTemplate.Query().
		Where(coupontemplate.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, couponV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *CouponTemplateRepo) Get(ctx context.Context, req *couponV1.GetCouponTemplateRequest) (*couponV1.CouponTemplate, error) {
	if req == nil {
		return nil, couponV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().CouponTemplate.Query()

	var whereCond []func(s *sql.Selector)
	switch req.GetQueryBy().(type) {
	default:
	case *couponV1.GetCouponTemplateRequest_Id:
		whereCond = append(whereCond, coupontemplate.IDEQ(req.GetId()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *CouponTemplateRepo) Create(ctx context.Context, req *couponV1.CreateCouponTemplateRequest) error {
	if req == nil || req.Data == nil {
		return couponV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().CouponTemplate.Create().
		SetNillableDiscountType(r.discountTypeConverter.ToEntity(req.Data.DiscountType)).
		SetNillableDiscountValue(req.Data.DiscountValue).
		SetNillableDiscountPercentage(req.Data.DiscountPercentage).
		SetNillableValidFrom(timeutil.TimestamppbToTime(req.Data.ValidFrom)).
		SetNillableValidUntil(timeutil.TimestamppbToTime(req.Data.ValidUntil)).
		SetNillableMaxRedemptions(req.Data.MaxRedemptions).
		SetNillableMaxRedemptionsPerUser(req.Data.MaxRedemptionsPerUser).
		SetNillableRedeemedCount(req.Data.RedeemedCount).
		SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
		SetNillableCurrency(req.Data.Currency).
		SetNillableTenantID(req.Data.TenantId).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert coupon template failed: %s", err.Error())
		return couponV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *CouponTemplateRepo) Update(ctx context.Context, req *couponV1.UpdateCouponTemplateRequest) error {
	if req == nil || req.Data == nil {
		return couponV1.ErrorBadRequest("invalid parameter")
	}
	if req.GetId() == 0 {
		return couponV1.ErrorBadRequest("id is required")
	}

	// 如果不存在则创建
	if req.GetAllowMissing() {
		exist, err := r.IsExist(ctx, req.GetId())
		if err != nil {
			return err
		}
		if !exist {
			createReq := &couponV1.CreateCouponTemplateRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	builder := r.entClient.Client().Debug().CouponTemplate.Update()

	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *couponV1.CouponTemplate) {
			builder.
				SetNillableDiscountType(r.discountTypeConverter.ToEntity(req.Data.DiscountType)).
				SetNillableDiscountValue(req.Data.DiscountValue).
				SetNillableDiscountPercentage(req.Data.DiscountPercentage).
				SetNillableValidFrom(timeutil.TimestamppbToTime(req.Data.ValidFrom)).
				SetNillableValidUntil(timeutil.TimestamppbToTime(req.Data.ValidUntil)).
				SetNillableMaxRedemptions(req.Data.MaxRedemptions).
				SetNillableMaxRedemptionsPerUser(req.Data.MaxRedemptionsPerUser).
				SetNillableRedeemedCount(req.Data.RedeemedCount).
				SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
				SetNillableCurrency(req.Data.Currency).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(coupontemplate.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *CouponTemplateRepo) Delete(ctx context.Context, req *couponV1.DeleteCouponTemplateRequest) error {
	if req == nil {
		return couponV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Debug().CouponTemplate.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(coupontemplate.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete coupon template failed: %s", err.Error())
		return couponV1.ErrorInternalServerError("delete coupon template failed")
	}

	return nil
}
