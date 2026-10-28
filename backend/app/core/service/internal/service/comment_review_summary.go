package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai"

	commentV1 "go-wind-shop/api/gen/go/comment/service/v1"
	"go-wind-shop/app/core/service/internal/data"
)

// 商品评论 AI 摘要（C端商品页"大家说"区块）：
//   - 输入：商品最近 ListApprovedCommentsForProduct 条已批准评论（内容+星级）；
//   - 生成：默认 provider（与 chat/agent 同源）单轮 completion，输出 2-3 句
//     中文购物洞察；评论数不足 minReviewSummaryComments 时 available=false；
//   - 缓存：进程内 6 小时（评论天然低频更新，够用；多实例部署可换 redis）；
//   - 降级：provider 未配置 / LLM 失败 → available=false（前端不渲染区块），不报错。

const (
	// minReviewSummaryComments 参与摘要的最低评论条数（太少没有摘要价值）。
	minReviewSummaryComments = 3

	// reviewSummaryInputMax 参与摘要的最大评论条数（控制 prompt 长度）。
	reviewSummaryInputMax = 50

	// reviewSummaryTimeout 单次生成的兜底时长。
	reviewSummaryTimeout = 60 * time.Second

	// reviewSummaryCacheTTL 摘要缓存时长。
	reviewSummaryCacheTTL = 6 * time.Hour

	// reviewSummaryMaxChars 单条评论送入 prompt 的截断长度。
	reviewSummaryMaxChars = 200
)

// reviewSummaryCache 商品评论摘要的进程内缓存。
type reviewSummaryEntry struct {
	summary   *commentV1.ProductReviewSummary
	expiresAt time.Time
}

var reviewSummaryCache = struct {
	sync.Mutex
	m map[uint32]reviewSummaryEntry
}{m: make(map[uint32]reviewSummaryEntry)}

// GetProductReviewSummary 商品评论 AI 摘要（匿名可读，走 BFF 白名单）。
func (s *CommentService) GetProductReviewSummary(ctx context.Context, req *commentV1.GetProductReviewSummaryRequest) (*commentV1.ProductReviewSummary, error) {
	productId := req.GetProductId()
	if productId == 0 {
		return nil, commentV1.ErrorBadRequest("product_id is required")
	}

	// 命中缓存直接返回。
	now := time.Now()
	reviewSummaryCache.Lock()
	if hit, ok := reviewSummaryCache.m[productId]; ok && now.Before(hit.expiresAt) {
		reviewSummaryCache.Unlock()
		return hit.summary, nil
	}
	reviewSummaryCache.Unlock()

	summary := s.buildReviewSummary(ctx, productId)

	// 只有生成成功的结果才值得缓存；不可用（评论不足/LLM 失败）下次再试。
	if summary.GetAvailable() {
		reviewSummaryCache.Lock()
		reviewSummaryCache.m[productId] = reviewSummaryEntry{summary: summary, expiresAt: now.Add(reviewSummaryCacheTTL)}
		reviewSummaryCache.Unlock()
	}
	return summary, nil
}

// buildReviewSummary 拉评论 → 拼 prompt → 调 LLM；任何一步不满足即返回不可用。
func (s *CommentService) buildReviewSummary(ctx context.Context, productId uint32) *commentV1.ProductReviewSummary {
	unavailable := &commentV1.ProductReviewSummary{Available: false}

	if s.aiDeps == nil || s.aiDeps.ProviderRepo == nil || s.repo == nil {
		return unavailable
	}

	reviews, err := s.repo.ListApprovedCommentsForProduct(ctx, productId, reviewSummaryInputMax)
	if err != nil || len(reviews) < minReviewSummaryComments {
		// 不可用时也带参与条数，便于前端/运营诊断（评论是否够量）
		unavailable.CommentCount = uint32(len(reviews))
		return unavailable
	}

	var sb strings.Builder
	for i, r := range reviews {
		content := r.Content
		if len(content) > reviewSummaryMaxChars {
			content = content[:reviewSummaryMaxChars]
		}
		sb.WriteString(fmt.Sprintf("%d. (%d星) %s\n", i+1, r.Rating, content))
	}

	provider, err := s.aiDeps.ProviderRepo.GetEnabledDefault(ctx)
	if err != nil {
		return unavailable
	}
	client, err := newOpenAIClientForProvider(provider)
	if err != nil {
		return unavailable
	}

	model := derefAgentStr(provider.ModelName)
	gctx, cancel := context.WithTimeout(ctx, reviewSummaryTimeout)
	defer cancel()

	resp, err := client.CreateChatCompletion(gctx, openai.ChatCompletionRequest{
		Model: model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleSystem,
				Content: "你是电商商品评论分析助手。基于买家评论输出 2-3 句中文购物洞察，" +
					"概括商品的主要优点与被提及的缺点，语气客观中立。直接输出正文，" +
					"不要任何前言、标题、列表符号或「根据评论」之类的套话。",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: "以下是某商品的买家评论（星号为评分）：\n" + sb.String(),
			},
		},
		Temperature: 0.3,
		// MaxTokens 需覆盖思考型模型（qwen3/deepseek-r1 等）的思考段 + 正文
		MaxTokens: 1000,
	})
	if err != nil {
		s.log.Errorf("review summary llm call failed: product=%d: %s", productId, err.Error())
		return unavailable
	}
	if len(resp.Choices) == 0 {
		return unavailable
	}

	// 用量流水（非对话调用，与反垃圾同口径；失败仅留日志）。
	if s.aiDeps.UsageLogRepo != nil && resp.Usage.TotalTokens > 0 {
		_ = s.aiDeps.UsageLogRepo.Create(ctx, &data.UsageRecord{
			ProviderID:       provider.ID,
			ModelName:        model,
			PromptTokens:     uint32(resp.Usage.PromptTokens),
			CompletionTokens: uint32(resp.Usage.CompletionTokens),
			TotalTokens:      uint32(resp.Usage.TotalTokens),
		})
	}

	text := strings.TrimSpace(resp.Choices[0].Message.Content)
	if text == "" {
		return unavailable
	}
	return &commentV1.ProductReviewSummary{
		Summary:      text,
		CommentCount: uint32(len(reviews)),
		Available:    true,
	}
}
