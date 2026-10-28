package service

import (
	"context"
	"io"

	"go-wind-shop/pkg/logx"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"github.com/tx7do/kratos-transport/transport/sse"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	aiV1 "go-wind-shop/api/gen/go/ai/service/v1"
	appV1 "go-wind-shop/api/gen/go/app/service/v1"

	"go-wind-shop/pkg/middleware/auth"
)

// AI chat SSE 事件名（与 admin 侧逐字对齐；C端未订阅 SSE 时事件无人消费）。
const (
	appSseEventAIChatChunk = "ai_chat_chunk"
	appSseEventAIChatTool  = "ai_chat_tool"
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
// chunk/tool 帧转 protojson 经 SSE 推给发起用户（C端 SSE 未接线时为 noop），
// done 帧作为同步 POST 响应返回。
type AiChatService struct {
	appV1.AiChatServiceHTTPServer

	log *logx.Helper

	aiChatServiceClient aiV1.AiChatServiceClient

	aiChatPublisher AiChatPublisher
}

func NewAiChatService(
	ctx *bootstrap.Context,
	aiChatServiceClient aiV1.AiChatServiceClient,
) *AiChatService {
	return &AiChatService{
		log:                 logx.New(ctx.NewLoggerHelper("ai_chat/app-service")),
		aiChatServiceClient: aiChatServiceClient,
		aiChatPublisher:     noopAiChatPublisher{},
	}
}

// RegisterAiChatPublisher app SSE 网关启动时注入真身（可选能力）。
func (s *AiChatService) RegisterAiChatPublisher(publisher AiChatPublisher) {
	if publisher == nil {
		return
	}
	s.aiChatPublisher = publisher
}

// Chat 一轮导购对话。
func (s *AiChatService) Chat(ctx context.Context, req *aiV1.ChatRequest) (*aiV1.ChatResponse, error) {
	operator, err := auth.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 流式 gRPC 调用不走 unary 客户端中间件，operator 元数据不会自动随请求转发，
	// 显式把操作者身份写进请求体（core 信任内部跳，前端传入值在此被覆盖）——
	// query_my_orders 等买家工具以此取 user_id。
	req.OperatorUserId = trans.Ptr(operator.UserId)
	req.OperatorTenantId = trans.Ptr(operator.GetTenantId())

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
			s.publishFrame(ctx, appSseEventAIChatChunk, f.Chunk)
		case *aiV1.ChatStreamFrame_Tool:
			s.publishFrame(ctx, appSseEventAIChatTool, f.Tool)
		case *aiV1.ChatStreamFrame_Done:
			return f.Done, nil
		}
	}
}

// publishFrame 向 SSE 网关广播一帧（protojson、camelCase 键）。
// C端 SSE 网关未接线时为 noop，静默丢弃。
func (s *AiChatService) publishFrame(ctx context.Context, eventName string, msg proto.Message) {
	if _, ok := s.aiChatPublisher.(noopAiChatPublisher); ok {
		return
	}
	data, mErr := protojson.Marshal(msg)
	if mErr != nil {
		s.log.Errorf("marshal %s event failed: %v", eventName, mErr)
		return
	}
	s.aiChatPublisher.Publish(ctx, sse.StreamID("*"), &sse.Event{
		Event: []byte(eventName),
		Data:  data,
	})
}
