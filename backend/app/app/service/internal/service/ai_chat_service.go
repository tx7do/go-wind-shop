package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

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

// ── 搜索查询 AI 改写 ────────────────────────────────────────────────

const (
	// searchRewriteMaxChars 输入长度钳制（匿名端点防滥用兜底）。
	searchRewriteMaxChars = 100

	// searchRewriteMinLen 输入低于该长度视为已是高效关键词，不做改写（省 LLM 调用）。
	searchRewriteMinLen = 6

	// searchRewriteTimeout 单次改写兜底时长。搜索路径上用户在等，但思考型模型
	// （qwen3/r1 类）思考段就要十几秒，给足余量；超时由前端静默降级原词。
	searchRewriteTimeout = 45 * time.Second
)

// RewriteSearchQuery 搜索查询 AI 改写：自然语言 → 检索关键词组。
// 任何失败都降级为 rewritten=false（前端沿用原词，不阻断搜索）。
func (s *AiChatService) RewriteSearchQuery(ctx context.Context, req *appV1.RewriteSearchQueryRequest) (*appV1.RewriteSearchQueryResponse, error) {
	resp := &appV1.RewriteSearchQueryResponse{Rewritten: false}

	query := strings.TrimSpace(req.GetQuery())
	if query == "" {
		return nil, appV1.ErrorBadRequest("query is required")
	}
	if len(query) > searchRewriteMaxChars {
		query = query[:searchRewriteMaxChars]
	}
	// 短输入（已是关键词形态）或纯 ASCII 单词，改写无增益，直接原样返回。
	if len([]rune(query)) < searchRewriteMinLen && !strings.Contains(query, " ") {
		return resp, nil
	}

	gctx, cancel := context.WithTimeout(ctx, searchRewriteTimeout)
	defer cancel()

	gen, err := s.aiChatServiceClient.Generate(gctx, &aiV1.GenerateRequest{
		System: trans.Ptr("你是电商搜索查询优化器。把用户的搜索请求改写成 2-4 个中文检索关键词" +
			"（提取品类、核心属性、价格意图；丢弃语气词）。只输出 JSON 数组，" +
			"格式：{\"keywords\": [\"关键词1\", \"关键词2\"]}，不要任何解释或代码块标记。"),
		Prompt:    trans.Ptr(query),
		// MaxTokens 需覆盖思考型模型的思考段 + 正文（同评论摘要的经验值）
		MaxTokens: trans.Ptr(uint32(800)),
	})
	if err != nil {
		s.log.Warnf("search rewrite llm call failed (degrade to original query): %v", err)
		return resp, nil
	}

	text := strings.TrimSpace(gen.GetText())
	if begin := strings.Index(text, "{"); begin >= 0 {
		if end := strings.LastIndex(text, "}"); end > begin {
			text = text[begin : end+1]
		}
	}
	var parsed struct {
		Keywords []string `json:"keywords"`
	}
	if err := json.Unmarshal([]byte(text), &parsed); err != nil || len(parsed.Keywords) == 0 {
		s.log.Warnf("search rewrite unparsable llm output (degrade): %q", text)
		return resp, nil
	}

	// 清洗：去空串/去重/每个再钳 30 字符，最多取 4 个
	seen := make(map[string]bool, len(parsed.Keywords))
	keywords := make([]string, 0, 4)
	for _, k := range parsed.Keywords {
		k = strings.TrimSpace(k)
		if k == "" || seen[k] {
			continue
		}
		if len([]rune(k)) > 30 {
			k = string([]rune(k)[:30])
		}
		seen[k] = true
		keywords = append(keywords, k)
		if len(keywords) >= 4 {
			break
		}
	}
	if len(keywords) == 0 {
		return resp, nil
	}
	resp.Keywords = keywords
	resp.Rewritten = true
	return resp, nil
}
