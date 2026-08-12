package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	adminV1 "go-wind-shop/api/gen/go/admin/service/v1"
	couponV1 "go-wind-shop/api/gen/go/coupon/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

// CouponTemplateService 优惠券模板管理（admin BFF，REST → gRPC 转发）。
type CouponTemplateService struct {
	adminV1.CouponTemplateServiceHTTPServer

	log *log.Helper

	couponTemplateServiceClient couponV1.CouponTemplateServiceClient
}

func NewCouponTemplateService(
	ctx *bootstrap.Context,
	couponTemplateServiceClient couponV1.CouponTemplateServiceClient,
) *CouponTemplateService {
	return &CouponTemplateService{
		log:                         ctx.NewLoggerHelper("coupon-template/service/admin-service"),
		couponTemplateServiceClient: couponTemplateServiceClient,
	}
}

func (s *CouponTemplateService) List(ctx context.Context, req *paginationV1.PagingRequest) (*couponV1.ListCouponTemplateResponse, error) {
	return s.couponTemplateServiceClient.List(ctx, req)
}

func (s *CouponTemplateService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*couponV1.CountCouponTemplateResponse, error) {
	return s.couponTemplateServiceClient.Count(ctx, req)
}

func (s *CouponTemplateService) Get(ctx context.Context, req *couponV1.GetCouponTemplateRequest) (*couponV1.CouponTemplate, error) {
	return s.couponTemplateServiceClient.Get(ctx, req)
}

func (s *CouponTemplateService) Create(ctx context.Context, req *couponV1.CreateCouponTemplateRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	return s.couponTemplateServiceClient.Create(ctx, req)
}

func (s *CouponTemplateService) Update(ctx context.Context, req *couponV1.UpdateCouponTemplateRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.Id = trans.Ptr(req.GetId())

	req.Data.UpdatedBy = trans.Ptr(operator.GetUserId())
	if req.UpdateMask != nil {
		req.UpdateMask.Paths = append(req.UpdateMask.Paths, "updated_by")
	}

	return s.couponTemplateServiceClient.Update(ctx, req)
}

func (s *CouponTemplateService) Delete(ctx context.Context, req *couponV1.DeleteCouponTemplateRequest) (*emptypb.Empty, error) {
	return s.couponTemplateServiceClient.Delete(ctx, req)
}

// UserCouponService 用户优惠券发放管理（admin BFF，REST → gRPC 转发）。
// 仅暴露 Create(发放)/List/Get/Delete，无 Update。admin 侧发放时 user_id 取自请求体
// （运营指定接收用户），core 的 TenantPrivacy 按运营 token 的 tenant_id 注入过滤。
type UserCouponService struct {
	adminV1.UserCouponServiceHTTPServer

	log *log.Helper

	userCouponServiceClient couponV1.UserCouponServiceClient
}

func NewUserCouponService(
	ctx *bootstrap.Context,
	userCouponServiceClient couponV1.UserCouponServiceClient,
) *UserCouponService {
	return &UserCouponService{
		log:                     ctx.NewLoggerHelper("user-coupon/service/admin-service"),
		userCouponServiceClient: userCouponServiceClient,
	}
}

func (s *UserCouponService) List(ctx context.Context, req *paginationV1.PagingRequest) (*couponV1.ListUserCouponResponse, error) {
	return s.userCouponServiceClient.List(ctx, req)
}

func (s *UserCouponService) Count(ctx context.Context, req *paginationV1.PagingRequest) (*couponV1.CountUserCouponResponse, error) {
	return s.userCouponServiceClient.Count(ctx, req)
}

func (s *UserCouponService) Get(ctx context.Context, req *couponV1.GetUserCouponRequest) (*couponV1.UserCoupon, error) {
	return s.userCouponServiceClient.Get(ctx, req)
}

func (s *UserCouponService) Create(ctx context.Context, req *couponV1.CreateUserCouponRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, adminV1.ErrorBadRequest("invalid parameter")
	}

	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	req.Data.CreatedBy = trans.Ptr(operator.UserId)

	return s.userCouponServiceClient.Create(ctx, req)
}

func (s *UserCouponService) Delete(ctx context.Context, req *couponV1.DeleteUserCouponRequest) (*emptypb.Empty, error) {
	return s.userCouponServiceClient.Delete(ctx, req)
}
