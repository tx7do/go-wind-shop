package service

import (
	"context"

	"go-wind-shop/pkg/logx"

	"github.com/tx7do/kratos-bootstrap/bootstrap"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"
)

// GrouponService 拼团前台服务（C端只读：活动/团）。
// 租户隔离由 core 侧 TenantPrivacy 按 viewer 自动注入。
type GrouponService struct {
	appV1.GrouponServiceHTTPServer

	log *logx.Helper

	grouponActivityServiceClient catalogV1.GrouponActivityServiceClient
}

func NewGrouponService(
	ctx *bootstrap.Context,
	grouponActivityServiceClient catalogV1.GrouponActivityServiceClient,
) *GrouponService {
	return &GrouponService{
		log:                          logx.New(ctx.NewLoggerHelper("groupon/app-service")),
		grouponActivityServiceClient: grouponActivityServiceClient,
	}
}

// ListActiveGroupons 当前生效中的拼团活动（匿名可读，商品页拼团价展示）。
func (s *GrouponService) ListActiveGroupons(ctx context.Context, req *catalogV1.ListActiveGrouponsRequest) (*catalogV1.ListGrouponActivityResponse, error) {
	return s.grouponActivityServiceClient.ListActiveGroupons(ctx, req)
}

// GetGroup 团详情（订单详情展示团状态）。
func (s *GrouponService) GetGroup(ctx context.Context, req *catalogV1.GetGrouponGroupRequest) (*catalogV1.GrouponGroup, error) {
	return s.grouponActivityServiceClient.GetGroup(ctx, req)
}

// ListGroups 活动下的团列表（参团列表）。
func (s *GrouponService) ListGroups(ctx context.Context, req *catalogV1.ListGrouponGroupsRequest) (*catalogV1.ListGrouponGroupResponse, error) {
	return s.grouponActivityServiceClient.ListGroups(ctx, req)
}
