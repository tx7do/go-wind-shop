<script setup lang="ts">
import { XIcon } from '@/plugins/xicon'
import { useAccessStore } from '@/stores/modules/core/access.state'
import { sendAssistantMessage } from '@/api/composables'

const { t } = useI18n()
const localePath = useLocalePath()
const accessStore = useAccessStore()

const isLogin = computed(() => {
  const token = accessStore.accessToken
  return !!token?.value && !accessStore.loginExpired
})

const open = ref(false)

interface ChatMsg {
  role: 'user' | 'assistant'
  content: string
}

const messages = ref<ChatMsg[]>([])
const input = ref('')
const sending = ref(false)
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

async function handleSend(text?: string) {
  const content = (text ?? input.value).trim()
  if (!content || sending.value || !isLogin.value) return

  messages.value.push({ role: 'user', content })
  input.value = ''
  sending.value = true
  nextTick(scrollToBottom)
  try {
    const resp = await sendAssistantMessage(content, conversationId.value)
    if (resp.conversation?.id) conversationId.value = resp.conversation.id
    messages.value.push({
      role: 'assistant',
      content: resp.message?.content || t('assistant.emptyReply'),
    })
  } catch {
    messages.value.push({ role: 'assistant', content: t('assistant.errorReply') })
  } finally {
    sending.value = false
    nextTick(scrollToBottom)
  }
}

function handleToggle() {
  if (!isLogin.value) {
    navigateTo(localePath('/login'))
    return
  }
  open.value = !open.value
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

    <div ref="listEl" class="flex flex-1 flex-col gap-2 overflow-y-auto px-3 py-3">
      <p v-if="messages.length === 0" class="mt-6 text-center text-xs text-muted-foreground">
        {{ t('assistant.welcome') }}
      </p>
      <div
        v-for="(msg, idx) in messages"
        :key="idx"
        class="max-w-[85%] rounded-2xl px-3 py-2 text-sm leading-relaxed whitespace-pre-wrap"
        :class="msg.role === 'user'
          ? 'self-end bg-primary text-primary-foreground'
          : 'self-start bg-muted text-foreground'"
      >
        {{ msg.content }}
      </div>
      <div v-if="sending" class="self-start rounded-2xl bg-muted px-3 py-2 text-sm text-muted-foreground">
        {{ t('assistant.thinking') }}
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
