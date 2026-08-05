package service

import (
	"context"
	"encoding/json"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	internalMessageV1 "go-wind-shop/api/gen/go/internal_message/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

// InternalMessageRecipientService 站内信收件箱前台服务（BFF 网关，REST → gRPC 转发）。
// recipient 表无 UserPrivacy 策略，行级隔离完全依赖本层从登录态强制注入
// recipientUserId / userId，客户端传入的归属字段一律被覆盖，防止越权读取/操作他人收件箱。
type InternalMessageRecipientService struct {
	appV1.InternalMessageRecipientServiceHTTPServer

	log *log.Helper

	internalMessageRecipientServiceClient internalMessageV1.InternalMessageRecipientServiceClient
}

func NewInternalMessageRecipientService(
	ctx *bootstrap.Context,
	internalMessageRecipientServiceClient internalMessageV1.InternalMessageRecipientServiceClient,
) *InternalMessageRecipientService {
	return &InternalMessageRecipientService{
		log:                                   ctx.NewLoggerHelper("internal-message-recipient/service/app-service"),
		internalMessageRecipientServiceClient: internalMessageRecipientServiceClient,
	}
}

// ListUserInbox 强制把 recipientUserId=当前用户 注入到分页 query，确保只返回本人的收件箱。
// 客户端可携带 status 等其它过滤字段，会与 recipientUserId 合并。
func (s *InternalMessageRecipientService) ListUserInbox(ctx context.Context, req *paginationV1.PagingRequest) (*internalMessageV1.ListUserInboxResponse, error) {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	userId := operator.GetUserId()

	// 解析现有 query（可能含 status 等过滤），注入 recipientUserId 后写回。
	queryMap := map[string]any{}
	if raw := req.GetQuery(); raw != "" {
		_ = json.Unmarshal([]byte(raw), &queryMap) // 解析失败则按空 map 处理，避免脏数据导致 500
	}
	queryMap["recipientUserId"] = userId
	if newJSON, err := json.Marshal(queryMap); err == nil {
		req.FilteringType = &paginationV1.PagingRequest_Query{Query: string(newJSON)}
	}

	return s.internalMessageRecipientServiceClient.ListUserInbox(ctx, req)
}

// MarkNotificationAsRead 强制以当前登录用户身份标记已读，忽略客户端传入的 userId。
func (s *InternalMessageRecipientService) MarkNotificationAsRead(ctx context.Context, req *internalMessageV1.MarkNotificationAsReadRequest) (*emptypb.Empty, error) {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	req.UserId = operator.GetUserId()

	return s.internalMessageRecipientServiceClient.MarkNotificationAsRead(ctx, req)
}

// DeleteNotificationFromInbox 强制以当前登录用户身份删除，忽略客户端传入的 userId。
func (s *InternalMessageRecipientService) DeleteNotificationFromInbox(ctx context.Context, req *internalMessageV1.DeleteNotificationFromInboxRequest) (*emptypb.Empty, error) {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}
	req.UserId = operator.GetUserId()

	return s.internalMessageRecipientServiceClient.DeleteNotificationFromInbox(ctx, req)
}
