package service

import (
	"context"

	"go-wind-shop/pkg/logx"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	catalogV1 "go-wind-shop/api/gen/go/catalog/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

// PointsService 积分前台服务（C端：余额 + 本人台账）。
// 台账读取走 UserPrivacy 行级隔离（普通用户只能看本人台账）。
type PointsService struct {
	appV1.PointsServiceHTTPServer

	log *logx.Helper

	pointsLedgerServiceClient catalogV1.PointsLedgerServiceClient
}

func NewPointsService(
	ctx *bootstrap.Context,
	pointsLedgerServiceClient catalogV1.PointsLedgerServiceClient,
) *PointsService {
	return &PointsService{
		log:                       logx.New(ctx.NewLoggerHelper("points/app-service")),
		pointsLedgerServiceClient: pointsLedgerServiceClient,
	}
}

// GetMyPointsBalance 我的积分余额（请求里的 userId 会被 core 侧按登录态覆盖语义使用；
// 此处 BFF 强制以登录态注入，防越权查询他人余额）。
func (s *PointsService) GetMyPointsBalance(ctx context.Context, req *catalogV1.GetUserPointsBalanceRequest) (*catalogV1.GetUserPointsBalanceResponse, error) {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.pointsLedgerServiceClient.GetUserPointsBalance(ctx, &catalogV1.GetUserPointsBalanceRequest{
		UserId: operator.UserId,
	})
}

// ListMyPointsLedgers 我的积分台账（隐私层钉定本人）。
func (s *PointsService) ListMyPointsLedgers(ctx context.Context, req *paginationV1.PagingRequest) (*catalogV1.ListPointsLedgerResponse, error) {
	if _, err := auth.FromContext(ctx); err != nil {
		return nil, err
	}
	return s.pointsLedgerServiceClient.List(ctx, req)
}

// CheckIn 每日签到（登录态；BFF 强制以登录态注入 userId）。
func (s *PointsService) CheckIn(ctx context.Context, req *catalogV1.CheckInRequest) (*catalogV1.CheckInResponse, error) {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.pointsLedgerServiceClient.CheckIn(ctx, &catalogV1.CheckInRequest{
		UserId: operator.UserId,
	})
}

// GetCheckInStatus 签到状态（登录态）。
func (s *PointsService) GetCheckInStatus(ctx context.Context, req *catalogV1.GetCheckInStatusRequest) (*catalogV1.GetCheckInStatusResponse, error) {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	return s.pointsLedgerServiceClient.GetCheckInStatus(ctx, &catalogV1.GetCheckInStatusRequest{
		UserId: operator.UserId,
	})
}
