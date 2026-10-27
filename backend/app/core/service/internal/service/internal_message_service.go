package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"go-wind-shop/pkg/logx"
	paginationV1 "github.com/tx7do/go-crud/api/gen/go/pagination/v1"
	"github.com/tx7do/go-utils/aggregator"
	"github.com/tx7do/go-utils/timeutil"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/emptypb"

	"go-wind-shop/app/core/service/internal/data"

	internalMessageV1 "go-wind-shop/api/gen/go/internal_message/service/v1"
	identityV1 "go-wind-shop/api/gen/go/identity/service/v1"
	notificationV1 "go-wind-shop/api/gen/go/notification/service/v1"
)

type InternalMessageService struct {
	internalMessageV1.UnimplementedInternalMessageServiceServer

	log *logx.Helper

	internalMessageRepo          *data.InternalMessageRepo
	internalMessageCategoryRepo  *data.InternalMessageCategoryRepo
	internalMessageRecipientRepo *data.InternalMessageRecipientRepo

	userRepo data.UserRepo

	// notifier 通知域投递缝（Notifier 接口，由 NotificationService 实现）。
	// 构造期留 nil，装配期由 wiring 在两个 service 都就绪后 RegisterNotifier 注入；
	// 未注入时站内信仅落收件行，不出站到渠道路由。
	notifier Notifier
}

// RegisterNotifier 注入通知域投递缝。
func (s *InternalMessageService) RegisterNotifier(notifier Notifier) {
	s.notifier = notifier
}

func NewInternalMessageService(
	ctx *bootstrap.Context,
	internalMessageRepo *data.InternalMessageRepo,
	internalMessageCategoryRepo *data.InternalMessageCategoryRepo,
	internalMessageRecipientRepo *data.InternalMessageRecipientRepo,
	userRepo data.UserRepo,
) *InternalMessageService {
	return &InternalMessageService{
		log:                          logx.New(ctx.NewLoggerHelper("internal-message/service/core-service")),
		internalMessageRepo:          internalMessageRepo,
		internalMessageCategoryRepo:  internalMessageCategoryRepo,
		internalMessageRecipientRepo: internalMessageRecipientRepo,
		userRepo:                     userRepo,
	}
}

func (s *InternalMessageService) extractRelationIDs(
	messages []*internalMessageV1.InternalMessage,
	categorySet aggregator.ResourceMap[uint32, *internalMessageV1.InternalMessageCategory],
) {
	for _, p := range messages {
		if p.GetCategoryId() > 0 {
			categorySet[p.GetCategoryId()] = nil
		}
	}
}

func (s *InternalMessageService) fetchRelationInfo(
	ctx context.Context,
	categorySet aggregator.ResourceMap[uint32, *internalMessageV1.InternalMessageCategory],
) error {
	if len(categorySet) > 0 {
		categoryIds := make([]uint32, 0, len(categorySet))
		for id := range categorySet {
			categoryIds = append(categoryIds, id)
		}

		categories, err := s.internalMessageCategoryRepo.ListCategoriesByIds(ctx, categoryIds)
		if err != nil {
			s.log.Errorf("query internal message category err: %v", err)
			return err
		}

		for _, g := range categories {
			categorySet[g.GetId()] = g
		}
	}

	return nil
}

func (s *InternalMessageService) bindRelations(
	messages []*internalMessageV1.InternalMessage,
	categorySet aggregator.ResourceMap[uint32, *internalMessageV1.InternalMessageCategory],
) {
	aggregator.Populate(
		messages,
		categorySet,
		func(ou *internalMessageV1.InternalMessage) uint32 { return ou.GetCategoryId() },
		func(ou *internalMessageV1.InternalMessage, c *internalMessageV1.InternalMessageCategory) {
			ou.CategoryName = c.Name
		},
	)
}

func (s *InternalMessageService) enrichRelations(ctx context.Context, messages []*internalMessageV1.InternalMessage) error {
	var categorySet = make(aggregator.ResourceMap[uint32, *internalMessageV1.InternalMessageCategory])
	s.extractRelationIDs(messages, categorySet)
	if err := s.fetchRelationInfo(ctx, categorySet); err != nil {
		return err
	}
	s.bindRelations(messages, categorySet)
	return nil
}

func (s *InternalMessageService) ListMessage(ctx context.Context, req *paginationV1.PagingRequest) (*internalMessageV1.ListInternalMessageResponse, error) {
	resp, err := s.internalMessageRepo.List(ctx, req)
	if err != nil {
		return nil, err
	}

	_ = s.enrichRelations(ctx, resp.Items)

	return resp, nil
}

func (s *InternalMessageService) GetMessage(ctx context.Context, req *internalMessageV1.GetInternalMessageRequest) (*internalMessageV1.InternalMessage, error) {
	resp, err := s.internalMessageRepo.Get(ctx, req)
	if err != nil {
		return nil, err
	}

	fakeItems := []*internalMessageV1.InternalMessage{resp}
	_ = s.enrichRelations(ctx, fakeItems)

	return resp, nil
}

func (s *InternalMessageService) CreateMessage(ctx context.Context, req *internalMessageV1.CreateInternalMessageRequest) (*internalMessageV1.InternalMessage, error) {
	if req.Data == nil {
		return nil, internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	var created *internalMessageV1.InternalMessage
	var err error
	if created, err = s.internalMessageRepo.Create(ctx, req); err != nil {
		return nil, err
	}

	return created, nil
}

func (s *InternalMessageService) UpdateMessage(ctx context.Context, req *internalMessageV1.UpdateInternalMessageRequest) (*emptypb.Empty, error) {
	if req.Data == nil {
		return nil, internalMessageV1.ErrorBadRequest("invalid parameter")
	}

	if err := s.internalMessageRepo.Update(ctx, req); err != nil {
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *InternalMessageService) DeleteMessage(ctx context.Context, req *internalMessageV1.DeleteInternalMessageRequest) (*emptypb.Empty, error) {
	if err := s.internalMessageRepo.Delete(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// RevokeMessage 撤销某条消息
func (s *InternalMessageService) RevokeMessage(ctx context.Context, req *internalMessageV1.RevokeMessageRequest) (*emptypb.Empty, error) {
	var err error
	if err = s.internalMessageRepo.Delete(ctx, req.GetMessageId()); err != nil {
		s.log.Errorf("delete internal message failed: [%d]", req.GetMessageId())
	}

	if err = s.internalMessageRecipientRepo.RevokeMessage(ctx, req); err != nil {
		s.log.Errorf("delete internal message inbox failed: [%d][%d]", req.GetMessageId(), req.GetUserId())
	}

	return &emptypb.Empty{}, err
}

// SendMessage 发送消息
func (s *InternalMessageService) SendMessage(ctx context.Context, req *internalMessageV1.SendMessageRequest) (*internalMessageV1.SendMessageResponse, error) {
	now := time.Now()

	var err error
	var msg *internalMessageV1.InternalMessage
	if msg, err = s.internalMessageRepo.Create(ctx, &internalMessageV1.CreateInternalMessageRequest{
		Data: &internalMessageV1.InternalMessage{
			Title:      req.Title,
			Content:    trans.Ptr(req.GetContent()),
			Status:     trans.Ptr(internalMessageV1.InternalMessage_PUBLISHED),
			Type:       trans.Ptr(req.GetType()),
			CategoryId: req.CategoryId,
			CreatedBy:  trans.Ptr(req.GetSendUserId()),
			CreatedAt:  timeutil.TimeToTimestamppb(&now),
		},
	}); err != nil {
		s.log.Errorf("create internal message failed: %s", err)
		return nil, err
	}

	if req.GetTargetAll() {
		users, err := s.userRepo.List(ctx, &paginationV1.PagingRequest{NoPaging: trans.Ptr(true)})
		if err != nil {
			s.log.Errorf("send message failed, list users failed, %s", err)
		} else {
			for _, user := range users.Items {
				_ = s.sendNotification(ctx, msg.GetId(), user.GetId(), req.GetSendUserId(), &now, msg.GetTitle(), msg.GetContent())
			}
		}
	} else {
		if req.RecipientUserId != nil {
			_ = s.sendNotification(ctx, msg.GetId(), req.GetRecipientUserId(), req.GetSendUserId(), &now, msg.GetTitle(), msg.GetContent())
		} else {
			if len(req.TargetUserIds) != 0 {
				for _, uid := range req.TargetUserIds {
					_ = s.sendNotification(ctx, msg.GetId(), uid, req.GetSendUserId(), &now, msg.GetTitle(), msg.GetContent())
				}
			}
		}
	}

	return &internalMessageV1.SendMessageResponse{
		MessageId: msg.GetId(),
	}, nil
}

// sendNotification 向客户端发送通知消息
func (s *InternalMessageService) sendNotification(ctx context.Context, messageId uint32, recipientUserId uint32, senderUserId uint32, now *time.Time, title, content string) error {
	recipient := &internalMessageV1.InternalMessageRecipient{
		MessageId:       trans.Ptr(messageId),
		RecipientUserId: trans.Ptr(recipientUserId),
		Status:          trans.Ptr(internalMessageV1.InternalMessageRecipient_SENT),
		CreatedBy:       trans.Ptr(senderUserId),
		CreatedAt:       timeutil.TimeToTimestamppb(now),
		Title:           trans.Ptr(title),
		Content:         trans.Ptr(content),
	}

	var err error
	var entity *internalMessageV1.InternalMessageRecipient
	if entity, err = s.internalMessageRecipientRepo.Create(ctx, recipient); err != nil {
		s.log.Errorf("send message failed, send to user failed, %s", err)
		return err
	}
	recipient.Id = entity.Id

	// 经通知域投递缝出站（INTERNAL_MESSAGE 事件）：渠道路由决定走邮箱还是仅站内行。
	// 单收件人投递失败不中断发送（收件行已落，出站失败在台账可查）；未接线时静默跳过。
	if s.notifier != nil {
		if _, dErr := s.deliverViaNotifier(ctx, &notificationV1.SendDirectNotificationRequest{
			EventType:       notificationV1.EventType_INTERNAL_MESSAGE,
			Target:          s.notificationTargetFor(ctx, recipientUserId),
			Title:           title,
			Content:         content,
			RecipientUserId: trans.Ptr(recipientUserId),
			RelatedId:       trans.Ptr(messageId),
		}); dErr != nil {
			s.log.Errorf("send message: notifier deliver to user [%d] failed: %v", recipientUserId, dErr)
		}
	}

	return nil
}

// deliverViaNotifier 经通知域投递缝发一条通知；未接线时返回 nil（静默跳过出站）。
func (s *InternalMessageService) deliverViaNotifier(ctx context.Context, req *notificationV1.SendDirectNotificationRequest) (*notificationV1.SendNotificationResponse, error) {
	if s.notifier == nil {
		return nil, nil
	}
	return s.notifier.SendDirect(ctx, req)
}

// notificationTargetFor 返回收件人的 SendDirect Target：
// INTERNAL_MESSAGE 事件的 EMAIL 路由需要邮箱地址，而本链路持有的是收件人
// uid——这里解析 sys_users.email；未配置邮箱回落 uid 字符串（既有行为，
// EMAIL 渠道投递以 FAILED 呈现而非误发）。INTERNAL 渠道不消费 Target。
func (s *InternalMessageService) notificationTargetFor(ctx context.Context, uid uint32) string {
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
