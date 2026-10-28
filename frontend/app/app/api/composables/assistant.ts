import { apiClient } from '@/api/client';
import type {
  aiservicev1_ChatRequest,
  aiservicev1_ChatResponse,
} from '@/api/generated/app/service/v1';

// ==============================
// AI 导购对话
// 同步 POST：done 帧（完整回复）作为响应返回；operator 身份由 BFF 从登录态
// 强制覆盖。conversationId 传回可实现多轮上下文（服务端按会话取历史）。
// ==============================
export async function sendAssistantMessage(
  content: string,
  conversationId?: number,
): Promise<aiservicev1_ChatResponse> {
  const request: aiservicev1_ChatRequest = { content };
  if (conversationId !== undefined && conversationId > 0) {
    request.conversationId = conversationId;
  }
  return await apiClient.aiChatService.Chat(request);
}
