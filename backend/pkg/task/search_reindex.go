package task

// ============================================================================
// 搜索重索引任务类型定义
//
// 用于 mall_products / mall_product_translations → Elasticsearch 的双写同步：
// product/product_translation 的 Create/Update/Delete 在 DB 事务提交后，入队一个
// search.reindex 任务，asynq worker 收到后从 DB 取最新数据写入/删除 ES 文档。
//
// 安全：
//   - payload 只含 id + op，不含文档内容，体积小、无敏感字段
//   - 实际文档内容由 worker 从 DB 取（带 SystemViewer 跨状态读），写入 ES
//   - 商品是全局共享目录（无 TenantID mixin），ES 文档无 tenant_id 字段
//   - 失败由 asynq 自动重试
//
// 与 go-wind-cms 的差异：CMS 的 SearchReindexPayload 含 TenantID（因 Post 有租户
// 隔离）；Shop 的 Product 无租户归属，故 payload 不含 TenantID。
// ============================================================================

const (
	// SearchReindexTaskType 单条重索引任务的 asynq 任务类型。
	SearchReindexTaskType = "search.reindex"
)

// SearchReindexPayload 单条重索引任务的 payload。
//
// Op 取值：
//   - "index"  ：product 或其翻译被创建/更新，worker 从 DB 取最新数据 upsert 到 ES
//   - "delete" ：product 被删除，worker 按已索引语言枚举删除 ES 中所有语言文档
//
// Entity 取值："product"（v1 仅此一种）。
type SearchReindexPayload struct {
	Entity string `json:"entity"`
	ID     uint32 `json:"id"`
	Op     string `json:"op"`
}
