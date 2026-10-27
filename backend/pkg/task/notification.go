package task

// 通知域 asynq 任务类型常量与载荷。
//
// 命名约束：任务类型互不为前缀——ServeMux 精确匹配失败后按前缀兜底路由，
// 互为前缀会让 payload 被路由到错误的 handler 解码失败（历史教训见 AGENTS.md）。
// notification.dispatch 与 notification.delivery.sweep 在 "notification.dis/del" 处分叉，安全。

// NotificationDispatchTaskType 单条通知异步投递任务（同步派发失败退回同步、异步规则入队）。
const NotificationDispatchTaskType = "notification.dispatch"

// NotificationDispatchTaskData 异步投递载荷。
// 台账不存正文快照且 target 已脱敏，所以载荷必须自带投递所需的正文原文；
// 重试不重新解析模板（SendDirect 落台账前就渲染好了）。
type NotificationDispatchTaskData struct {
	DeliveryId uint32            `json:"deliveryId"`
	Target     string            `json:"target"`
	Title      string            `json:"title"`
	Content    string            `json:"content"`
	Params     map[string]string `json:"params,omitempty"`
}

// NotificationDeliverySweepTaskType 台账清扫周期任务：把超期停留在 SENDING 的行定案 FAILED（只定案不补投）。
const NotificationDeliverySweepTaskType = "notification.delivery.sweep"

// NotificationDeliverySweepCronSpec 清扫周期：每 5 分钟。
const NotificationDeliverySweepCronSpec = "*/5 * * * *"

// NotificationDeliverySweepTaskData 空载荷：清扫阈值走环境变量（NOTIFICATION_DELIVERY_STALE_MINUTES）。
type NotificationDeliverySweepTaskData struct{}
// AiDocReindexTaskType 知识库切片重索引任务的 asynq 任务类型。
const AiDocReindexTaskType = "ai.doc.reindex"

// AiDocReindexPayload 知识库切片重索引任务 payload。
// BaseID=0 表示对全部知识库重索引；>0 只处理指定库。
type AiDocReindexPayload struct {
	BaseID uint32 `json:"base_id"`
}

// AiQueryRunTaskType 定时问数任务的 asynq 任务类型。
// 执行一次问数并把结果以站内信推送给平台侧用户。
const AiQueryRunTaskType = "ai.query.run"

// AiQueryRunPayload 定时问数任务 payload。
type AiQueryRunPayload struct {
	Question string `json:"question"`
	Lang     string `json:"lang"`
}

// AiPostIndexTaskType 帖子语义索引任务的 asynq 任务类型。
// 把帖子标题+摘要向量化写入 AI 全局搜索索引（sys_ai_search_index）。
const AiPostIndexTaskType = "ai.post.index"

// AiPostIndexPayload 帖子语义索引任务 payload。
type AiPostIndexPayload struct {
	PostID   uint32 `json:"post_id"`
	TenantID uint32 `json:"tenant_id"`
	Op       string `json:"op"` // index / delete
}
