package service

import (
	"context"
	"encoding/json"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	internalMessageV1 "go-wind-shop/api/gen/go/internal_message/service/v1"
)

type InternalMessageRecipientService struct {
	internalMessageV1.UnimplementedInternalMessageRecipientServiceServer

	log *log.Helper

	internalMessageRepo          *data.InternalMessageRepo
	internalMessageRecipientRepo *data.InternalMessageRecipientRepo
}

func NewInternalMessageRecipientService(
	ctx *bootstrap.Context,
	internalMessageRepo *data.InternalMessageRepo,
	internalMessageRecipientRepo *data.InternalMessageRecipientRepo,
) *InternalMessageRecipientService {
	return &InternalMessageRecipientService{
		log:                          ctx.NewLoggerHelper("internal-message-recipient/service/core-service"),
		internalMessageRepo:          internalMessageRepo,
		internalMessageRecipientRepo: internalMessageRecipientRepo,
	}
}

// ListUserInbox 获取用户的收件箱列表 (通知类)
//
// 行级隔离由 schema Policy（UserPrivacy，列名 recipient_user_id）兜底：
// 普通 viewer 的查询自动注入 recipient_user_id = uid 的 WHERE。
// 此处再补一层显式注入作为纵深防御——即使 BFF 注入被绕过或 core gRPC
// 直连，仍强制按 recipientUserId 过滤。系统/平台视图放行。
func (s *InternalMessageRecipientService) ListUserInbox(ctx context.Context, req *paginationV1.PagingRequest) (*internalMessageV1.ListUserInboxResponse, error) {
	// 纵深防御：普通 viewer 强制注入 recipientUserId 到 query
	if viewerUid, ok := viewerUserIDFromContext(ctx); ok {
		queryMap := map[string]any{}
		if raw := req.GetQuery(); raw != "" {
			if jErr := json.Unmarshal([]byte(raw), &queryMap); jErr != nil {
				// 解析失败 fail-closed，避免脏 query 越权
				return nil, internalMessageV1.ErrorInternalServerError("internal error")
			}
		}
		queryMap["recipientUserId"] = viewerUid
		newJSON, mErr := json.Marshal(queryMap)
		if mErr != nil {
			return nil, internalMessageV1.ErrorInternalServerError("internal error")
		}
		req.FilteringType = &paginationV1.PagingRequest_Query{Query: string(newJSON)}
	}

	resp, err := s.internalMessageRecipientRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	for _, d := range resp.Items {
		if d.MessageId == nil {
			continue
		}

		msg, err := s.internalMessageRepo.Get(ctx, &internalMessageV1.GetInternalMessageRequest{
			QueryBy: &internalMessageV1.GetInternalMessageRequest_Id{
				Id: d.GetMessageId(),
			},
		})
		if err != nil {
			s.log.Errorf("list user inbox failed, get message failed: %s", err)
			continue
		}

		d.Title = msg.Title
		d.Content = msg.Content
	}

	return resp, nil
}

func (s *InternalMessageRecipientService) DeleteNotificationFromInbox(ctx context.Context, req *internalMessageV1.DeleteNotificationFromInboxRequest) (*emptypb.Empty, error) {
	var err error
	err = s.internalMessageRecipientRepo.DeleteNotificationFromInbox(ctx, req)
	return &emptypb.Empty{}, err
}

// MarkNotificationAsRead 将通知标记为已读
func (s *InternalMessageRecipientService) MarkNotificationAsRead(ctx context.Context, req *internalMessageV1.MarkNotificationAsReadRequest) (*emptypb.Empty, error) {
	var err error
	err = s.internalMessageRecipientRepo.MarkNotificationAsRead(ctx, req)
	return &emptypb.Empty{}, err
}

// MarkNotificationsStatus 标记特定用户的某些或所有通知的状态
func (s *InternalMessageRecipientService) MarkNotificationsStatus(ctx context.Context, req *internalMessageV1.MarkNotificationsStatusRequest) (*emptypb.Empty, error) {
	var err error
	err = s.internalMessageRecipientRepo.MarkNotificationsStatus(ctx, req)
	return &emptypb.Empty{}, err
}
