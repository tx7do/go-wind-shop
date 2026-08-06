<script setup lang="ts">
import { computed, ref } from 'vue';
import { toast } from 'vue-sonner';
import { XIcon } from '@/plugins/xicon';
import {
  useListUserInbox,
  useMarkNotificationAsRead,
  useDeleteNotificationFromInbox,
  invalidateUserInbox,
} from '@/api/composables';
import { useAccessStore } from '@/stores/modules/core/access.state';

const { t } = useI18n();
const localePath = useLocalePath();

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
    icon="carbon:message"
    size="sm"
  />

  <LayoutSectionContainer>
    <!-- 未登录 -->
    <div
      v-if="!isLogin"
      class="flex flex-col items-center gap-6 rounded-2xl border border-border bg-card p-12 text-center"
    >
      <XIcon icon="carbon:locked" :size="48" class="text-muted-foreground" />
      <p class="text-lg text-muted-foreground">{{ t('authentication.login.please_login') }}</p>
      <UiButton @click="navigateTo(localePath('/login'))">
        {{ t('navbar.user.login') }}
      </UiButton>
    </div>

    <!-- 加载中 -->
    <div v-else-if="loading" class="flex flex-col gap-3">
      <div v-for="i in 4" :key="i" class="rounded-2xl border border-border bg-card p-5">
        <div class="flex items-center gap-4">
          <UiSkeleton class="h-4 w-3/4" />
          <UiSkeleton class="h-4 w-20 ml-auto" />
        </div>
      </div>
    </div>

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
    <div v-else class="flex flex-col gap-3">
      <!-- 未读计数提示 -->
      <div v-if="unreadCount > 0" class="text-sm text-muted-foreground">
        {{ t('messages.unread', { count: unreadCount }) }}
      </div>

      <div
        v-for="msg in messages"
        :key="msg.id"
        class="rounded-2xl border border-border bg-card p-5 transition-colors"
        :class="isUnread(msg) ? 'border-primary/40' : ''"
      >
        <div class="flex flex-wrap items-start justify-between gap-3">
          <button
            type="button"
            class="flex min-w-0 flex-1 items-center gap-2 text-left"
            @click="toggleDetail(msg)"
          >
            <span
              v-if="isUnread(msg)"
              class="h-2 w-2 shrink-0 rounded-full bg-primary"
              aria-hidden="true"
            />
            <span class="min-w-0 flex-1">
              <span class="block truncate text-sm font-semibold text-foreground">
                {{ msg.title || '—' }}
              </span>
              <span v-if="!isUnread(msg)" class="mt-0.5 block truncate text-xs text-muted-foreground">
                {{ msg.content || '' }}
              </span>
            </span>
          </button>

          <div class="flex shrink-0 items-center gap-2">
            <span class="text-[10px] text-muted-foreground">{{ formatTime(msg.receivedAt || msg.createdAt) }}</span>
            <span
              class="rounded-full px-2 py-0.5 text-[10px] font-medium"
              :class="isUnread(msg) ? 'bg-primary/15 text-primary' : 'bg-muted text-muted-foreground'"
            >
              {{ statusLabel(msg.status) }}
            </span>
          </div>
        </div>

        <!-- 展开的详情 -->
        <div v-if="expandedId === msg.id && msg.content" class="mt-4 border-t border-border pt-4">
          <p class="whitespace-pre-wrap text-sm text-foreground">{{ msg.content }}</p>
        </div>

        <!-- 操作 -->
        <div class="mt-3 flex items-center gap-2">
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
  </LayoutSectionContainer>
</template>
