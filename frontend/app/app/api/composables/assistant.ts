import { apiClient } from '@/api/client';
import type {
  aiservicev1_ChatRequest,
  aiservicev1_ChatResponse,
} from '@/api/generated/app/service/v1';

// ==============================
// AI 导购对话
// 同步 POST：done 帧（完整回复）作为响应返回；operator 身份由 BFF 从登录态
// 强制覆盖。conversationId 传回可实现多轮上下文（服务端按会话取历史）；
// contextProductId 携带商品页上下文（core 注入商品摘要到 system prompt）。
// 增量 chunk/tool 帧经 SSE 网关（NUXT_PUBLIC_SSE_URL）以
// `ai_chat_chunk` / `ai_chat_tool` 事件推送，本响应为最终事实。
// ==============================
export async function sendAssistantMessage(
  content: string,
  conversationId?: number,
  contextProductId?: number,
): Promise<aiservicev1_ChatResponse> {
  const request: aiservicev1_ChatRequest = { content };
  if (conversationId !== undefined && conversationId > 0) {
    request.conversationId = conversationId;
  }
  if (contextProductId !== undefined && contextProductId > 0) {
    request.contextProductId = contextProductId;
  }
  return await apiClient.aiChatService.Chat(request);
}

// ==============================
// 搜索查询 AI 改写（匿名可用）
// 自然语言 → 2-4 个检索关键词；rewritten=false 表示无需改写/降级，沿用原词。
// ==============================
export async function rewriteSearchQuery(query: string): Promise<{
  keywords: string[];
  rewritten: boolean;
}> {
  return await apiClient.aiChatService.RewriteSearchQuery({ query });
}
