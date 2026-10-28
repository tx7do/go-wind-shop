<script setup lang="ts">
import { XIcon } from '@/plugins/xicon'
import { useAccessStore } from '@/stores/modules/core/access.state'
import { globalSSEClient } from '@/core/transport/sse'
import { SSE_EVENT } from '@/core/transport/sse/event'
import { sendAssistantMessage } from '@/api/composables'

const { t } = useI18n()
const localePath = useLocalePath()
const route = useRoute()
const runtimeConfig = useRuntimeConfig()
const accessStore = useAccessStore()

const isLogin = computed(() => {
  const token = accessStore.accessToken
  return !!token?.value && !accessStore.loginExpired
})

// 商品页上下文：/product/:id 路由自动携带商品ID，让导购围绕当前商品作答
const contextProductId = computed<number | undefined>(() => {
  const m = route.path.match(/\/product\/(\d+)/)
  return m ? Number(m[1]) : undefined
})

const open = ref(false)

interface ProductCard {
  id: number
  name: string
  shortDesc: string
}

interface ChatMsg {
  role: 'user' | 'assistant'
  content: string
  cards?: ProductCard[]
}

const messages = ref<ChatMsg[]>([])
const input = ref('')
const sending = ref(false)
const toolStatus = ref('')
const conversationId = ref<number | undefined>(undefined)
const listEl = ref<HTMLElement | null>(null)

const suggestions = computed(() => [
  t('assistant.suggest1'),
  t('assistant.suggest2'),
  t('assistant.suggest3'),
])

watch(open, (v) => {
  if (v) nextTick(scrollToBottom)
})

function scrollToBottom() {
  if (listEl.value) listEl.value.scrollTop = listEl.value.scrollHeight
}

// ── SSE 流式接线：chunk 逐字渲染，tool 帧显示"正在查询…"并解析商品卡片 ──

let sseBound = false

function sseTargetUrl(): string | null {
  const base = runtimeConfig.public?.NUXT_PUBLIC_SSE_URL as string | undefined
  const token = String(accessStore.accessToken?.value ?? '')
  if (!base || !token) return null
  return `${base}?stream=${encodeURIComponent(token)}&token=${encodeURIComponent(token)}`
}

function ensureSSE() {
  if (sseBound) return
  const url = sseTargetUrl()
  if (!url) return
  const token = String(accessStore.accessToken?.value ?? '')
  globalSSEClient.setHeaders({ Authorization: `Bearer ${token}` })
  globalSSEClient.on(SSE_EVENT.AIChatChunk, handleChunkEvent)
  globalSSEClient.on(SSE_EVENT.AIChatTool, handleToolEvent)
  globalSSEClient.connect(url)
  sseBound = true
}

function teardownSSE() {
  if (!sseBound) return
  globalSSEClient.off(SSE_EVENT.AIChatChunk, handleChunkEvent)
  globalSSEClient.off(SSE_EVENT.AIChatTool, handleToolEvent)
  globalSSEClient.close()
  sseBound = false
}

// 当前轮的流式占位消息：chunk 追加其 content，POST 响应到达后以响应为准校正
let streamingMsg: ChatMsg | null = null

function handleChunkEvent(data: unknown) {
  if (!sending.value || !streamingMsg) return
  const delta = (data as any)?.delta
  if (typeof delta === 'string' && delta) {
    streamingMsg.content += delta
    nextTick(scrollToBottom)
  }
}

function toolLabel(name: string): string {
  const map: Record<string, string> = {
    search_products: 'assistant.toolSearch',
    get_product_detail: 'assistant.toolDetail',
    query_my_orders: 'assistant.toolOrders',
    query_active_coupons: 'assistant.toolCoupons',
  }
  const key = map[name]
  return key ? t(key) : t('assistant.toolGeneric')
}

// search_products 工具输出的竖线分隔文本 → 商品卡片（id | name | short_desc）
function parseProductCards(name: string, result: string): ProductCard[] {
  if (name !== 'search_products' || !result || result.startsWith('tool error')) return []
  const lines = result.split('\n').filter((l) => l.trim() !== '')
  if (lines.length < 2) return []
  const cards: ProductCard[] = []
  for (const line of lines.slice(1)) {
    const cells = line.split('|').map((c) => c.trim())
    const id = Number(cells[0])
    if (!Number.isInteger(id) || id <= 0) continue
    cards.push({ id, name: cells[1] ?? '', shortDesc: cells[2] ?? '' })
  }
  return cards
}

function handleToolEvent(data: unknown) {
  if (!sending.value || !streamingMsg) return
  const evt = data as any
  if (typeof evt?.name !== 'string') return
  toolStatus.value = toolLabel(evt.name)
  const cards = parseProductCards(evt.name, String(evt?.result ?? ''))
  if (cards.length > 0) {
    streamingMsg.cards = [...(streamingMsg.cards ?? []), ...cards]
  }
  nextTick(scrollToBottom)
}

onBeforeUnmount(teardownSSE)

// ── 发送 ──

async function handleSend(text?: string) {
  const content = (text ?? input.value).trim()
  if (!content || sending.value || !isLogin.value) return

  messages.value.push({ role: 'user', content })
  input.value = ''
  sending.value = true
  toolStatus.value = t('assistant.thinking')
  nextTick(scrollToBottom)

  ensureSSE()

  const placeholder: ChatMsg = { role: 'assistant', content: '' }
  messages.value.push(placeholder)
  streamingMsg = placeholder

  try {
    const resp = await sendAssistantMessage(content, conversationId.value, contextProductId.value)
    if (resp.conversation?.id) conversationId.value = resp.conversation.id
    // POST 响应为最终事实（覆盖流式累积，防丢帧/乱序）
    placeholder.content = resp.message?.content || t('assistant.emptyReply')
  } catch (e: any) {
    // 限额超额（403）单独提示
    const code = e?.response?.data?.code ?? e?.code
    placeholder.content = code === 403
      ? t('assistant.quotaExceeded')
      : t('assistant.errorReply')
  } finally {
    streamingMsg = null
    sending.value = false
    toolStatus.value = ''
    nextTick(scrollToBottom)
  }
}

function handleToggle() {
  if (!isLogin.value) {
    navigateTo(localePath('/login'))
    return
  }
  open.value = !open.value
  if (open.value) ensureSSE()
}

function productHref(id: number) {
  return localePath(`/product/${id}`)
}
</script>

<template>
  <!-- 悬浮入口 -->
  <button
    type="button"
    class="fixed bottom-6 right-6 z-900 flex h-12 w-12 items-center justify-center rounded-full bg-primary text-primary-foreground shadow-lg transition-transform hover:scale-105"
    :aria-label="t('assistant.title')"
    @click="handleToggle"
  >
    <XIcon :icon="open ? 'lucide:x' : 'lucide:bot'" width="22" height="22" />
  </button>

  <!-- 聊天面板 -->
  <div
    v-if="open"
    class="fixed bottom-22 right-6 z-900 flex h-100 w-85 max-w-[calc(100vw-3rem)] flex-col overflow-hidden rounded-2xl border border-border bg-card shadow-xl"
  >
    <div class="flex items-center gap-2 border-b border-border px-4 py-3">
      <XIcon icon="lucide:bot" width="18" height="18" class="text-primary" />
      <span class="text-sm font-semibold">{{ t('assistant.title') }}</span>
      <span class="text-xs text-muted-foreground">{{ t('assistant.subtitle') }}</span>
    </div>

    <!-- 商品上下文标签 -->
    <div
      v-if="contextProductId"
      class="flex items-center gap-1.5 border-b border-border bg-primary/5 px-4 py-1.5 text-xs text-primary"
    >
      <XIcon icon="lucide:package" width="12" height="12" />
      <span class="truncate">{{ t('assistant.contextTag') }} #{{ contextProductId }}</span>
    </div>

    <div ref="listEl" class="flex flex-1 flex-col gap-2 overflow-y-auto px-3 py-3">
      <p v-if="messages.length === 0" class="mt-6 text-center text-xs text-muted-foreground">
        {{ t('assistant.welcome') }}
      </p>
      <div
        v-for="(msg, idx) in messages"
        :key="idx"
        class="max-w-[92%]"
        :class="msg.role === 'user' ? 'self-end' : 'self-start'"
      >
        <div
          class="rounded-2xl px-3 py-2 text-sm leading-relaxed whitespace-pre-wrap"
          :class="msg.role === 'user'
            ? 'bg-primary text-primary-foreground'
            : 'bg-muted text-foreground'"
        >
          {{ msg.content }}<span
            v-if="sending && msg === messages[messages.length - 1] && msg.role === 'assistant' && !msg.content && !toolStatus"
            class="text-muted-foreground"
          >{{ t('assistant.thinking') }}</span>
        </div>

        <!-- 商品卡片（search_products 工具结果） -->
        <div v-if="msg.cards?.length" class="mt-1.5 flex flex-col gap-1.5">
          <p class="text-xs text-muted-foreground">{{ t('assistant.relatedProducts') }}</p>
          <NuxtLink
            v-for="card in msg.cards"
            :key="`${idx}-${card.id}`"
            :to="productHref(card.id)"
            class="flex items-center gap-2 rounded-xl border border-border bg-background px-2.5 py-2 transition-colors hover:border-primary/50"
          >
            <XIcon icon="lucide:package" width="16" height="16" class="shrink-0 text-primary" />
            <span class="min-w-0 flex-1">
              <span class="block truncate text-xs font-medium">{{ card.name || `#${card.id}` }}</span>
              <span v-if="card.shortDesc" class="block truncate text-xs text-muted-foreground">{{ card.shortDesc }}</span>
            </span>
            <span class="shrink-0 text-xs text-primary">{{ t('assistant.viewProduct') }}</span>
          </NuxtLink>
        </div>
      </div>

      <!-- 工具调用状态 -->
      <div
        v-if="sending && toolStatus"
        class="self-start flex items-center gap-1.5 rounded-2xl bg-muted px-3 py-2 text-xs text-muted-foreground"
      >
        <XIcon icon="lucide:loader-circle" width="12" height="12" class="animate-spin" />
        {{ toolStatus }}
      </div>
    </div>

    <!-- 快捷提问（空会话时展示） -->
    <div v-if="messages.length === 0" class="flex flex-wrap gap-1.5 px-3 pb-2">
      <button
        v-for="s in suggestions"
        :key="s"
        type="button"
        class="rounded-full border border-border px-2.5 py-1 text-xs text-muted-foreground transition-colors hover:bg-primary/10 hover:text-primary"
        @click="handleSend(s)"
      >
        {{ s }}
      </button>
    </div>

    <div class="flex items-center gap-2 border-t border-border p-2.5">
      <input
        v-model="input"
        type="text"
        class="h-9 flex-1 rounded-lg border border-border bg-background px-3 text-sm outline-none focus:border-primary"
        :placeholder="t('assistant.placeholder')"
        :disabled="sending"
        @keydown.enter.prevent="handleSend()"
      />
      <button
        type="button"
        class="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-primary text-primary-foreground transition-opacity disabled:opacity-50"
        :disabled="sending || !input.trim()"
        :aria-label="t('assistant.send')"
        @click="handleSend()"
      >
        <XIcon icon="lucide:send" width="16" height="16" />
      </button>
    </div>
  </div>
</template>
