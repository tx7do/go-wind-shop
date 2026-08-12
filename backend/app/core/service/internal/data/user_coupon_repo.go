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
	"go-wind-shop/app/core/service/internal/data/ent/predicate"
	"go-wind-shop/app/core/service/internal/data/ent/usercoupon"

	couponV1 "go-wind-shop/api/gen/go/coupon/service/v1"
)

type UserCouponRepo struct {
	entClient *entCrud.EntClient[*ent.Client]
	log       *log.Helper

	mapper *mapper.CopierMapper[couponV1.UserCoupon, ent.UserCoupon]

	statusConverter *mapper.EnumTypeConverter[couponV1.UserCoupon_Status, usercoupon.Status]

	repository *entCrud.Repository[
		ent.UserCouponQuery, ent.UserCouponSelect,
		ent.UserCouponCreate, ent.UserCouponCreateBulk,
		ent.UserCouponUpdate, ent.UserCouponUpdateOne,
		ent.UserCouponDelete,
		predicate.UserCoupon,
		couponV1.UserCoupon, ent.UserCoupon,
	]
}

func NewUserCouponRepo(ctx *bootstrap.Context, entClient *entCrud.EntClient[*ent.Client]) *UserCouponRepo {
	repo := &UserCouponRepo{
		log:             ctx.NewLoggerHelper("user-coupon/repo/core-service"),
		entClient:       entClient,
		mapper:          mapper.NewCopierMapper[couponV1.UserCoupon, ent.UserCoupon](),
		statusConverter: mapper.NewEnumTypeConverter[couponV1.UserCoupon_Status, usercoupon.Status](couponV1.UserCoupon_Status_name, couponV1.UserCoupon_Status_value),
	}

	repo.init()

	return repo
}

func (r *UserCouponRepo) init() {
	r.repository = entCrud.NewRepository[
		ent.UserCouponQuery, ent.UserCouponSelect,
		ent.UserCouponCreate, ent.UserCouponCreateBulk,
		ent.UserCouponUpdate, ent.UserCouponUpdateOne,
		ent.UserCouponDelete,
		predicate.UserCoupon,
		couponV1.UserCoupon, ent.UserCoupon,
	](r.mapper)

	r.mapper.AppendConverters(copierutil.NewTimeStringConverterPair())
	r.mapper.AppendConverters(copierutil.NewTimeTimestamppbConverterPair())

	r.mapper.AppendConverters(r.statusConverter.NewConverterPair())
}

func (r *UserCouponRepo) Count(ctx context.Context, req *paginationV1.PagingRequest) (int, error) {
	builder := r.entClient.Client().UserCoupon.Query()

	whereSelectors, _, err := r.repository.BuildListSelectorWithPaging(builder, req)
	if len(whereSelectors) != 0 {
		builder.Modify(whereSelectors...)
	}

	count, err := builder.Count(ctx)
	if err != nil {
		r.log.Errorf("query user coupon count failed: %s", err.Error())
		return 0, couponV1.ErrorInternalServerError("query count failed")
	}

	return count, nil
}

func (r *UserCouponRepo) List(ctx context.Context, req *paginationV1.PagingRequest) (*couponV1.ListUserCouponResponse, error) {
	if req == nil {
		return nil, couponV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().UserCoupon.Query()

	ret, err := r.repository.ListWithPaging(ctx, builder, builder.Clone(), req)
	if err != nil {
		return nil, err
	}
	if ret == nil {
		return &couponV1.ListUserCouponResponse{Total: 0, Items: nil}, nil
	}

	return &couponV1.ListUserCouponResponse{
		Total: ret.Total,
		Items: ret.Items,
	}, nil
}

func (r *UserCouponRepo) IsExist(ctx context.Context, id uint32) (bool, error) {
	exist, err := r.entClient.Client().UserCoupon.Query().
		Where(usercoupon.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		r.log.Errorf("query exist failed: %s", err.Error())
		return false, couponV1.ErrorInternalServerError("query exist failed")
	}
	return exist, nil
}

func (r *UserCouponRepo) Get(ctx context.Context, req *couponV1.GetUserCouponRequest) (*couponV1.UserCoupon, error) {
	if req == nil {
		return nil, couponV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().UserCoupon.Query()

	var whereCond []func(s *sql.Selector)
	switch req.GetQueryBy().(type) {
	default:
	case *couponV1.GetUserCouponRequest_Id:
		whereCond = append(whereCond, usercoupon.IDEQ(req.GetId()))
	}

	dto, err := r.repository.Get(ctx, builder, req.GetViewMask(), whereCond...)
	if err != nil {
		return nil, err
	}

	return dto, err
}

func (r *UserCouponRepo) Create(ctx context.Context, req *couponV1.CreateUserCouponRequest) error {
	if req == nil || req.Data == nil {
		return couponV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().UserCoupon.Create().
		SetNillableUserID(req.Data.UserId).
		SetNillableCouponTemplateID(req.Data.CouponTemplateId).
		SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
		SetNillableRedeemedAt(timeutil.TimestamppbToTime(req.Data.RedeemedAt)).
		SetNillableRedeemedOrderID(req.Data.RedeemedOrderId).
		SetNillableAppliedDiscountAmount(req.Data.AppliedDiscountAmount).
		SetNillableTenantID(req.Data.TenantId).
		SetNillableCreatedBy(req.Data.CreatedBy).
		SetCreatedAt(time.Now())

	if req.Data.Id != nil {
		builder.SetID(req.GetData().GetId())
	}

	if err := builder.Exec(ctx); err != nil {
		r.log.Errorf("insert user coupon failed: %s", err.Error())
		return couponV1.ErrorInternalServerError("insert data failed")
	}

	return nil
}

func (r *UserCouponRepo) Update(ctx context.Context, req *couponV1.UpdateUserCouponRequest) error {
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
			createReq := &couponV1.CreateUserCouponRequest{Data: req.Data}
			createReq.Data.CreatedBy = createReq.Data.UpdatedBy
			createReq.Data.UpdatedBy = nil
			return r.Create(ctx, createReq)
		}
	}

	builder := r.entClient.Client().Debug().UserCoupon.Update()

	err := r.repository.UpdateX(ctx, builder, req.Data, req.GetUpdateMask(),
		func(dto *couponV1.UserCoupon) {
			builder.
				SetNillableUserID(req.Data.UserId).
				SetNillableCouponTemplateID(req.Data.CouponTemplateId).
				SetNillableStatus(r.statusConverter.ToEntity(req.Data.Status)).
				SetNillableRedeemedAt(timeutil.TimestamppbToTime(req.Data.RedeemedAt)).
				SetNillableRedeemedOrderID(req.Data.RedeemedOrderId).
				SetNillableAppliedDiscountAmount(req.Data.AppliedDiscountAmount).
				SetNillableUpdatedBy(req.Data.UpdatedBy).
				SetUpdatedAt(time.Now())
		},
		func(s *sql.Selector) {
			s.Where(sql.EQ(usercoupon.FieldID, req.GetId()))
		},
	)

	return err
}

func (r *UserCouponRepo) Delete(ctx context.Context, req *couponV1.DeleteUserCouponRequest) error {
	if req == nil {
		return couponV1.ErrorBadRequest("invalid parameter")
	}

	builder := r.entClient.Client().Debug().UserCoupon.Delete()

	var err error
	_, err = r.repository.Delete(ctx, builder, func(s *sql.Selector) {
		s.Where(sql.EQ(usercoupon.FieldID, req.GetId()))
	})
	if err != nil {
		r.log.Errorf("delete user coupon failed: %s", err.Error())
		return couponV1.ErrorInternalServerError("delete user coupon failed")
	}

	return nil
}
