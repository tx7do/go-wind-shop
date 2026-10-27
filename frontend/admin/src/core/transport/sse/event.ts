/**
 * SSE 事件名注册表（vue-element 端副本）。
 *
 * 真相源是后端 SSE 推送处的 `event:` 行（本仓库目前只有站内信 notification，
 * 见 app/admin/service/internal/service/internal_message_service.go）：
 * 后端把常量写进帧的 `event:` 行，这里按同名注册回调，中间没有编译期检查
 * （`SSEEventName` 带 `| string`，等于不约束）。新增事件要两边同时改；
 * 已发布事件的取值不可改——改名等于对线上前端静默断流。
 */
export const SSE_EVENT = {
  /**
   * 站内信收件行推送。
   *
   * data 为 `InternalMessageRecipient` 的 protojson：camelCase 键（`messageId` /
   * `recipientUserId` / `createdAt`），`status` 是枚举名字符串（`RECEIVED` / `READ`），
   * 与 REST 收件箱接口返回的形状一致。
   */
  Notification: "notification",

  // 上游（go-wind-admin）还有 AIChatChunk / AIChatTool / AgentRun 三个 AI 相关事件，
  // 本仓库后端没有对应推送源，暂不引入；接入 AI 推送时再同步。
} as const;
