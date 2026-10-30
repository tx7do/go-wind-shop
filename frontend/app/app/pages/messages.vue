<script setup lang="ts">
definePageMeta({
  layout: 'account',
  middleware: 'auth',
})
import { computed, ref } from 'vue';
import { toast } from 'vue-sonner';
import {
  useListUserInbox,
  useMarkNotificationAsRead,
  useDeleteNotificationFromInbox,
  invalidateUserInbox,
} from '@/api/composables';
import { useAccessStore } from '@/stores/modules/core/access.state';

const { t } = useI18n();

useHead({ title: t('messages.title') });

const accessStore = useAccessStore();
const isLogin = computed(() => {
  const token = accessStore.accessToken;
  return !!token?.value && !accessStore.loginExpired;
});

type MessageStatus = 'SENT' | 'RECEIVED' | 'READ' | 'REVOKED' | 'DELETED';
type MessageEntity = {
  id?: number;
  messageId?: number;
  title?: string;
  content?: string;
  status?: MessageStatus;
  receivedAt?: string;
  readAt?: string;
  createdAt?: string;
};

// 收件箱列表：后端网关强制注入 recipientUserId，前端只传分页。
// enabled 守卫：未登录时不发请求，避免预 hydrate 闪空。
const inboxQuery = useListUserInbox(
  computed(() => ({
    page: 1,
    pageSize: 50,
    noPaging: false,
    sorting: [{ field: 'id', direction: 'DESC' }],
  })),
  { enabled: isLogin },
);
const messages = computed<MessageEntity[]>(() => {
  const items = (inboxQuery.data?.value as any)?.items ?? [];
  return (items as MessageEntity[]) ?? [];
});
const loading = computed(() => inboxQuery.isPending.value);
const loadError = computed(() => inboxQuery.isError.value);

// 未读数（SENT/RECEIVED 视为未读）
const unreadCount = computed(
  () => messages.value.filter((m) => m.status === 'SENT' || m.status === 'RECEIVED').length,
);

function isUnread(m: MessageEntity): boolean {
  return m.status === 'SENT' || m.status === 'RECEIVED';
}

const STATUS_LABEL_KEY: Record<MessageStatus, string> = {
  SENT: 'messages.status.sent',
  RECEIVED: 'messages.status.received',
  READ: 'messages.status.read',
  REVOKED: 'messages.status.revoked',
  DELETED: 'messages.status.deleted',
};
function statusLabel(s: MessageStatus | undefined): string {
  return t(STATUS_LABEL_KEY[s ?? 'SENT']);
}

// 徽标色调：未读用品牌色，撤回用警示色，其余中性
const STATUS_TONE: Record<MessageStatus, 'neutral' | 'primary' | 'warning' | 'danger'> = {
  SENT: 'primary',
  RECEIVED: 'primary',
  READ: 'neutral',
  REVOKED: 'warning',
  DELETED: 'danger',
};
function statusTone(s: MessageStatus | undefined) {
  return STATUS_TONE[s ?? 'SENT'];
}

function formatTime(ts: string | undefined): string {
  if (!ts) return '—';
  try {
    return new Date(ts).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' });
  } catch {
    return '—';
  }
}

// 展开的详情
const expandedId = ref<number | null>(null);
function toggleDetail(m: MessageEntity) {
  const id = m.id ?? 0;
  expandedId.value = expandedId.value === id ? null : id;
  // 首次展开未读消息时自动标记已读
  if (expandedId.value === id && isUnread(m) && id) {
    markReadMutation.mutate([id]);
  }
}

const markReadMutation = useMarkNotificationAsRead({
  onSuccess: () => {
    toast.success(t('messages.result.markedRead'));
    invalidateUserInbox();
  },
  onError: (err: any) => toast.error(err?.message || t('messages.errors.markReadFailed')),
});

const deleteMutation = useDeleteNotificationFromInbox({
  onSuccess: () => {
    toast.success(t('messages.result.deleted'));
    invalidateUserInbox();
  },
  onError: (err: any) => toast.error(err?.message || t('messages.errors.deleteFailed')),
});

function handleMarkRead(m: MessageEntity) {
  if (!m.id || !isUnread(m)) return;
  markReadMutation.mutate([m.id]);
}

function handleDelete(m: MessageEntity) {
  if (!m.id) return;
  if (!window.confirm(t('messages.confirm.delete'))) return;
  deleteMutation.mutate([m.id]);
}

const anyPending = computed(
  () => markReadMutation.isPending.value || deleteMutation.isPending.value,
);
</script>

<template>
  <LayoutPageHero
    :title="t('messages.title')"
    :description="t('messages.subtitle')"
    icon="lucide:message-square"
    size="sm"
  >
    <UiStatusBadge
      v-if="unreadCount > 0"
      tone="primary"
      dot
      :label="t('messages.unread', { count: unreadCount })"
    />
  </LayoutPageHero>

  <LayoutSectionContainer>
    <UiAuthGate v-if="!isLogin" />

    <!-- 加载中 -->
    <UiListSkeleton v-else-if="loading" :rows="5" :header="false" />

    <!-- 错误态 -->
    <UiAppEmpty
      v-else-if="loadError"
      variant="error"
    >
      <template #action>
        <UiButton variant="outline" size="sm" @click="inboxQuery.refetch()">
          {{ t('ui.button.retry') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <!-- 空列表 -->
    <UiAppEmpty
      v-else-if="messages.length === 0"
      variant="noData"
      :description="t('messages.empty')"
    />

    <!-- 消息列表 -->
    <div v-else class="overflow-hidden rounded-lg border border-border bg-card shadow-card">
      <ul class="divide-y divide-border">
        <li
          v-for="msg in messages"
          :key="msg.id"
          class="transition-colors hover:bg-muted"
        >
          <div class="flex items-start gap-3 px-4 py-4">
            <!-- 未读点 -->
            <span
              class="mt-1.5 h-2 w-2 shrink-0 rounded-full"
              :class="isUnread(msg) ? 'bg-primary' : 'bg-border'"
              aria-hidden="true"
            />

            <div class="min-w-0 flex-1">
              <button
                type="button"
                class="block w-full text-left"
                @click="toggleDetail(msg)"
              >
                <span class="flex flex-wrap items-center gap-x-2 gap-y-1">
                  <span
                    class="min-w-0 flex-1 truncate text-sm text-foreground"
                    :class="isUnread(msg) ? 'font-semibold' : 'font-medium'"
                  >
                    {{ msg.title || '—' }}
                  </span>
                  <UiStatusBadge
                    size="sm"
                    :tone="statusTone(msg.status)"
                    :label="statusLabel(msg.status)"
                  />
                  <span class="shrink-0 text-[11px] tabular-nums text-muted-foreground">
                    {{ formatTime(msg.receivedAt || msg.createdAt) }}
                  </span>
                </span>
                <span class="mt-1 line-clamp-2 block text-xs leading-relaxed text-muted-foreground">
                  {{ msg.content || '—' }}
                </span>
              </button>

              <!-- 展开的详情 -->
              <div v-if="expandedId === msg.id && msg.content" class="mt-3 rounded-md bg-muted p-3">
                <p class="whitespace-pre-wrap text-sm leading-relaxed text-foreground">{{ msg.content }}</p>
              </div>

              <!-- 操作 -->
              <div class="mt-2 flex items-center justify-end gap-1">
                <UiButton
                  v-if="isUnread(msg)"
                  variant="ghost"
                  size="sm"
                  :disabled="anyPending"
                  @click="handleMarkRead(msg)"
                >
                  {{ t('messages.markRead') }}
                </UiButton>
                <UiButton
                  variant="ghost"
                  size="sm"
                  class="text-destructive hover:text-destructive"
                  :disabled="anyPending"
                  @click="handleDelete(msg)"
                >
                  {{ t('messages.delete') }}
                </UiButton>
              </div>
            </div>
          </div>
        </li>
      </ul>
    </div>
  </LayoutSectionContainer>
</template>
