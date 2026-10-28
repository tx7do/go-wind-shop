package constants

// sys_config 平台参数键与默认值：口令复杂度/有效期/历史口令策略（等保要求）。
// 阈值经参数管理页（sys_configs，is_built_in）可改，修改落库后即时生效
// （ConfigRepo 缓存写路径同步失效）。口令校验消费方（复杂度/有效期/历史检查）
// 为上游等保批能力，CMS 当前未移植——键与默认值先随配置域落位，消费方移植时即插即用。
const (
	// ConfigKeyPasswordMinLen 最小长度，默认 DefaultPasswordMinLen。
	ConfigKeyPasswordMinLen = "sys.password.minLen"
	// ConfigKeyPasswordMaxAgeDays 有效期天数；<=0 表示不启用有效期。
	ConfigKeyPasswordMaxAgeDays = "sys.password.maxAgeDays"
	// ConfigKeyPasswordHistoryCount 历史口令保留条数；<=0 表示不启用历史检查。
	ConfigKeyPasswordHistoryCount = "sys.password.historyCount"
)

const (
	DefaultPasswordMinLen       = 8
	DefaultPasswordMaxAgeDays   = 90
	DefaultPasswordHistoryCount = 3

	// 评论 AI 反垃圾开关键（AI 域接入后启用对应默认配置行）
	ConfigKeyCommentAIModeration = "comment.ai_moderation.enabled"
)

// AI 导购/生成域平台参数（sys_configs 内置种子，参数管理页可改、即时生效）。
const (
	// ConfigKeyAiAssistantKnowledgeBaseId C端导购助手默认绑定的知识库ID；
	// 空/0 表示不注入知识库上下文。ChatRequest.knowledgeBaseId=0 时（仅 app 渠道）
	// 回落到该值——售后政策/FAQ 知识库建好后把库ID填进来即可让助手答政策题。
	ConfigKeyAiAssistantKnowledgeBaseId = "ai.assistant.knowledge_base_id"

	// ConfigKeyAiChatDailyTokenLimit C端单用户每日对话 token 总限额；
	// <=0 表示不限。超额返回 429 语义的错误文本。
	ConfigKeyAiChatDailyTokenLimit = "ai.chat.daily_token_limit"

	// DefaultAiChatDailyTokenLimit 默认限额（20 万 token/人/天，防滥用兜底）。
	DefaultAiChatDailyTokenLimit = 200000
)
