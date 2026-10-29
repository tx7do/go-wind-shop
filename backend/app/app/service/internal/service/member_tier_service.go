package service

import (
	"context"

	"go-wind-shop/pkg/logx"

	"github.com/tx7do/kratos-bootstrap/bootstrap"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

type MemberTierService struct {
	appV1.MemberTierServiceHTTPServer
	log *logx.Helper
	memberTierServiceClient catalogV1.MemberTierServiceClient
}

func NewMemberTierService(ctx *bootstrap.Context, memberTierServiceClient catalogV1.MemberTierServiceClient) *MemberTierService {
	return &MemberTierService{
		log:                      logx.New(ctx.NewLoggerHelper("member-tier/app-service")),
		memberTierServiceClient:  memberTierServiceClient,
	}
}

// GetMyTierInfo 我的等级信息（登录态；BFF 强制注入 userId）。
func (s *MemberTierService) GetMyTierInfo(ctx context.Context, req *catalogV1.GetUserTierInfoRequest) (*catalogV1.GetUserTierInfoResponse, error) {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.memberTierServiceClient.GetUserTierInfo(ctx, &catalogV1.GetUserTierInfoRequest{
		UserId: operator.UserId,
	})
}

// ListTiers 全部等级列表（匿名可读）。
func (s *MemberTierService) ListTiers(ctx context.Context, req *catalogV1.ListActiveMemberTiersRequest) (*catalogV1.ListMemberTierResponse, error) {
	return s.memberTierServiceClient.ListActiveMemberTiers(ctx, req)
}
