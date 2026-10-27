package service

import (
	"context"

	"go-wind-shop/pkg/logx"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	adminV1 "go-wind-shop/api/gen/go/admin/service/v1"
	notifyV1 "go-wind-shop/api/gen/go/notification/service/v1"
)

// NotificationRuleService 通知路由规则管理（admin 门面，透传 core）。
type NotificationRuleService struct {
	adminV1.NotificationRuleServiceHTTPServer

	log *logx.Helper
	ruleCli  notifyV1.NotificationRuleServiceClient
}

func NewNotificationRuleService(ctx *bootstrap.Context, ruleClient notifyV1.NotificationRuleServiceClient) *NotificationRuleService {
	return &NotificationRuleService{
		log:     logx.New(ctx.NewLoggerHelper("notification-rule/service/admin-service")),
		ruleCli: ruleClient,
	}
}

func (s *NotificationRuleService) ListNotificationRule(ctx context.Context, req *paginationV1.PagingRequest) (*notifyV1.ListNotificationRuleResponse, error) {
	return s.ruleCli.List(ctx, req)
}

func (s *NotificationRuleService) GetNotificationRule(ctx context.Context, req *notifyV1.GetNotificationRuleRequest) (*notifyV1.NotificationRule, error) {
	return s.ruleCli.Get(ctx, req)
}

func (s *NotificationRuleService) CreateNotificationRule(ctx context.Context, req *notifyV1.CreateNotificationRuleRequest) (*notifyV1.NotificationRule, error) {
	return s.ruleCli.Create(ctx, req)
}

func (s *NotificationRuleService) UpdateNotificationRule(ctx context.Context, req *notifyV1.UpdateNotificationRuleRequest) (*notifyV1.NotificationRule, error) {
	return s.ruleCli.Update(ctx, req)
}

func (s *NotificationRuleService) DeleteNotificationRule(ctx context.Context, req *notifyV1.DeleteNotificationRuleRequest) (*emptypb.Empty, error) {
	return s.ruleCli.Delete(ctx, req)
}

func (s *NotificationRuleService) TestDispatchNotification(ctx context.Context, req *notifyV1.TestDispatchNotificationRequest) (*notifyV1.TestDispatchNotificationResponse, error) {
	return s.ruleCli.TestDispatch(ctx, req)
}
