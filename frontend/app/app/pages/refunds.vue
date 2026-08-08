<script setup lang="ts">
definePageMeta({
  layout: 'account',
  middleware: 'auth',
})
import { computed } from 'vue';
import { XIcon } from '@/plugins/xicon';
import { useListPaymentRefunds } from '@/api/composables';
import { useAccessStore } from '@/stores/modules/core/access.state';
import { useUserStore } from '@/stores/modules/core/user.state';

const { t } = useI18n();
const localePath = useLocalePath();

useHead({ title: t('refunds.title') });

const accessStore = useAccessStore();
const userStore = useUserStore();

const isLogin = computed(() => {
  const token = accessStore.accessToken;
  return !!token?.value && !accessStore.loginExpired;
});
// 退款单没有 userId 字段，但携带 OperatorID mixin 的 created_by（创建者ID）。
// 买家发起的退款 created_by 即其本人，故按 createdBy 过滤等价于“我的退款”。
const currentUserId = computed(() => userStore.user?.id ?? 0);

type RefundStatus = 'PENDING' | 'SUCCEEDED' | 'FAILED' | 'STATUS_UNSPECIFIED';
type RefundEntity = {
  id?: number;
  transactionId?: number;
  amount?: number;
  currency?: string;
  status?: RefundStatus;
  createdAt?: string;
};

// 退款单按 created_by 过滤（创建者即本人）。退款实体无 UserPrivacy 策略，
// 故 created_by 是该接口唯一的用户行级隔离条件 —— 必须带且不可由客户端绕过。
// enabled 守卫：未登录 / currentUserId 尚未 hydrate（=0）时不发请求，
// 避免预 hydrate 窗口以 createdBy:0 发出无过滤请求拉到他人退款。
const PAGE_SIZE = 10;
const page = ref(1);
const refundsQuery = useListPaymentRefunds(
  computed(() => ({
    page: page.value,
    pageSize: PAGE_SIZE,
    noPaging: false,
    sorting: [{ field: 'id', direction: 'DESC' }],
    query: JSON.stringify({ createdBy: currentUserId.value }),
  })),
  { enabled: computed(() => isLogin.value && currentUserId.value > 0) },
);
const refunds = computed<RefundEntity[]>(() => {
  const items = (refundsQuery.data?.value as any)?.items ?? [];
  return (items as RefundEntity[]) ?? [];
});
const refundsLoading = computed(() => refundsQuery.isPending.value);
const refundsError = computed(() => refundsQuery.isError.value);

const totalCount = computed(() => (refundsQuery.data?.value as any)?.total ?? 0);
const totalPages = computed(() => Math.max(1, Math.ceil(totalCount.value / PAGE_SIZE)));

function goToPage(p: number) {
  const clamped = Math.min(Math.max(1, p), totalPages.value);
  if (clamped !== page.value) {
    page.value = clamped;
  }
}

const STATUS_LABEL_KEY: Record<RefundStatus, string> = {
  PENDING: 'refunds.status.pending',
  SUCCEEDED: 'refunds.status.succeeded',
  FAILED: 'refunds.status.failed',
  STATUS_UNSPECIFIED: 'refunds.status.status_unspecified',
};
const STATUS_TAG_CLASS: Record<RefundStatus, string> = {
  PENDING: 'bg-amber-500/15 text-amber-600 dark:text-amber-400',
  SUCCEEDED: 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400',
  FAILED: 'bg-destructive/15 text-destructive',
  STATUS_UNSPECIFIED: 'bg-muted text-muted-foreground',
};

function statusLabel(s: RefundStatus | undefined): string {
  return t(STATUS_LABEL_KEY[s ?? 'STATUS_UNSPECIFIED']);
}
function statusTagClass(s: RefundStatus | undefined): string {
  return STATUS_TAG_CLASS[s ?? 'STATUS_UNSPECIFIED'];
}

function formatCreatedAt(ts: string | undefined): string {
  if (!ts) return '—';
  try {
    return new Date(ts).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' });
  } catch {
    return '—';
  }
}

function displayAmount(v: number | undefined, currency?: string): string {
  const prefix = currency === 'CNY' ? t('mall.product.currencyCny') : '';
  return prefix + (v ?? 0);
}
</script>

<template>
  <LayoutPageHero
    :title="t('refunds.title')"
    :description="t('refunds.subtitle')"
    icon="lucide:rotate-ccw"
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
    <div v-else-if="refundsLoading" class="overflow-hidden rounded-2xl border border-border bg-card">
      <div class="border-b border-border bg-muted/40 px-6 py-3">
        <UiSkeleton class="h-4 w-48" />
      </div>
      <div v-for="i in 4" :key="i" class="border-b border-border px-6 py-4 last:border-b-0">
        <div class="flex items-center gap-4">
          <UiSkeleton class="h-4 w-10" />
          <UiSkeleton class="h-5 w-20" />
          <UiSkeleton class="h-4 w-14 ml-auto" />
          <UiSkeleton class="h-4 w-24" />
        </div>
      </div>
    </div>

    <!-- 错误态：网络/服务端错误，与"空"区分开，并提供重试 -->
    <UiAppEmpty
      v-else-if="refundsError"
      variant="error"
    >
      <template #action>
        <UiButton variant="outline" size="sm" @click="refundsQuery.refetch()">
          {{ t('ui.button.retry') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <!-- 已加载：空 / 列表 -->
    <div v-else class="flex flex-col gap-4">
      <!-- 空态 -->
      <UiAppEmpty
        v-if="refunds.length === 0"
        variant="noData"
        :description="t('refunds.empty')"
      />

      <!-- 退款列表 -->
      <div v-else class="overflow-x-auto rounded-2xl border border-border bg-card">
        <div class="min-w-[640px]">
          <div class="border-b border-border bg-muted/40 px-6 py-3 text-xs font-medium uppercase tracking-wide text-muted-foreground">
            <div class="grid grid-cols-[80px_1fr_120px_140px] items-center gap-4">
              <span>{{ t('refunds.table.id') }}</span>
              <span>{{ t('refunds.table.status') }}</span>
              <span class="text-right">{{ t('refunds.table.amount') }}</span>
              <span class="text-right">{{ t('refunds.table.createdAt') }}</span>
            </div>
          </div>

          <div
            v-for="refund in refunds"
            :key="refund.id"
            class="border-b border-border px-6 py-4 last:border-b-0"
          >
            <div class="grid grid-cols-[80px_1fr_120px_140px] items-center gap-4">
              <span class="text-sm tabular-nums text-foreground">#{{ refund.id ?? '—' }}</span>

              <div class="flex items-center gap-2">
                <span
                  class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-medium"
                  :class="statusTagClass(refund.status)"
                >
                  {{ statusLabel(refund.status) }}
                </span>
                <span class="text-[10px] text-muted-foreground">
                  {{ t('refunds.table.transactionId') }} #{{ refund.transactionId ?? '—' }}
                </span>
              </div>

              <span class="text-right text-sm tabular-nums text-foreground">
                {{ displayAmount(refund.amount, refund.currency) }}
              </span>
              <span class="text-right text-xs tabular-nums text-muted-foreground">
                {{ formatCreatedAt(refund.createdAt) }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- 分页：仅当超过一页时显示 -->
      <div
        v-if="totalPages > 1"
        class="flex items-center justify-center gap-4 py-2"
      >
        <UiButton
          variant="outline"
          size="sm"
          :disabled="page <= 1"
          @click="goToPage(page - 1)"
        >
          {{ t('ui.pagination.previous') }}
        </UiButton>
        <span class="text-xs tabular-nums text-muted-foreground">
          {{ t('ui.pagination.page', { current: page, total: totalPages }) }}
        </span>
        <UiButton
          variant="outline"
          size="sm"
          :disabled="page >= totalPages"
          @click="goToPage(page + 1)"
        >
          {{ t('ui.pagination.next') }}
        </UiButton>
      </div>
    </div>
  </LayoutSectionContainer>
</template>
