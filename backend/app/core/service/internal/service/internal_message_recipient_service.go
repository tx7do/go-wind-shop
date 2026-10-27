package service

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	internalMessageV1 "go-wind-shop/api/gen/go/internal_message/service/v1"
	identityV1 "go-wind-shop/api/gen/go/identity/service/v1"
	notificationV1 "go-wind-shop/api/gen/go/notification/service/v1"
)

type InternalMessageRecipientService struct {
	internalMessageV1.UnimplementedInternalMessageRecipientServiceServer

	log *log.Helper

	internalMessageRepo          *data.InternalMessageRepo
	internalMessageRecipientRepo *data.InternalMessageRecipientRepo

	userRepo data.UserRepo

	// notifier 通知域投递缝（Notifier 接口，由 NotificationService 实现）。
	// 构造期留 nil，装配期由 wiring 注入；未注入时收件行仅落库，不出站到渠道路由。
	notifier Notifier
}

// RegisterNotifier 注入通知域投递缝。
func (s *InternalMessageRecipientService) RegisterNotifier(notifier Notifier) {
	s.notifier = notifier
}

// Create 落一条收件行（BFF 发送编排逐收件人调用）。
// 落库成功后经通知域投递缝出站（INTERNAL_MESSAGE 事件，related_id=消息ID）：
// 渠道路由决定走邮箱还是仅站内行；出站失败不影响收件行（台账可查，异步可重试）。
func (s *InternalMessageRecipientService) Create(ctx context.Context, req *internalMessageV1.CreateInternalMessageRecipientRequest) (*internalMessageV1.InternalMessageRecipient, error) {
	if req == nil || req.Data == nil {
		return nil, internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	entity, err := s.internalMessageRecipientRepo.Create(ctx, req.Data)
	if err != nil {
		s.log.Errorf("create recipient failed: %v", err)
		return nil, err
	}

	if s.notifier != nil && entity.GetRecipientUserId() > 0 {
		// title/content 取请求入参（shop 收件行表无这两列，repo 往返后 DTO 不携带）
		if _, dErr := s.notifier.SendDirect(ctx, &notificationV1.SendDirectNotificationRequest{
			EventType:       notificationV1.EventType_INTERNAL_MESSAGE,
			Target:          s.notificationTargetFor(ctx, entity.GetRecipientUserId()),
			Title:           req.Data.GetTitle(),
			Content:         req.Data.GetContent(),
			RecipientUserId: entity.RecipientUserId,
			RelatedId:       entity.MessageId,
		}); dErr != nil {
			s.log.Errorf("recipient [%d] notifier deliver failed (delivery ledger records attempt): %v", entity.GetId(), dErr)
		}
	}

	return entity, nil
}

// notificationTargetFor 返回收件人的 SendDirect Target：
// INTERNAL_MESSAGE 事件的 EMAIL 路由需要邮箱地址，而本链路持有的是收件人
// uid——这里解析 sys_users.email；未配置邮箱回落 uid 字符串（既有行为，
// EMAIL 渠道投递以 FAILED 呈现而非误发）。INTERNAL 渠道不消费 Target。
func (s *InternalMessageRecipientService) notificationTargetFor(ctx context.Context, uid uint32) string {
	u, uErr := s.userRepo.Get(ctx, &identityV1.GetUserRequest{
		QueryBy: &identityV1.GetUserRequest_Id{Id: uid},
	})
	if uErr != nil || u == nil {
		s.log.Warnf("notification target: read user [%d] failed: %v", uid, uErr)
		return strconv.FormatUint(uint64(uid), 10)
	}
	if strings.TrimSpace(u.GetEmail()) != "" {
		return u.GetEmail()
	}
	return strconv.FormatUint(uint64(uid), 10)
}

func NewInternalMessageRecipientService(
	ctx *bootstrap.Context,
	internalMessageRepo *data.InternalMessageRepo,
	internalMessageRecipientRepo *data.InternalMessageRecipientRepo,
	userRepo data.UserRepo,
) *InternalMessageRecipientService {
	return &InternalMessageRecipientService{
		log:                          ctx.NewLoggerHelper("internal-message-recipient/service/core-service"),
		internalMessageRepo:          internalMessageRepo,
		internalMessageRecipientRepo: internalMessageRecipientRepo,
		userRepo:                     userRepo,
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
