package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	couponV1 "go-wind-shop/api/gen/go/coupon/service/v1"
)

// CouponTemplateService 优惠券模板前台服务（app BFF，REST → gRPC 转发）。
//
// 裁剪版：仅 List（领券中心浏览可领模板），无 Create/Update/Delete/Count/Get
// （运营配置不可由买家侧变更）。
//
// List 纯透传到 core。core 按 viewer 类型分流：平台视图返回全量模板；普通用户/
// 匿名仅返回 claimable=true AND status=ACTIVE 且在有效窗口内的模板（详见 core
// coupon_template_service.List）。匿名可读由 TenantPrivacy 对无 viewer 放行保证
// （与商品 List 同机制），故本接口入 rest_server 白名单。
type CouponTemplateService struct {
	appV1.CouponTemplateServiceHTTPServer

	log *log.Helper

	couponTemplateServiceClient couponV1.CouponTemplateServiceClient
}

func NewCouponTemplateService(
	ctx *bootstrap.Context,
	couponTemplateServiceClient couponV1.CouponTemplateServiceClient,
) *CouponTemplateService {
	return &CouponTemplateService{
		log:                         ctx.NewLoggerHelper("coupon-template/service/app-service"),
		couponTemplateServiceClient: couponTemplateServiceClient,
	}
}

func (s *CouponTemplateService) List(ctx context.Context, req *paginationV1.PagingRequest) (*couponV1.ListCouponTemplateResponse, error) {
	return s.couponTemplateServiceClient.List(ctx, req)
}
