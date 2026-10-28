/**
 * SSE 事件名注册表（C端副本）。
 *
 * 真相源是后端 `backend/pkg/sseevent/sseevent.go`：它把常量写进帧的 `event:` 行，
 * 这里按同名注册回调，中间没有编译期检查（`SSEEventName` 带 `| string`，等于不约束）。
 * 新增事件要两边同时改；已发布事件的取值不可改——改名等于对线上前端静默断流。
 */
export const SSE_EVENT = {
  /**
   * AI 对话流式片段推送。
   *
   * data 为 `ChatChunkEvent` 的 protojson：camelCase 键（`conversationId` /
   * `seq` / `delta`）。一次对话产生多个 chunk 帧，按 seq 顺序累积渲染；
   * POST /app/v1/ai/chat/completions 的同步响应携带完整回复，以响应为准校正。
   */
  AIChatChunk: "ai_chat_chunk",

  /**
   * AI 对话的工具调用可见化推送。
   *
   * data 为 `ChatToolEvent` 的 protojson：camelCase 键（`conversationId` /
   * `name` / `arguments` / `result`）。前端据此展示"模型正在调工具"，
   * 并把 search_products 的结果解析成商品卡片。
   */
  AIChatTool: "ai_chat_tool",
} as const;
