package service

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
	"time"

	"go-wind-shop/pkg/logx"
	"github.com/sashabaranov/go-openai"
	"github.com/tx7do/go-crud/viewer"
	"github.com/tx7do/go-utils/trans"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"google.golang.org/protobuf/types/known/timestamppb"

	internalMessageV1 "go-wind-shop/api/gen/go/internal_message/service/v1"
	"go-wind-shop/app/core/service/internal/data/ent/user"
	appViewer "go-wind-shop/pkg/entgo/viewer"
	"go-wind-shop/pkg/task"

	entCrud "github.com/tx7do/go-crud/entgo"

	aiV1 "go-wind-shop/api/gen/go/ai/service/v1"
	"go-wind-shop/app/core/service/internal/data"
	"go-wind-shop/app/core/service/internal/data/ent"
)

// ── 智能问数安全边界 ────────────────────────────────────────────────

const (
	// aiQueryRowLimit 结果行硬上限（缺 LIMIT 自动补；超限的显式 LIMIT 也被改写）。
	aiQueryRowLimit = 100

	// aiQueryCallTimeout SQL 生成与结论生成的单次 LLM 调用兜底时长。
	aiQueryCallTimeout = 90 * time.Second
)

// aiQueryTableWhitelist 平台表白名单：NL→SQL 只允许触达这些表。
// 白名单按本仓商城真实 schema 适配（列清单经 information_schema 核对，
// 枚举值经 ent schema 核对）。
// 键为表名，值为给 LLM 的列说明（拼进提示词当数据字典）。
// 金额口径：除 mall_sku_prices.amount（varchar 数字串）外，金额列一律为
// 最小货币单位（分），给业务结论时须换算成元。
var aiQueryTableWhitelist = map[string]string{
	"mall_products":             "商品主表：id, status(PRODUCT_STATUS_DRAFT草稿/PRODUCT_STATUS_ACTIVE在售/PRODUCT_STATUS_INACTIVE下架), category_id(分类ID), brand_id(品牌ID), sort_order, created_at；注意：名称/简介在 mall_product_translations 表（经 product_id 关联），mall_products 表没有 name 列",
	"mall_product_translations": "商品翻译表：product_id(商品ID), language_code(语言代码,如 zh-CN), name(商品名), slug, short_description(简介), long_description(详情), created_at；查商品名/简介必须 JOIN 此表",
	"mall_categories":           "商品分类表：id, parent_id(父分类), path, depth, sort_order, image_url, created_at；分类名在 mall_category_translations（经 category_id 关联），本表没有 name 列",
	"mall_category_translations": "分类翻译表：category_id(分类ID), language_code, name(分类名), created_at",
	"mall_brands":               "品牌表：id, sort_order, logo_url, created_at；品牌名在 mall_brand_translations（经 brand_id 关联），本表没有 name 列",
	"mall_brand_translations":   "品牌翻译表：brand_id(品牌ID), language_code, name(品牌名), created_at",
	"mall_skus":                 "SKU 表：id, product_id(商品ID), sku_code(唯一编码), stock_qty(库存整数), created_at",
	"mall_sku_prices":           "SKU 价格表：sku_id(SKU ID), currency(币种,如 CNY), amount(价格，varchar 数字串，单位分，SQL 里需 ::numeric 换算)，取价须过滤 currency='CNY'",
	"mall_orders":               "订单主表：id, user_id(买家用户ID), status(PENDING_PAYMENT待支付/PAID已支付/CANCELLED已取消/FULFILLED已发货/CLOSED已关闭), total_amount(实付总额,分), original_amount(原价总额,分), discount_amount(优惠金额,分), shipping_fee(运费,分), tax_amount(税费,分), currency, recipient_name, recipient_phone, shipping_region, created_at（销售额统计用 total_amount，取消单 status=CANCELLED 不计）",
	"mall_order_items":          "订单明细表：id, order_id(订单ID), user_id, sku_id, sku_snapshot(下单时SKU快照JSON), quantity(数量), unit_price(成交单价,分), subtotal(小计,分), created_at；算商品销量/热卖榜用本表聚合",
	"mall_payment_transactions": "支付流水表：id, order_id, user_id, amount(金额,分), currency, payment_method(支付方式), status(含 SUCCEEDED=成功), business_type, created_at；营收对账以 status='SUCCEEDED' 为准",
	"mall_payment_refunds":      "退款表：id, transaction_id(支付流水ID), user_id, amount(退款额,分), currency, status(PENDING/SUCCEEDED/FAILED), created_at",
	"mall_coupon_templates":     "优惠券模板表：id, discount_type(优惠类型: FIXED固定额/PERCENTAGE百分比), discount_value(固定优惠额,分), discount_percentage(百分比,整数), currency, valid_from, valid_until, max_redemptions(发放上限), max_redemptions_per_user, redeemed_count(已领取数), claimable(是否可自助领取), status(ACTIVE=进行中), created_at",
	"mall_user_coupons":         "用户券实例表：id, user_id, coupon_template_id, status(UNUSED未用/USED已用/EXPIRED过期), redeemed_at, redeemed_order_id, applied_discount_amount(实际抵扣,分), created_at；核销率=USED数/总数",
	"mall_promotions":           "促销活动表：id, name(活动名), promo_type(SPECIAL_PRICE专价/FULL_REDUCTION满减), product_id(专价生效商品,0=店铺级), special_price(专价,分), threshold_amount(满减门槛,分), reduction_amount(满减额,分), currency, priority, valid_from, valid_until, status(ACTIVE=进行中), created_at",
	"mall_comments":             "商品评论表：id, content_type(='CONTENT_TYPE_PRODUCT'), object_id(=商品ID), content(评论文本), status(STATUS_APPROVED已通过/STATUS_PENDING待审/STATUS_REJECTED已拒绝), rating(1-5星), parent_id(商家回复的父评论ID), created_at",
	"mall_comment_likes":        "评论点赞表：id, user_id, comment_id, created_at",
	"mall_cart_items":           "购物车明细表：id, cart_id, user_id, sku_id, quantity, created_at",
	"mall_carts":                "购物车主表：id, user_id, created_at",
	"mall_wishlist":             "收藏表：id, user_id, product_id, created_at（注意表名单数 wishlist）",
	"mall_invoices":             "发票表：id, order_id, user_id, invoice_type, status(PENDING/ISSUED已开票/CANCELLED), amount(金额,分), buyer_name, buyer_tax_id, created_at",
	"mall_shipments":            "物流单表：id, order_id, user_id, carrier(承运商), tracking_number, status(PENDING/SHIPPED已发货/DELIVERED已送达), tracking_events(JSON), created_at",
	"mall_shipping_rates":       "运费模板表：id, region(区域), base_fee(基础运费,分), per_unit_fee(续重费,分), currency, status, created_at",
	"mall_stock_op_logs":        "库存操作流水表（append-only）：id, sku_id, product_id, delta(变动量正负), stock_before, stock_after, reason(原因,如 order_deduct/cancel_restore/manual_adjust), ref_id(关联单据ID), remark, created_at",
	"mall_stock_alerts":         "库存预警表：id, sku_id, threshold(阈值), stock_qty_at_trigger, status, created_at",
	"sys_users":      "用户表：id, username(登录名), nickname(昵称), email, status, last_login_at, created_at；用户与角色是多对多，经 sys_user_roles 关联，sys_users 表没有 role_id 列",
	"sys_roles":      "角色表：id, name(角色名), code(角色编码), status, created_at",
	"sys_user_roles": "用户-角色关联表：user_id, role_id",
	"sys_operation_audit_logs": "操作审计日志：id, user_id, username, action(CREATE/UPDATE/DELETE/READ/EXPORT/…), resource_type, success(bool), failure_reason, created_at",
	"sys_login_audit_logs":     "登录审计日志：id, user_id, username, status(SUCCESS/FAILED), ip_address, created_at",
	"sys_ai_usage_logs":        "AI 用量流水：id, user_id, tenant_id, model_name, prompt_tokens, completion_tokens, total_tokens, duration_ms, created_at",
}

// aiQueryTenantTableWhitelist 租户表白名单：只含带 tenant_id 列、可安全暴露给
// 租户自助分析的本租户数据表（平台级表不在其中）。
var aiQueryTenantTableWhitelist = map[string]string{
	"mall_products":             "商品表（本租户）：id, status(PRODUCT_STATUS_*), category_id, brand_id, created_at；名称在 mall_product_translations（经 product_id 关联）",
	"mall_product_translations": "商品翻译表（本租户）：product_id, language_code, name, short_description, created_at",
	"mall_categories":           "分类表（本租户）：id, parent_id, sort_order, created_at；分类名在 mall_category_translations",
	"mall_category_translations": "分类翻译表（本租户）：category_id, language_code, name, created_at",
	"mall_brands":               "品牌表（本租户）：id, sort_order, created_at；品牌名在 mall_brand_translations",
	"mall_brand_translations":   "品牌翻译表（本租户）：brand_id, language_code, name, created_at",
	"mall_skus":                 "SKU 表（本租户）：id, product_id, sku_code, stock_qty, created_at",
	"mall_sku_prices":           "SKU 价格表（本租户）：sku_id, currency, amount(varchar 数字串,单位分)",
	"mall_orders":               "订单表（本租户）：id, user_id, status(PENDING_PAYMENT/PAID/CANCELLED/FULFILLED/CLOSED), total_amount(分), discount_amount(分), shipping_fee(分), created_at",
	"mall_order_items":          "订单明细表（本租户）：id, order_id, sku_id, sku_snapshot, quantity, unit_price(分), subtotal(分), created_at",
	"mall_coupon_templates":     "优惠券模板表（本租户）：id, discount_type, discount_value(分), discount_percentage, valid_from, valid_until, max_redemptions, redeemed_count, claimable, status, created_at",
	"mall_user_coupons":         "用户券实例表（本租户）：id, user_id, coupon_template_id, status(UNUSED/USED/EXPIRED), applied_discount_amount(分), created_at",
	"mall_promotions":           "促销活动表（本租户）：id, name, promo_type(SPECIAL_PRICE/FULL_REDUCTION), product_id, special_price(分), threshold_amount(分), reduction_amount(分), valid_from, valid_until, status, created_at",
	"mall_comments":             "商品评论表（本租户）：id, content_type, object_id(=商品ID), content, status(STATUS_APPROVED/STATUS_PENDING/STATUS_REJECTED), rating(1-5), parent_id, created_at",
	"mall_cart_items":           "购物车明细表（本租户）：id, user_id, sku_id, quantity, created_at",
	"mall_wishlist":             "收藏表（本租户）：id, user_id, product_id, created_at",
	"mall_invoices":             "发票表（本租户）：id, order_id, status(PENDING/ISSUED/CANCELLED), amount(分), created_at",
	"mall_shipments":            "物流单表（本租户）：id, order_id, carrier, tracking_number, status(PENDING/SHIPPED/DELIVERED), created_at",
	"mall_stock_op_logs":        "库存操作流水表（本租户）：id, sku_id, product_id, delta, stock_before, stock_after, reason, ref_id, created_at",
	"sys_operation_audit_logs": "操作审计日志（本租户）：id, user_id, username, action, success, failure_reason, created_at",
	"sys_login_audit_logs":     "登录审计日志（本租户）：id, user_id, username, status(SUCCESS/FAILED), ip_address, created_at",
	"sys_ai_usage_logs":        "AI 用量流水（本租户）：id, user_id, model_name, prompt_tokens, completion_tokens, total_tokens, created_at",
}

// tenantIDPatternFor 构造"tenant_id = {id}"存在性校验的正则（容忍别名与空白）。
func tenantIDPatternFor(tenantID uint32) *regexp.Regexp {
	return regexp.MustCompile(`(?i)\btenant_id\s*=\s*` + fmt.Sprintf("%d", tenantID) + `\b`)
}

// aiQueryForbiddenPatterns 生成 SQL 的硬拒绝模式：写操作与危险语句。
var aiQueryForbiddenPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\b(INSERT|UPDATE|DELETE|DROP|TRUNCATE|ALTER|CREATE|GRANT|REVOKE|COPY|VACUUM|REINDEX|CALL|DO|SET)\b`),
	regexp.MustCompile(`(?i)\b(pg_read_file|pg_ls_dir|pg_sleep|lo_import|lo_export|dblink)\b`),
	regexp.MustCompile(`\;`),     // 多语句
	regexp.MustCompile(`--|/\*`), // 注释（防注释截断绕过）
}

// aiQuerySelectPattern 必须以 SELECT 或 WITH 开头（含前导空白/括号容忍）。
var aiQuerySelectPattern = regexp.MustCompile(`(?is)^\s*\(?\s*(SELECT|WITH)\b`)

// ── 服务 ───────────────────────────────────────────────────────────

// AiQueryService 智能问数：NL → 只读 SQL → 结构化结果 →（可选）自然语言结论。
// 安全边界见 ai_query.proto：只读校验/表白名单/LIMIT 钳制/只读事务四重防线。
type AiQueryService struct {
	aiV1.UnimplementedAiQueryServiceServer
	log *logx.Helper
	providerRepo     *data.AiProviderRepo
	usageRepo        *data.AiUsageLogRepo
	entClient        *entCrud.EntClient[*ent.Client]
	internalMsgRepo  *data.InternalMessageRepo
	internalMessages *InternalMessageService
}

func NewAiQueryService(
	ctx *bootstrap.Context,
	providerRepo *data.AiProviderRepo,
	usageRepo *data.AiUsageLogRepo,
	entClient *entCrud.EntClient[*ent.Client],
	internalMsgRepo *data.InternalMessageRepo,
	internalMessages *InternalMessageService,
) *AiQueryService {
	return &AiQueryService{
		log:              logx.New(ctx.NewLoggerHelper("ai_query/service/core-service")),
		providerRepo:     providerRepo,
		usageRepo:        usageRepo,
		entClient:        entClient,
		internalMsgRepo:  internalMsgRepo,
		internalMessages: internalMessages,
	}
}

// Ask 智能问数主链路：NL→SQL（第一次调用）→ 护栏校验 → 只读执行 → NL 结论（第二次调用）。
func (s *AiQueryService) Ask(ctx context.Context, req *aiV1.AskAiQueryRequest) (*aiV1.AskAiQueryResponse, error) {
	// 双白名单：平台用户查全平台表；租户用户查本租户数据表子集，
	// 且生成的 SQL 必须自带 tenant_id = 自身租户 过滤（后置校验 fail-closed）。
	tenantID, _ := s.operatorTenant(ctx)
	whitelist := aiQueryTableWhitelist
	tenantPattern := (*regexp.Regexp)(nil)
	if tenantID != 0 {
		whitelist = aiQueryTenantTableWhitelist
		tenantPattern = tenantIDPatternFor(tenantID)
	}
	question := strings.TrimSpace(req.GetQuestion())
	if question == "" {
		return nil, aiV1.ErrorBadRequest("question is required")
	}

	provider, err := s.providerRepo.GetEnabledDefault(ctx)
	if err != nil {
		return nil, err
	}
	client, err := newOpenAIClientForProvider(provider)
	if err != nil {
		return nil, err
	}

	lang := req.GetLang()
	if lang == "" {
		lang = "zh-CN"
	}

	// 1. NL → SQL（携带对话历史以消解追问里的指代，如"那只看 admin 的"）
	history := make([]aiV1.AiQueryHistoryItem, 0, len(req.GetHistory()))
	for _, h := range req.GetHistory() {
		history = append(history, aiV1.AiQueryHistoryItem{
			Question:      h.GetQuestion(),
			Sql:           h.GetSql(),
			ResultSummary: h.GetResultSummary(),
		})
	}
	sqlGenerated, genTokens, err := s.generateSQL(ctx, client, provider, question, lang, history, whitelist, tenantID)
	if err != nil {
		return nil, err
	}

	// 2. 四重护栏：只读语句 / 白名单表 / LIMIT 钳制 / 租户谓词复核
	cleaned, err := sanitizeSQLFor(sqlGenerated, whitelist, tenantPattern)
	if err != nil {
		return nil, aiV1.ErrorBadRequest("generated sql rejected: %v", err)
	}

	// 3. 只读事务执行
	columns, rows, execErr := s.executeReadOnly(ctx, cleaned)

	resp := &aiV1.AskAiQueryResponse{
		Sql:     cleaned,
		Columns: columns,
		Rows:    rows,
	}
	if execErr != nil {
		// 执行失败把数据库错误回给前端（便于用户改问题重试），SQL 照样返回
		s.log.Errorf("ai query execute failed: %v", execErr)
		resp.ErrorMessage = trans.Ptr(execErr.Error())
		return resp, nil
	}
	resp.RowCount = uint32(len(rows))

	// 4.（可选）结果 → 自然语言结论
	if !req.GetWithAnswer() || len(rows) == 0 {
		return resp, nil
	}
	answer, answerTokens, err := s.generateAnswer(ctx, client, provider, question, cleaned, columns, rows, lang)
	if err != nil {
		s.log.Errorf("ai query answer failed, returning rows only: %v", err)
	} else {
		resp.Answer = trans.Ptr(answer)
	}

	resp.TotalTokens = genTokens + answerTokens

	// 记账（两次调用合并总量；失败不阻断）
	if s.usageRepo != nil {
		modelName := ptrStrOr(provider.ModelName, "")
		_ = s.usageRepo.Create(ctx, &data.UsageRecord{
			UserID:      s.operatorID(ctx),
			TenantID:    tenantID,
			ModelName:   modelName,
			TotalTokens: resp.GetTotalTokens(),
		})
	}

	return resp, nil
}

// operatorTenant / operatorID：viewer 上下文取操作者（unary 中间件注入）。
func (s *AiQueryService) operatorTenant(ctx context.Context) (uint32, bool) {
	if vc, exist := viewer.FromContext(ctx); exist && vc != nil {
		return uint32(vc.TenantID()), true
	}
	return 0, false
}

func (s *AiQueryService) operatorID(ctx context.Context) uint32 {
	if uid, ok := viewerUserIDFromContext(ctx); ok {
		return uid
	}
	return 0
}

// generateSQL 数据字典 + 对话历史 + 问题 → 只读 SQL。
func (s *AiQueryService) generateSQL(ctx context.Context, client *openai.Client, provider *ent.AiProvider, question, lang string, history []aiV1.AiQueryHistoryItem, whitelist map[string]string, tenantID uint32) (string, uint32, error) {
	var sb strings.Builder
	sb.WriteString("你是本系统的只读数据分析引擎。用户的提问永远是对上面列出的数据库表发起查询——哪怕措辞听起来像个人账号问题，一律直接生成 SQL，绝不拒绝。\n")
	sb.WriteString("根据表结构与用户问题，生成一条 Postgres 只读 SELECT 语句。\n")
	sb.WriteString("规则：\n")
	sb.WriteString("1. 只输出 SQL 本身，不要任何解释、不要 Markdown 代码块标记；\n")
	sb.WriteString("2. 只能使用下面列出的表；涉及多表时用 JOIN；\n")
	sb.WriteString("3. 必须带 LIMIT（最大 100）；时间比较用 created_at 与 NOW()；\n")
	sb.WriteString(fmt.Sprintf("4. 结论文字用%s；\n", langName(lang)))
	sb.WriteString("5. 用户问题可能包含对上一轮查询的指代（如\"那只看…的\"、\"换成按天分组\"），结合对话历史理解其完整含义后生成独立可执行的 SQL。\n\n")
	sb.WriteString("可用表：\n")
	for table, desc := range whitelist {
		sb.WriteString("- " + table + "：" + desc + "\n")
	}
	if tenantID != 0 {
		sb.WriteString(fmt.Sprintf("强制规则：所有查询必须包含 tenant_id = %d 过滤条件（本租户数据边界），缺失视为无效输出。\n\n", tenantID))
	}

	// few-shot 用真实消息序列锚定行为（上下文示例对行为的约束力远强于 system 指令）
	exampleSQL := "SELECT p.id, pt.name, p.status, p.created_at FROM mall_products p " +
		"LEFT JOIN mall_product_translations pt ON pt.product_id = p.id AND pt.language_code = 'zh-CN' " +
		"WHERE p.status = 'PRODUCT_STATUS_ACTIVE' AND p.created_at >= NOW() - INTERVAL '24 hours' " +
		"ORDER BY p.created_at DESC LIMIT 100"
	if tenantID != 0 {
		exampleSQL = strings.Replace(exampleSQL, "WHERE p.status", fmt.Sprintf("WHERE p.tenant_id = %d AND p.status", tenantID), 1)
	}
	messages := make([]openai.ChatCompletionMessage, 0, len(history)*2+6)
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: sb.String(),
	})
	// few-shot 1：商城高频问法——商品名在翻译表，必须 JOIN（锚定本仓 schema 的最大陷阱）
	messages = append(messages,
		openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: "最近 24 小时上架的商品有哪些？",
		},
		openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: exampleSQL,
		},
	)
	// few-shot 2：热卖榜——订单明细聚合 + 金额分转元（锚定 sku_snapshot 汇总与金额口径）
	messages = append(messages,
		openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: "卖得最好的 5 个商品是哪些？各卖了多少",
		},
		openai.ChatCompletionMessage{
			Role: openai.ChatMessageRoleAssistant,
			Content: "SELECT oi.sku_id, oi.sku_snapshot, SUM(oi.quantity) AS total_qty, " +
				"ROUND(SUM(oi.subtotal)::numeric/100, 2) AS total_sales_yuan " +
				"FROM mall_order_items oi " +
				"JOIN mall_orders o ON o.id = oi.order_id AND o.status IN ('PAID','FULFILLED') " +
				"GROUP BY oi.sku_id, oi.sku_snapshot ORDER BY total_qty DESC LIMIT 5",
		},
	)
	// 下标访问而非按值 range：AiQueryHistoryItem 内嵌 protoimpl.MessageState
	// （含 sync.Mutex），按值拷贝会连带复制锁。
	for i := range history {
		messages = append(messages,
			openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: history[i].GetQuestion()},
			openai.ChatCompletionMessage{Role: openai.ChatMessageRoleAssistant, Content: history[i].GetSql()},
		)
	}
	messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: question})

	qCtx, cancel := context.WithTimeout(ctx, aiQueryCallTimeout)
	defer cancel()
	resp, err := client.CreateChatCompletion(qCtx, openai.ChatCompletionRequest{
		Model:    ptrStrOr(provider.ModelName, "gpt-4o-mini"),
		Messages: messages,
	})
	if err != nil {
		return "", 0, aiV1.ErrorInternalServerError("ai query generate failed: %v", err)
	}
	tokens := uint32(resp.Usage.TotalTokens)
	if len(resp.Choices) == 0 {
		return "", tokens, aiV1.ErrorInternalServerError("ai query generate empty")
	}
	raw := resp.Choices[0].Message.Content
	s.log.Infof("ai query raw model output: %s", raw)
	return raw, tokens, nil
}

// sanitizeSQLFor 带白名单与租户谓词校验的四重护栏：
// SQL 提取 → 只读校验 → 危险模式拒绝 → 白名单表校验（含表名纠正）→ LIMIT 钳制。
func sanitizeSQLFor(raw string, whitelist map[string]string, tenantPattern *regexp.Regexp) (string, error) {
	sqlText := extractSQL(raw)
	if sqlText == "" {
		return "", fmt.Errorf("no SQL statement found in model output")
	}

	if !aiQuerySelectPattern.MatchString(sqlText) {
		return "", fmt.Errorf("only SELECT/WITH statements are allowed")
	}
	for _, pat := range aiQueryForbiddenPatterns {
		if pat.MatchString(sqlText) {
			return "", fmt.Errorf("forbidden pattern detected")
		}
	}
	// 白名单：提取 FROM/JOIN 后的表名逐一校验；模型偶发丢前缀或写单复数，依候选序列纠正
	tablePattern := regexp.MustCompile(`(?i)\b(FROM|JOIN)\s+([a-zA-Z_][a-zA-Z0-9_]*)`)
	matches := tablePattern.FindAllStringSubmatch(sqlText, -1)
	for _, m := range matches {
		table := strings.ToLower(m[2])
		if _, ok := whitelist[table]; ok {
			continue
		}
		corrected, ok := correctTableName(table, whitelist)
		if !ok {
			return "", fmt.Errorf("table %q is not in the whitelist", m[2])
		}
		fixPattern := regexp.MustCompile(`(?i)\b` + regexp.QuoteMeta(table) + `\b`)
		sqlText = fixPattern.ReplaceAllStringFunc(sqlText, func(string) string {
			return corrected
		})
	}
	if tenantPattern != nil && !tenantPattern.MatchString(sqlText) {
		return "", fmt.Errorf("missing mandatory tenant_id filter")
	}
	// LIMIT 钳制：无 LIMIT 补 aiQueryRowLimit；LIMIT N>上限改写
	lower := strings.ToLower(sqlText)
	if !regexp.MustCompile(`(?i)\blimit\s+\d+`).MatchString(lower) {
		sqlText = fmt.Sprintf("SELECT * FROM (%s) AS _q LIMIT %d", sqlText, aiQueryRowLimit)
	} else {
		reLimit := regexp.MustCompile(`(?i)\blimit\s+(\d+)`)
		sqlText = reLimit.ReplaceAllStringFunc(sqlText, func(m string) string {
			n := 0
			fmt.Sscanf(strings.TrimSpace(m[5:]), "%d", &n)
			if n > aiQueryRowLimit {
				return fmt.Sprintf("LIMIT %d", aiQueryRowLimit)
			}
			return m
		})
	}
	return sqlText, nil
}

// executeReadOnly 只读事务执行（数据库层兜底：READ ONLY 事务里写操作必败）。
func (s *AiQueryService) executeReadOnly(ctx context.Context, sqlText string) ([]string, []*aiV1.AiQueryRow, error) {
	db := s.entClient.DB()
	txCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	tx, err := db.BeginTx(txCtx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, nil, fmt.Errorf("begin read-only tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	resultRows, err := tx.QueryContext(txCtx, sqlText)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = resultRows.Close() }()

	columns, err := resultRows.Columns()
	if err != nil {
		return nil, nil, err
	}

	out := make([]*aiV1.AiQueryRow, 0, 16)
	for resultRows.Next() {
		values := make([]interface{}, len(columns))
		ptrs := make([]interface{}, len(columns))
		for i := range values {
			ptrs[i] = &values[i]
		}
		if err = resultRows.Scan(ptrs...); err != nil {
			return nil, nil, err
		}
		row := &aiV1.AiQueryRow{Values: make([]string, 0, len(columns))}
		for _, v := range values {
			row.Values = append(row.Values, stringifyCellValue(v))
		}
		out = append(out, row)
	}
	if err = resultRows.Err(); err != nil {
		return nil, nil, err
	}
	return columns, out, nil
}

// generateAnswer 结果集 → 自然语言结论（第二次调用）。
func (s *AiQueryService) generateAnswer(ctx context.Context, client *openai.Client, provider *ent.AiProvider, question, sqlText string, columns []string, rows []*aiV1.AiQueryRow, lang string) (string, uint32, error) {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("用户问题：%s\n\n执行的 SQL：%s\n\n查询结果（列：%s）：\n", question, sqlText, strings.Join(columns, ", ")))
	maxRows := len(rows)
	if maxRows > 20 {
		maxRows = 20 // 结论只看前 20 行，避免超长
	}
	for _, r := range rows[:maxRows] {
		sb.WriteString(strings.Join(r.Values, " | "))
		sb.WriteString("\n")
	}

	langLine := "用中文回答。"
	if strings.HasPrefix(lang, "en") {
		langLine = "Answer in English."
	}

	aCtx, cancel := context.WithTimeout(ctx, aiQueryCallTimeout)
	defer cancel()
	resp, err := client.CreateChatCompletion(aCtx, openai.ChatCompletionRequest{
		Model: ptrStrOr(provider.ModelName, "gpt-4o-mini"),
		Messages: []openai.ChatCompletionMessage{
			{
				Role: openai.ChatMessageRoleSystem,
				Content: "你是数据分析助手。根据用户问题、SQL 与查询结果，" +
					"给出简明的结论：直接回答问题本身，引用具体数字；" +
					"结果为空就明确说明没有匹配数据。" + langLine,
			},
			{Role: openai.ChatMessageRoleUser, Content: sb.String()},
		},
	})
	if err != nil {
		return "", 0, err
	}
	tokens := uint32(resp.Usage.TotalTokens)
	if len(resp.Choices) == 0 {
		return "", tokens, fmt.Errorf("empty answer")
	}
	return strings.TrimSpace(resp.Choices[0].Message.Content), tokens, nil
}

// aiQueryExtractPattern 提取用：非锚定定位第一条 SELECT/WITH。
var aiQueryExtractPattern = regexp.MustCompile(`(?is)\b(SELECT|WITH)\b`)

// extractSQL 从模型输出中提取 SQL（剥围栏、截分号、剥行注释）。
func extractSQL(raw string) string {
	s := stripSQLFence(raw)
	loc := aiQueryExtractPattern.FindStringIndex(s)
	if loc == nil {
		return ""
	}
	s = s[loc[0]:]
	if idx := strings.Index(s, ";"); idx >= 0 {
		s = s[:idx]
	}
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if idx := strings.Index(line, "--"); idx >= 0 {
			lines[i] = line[:idx]
		}
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

// stripSQLFence 剥掉模型可能输出的 Markdown 代码块围栏。
func stripSQLFence(s string) string {
	s = strings.TrimSpace(s)
	for _, fence := range []string{"```sql", "```SQL", "```"} {
		if strings.HasPrefix(s, fence) {
			s = strings.TrimPrefix(s, fence)
			if idx := strings.LastIndex(s, "```"); idx >= 0 {
				s = s[:idx]
			}
			break
		}
	}
	return strings.TrimSpace(s)
}

// stringifyCellValue 数据库原始值 → 展示字符串（time/[]byte/nil 特判）。
func stringifyCellValue(v interface{}) string {
	switch val := v.(type) {
	case nil:
		return ""
	case time.Time:
		return val.Format("2006-01-02 15:04:05")
	case []byte:
		return string(val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// correctTableName 表名纠正：依候选序列匹配白名单（前缀 / 复数 s / 组合）。
func correctTableName(table string, whitelist map[string]string) (string, bool) {
	if _, ok := whitelist[table]; ok {
		return table, true
	}
	for _, candidate := range []string{
		"sys_" + table,
		table + "s",
		"sys_" + table + "s",
		strings.TrimSuffix(table, "s"),
		strings.TrimSuffix(table, "s") + "es",
	} {
		if _, ok := whitelist[candidate]; ok {
			return candidate, true
		}
	}
	return "", false
}

// ── 定时问数（asynq：ai.query.run） ─────────────────────────────────

// RunScheduledQuery 执行一次问数并返回格式化的文本结果（平台白名单）。
func (s *AiQueryService) RunScheduledQuery(ctx context.Context, question, lang string) (string, error) {
	provider, err := s.providerRepo.GetEnabledDefault(ctx)
	if err != nil {
		return "", err
	}
	client, err := newOpenAIClientForProvider(provider)
	if err != nil {
		return "", err
	}

	var history []aiV1.AiQueryHistoryItem
	sqlText, _, err := s.generateSQL(ctx, client, provider, question, lang, history, aiQueryTableWhitelist, 0)
	if err != nil {
		return "", err
	}
	cleaned, err := sanitizeSQLFor(sqlText, aiQueryTableWhitelist, nil)
	if err != nil {
		return "", err
	}
	_, rows, execErr := s.executeReadOnly(ctx, cleaned)
	if execErr != nil {
		return "", fmt.Errorf("execute: %w", execErr)
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("问题：%s\n", question))
	sb.WriteString(fmt.Sprintf("SQL：%s\n", cleaned))
	if len(rows) == 0 {
		sb.WriteString("结果：无匹配数据")
	} else {
		sb.WriteString(fmt.Sprintf("结果（%d 行）：\n", len(rows)))
		for _, r := range rows {
			sb.WriteString("  " + strings.Join(r.Values, " | ") + "\n")
		}
	}
	return sb.String(), nil
}

// AsyncAiQueryRun 定时问数任务 handler：执行问题 → 站内信推送结果给平台侧用户。
// asynq ctx 无 viewer，走系统查看器。
func (s *AiQueryService) AsyncAiQueryRun(taskType string, data *task.AiQueryRunPayload) error {
	ctx := appViewer.NewSystemViewerContext(context.Background())
	question := data.Question
	if question == "" {
		return fmt.Errorf("question is required")
	}

	result, err := s.RunScheduledQuery(ctx, question, "zh-CN")
	if err != nil {
		return err
	}

	title := "AI 问数结果：" + question
	if err = s.pushInternalMessage(ctx, title, result); err != nil {
		s.log.Errorf("ai query run: push result failed: %v", err)
	}
	return nil
}

// pushInternalMessage 创建站内信（平台侧，NOTIFICATION 类型）并给每个
// 平台用户（tenant_id=0）建收件行。
func (s *AiQueryService) pushInternalMessage(ctx context.Context, title, content string) error {
	now := time.Now()
	msg, err := s.internalMsgRepo.Create(ctx, &internalMessageV1.CreateInternalMessageRequest{
		Data: &internalMessageV1.InternalMessage{
			Title:     &title,
			Content:   &content,
			Status:    trans.Ptr(internalMessageV1.InternalMessage_PUBLISHED),
			Type:      trans.Ptr(internalMessageV1.InternalMessage_NOTIFICATION),
			CreatedBy: trans.Ptr(uint32(0)),
			CreatedAt: timestamppb.New(now),
		},
	})
	if err != nil {
		return err
	}

	// 投递给平台侧（tenant_id=0）用户
	users, err := s.entClient.Client().User.Query().
		Where(user.TenantIDEQ(0), user.DeletedAtIsNil()).
		All(ctx)
	if err != nil {
		return err
	}
	for _, u := range users {
		if err = s.internalMessages.sendNotification(ctx, msg.GetId(), u.ID, 0, &now, title, content); err != nil {
			s.log.Errorf("push to user %d failed: %v", u.ID, err)
		}
	}
	return nil
}
