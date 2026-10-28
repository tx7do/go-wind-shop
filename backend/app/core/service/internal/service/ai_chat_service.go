package service

import (
	"go-wind-shop/pkg/logx"
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/grpc"
	entCrud "github.com/tx7do/go-crud/entgo"
	"github.com/tx7do/go-crud/viewer"

	aiV1 "go-wind-shop/api/gen/go/ai/service/v1"
	"go-wind-shop/app/core/service/internal/data"
	"go-wind-shop/app/core/service/internal/data/ent"
	"go-wind-shop/app/core/service/internal/data/ent/aimessage"
	entmw "go-wind-shop/pkg/middleware/ent"
	"go-wind-shop/pkg/constants"
)

const (
	// chatContextMaxMessages LLM 上下文最多携带的历史消息条数。
	chatContextMaxMessages = 20

	// conversationTitleMaxRunes 新会话默认标题截断长度（按字符，中文友好）。
	conversationTitleMaxRunes = 30

	// chatCallTimeout 单轮流式对话的兜底时长：覆盖整轮（历史拼装 + 模型生成 + 落库）。
	chatCallTimeout = 5 * time.Minute

	// generateCallTimeout Generate 单次非流式调用的兜底时长。
	generateCallTimeout = 90 * time.Second

	// channelApp C端导购渠道标识（BFF 注入 req.Channel；其余值按 admin 处理）。
	channelApp = "app"
)

// AiChatService AI 对话：core 侧 server-streaming——逐帧吐 chunk/tool，
// done 帧携带完整回复；SSE 推送由 admin BFF 消费该流后完成（见 i_ai_chat）。
type AiChatService struct {
	aiV1.UnimplementedAiChatServiceServer
	log *logx.Helper
	conversationRepo *data.AiConversationRepo
	messageRepo      *data.AiMessageRepo
	providerRepo     *data.AiProviderRepo
	usageLogRepo     *data.AiUsageLogRepo
	knowledgeRepo    *data.AiKnowledgeRepo
	configRepo       *data.ConfigRepo
	entClient        *entCrud.EntClient[*ent.Client]
}

func NewAiChatService(
	ctx *bootstrap.Context,
	conversationRepo *data.AiConversationRepo,
	messageRepo *data.AiMessageRepo,
	providerRepo *data.AiProviderRepo,
	usageLogRepo *data.AiUsageLogRepo,
	knowledgeRepo *data.AiKnowledgeRepo,
	configRepo *data.ConfigRepo,
	entClient *entCrud.EntClient[*ent.Client],
) *AiChatService {
	return &AiChatService{
		log:              logx.New(ctx.NewLoggerHelper("ai_chat/service/core-service")),
		conversationRepo: conversationRepo,
		messageRepo:      messageRepo,
		providerRepo:     providerRepo,
		usageLogRepo:     usageLogRepo,
		knowledgeRepo:    knowledgeRepo,
		configRepo:       configRepo,
		entClient:        entClient,
	}
}

// Chat 一轮对话：落 user 消息 → 流式调模型（chunk/tool 帧实时吐给 BFF）→
// 落 assistant 消息与用量流水 → done 帧返回完整回复。
func (s *AiChatService) Chat(req *aiV1.ChatRequest, stream grpc.ServerStreamingServer[aiV1.ChatStreamFrame]) error {
	ctx := stream.Context()
	// unary 的 ent viewer 中间件不覆盖流式 RPC，这里自行注入 viewer
	// （BFF 转发的 operator 元数据 → UserViewer；异常回落 SystemViewer）。
	ctx = entmw.InjectViewerForStream(ctx)

	// BFF 注入的 operator 字段优先（流式 RPC 的元数据转发不可靠，见 proto 注释）；
	// 缺失时回落 viewer 上下文。
	var userId, tenantId uint32
	if uid := req.GetOperatorUserId(); uid > 0 {
		userId = uid
		tenantId = req.GetOperatorTenantId()
	} else {
		if vc, exist := viewer.FromContext(ctx); exist && vc != nil {
			tenantId = uint32(vc.TenantID())
		}
		if uid, hasUser := viewerUserIDFromContext(ctx); hasUser {
			userId = uid
		}
	}

	content := req.GetContent()
	if content == "" {
		return aiV1.ErrorBadRequest("content is required")
	}

	// 渠道策略：app（C端导购）生效每日 token 限额与默认知识库绑定；admin 不受限。
	isAppChannel := req.GetChannel() == channelApp
	if isAppChannel {
		if err := s.enforceDailyTokenLimit(ctx, userId); err != nil {
			return err
		}
	}
	knowledgeBaseId := req.GetKnowledgeBaseId()
	if knowledgeBaseId == 0 && isAppChannel && s.configRepo != nil {
		knowledgeBaseId = uint32(s.configRepo.GetConfigInt(ctx, constants.ConfigKeyAiAssistantKnowledgeBaseId, 0))
	}

	// 1. 解析提供商（显式指定或默认）
	provider, err := s.resolveProvider(ctx, req.GetProviderId())
	if err != nil {
		return err
	}

	// 2. 会话归属：复用或新建
	conversation, err := s.resolveConversation(ctx, userId, tenantId, req.GetConversationId(), provider, content)
	if err != nil {
		return err
	}

	// 3. 落 user 消息
	userEntity, err := s.messageRepo.Create(ctx, &aiV1.AiMessage{
		ConversationId: &conversation.ID,
		Role:           aiV1.AiRole_USER.Enum(),
		Content:        &content,
		UserId:         &userId,
		TenantId:       trans.Ptr(tenantId),
		CreatedBy:      &userId,
	})
	if err != nil {
		return err
	}

	// 4. LLM 上下文：system（商品上下文 + 知识库检索注入 RAG + 提供商 system prompt）+ 历史 + 本轮。
	messages := make([]openai.ChatCompletionMessage, 0, chatContextMaxMessages+2)
	systemParts := make([]string, 0, 3)
	if pid := req.GetContextProductId(); pid > 0 {
		if summary := s.productContextSummary(ctx, pid); summary != "" {
			systemParts = append(systemParts, "用户当前正在浏览以下商品，回答应围绕该商品展开；用户问题与商品无关时正常回答：\n"+summary)
		}
	}
	if knowledgeBaseId > 0 && s.knowledgeRepo != nil {
		hits, searchErr := searchKnowledgeBase(ctx, s.knowledgeRepo, s.providerRepo, s.usageLogRepo, s.log, userId, tenantId, knowledgeBaseId, content, ragTopK)
		if searchErr != nil {
			// 检索失败不阻断对话：降级为无知识库上下文
			s.log.Errorf("knowledge search failed, degrade to plain chat: base=%d: %v", knowledgeBaseId, searchErr)
		} else if len(hits) > 0 {
			var kb strings.Builder
			kb.WriteString("请优先依据以下知识库片段回答用户问题，片段无关时可忽略：\n")
			for i, hit := range hits {
				kb.WriteString(fmt.Sprintf("[片段%d] %s\n", i+1, hit.Content))
			}
			systemParts = append(systemParts, kb.String())
		}
	}
	if provider.SystemPrompt != nil && *provider.SystemPrompt != "" {
		systemParts = append(systemParts, *provider.SystemPrompt)
	}
	if len(systemParts) > 0 {
		messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleSystem, Content: strings.Join(systemParts, "\n\n")})
	}

	// 多取一条：第 3 步刚落库的本轮消息必是表内最新一行，会被这次历史查询带出，
	// 循环里按 ID 剔除后恰好剩 N-1 条历史；本轮消息在下方显式 append，不剔就会重复计两次。
	history, err := s.messageRepo.ListRecentByConversation(ctx, conversation.ID, chatContextMaxMessages)
	if err != nil {
		return err
	}
	for _, m := range history {
		if m.Role == nil || m.ID == userEntity.ID {
			continue
		}
		// ent 枚举是大写（USER/ASSISTANT/SYSTEM），OpenAI 协议要求小写 role——
		// 真实厂商（DeepSeek 等）会 422 拒绝大写值，必须映射。
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    aiRoleToOpenAI(*m.Role),
			Content: ptrStr(m.Content),
		})
	}
	messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: content})

	// 5. 客户端 + Function Call 多轮循环：模型请求工具 → 本地执行 → 结果回传，
	//    直到给出文本答案（最终答案轮才经 chunk 帧吐给前端）。
	client, err := newOpenAIClientForProvider(provider)
	if err != nil {
		return err
	}

	streamCtx, cancel := context.WithTimeout(ctx, chatCallTimeout)
	defer cancel()

	var seq uint32
	var usage openai.Usage
	start := time.Now()

	runner := &aiToolLoop{
		client:    client,
		model:     ptrStrOr(provider.ModelName, "gpt-4o-mini"),
		tools:     append(aiBuiltinToolDefs(), shopToolDefs()...),
		exec: func(tctx context.Context, name, args string) (string, error) {
			if out, handled, err := s.execShopTool(tctx, userId, tenantId, name, args); handled {
				return out, err
			}
			return execAiTool(tctx, name, args)
		},
		maxRounds: aiToolMaxRounds,
		onDelta: func(delta string) {
			_ = stream.Send(&aiV1.ChatStreamFrame{Frame: &aiV1.ChatStreamFrame_Chunk{
				Chunk: &aiV1.ChatChunkEvent{
					ConversationId: conversation.ID,
					Seq:            seq,
					Delta:          delta,
				},
			}})
			seq++
		},
		onToolCall: func(name, arguments, result string) {
			_ = stream.Send(&aiV1.ChatStreamFrame{Frame: &aiV1.ChatStreamFrame_Tool{
				Tool: &aiV1.ChatToolEvent{
					ConversationId: conversation.ID,
					Name:           name,
					Arguments:      arguments,
					Result:         result,
				},
			}})
		},
	}
	fullText, usage, toolErr := runner.run(streamCtx, messages)
	if toolErr != nil {
		s.log.Errorf("ai chat stream failed: conversation=%d: %v", conversation.ID, toolErr)
		return aiV1.ErrorInternalServerError("ai chat failed: %v", toolErr)
	}
	durationMs := uint32(time.Since(start).Milliseconds())

	// 6. 落 assistant 消息与用量流水
	assistantContent := fullText
	modelName := ptrStrOr(provider.ModelName, "")
	promptTokens := uint32(usage.PromptTokens)
	completionTokens := uint32(usage.CompletionTokens)
	totalTokens := uint32(usage.TotalTokens)

	assistantEntity, err := s.messageRepo.Create(ctx, &aiV1.AiMessage{
		ConversationId:   &conversation.ID,
		Role:             aiV1.AiRole_ASSISTANT.Enum(),
		Content:          &assistantContent,
		ModelName:        &modelName,
		PromptTokens:     &promptTokens,
		CompletionTokens: &completionTokens,
		DurationMs:       &durationMs,
		UserId:           &userId,
		TenantId:         trans.Ptr(tenantId),
		CreatedBy:        &userId,
	})
	if err != nil {
		return err
	}

	if err = s.usageLogRepo.Create(ctx, &data.UsageRecord{
		ProviderID:       provider.ID,
		UserID:           userId,
		TenantID:         tenantId,
		ModelName:        modelName,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
		DurationMs:       durationMs,
	}); err != nil {
		// 记账失败不阻断对话返回：流水缺失只影响配额统计精度，记日志跟进。
		s.log.Errorf("record ai usage log failed: conversation=%d: %v", conversation.ID, err)
	}

	if err = s.conversationRepo.TouchLastMessage(ctx, conversation.ID); err != nil {
		s.log.Errorf("touch conversation last message failed: conversation=%d: %v", conversation.ID, err)
	}

	conversationDto := s.conversationRepo.ToDTO(conversation)
	messageDto := s.messageRepo.ToDTO(assistantEntity)
	if conversationDto == nil || messageDto == nil {
		return aiV1.ErrorInternalServerError("map ai chat result failed")
	}

	return stream.Send(&aiV1.ChatStreamFrame{Frame: &aiV1.ChatStreamFrame_Done{
		Done: &aiV1.ChatResponse{
			Conversation: conversationDto,
			Message:      messageDto,
		},
	}})
}

// resolveProvider 解析本次对话使用的提供商。
func (s *AiChatService) resolveProvider(ctx context.Context, providerId uint32) (*ent.AiProvider, error) {
	if providerId > 0 {
		entity, err := s.providerRepo.GetEntityByID(ctx, providerId)
		if err != nil {
			return nil, err
		}
		if entity.IsEnabled == nil || !*entity.IsEnabled {
			return nil, aiV1.ErrorBadRequest("ai provider is disabled")
		}
		return entity, nil
	}
	return s.providerRepo.GetEnabledDefault(ctx)
}

// enforceDailyTokenLimit C端单用户每日 token 限额闸门
// （sys_configs: ai.chat.daily_token_limit，<=0 不限）。聚合当日零点以来该用户
// 全部 AI 调用（对话 + Generate）的用量流水；聚合失败 fail-open 放行（只影响
// 限额精度，不应阻断对话）。
func (s *AiChatService) enforceDailyTokenLimit(ctx context.Context, userId uint32) error {
	if s.configRepo == nil || s.usageLogRepo == nil || userId == 0 {
		return nil
	}
	limit := s.configRepo.GetConfigInt(ctx, constants.ConfigKeyAiChatDailyTokenLimit, constants.DefaultAiChatDailyTokenLimit)
	if limit <= 0 {
		return nil
	}
	now := time.Now()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	used, err := s.usageLogRepo.UserTokensSince(ctx, userId, dayStart)
	if err != nil {
		s.log.Errorf("sum daily usage failed, fail-open: user=%d: %v", userId, err)
		return nil
	}
	if used >= uint64(limit) {
		return aiV1.ErrorForbidden("今日 AI 助手使用额度已用完，请明天再来")
	}
	return nil
}

// productContextSummary 拉取商品上下文摘要（名称/状态 + 在售 SKU 与价格），
// 注入 system prompt 让导购围绕当前浏览商品作答。查询失败返回空串——
// 降级为无商品上下文，不阻断对话。
func (s *AiChatService) productContextSummary(ctx context.Context, productId uint32) string {
	if s.entClient == nil {
		return ""
	}
	db := s.entClient.DB()

	var name, shortDesc string
	row := db.QueryRowContext(ctx, `
		SELECT COALESCE(t.name, ''), COALESCE(t.short_description, '')
		FROM mall_products p
		LEFT JOIN mall_product_translations t ON t.product_id = p.id AND t.language_code = 'zh-CN'
		WHERE p.id = $1`, productId)
	if err := row.Scan(&name, &shortDesc); err != nil {
		if err != sql.ErrNoRows {
			s.log.Errorf("product context query failed: product=%d: %v", productId, err)
		}
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("商品ID：%d；名称：%s", productId, name))
	if shortDesc != "" {
		sb.WriteString("；简介：" + shortDesc)
	}
	skuTable, err := s.queryRowsAsString(ctx, db, `
		SELECT s.sku_code, s.stock_qty AS stock,
			COALESCE(pr.amount, '0') AS price_fen,
			ROUND(COALESCE(pr.amount, '0')::numeric/100, 2) AS price_yuan
		FROM mall_skus s
		LEFT JOIN mall_sku_prices pr ON pr.sku_id = s.id AND pr.currency = 'CNY'
		WHERE s.product_id = $1`, productId)
	if err == nil && skuTable != "" {
		sb.WriteString("\n在售规格（价格单位：元）：\n" + skuTable)
	}
	return sb.String()
}

// Generate 通用单轮文本生成（非会话、不落库）：BFF 层组装 prompt，core 负责
// 解析默认提供商、调用模型、记录用量流水（与对话共用同一套记账）。
func (s *AiChatService) Generate(ctx context.Context, req *aiV1.GenerateRequest) (*aiV1.GenerateResponse, error) {
	// BFF 注入的 operator 字段优先；缺失时回落 viewer 上下文。
	var userId, tenantId uint32
	if uid := req.GetOperatorUserId(); uid > 0 {
		userId = uid
		tenantId = req.GetOperatorTenantId()
	} else if vc, exist := viewer.FromContext(ctx); exist && vc != nil {
		tenantId = uint32(vc.TenantID())
		if uid, hasUser := viewerUserIDFromContext(ctx); hasUser {
			userId = uid
		}
	}

	prompt := req.GetPrompt()
	if prompt == "" {
		return nil, aiV1.ErrorBadRequest("prompt is required")
	}

	provider, err := s.resolveProvider(ctx, 0)
	if err != nil {
		return nil, err
	}
	client, err := newOpenAIClientForProvider(provider)
	if err != nil {
		return nil, err
	}

	messages := make([]openai.ChatCompletionMessage, 0, 2)
	if sys := req.GetSystem(); sys != "" {
		messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleSystem, Content: sys})
	}
	messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: prompt})

	completionReq := openai.ChatCompletionRequest{
		Model:    ptrStrOr(provider.ModelName, "gpt-4o-mini"),
		Messages: messages,
	}
	if mt := req.GetMaxTokens(); mt > 0 {
		completionReq.MaxTokens = int(mt)
	}

	gctx, cancel := context.WithTimeout(ctx, generateCallTimeout)
	defer cancel()

	resp, err := client.CreateChatCompletion(gctx, completionReq)
	if err != nil {
		s.log.Errorf("ai generate failed: user=%d: %v", userId, err)
		return nil, aiV1.ErrorInternalServerError("ai generate failed: %v", err)
	}

	// 用量流水（记账失败不阻断返回，与对话同口径）。
	modelName := completionReq.Model
	if err = s.usageLogRepo.Create(ctx, &data.UsageRecord{
		ProviderID:       provider.ID,
		UserID:           userId,
		TenantID:         tenantId,
		ModelName:        modelName,
		PromptTokens:     uint32(resp.Usage.PromptTokens),
		CompletionTokens: uint32(resp.Usage.CompletionTokens),
		TotalTokens:      uint32(resp.Usage.TotalTokens),
		DurationMs:       0,
	}); err != nil {
		s.log.Errorf("record ai generate usage failed: user=%d: %v", userId, err)
	}

	text := ""
	if len(resp.Choices) > 0 {
		text = resp.Choices[0].Message.Content
	}
	return &aiV1.GenerateResponse{
		Text:        text,
		ModelName:   modelName,
		TotalTokens: uint32(resp.Usage.TotalTokens),
	}, nil
}

// resolveConversation 会话复用或新建（新建时标题取首条消息摘要）。
func (s *AiChatService) resolveConversation(ctx context.Context, userId, tenantId, conversationId uint32, provider *ent.AiProvider, content string) (*ent.AiConversation, error) {
	if conversationId > 0 {
		entity, err := s.conversationRepo.GetEntityByID(ctx, conversationId)
		if err != nil {
			return nil, err
		}
		if entity.UserID == nil || *entity.UserID != userId {
			return nil, aiV1.ErrorNotFound("ai conversation not found")
		}
		return entity, nil
	}

	title := truncateRunes(content, conversationTitleMaxRunes)
	created, err := s.conversationRepo.Create(ctx, &aiV1.AiConversation{
		Title:      &title,
		ProviderId: &provider.ID,
		UserId:     &userId,
		TenantId:   trans.Ptr(tenantId),
		CreatedBy:  &userId,
	})
	if err != nil {
		return nil, err
	}
	return created, nil
}

// aiRoleToOpenAI ent 大写枚举 → OpenAI 小写 role。
func aiRoleToOpenAI(role aimessage.Role) string {
	switch role {
	case aimessage.RoleUSER:
		return openai.ChatMessageRoleUser
	case aimessage.RoleASSISTANT:
		return openai.ChatMessageRoleAssistant
	case aimessage.RoleSYSTEM:
		return openai.ChatMessageRoleSystem
	default:
		return openai.ChatMessageRoleUser
	}
}

// truncateRunes 按字符截断（中文友好）。
func truncateRunes(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max])
}

func ptrStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
