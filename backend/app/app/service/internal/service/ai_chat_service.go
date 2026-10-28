package service

import (
	"context"
	"io"
	"net/http"

	"go-wind-shop/pkg/logx"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"github.com/tx7do/kratos-transport/transport/sse"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	aiV1 "go-wind-shop/api/gen/go/ai/service/v1"
	appV1 "go-wind-shop/api/gen/go/app/service/v1"
	authenticationV1 "go-wind-shop/api/gen/go/authentication/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

// AI chat SSE 事件名（与 admin 侧逐字对齐；C端未订阅 SSE 时事件无人消费）。
const (
	appSseEventAIChatChunk = "ai_chat_chunk"
	appSseEventAIChatTool  = "ai_chat_tool"

	// appChannel C端渠道标识（core 据此启用每日限额/默认知识库等 C端策略）。
	appChannel = "app"
)

// AiChatPublisher SSE 推送接口（app SSE 网关未接线时为 noop——对话仍完整可用，
// done 帧作为同步 POST 响应返回）。
type AiChatPublisher interface {
	Publish(ctx context.Context, streamId sse.StreamID, event *sse.Event)
}

// noopAiChatPublisher 占位实现。
type noopAiChatPublisher struct{}

func (noopAiChatPublisher) Publish(_ context.Context, _ sse.StreamID, _ *sse.Event) {}

// AiChatService AI 导购对话（app BFF 编排）：消费 core 的 ChatStreamFrame 流——
// chunk/tool 帧转 protojson 经 SSE 推给发起用户（逐 access-token 流隔离，
// 与 admin 同模式；未订阅时 noop 静默丢弃），done 帧作为同步 POST 响应返回。
type AiChatService struct {
	appV1.AiChatServiceHTTPServer

	log *logx.Helper

	aiChatServiceClient         aiV1.AiChatServiceClient
	authenticationServiceClient authenticationV1.AuthenticationServiceClient

	aiChatPublisher AiChatPublisher
}

func NewAiChatService(
	ctx *bootstrap.Context,
	aiChatServiceClient aiV1.AiChatServiceClient,
	authenticationServiceClient authenticationV1.AuthenticationServiceClient,
) *AiChatService {
	return &AiChatService{
		log:                         logx.New(ctx.NewLoggerHelper("ai_chat/app-service")),
		aiChatServiceClient:         aiChatServiceClient,
		authenticationServiceClient: authenticationServiceClient,
		aiChatPublisher:             noopAiChatPublisher{},
	}
}

// RegisterAiChatPublisher app SSE 网关启动时注入真身（可选能力）。
func (s *AiChatService) RegisterAiChatPublisher(publisher AiChatPublisher) {
	if publisher == nil {
		return
	}
	s.aiChatPublisher = publisher
}

// HandleAuthorize SSE 连接鉴权：校验 accessToken 归属与有效性
// （与 admin InternalMessageService.HandleAuthorize 同模式，clientType=app）。
func (s *AiChatService) HandleAuthorize(_ *http.Request, token string) error {
	resp, err := s.authenticationServiceClient.ValidateToken(context.Background(), &authenticationV1.ValidateTokenRequest{
		ClientType:    authenticationV1.ClientType_app,
		Token:         token,
		TokenCategory: authenticationV1.TokenCategory_ACCESS,
	})
	if err != nil {
		s.log.Errorf("sse token authentication failed: %s", err.Error())
		return err
	}
	if resp.GetIsBlocked() {
		return authenticationV1.ErrorForbidden("token is blocked")
	}
	if !resp.GetIsValid() {
		return authenticationV1.ErrorUnauthorized("invalid token")
	}
	return nil
}

// Chat 一轮导购对话。
func (s *AiChatService) Chat(ctx context.Context, req *aiV1.ChatRequest) (*aiV1.ChatResponse, error) {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 发起用户的全部 SSE 流（同一账号多标签页各持一个 accessToken 流）；
	// 拿不到就降级为无 chunk 对话（同步响应兜底）。
	streamIds, err := s.authenticationServiceClient.GetAccessTokens(ctx, &authenticationV1.GetAccessTokensRequest{
		UserId:     operator.UserId,
		ClientType: authenticationV1.ClientType_app,
	})
	if err != nil {
		s.log.Warnf("get user access tokens failed, degrade to no-chunk chat: %v", err)
		streamIds = nil
	}

	// 流式 gRPC 调用不走 unary 客户端中间件，operator 元数据不会自动随请求转发，
	// 显式把操作者身份写进请求体（core 信任内部跳，前端传入值在此被覆盖）——
	// query_my_orders 等买家工具以此取 user_id；渠道固定 app（C端策略开关依据）。
	req.OperatorUserId = trans.Ptr(operator.UserId)
	req.OperatorTenantId = trans.Ptr(operator.GetTenantId())
	req.Channel = trans.Ptr(appChannel)

	stream, err := s.aiChatServiceClient.Chat(ctx, req)
	if err != nil {
		return nil, err
	}

	for {
		frame, recvErr := stream.Recv()
		if recvErr != nil {
			if recvErr == io.EOF {
				return nil, aiV1.ErrorInternalServerError("ai chat stream ended without done frame")
			}
			return nil, recvErr
		}
		switch f := frame.Frame.(type) {
		case *aiV1.ChatStreamFrame_Chunk:
			s.publishToUserStreams(ctx, streamIds, appSseEventAIChatChunk, f.Chunk)
		case *aiV1.ChatStreamFrame_Tool:
			s.publishToUserStreams(ctx, streamIds, appSseEventAIChatTool, f.Tool)
		case *aiV1.ChatStreamFrame_Done:
			return f.Done, nil
		}
	}
}

// publishToUserStreams 向用户全部流推送一帧事件（protojson、camelCase 键）。
// 流 ID = accessToken，与 C端订阅的 `?stream=<token>` 一一对应，跨用户隔离。
func (s *AiChatService) publishToUserStreams(ctx context.Context, streamIds *authenticationV1.GetAccessTokensResponse, eventName string, msg proto.Message) {
	if streamIds == nil || len(streamIds.AccessTokens) == 0 {
		return
	}
	if _, ok := s.aiChatPublisher.(noopAiChatPublisher); ok {
		return
	}
	data, mErr := protojson.Marshal(msg)
	if mErr != nil {
		s.log.Errorf("marshal %s event failed: %v", eventName, mErr)
		return
	}
	for _, streamId := range streamIds.AccessTokens {
		s.aiChatPublisher.Publish(ctx, sse.StreamID(streamId), &sse.Event{
			Event: []byte(eventName),
			Data:  data,
		})
	}
}
