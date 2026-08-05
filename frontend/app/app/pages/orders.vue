<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { XIcon } from '@/plugins/xicon';
import { useListOrders } from '@/api/composables';
import { useAccessStore } from '@/stores/modules/core/access.state';
import { useUserStore } from '@/stores/modules/core/user.state';

const { t } = useI18n();
const localePath = useLocalePath();
const route = useRoute();
const router = useRouter();

useHead({ title: t('mall.orders.title') });

const accessStore = useAccessStore();
const userStore = useUserStore();

const isLogin = computed(() => {
  const token = accessStore.accessToken;
  return !!token?.value && !accessStore.loginExpired;
});
const currentUserId = computed(() => userStore.user?.id ?? 0);

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

const ordersQuery = useListOrders(
  computed(() => ({
    page: 1,
    pageSize: 50,
    noPaging: false,
    sorting: [{ field: 'id', direction: 'DESC' }],
    query: JSON.stringify({ userId: currentUserId.value }),
  })),
);
const orders = computed<OrderEntity[]>(() => {
  const items = (ordersQuery.data?.value as any)?.items ?? [];
  return (items as OrderEntity[]) ?? [];
});
const ordersLoading = computed(() => ordersQuery.isLoading.value);

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

// ---------- 状态筛选 Tab ----------
// 订单数据全量拉取当前用户的订单（pageSize 50），筛选项在前端过滤，
// 既不依赖后端 query DSL 对 status 的支持，切换也即时无网络抖动。
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
});
function filterLabel(key: FilterKey): string {
  if (key === 'all') return t('orders.filter.all');
  // 复用 mall.orderStatus.* 的状态文案
  const statusKey = STATUS_LABEL_KEY[key as OrderStatus];
  return t('mall.' + statusKey);
}
const filteredOrders = computed<OrderEntity[]>(() => {
  if (activeFilter.value === 'all') return orders.value;
  return orders.value.filter((o) => o.status === activeFilter.value);
});
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

    <!-- 已加载：Tab 条 + 内容 -->
    <div v-else class="flex flex-col gap-4">
      <!-- 状态筛选 Tab（仅有订单时显示） -->
      <div
        v-if="orders.length > 0"
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

      <!-- 全空（一个订单都没有） -->
      <div
        v-if="orders.length === 0"
        class="rounded-2xl border border-border bg-card p-16 text-center"
      >
        <XIcon icon="carbon:document" :size="48" class="mx-auto mb-4 text-muted-foreground" />
        <p class="text-lg text-muted-foreground">{{ t('orders.empty') }}</p>
        <UiButton variant="outline" class="mt-6" @click="navigateTo(localePath('/'))">
          {{ t('cart.continueShopping') }}
        </UiButton>
      </div>

      <!-- 筛选后无结果 -->
      <div
        v-else-if="filteredOrders.length === 0"
        class="rounded-2xl border border-border bg-card p-16 text-center"
      >
        <XIcon icon="carbon:document" :size="48" class="mx-auto mb-4 text-muted-foreground" />
        <p class="text-lg text-muted-foreground">{{ t('orders.emptyFiltered') }}</p>
        <UiButton variant="outline" class="mt-6" @click="activeFilter = 'all'">
          {{ t('orders.filter.all') }}
        </UiButton>
      </div>

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
        v-for="order in filteredOrders"
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
    </div>
  </LayoutSectionContainer>
</template>
