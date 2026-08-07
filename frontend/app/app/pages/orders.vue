<script setup lang="ts">
definePageMeta({
  layout: 'account',
  middleware: 'auth',
})
import { computed, ref, watch } from 'vue';
import { XIcon } from '@/plugins/xicon';
import { useListOrders } from '@/api/composables';
import { useAccessStore } from '@/stores/modules/core/access.state';

const { t } = useI18n();
const localePath = useLocalePath();
const route = useRoute();
const router = useRouter();

useHead({ title: t('mall.orders.title') });

const accessStore = useAccessStore();

const isLogin = computed(() => {
  const token = accessStore.accessToken;
  return !!token?.value && !accessStore.loginExpired;
});

type OrderStatus =
  | 'STATUS_UNSPECIFIED'
  | 'PENDING_PAYMENT'
  | 'PAID'
  | 'CANCELLED'
  | 'FULFILLED'
  | 'CLOSED';

type OrderEntity = {
  id?: number;
  status?: OrderStatus;
  currency?: string;
  totalAmount?: number;
  createdAt?: string;
};

// ---------- 状态筛选 Tab ----------
// 注：订单按 user_id 过滤由后端 UserPrivacy 策略强制注入（钉定为当前登录用户），
// 前端无需、也不应再带 userId。状态过滤走后端 DSL（order.status 可过滤，枚举值
// 与前端串一致），故切换 Tab 会触发服务端重新分页请求。
type FilterKey = 'all' | OrderStatus;
const FILTER_TABS: FilterKey[] = [
  'all',
  'PENDING_PAYMENT',
  'PAID',
  'FULFILLED',
  'CLOSED',
  'CANCELLED',
];
const activeFilter = ref<FilterKey>('all');
// 支持 URL ?status=xxx 预选筛选项（来自个人中心等入口的快捷跳转）。
// 非法值回退到 'all'，避免恶意/错误参数。
function normalizeFilter(v: unknown): FilterKey {
  return (FILTER_TABS as string[]).includes(v as string) ? (v as FilterKey) : 'all';
}
const initialStatus = route.query.status;
if (initialStatus) activeFilter.value = normalizeFilter(initialStatus);
// 切换 Tab 时同步到 URL query（replace，不入历史栈），便于刷新保持与分享。
watch(activeFilter, (v) => {
  const query = { ...route.query };
  if (v === 'all') delete query.status;
  else query.status = v;
  router.replace({ query });
  page.value = 1;
});

const PAGE_SIZE = 10;
const page = ref(1);
// 查询参数随 activeFilter / page 响应式变化。
const ordersQuery = useListOrders(
  computed(() => {
    const base: Record<string, unknown> = {};
    if (activeFilter.value !== 'all') {
      base.status = activeFilter.value;
    }
    return {
      page: page.value,
      pageSize: PAGE_SIZE,
      noPaging: false,
      sorting: [{ field: 'id', direction: 'DESC' }],
      query: JSON.stringify(base),
    };
  }),
  { enabled: isLogin },
);
const orders = computed<OrderEntity[]>(() => {
  const items = (ordersQuery.data?.value as any)?.items ?? [];
  return (items as OrderEntity[]) ?? [];
});
const ordersLoading = computed(() => ordersQuery.isPending.value);
const ordersError = computed(() => ordersQuery.isError.value);

// 总条数与总页数（取自后端响应的 total 字段）。
const totalCount = computed(() => (ordersQuery.data?.value as any)?.total ?? 0);
const totalPages = computed(() => Math.max(1, Math.ceil(totalCount.value / PAGE_SIZE)));

const STATUS_LABEL_KEY: Record<OrderStatus, string> = {
  STATUS_UNSPECIFIED: 'orderStatus.status_unspecified',
  PENDING_PAYMENT: 'orderStatus.pending_payment',
  PAID: 'orderStatus.paid',
  CANCELLED: 'orderStatus.cancelled',
  FULFILLED: 'orderStatus.fulfilled',
  CLOSED: 'orderStatus.closed',
};

const STATUS_TAG_CLASS: Record<OrderStatus, string> = {
  STATUS_UNSPECIFIED: 'bg-muted text-muted-foreground',
  PENDING_PAYMENT: 'bg-amber-500/15 text-amber-600 dark:text-amber-400',
  PAID: 'bg-emerald-500/15 text-emerald-600 dark:text-emerald-400',
  CANCELLED: 'bg-destructive/15 text-destructive',
  FULFILLED: 'bg-sky-500/15 text-sky-600 dark:text-sky-400',
  CLOSED: 'bg-muted text-muted-foreground',
};

function statusLabel(s: OrderStatus | undefined): string {
  const key = STATUS_LABEL_KEY[s ?? 'STATUS_UNSPECIFIED'];
  return t('mall.' + key);
}
function statusTagClass(s: OrderStatus | undefined): string {
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

function displayTotal(order: OrderEntity): string {
  const currency = order.currency === 'CNY' ? t('mall.product.currencyCny') : '';
  return currency + (order.totalAmount ?? 0);
}

function filterLabel(key: FilterKey): string {
  if (key === 'all') return t('orders.filter.all');
  // 复用 mall.orderStatus.* 的状态文案
  const statusKey = STATUS_LABEL_KEY[key as OrderStatus];
  return t('mall.' + statusKey);
}

// 分页跳转：超出范围时夹紧到 [1, totalPages]，并触发服务端重新请求。
function goToPage(p: number) {
  const clamped = Math.min(Math.max(1, p), totalPages.value);
  if (clamped !== page.value) {
    page.value = clamped;
  }
}
</script>

<template>
  <LayoutPageHero
    :title="t('mall.orders.title')"
    :description="t('mall.orders.subtitle')"
    icon="carbon:document"
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
    <div v-else-if="ordersLoading" class="overflow-hidden rounded-2xl border border-border bg-card">
      <div class="border-b border-border bg-muted/40 px-6 py-3">
        <UiSkeleton class="h-4 w-48" />
      </div>
      <div v-for="i in 4" :key="i" class="border-b border-border px-6 py-4 last:border-b-0">
        <div class="flex items-center gap-4">
          <UiSkeleton class="h-4 w-10" />
          <UiSkeleton class="h-5 w-16" />
          <UiSkeleton class="h-4 w-14 ml-auto" />
          <UiSkeleton class="h-4 w-24" />
          <UiSkeleton class="h-4 w-8" />
        </div>
      </div>
    </div>

    <!-- 错误态：网络/服务端错误，与"空"区分开，并提供重试 -->
    <UiAppEmpty
      v-else-if="ordersError"
      variant="error"
    >
      <template #action>
        <UiButton variant="outline" size="sm" @click="ordersQuery.refetch()">
          {{ t('ui.button.retry') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <!-- 已加载：Tab 条 + 内容 -->
    <div v-else class="flex flex-col gap-4">
      <!-- 状态筛选 Tab -->
      <div
        class="flex flex-wrap items-center gap-2 rounded-2xl border border-border bg-card p-3"
      >
        <button
          v-for="key in FILTER_TABS"
          :key="key"
          type="button"
          :class="[
            'rounded-full px-4 py-1.5 text-xs font-medium transition-colors',
            activeFilter === key
              ? 'bg-primary text-primary-foreground'
              : 'text-muted-foreground hover:bg-muted hover:text-foreground',
          ]"
          @click="activeFilter = key"
        >
          {{ filterLabel(key) }}
        </button>
      </div>

      <!-- 空态：当前过滤条件下无订单 -->
      <UiAppEmpty
        v-if="orders.length === 0"
        variant="noData"
        :description="activeFilter === 'all' ? t('orders.empty') : t('orders.emptyFiltered')"
      />

      <!-- 订单列表 -->
      <div v-else class="overflow-x-auto rounded-2xl border border-border bg-card">
        <div class="min-w-[700px]">
        <div class="border-b border-border bg-muted/40 px-6 py-3 text-xs font-medium uppercase tracking-wide text-muted-foreground">
          <div class="grid grid-cols-[80px_1fr_120px_140px_80px] items-center gap-4">
            <span>{{ t('orders.table.id') }}</span>
            <span>{{ t('orders.table.status') }}</span>
            <span class="text-right">{{ t('orders.table.total') }}</span>
            <span class="text-right">{{ t('orders.table.createdAt') }}</span>
            <span class="text-right">{{ t('orders.table.action') }}</span>
          </div>
        </div>

      <NuxtLink
        v-for="order in orders"
        :key="order.id"
        :to="localePath('/orders/' + order.id)"
        class="block border-b border-border px-6 py-4 transition-colors last:border-b-0 hover:bg-muted/30"
      >
        <div class="grid grid-cols-[80px_1fr_120px_140px_80px] items-center gap-4">
          <span class="text-sm tabular-nums text-foreground">#{{ order.id ?? '—' }}</span>

          <span>
            <span
              class="inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-medium"
              :class="statusTagClass(order.status)"
            >
              {{ statusLabel(order.status) }}
            </span>
          </span>

          <span class="text-right text-sm tabular-nums text-muted-foreground">
            {{ displayTotal(order) }}
          </span>
          <span class="text-right text-xs tabular-nums text-muted-foreground">
            {{ formatCreatedAt(order.createdAt) }}
          </span>
          <span class="flex justify-end text-xs text-primary hover:underline">
            {{ t('orders.detail') }}
          </span>
        </div>
      </NuxtLink>
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
