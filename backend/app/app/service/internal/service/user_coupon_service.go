package service

import (
	"context"
	"encoding/json"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	couponV1 "go-wind-shop/api/gen/go/coupon/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

// UserCouponService 用户优惠券前台服务（app BFF，REST → gRPC 转发）。
//
// 裁剪版：仅 List/Get/Quote + Claim 领取，无 Create/Update/Delete（admin 分配制不可对买家暴露）。
// 用户隔离双重保障：
//   - BFF fail-closed：List 强制注入 userId=当前登录用户，JSON 解析失败即拒（仿
//     internal_message_recipient_service.go 的 ListUserInbox）。
//   - core UserPrivacy：user_coupon 表注入 UserPrivacy，按 viewer.user_id 自动注入
//     WHERE 条件。Get/Quote 的隔离完全依赖 core UserPrivacy（请求体无 user_id 字段
//     可注入，但 ent privacy 层强制兜底，跨用户 coupon_id 查不到）。
//
// Claim：BFF 纯透传 template_id，user_id 由 core 从 viewer 强制（事务内 ForUpdate
// 原子校验 claimable/status/窗口/限领）。强制 auth（rest_server 不入白名单）。
type UserCouponService struct {
	appV1.UserCouponServiceHTTPServer

	log *log.Helper

	userCouponServiceClient couponV1.UserCouponServiceClient
}

func NewUserCouponService(
	ctx *bootstrap.Context,
	userCouponServiceClient couponV1.UserCouponServiceClient,
) *UserCouponService {
	return &UserCouponService{
		log:                     ctx.NewLoggerHelper("user-coupon/service/app-service"),
		userCouponServiceClient: userCouponServiceClient,
	}
}

// List 强制把 userId=当前用户 注入到分页 query，确保只返回本人的券。
// JSON 解析/序列化任一失败均 fail-closed 返回错误，避免保留客户端原 query 越权。
func (s *UserCouponService) List(ctx context.Context, req *paginationV1.PagingRequest) (*couponV1.ListUserCouponResponse, error) {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	userId := operator.GetUserId()

	queryMap := map[string]any{}
	if raw := req.GetQuery(); raw != "" {
		if jErr := json.Unmarshal([]byte(raw), &queryMap); jErr != nil {
			return nil, appV1.ErrorInternalServerError("internal error")
		}
	}
	queryMap["userId"] = userId
	newJSON, mErr := json.Marshal(queryMap)
	if mErr != nil {
		return nil, appV1.ErrorInternalServerError("internal error")
	}
	req.FilteringType = &paginationV1.PagingRequest_Query{Query: string(newJSON)}

	return s.userCouponServiceClient.List(ctx, req)
}

// Get 透传。隔离靠 core UserPrivacy（按 viewer.user_id 自动注入 WHERE，跨用户
// coupon_id 查不到→NotFound）。BFF 侧请求体无 user_id 字段可注入。
func (s *UserCouponService) Get(ctx context.Context, req *couponV1.GetUserCouponRequest) (*couponV1.UserCoupon, error) {
	return s.userCouponServiceClient.Get(ctx, req)
}

// Quote 透传。隔离靠 core UserPrivacy（同 Get）。试算不持锁不落库，
// 最终抵扣以下单时事务内校验为准。
func (s *UserCouponService) Quote(ctx context.Context, req *couponV1.QuoteRequest) (*couponV1.QuoteResponse, error) {
	return s.userCouponServiceClient.Quote(ctx, req)
}

// Claim 透传。user_id 由 core 从 viewer 强制（BFF 只透传 template_id），
// core 事务内 ForUpdate 原子校验 claimable/status/窗口/限领。强制 auth。
func (s *UserCouponService) Claim(ctx context.Context, req *couponV1.ClaimCouponRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, appV1.ErrorBadRequest("invalid parameter")
	}
	return s.userCouponServiceClient.Claim(ctx, req)
}
