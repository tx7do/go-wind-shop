<script setup lang="ts">
definePageMeta({
  layout: 'account',
  middleware: 'auth',
})
import { computed } from 'vue';
import { useListPaymentRefunds } from '@/api/composables';
import { useAccessStore } from '@/stores/modules/core/access.state';
import { useUserStore } from '@/stores/modules/core/user.state';

const { t } = useI18n();

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
  reason?: string;
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
type BadgeTone = 'neutral' | 'primary' | 'success' | 'warning' | 'danger';
const STATUS_TONE: Record<RefundStatus, BadgeTone> = {
  PENDING: 'warning',
  SUCCEEDED: 'success',
  FAILED: 'danger',
  STATUS_UNSPECIFIED: 'neutral',
};

function statusLabel(s: RefundStatus | undefined): string {
  return t(STATUS_LABEL_KEY[s ?? 'STATUS_UNSPECIFIED']);
}
function statusTone(s: RefundStatus | undefined): BadgeTone {
  return STATUS_TONE[s ?? 'STATUS_UNSPECIFIED'];
}

function formatCreatedAt(ts: string | undefined): string {
  if (!ts) return '—';
  try {
    return new Date(ts).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' });
  } catch {
    return '—';
  }
}

// 退款金额单位「分」，uint64 经 JSON 下成字符串，先 Number 再交给价格标签。
function toFen(v: number | undefined): number {
  const fen = Number(v ?? 0);
  return Number.isFinite(fen) ? fen : 0;
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
    <UiAuthGate v-if="!isLogin" />

    <UiListSkeleton v-else-if="refundsLoading" :rows="5" variant="row" />

    <!-- 错误态：网络/服务端错误，与"空"区分开，并提供重试 -->
    <UiAppEmpty v-else-if="refundsError" variant="error">
      <template #action>
        <UiButton variant="outline" size="sm" @click="refundsQuery.refetch()">
          {{ t('ui.button.retry') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <UiAppEmpty
      v-else-if="refunds.length === 0"
      variant="noData"
      :description="t('refunds.empty')"
    />

    <template v-else>
      <div class="overflow-hidden rounded-lg border border-border bg-card shadow-card">
        <div
          class="hidden grid-cols-[1fr_140px_170px] items-center gap-4 border-b border-border px-5 py-2.5 text-xs font-medium text-muted-foreground md:grid"
        >
          <span>{{ t('refunds.table.status') }}</span>
          <span class="text-right">{{ t('refunds.table.amount') }}</span>
          <span class="text-right">{{ t('refunds.table.createdAt') }}</span>
        </div>

        <div
          v-for="refund in refunds"
          :key="refund.id"
          class="grid grid-cols-[1fr_auto] items-center gap-x-3 gap-y-1.5 border-b border-border px-4 py-3.5 last:border-b-0 md:grid-cols-[1fr_140px_170px] md:gap-x-4 md:px-5"
        >
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-x-2 gap-y-1">
              <span class="text-sm font-semibold tabular-nums text-foreground">
                #{{ refund.id ?? '—' }}
              </span>
              <UiStatusBadge size="sm" dot :tone="statusTone(refund.status)" :label="statusLabel(refund.status)" />
            </div>
            <p class="mt-1 truncate text-[11px] text-muted-foreground md:text-xs">
              {{ t('refunds.table.transactionId') }} #{{ refund.transactionId ?? '—' }}
            </p>
            <p
              v-if="refund.reason"
              class="mt-0.5 line-clamp-2 break-words text-[11px] text-muted-foreground md:text-xs"
            >
              {{ t('refunds.table.reason') }}：{{ refund.reason }}
            </p>
          </div>

          <div class="flex justify-end">
            <UiPriceTag v-if="toFen(refund.amount) > 0" :amount="toFen(refund.amount)" size="md" />
            <span v-else class="text-sm tabular-nums text-muted-foreground">¥0.00</span>
          </div>

          <p class="col-span-2 text-[11px] tabular-nums text-muted-foreground md:col-span-1 md:text-right md:text-xs">
            {{ formatCreatedAt(refund.createdAt) }}
          </p>
        </div>
      </div>

      <UiPager
        :page="page"
        :total-pages="totalPages"
        class="mt-5"
        @update:page="goToPage"
      />
    </template>
  </LayoutSectionContainer>
</template>
