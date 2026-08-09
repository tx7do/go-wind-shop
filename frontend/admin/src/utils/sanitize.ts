/**
 * HTML 消毒工具
 *
 * 用于净化服务端返回的富文本内容（通知详情、站内信正文等），防止存储型 XSS。
 * 白名单按"通知/消息展示"场景收窄到基础排版标签，移除 iframe/video/svg/math/script
 * 等高风险标签与事件属性。
 */
import DOMPurify from "dompurify";

/**
 * 净化 HTML 富文本，仅保留基础排版标签。
 * 用于 v-html 渲染前的消毒。
 */
export function sanitizeHtml(html: string): string {
  return DOMPurify.sanitize(html, {
    // 仅允许基础排版/格式标签，不含 iframe/video/svg/math/script/form 等
    ALLOWED_TAGS: [
      "p", "br", "hr", "span", "div",
      "strong", "b", "em", "i", "u", "del", "s",
      "a", "ul", "ol", "li", "blockquote",
    ],
    ALLOWED_ATTR: ["href", "title", "target", "rel", "class", "style"],
    // 禁止 <a> 的 javascript: 协议
    ALLOWED_URI_REGEXP: /^(?:(?:https?|mailto):|#|\/)/i,
    KEEP_CONTENT: true,
  }) as unknown as string;
}

/**
 * 将 HTML 转为纯文本（去除所有标签），用于列表预览等不需要富文本的场景。
 * 比直接 innerHTML 更安全：DOMPurify 以空白名单清理后再取 textContent，
 * 避免触发 <img onerror> 等事件处理器。
 */
export function sanitizeToPlainText(html: string): string {
  const cleaned = DOMPurify.sanitize(html, {
    ALLOWED_TAGS: [],
    ALLOWED_ATTR: [],
  }) as unknown as string;
  // cleaned 已无任何标签，取 textContent 安全
  const tmp = document.createElement("div");
  tmp.textContent = cleaned;
  return tmp.textContent || "";
}
