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

useHead({ title: t('orders.title') });

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

type BadgeTone = 'neutral' | 'primary' | 'success' | 'warning' | 'danger';
const STATUS_TONE: Record<OrderStatus, BadgeTone> = {
  STATUS_UNSPECIFIED: 'neutral',
  PENDING_PAYMENT: 'warning',
  PAID: 'primary',
  FULFILLED: 'success',
  CANCELLED: 'danger',
  CLOSED: 'neutral',
};

function statusLabel(s: OrderStatus | undefined): string {
  const key = STATUS_LABEL_KEY[s ?? 'STATUS_UNSPECIFIED'];
  return t('mall.' + key);
}
function statusTone(s: OrderStatus | undefined): BadgeTone {
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

// 后端金额为最小货币单位（分），且 uint64 经 JSON 是字符串，先 Number 再交给价格标签。
function orderTotalFen(order: OrderEntity): number {
  const fen = Number(order.totalAmount ?? 0);
  return Number.isFinite(fen) ? fen : 0;
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

// pages/orders/[id].vue 是 pages/orders.vue 的子路由（Nuxt 按同名目录嵌套），
// 父级不透出 NuxtPage 时订单详情永远渲染不出来，这里按有无 id 二选一。
const isDetailRoute = computed(() => route.params.id !== undefined);
</script>

<template>
  <NuxtPage v-if="isDetailRoute" />

  <LayoutPageHero
    v-if="!isDetailRoute"
    :title="t('orders.title')"
    :description="t('orders.subtitle')"
    icon="carbon:document"
    size="sm"
  />

  <LayoutSectionContainer v-if="!isDetailRoute">
    <UiAuthGate v-if="!isLogin" />

    <template v-else>
      <!-- 状态筛选：窄屏横向滚动，不做换行挤压 -->
      <div class="-mx-4 mb-4 overflow-x-auto px-4 md:mx-0 md:mb-6 md:px-0">
        <div class="flex w-max items-center gap-1.5 rounded-full border border-border bg-card p-1">
          <button
            v-for="key in FILTER_TABS"
            :key="key"
            type="button"
            :class="[
              'whitespace-nowrap rounded-full px-3.5 py-1.5 text-xs font-medium transition-colors',
              activeFilter === key
                ? 'bg-primary text-primary-foreground'
                : 'text-muted-foreground hover:bg-muted hover:text-foreground',
            ]"
            @click="activeFilter = key"
          >
            {{ filterLabel(key) }}
          </button>
        </div>
      </div>

      <UiListSkeleton v-if="ordersLoading" :rows="6" variant="row" />

      <!-- 错误态：网络/服务端错误，与"空"区分开，并提供重试 -->
      <UiAppEmpty v-else-if="ordersError" variant="error">
        <template #action>
          <UiButton variant="outline" size="sm" @click="ordersQuery.refetch()">
            {{ t('ui.button.retry') }}
          </UiButton>
        </template>
      </UiAppEmpty>

      <!-- 空态：当前过滤条件下无订单 -->
      <UiAppEmpty
        v-else-if="orders.length === 0"
        variant="noData"
        :description="activeFilter === 'all' ? t('orders.empty') : t('orders.emptyFiltered')"
      />

      <template v-else>
        <!-- 移动端：卡片 -->
        <div class="flex flex-col gap-2.5 md:hidden">
          <NuxtLink
            v-for="order in orders"
            :key="order.id"
            :to="localePath('/orders/' + order.id)"
            class="rounded-lg border border-border bg-card p-3.5 shadow-card transition-colors active:bg-muted"
          >
            <div class="flex items-center justify-between gap-2">
              <span class="text-sm font-semibold tabular-nums text-foreground">
                #{{ order.id ?? '—' }}
              </span>
              <UiStatusBadge
                size="sm"
                dot
                :tone="statusTone(order.status)"
                :label="statusLabel(order.status)"
              />
            </div>

            <div class="mt-3 flex items-end justify-between gap-3">
              <div class="min-w-0">
                <p class="text-[11px] leading-none text-muted-foreground">
                  {{ t('orders.table.createdAt') }}
                </p>
                <p class="mt-1 truncate text-xs leading-tight tabular-nums text-muted-foreground">
                  {{ formatCreatedAt(order.createdAt) }}
                </p>
              </div>
              <UiPriceTag v-if="orderTotalFen(order) > 0" :amount="orderTotalFen(order)" size="md" class="shrink-0" />
              <span v-else class="shrink-0 text-sm tabular-nums text-muted-foreground">¥0.00</span>
            </div>

            <div class="mt-3 flex items-center justify-between border-t border-border pt-2.5">
              <span class="text-xs font-medium text-primary">{{ t('orders.detail') }}</span>
              <XIcon icon="lucide:chevron-right" :size="14" class="text-muted-foreground" />
            </div>
          </NuxtLink>
        </div>

        <!-- 桌面端：行式列表 -->
        <div class="hidden overflow-hidden rounded-lg border border-border bg-card shadow-card md:block">
          <div
            class="grid grid-cols-[110px_1fr_140px_170px_70px] items-center gap-4 border-b border-border px-5 py-2.5 text-xs font-medium text-muted-foreground"
          >
            <span>{{ t('orders.table.id') }}</span>
            <span>{{ t('orders.table.status') }}</span>
            <span class="text-right">{{ t('orders.table.total') }}</span>
            <span class="text-right">{{ t('orders.table.createdAt') }}</span>
            <span class="text-right">{{ t('orders.table.action') }}</span>
          </div>

          <NuxtLink
            v-for="order in orders"
            :key="order.id"
            :to="localePath('/orders/' + order.id)"
            class="grid grid-cols-[110px_1fr_140px_170px_70px] items-center gap-4 border-b border-border px-5 py-3.5 transition-colors last:border-b-0 hover:bg-muted"
          >
            <span class="text-sm font-medium tabular-nums text-foreground">
              #{{ order.id ?? '—' }}
            </span>
            <span>
              <UiStatusBadge
                size="sm"
                dot
                :tone="statusTone(order.status)"
                :label="statusLabel(order.status)"
              />
            </span>
            <span class="flex justify-end">
              <UiPriceTag v-if="orderTotalFen(order) > 0" :amount="orderTotalFen(order)" size="sm" />
              <span v-else class="text-sm tabular-nums text-muted-foreground">¥0.00</span>
            </span>
            <span class="text-right text-xs tabular-nums text-muted-foreground">
              {{ formatCreatedAt(order.createdAt) }}
            </span>
            <span class="flex items-center justify-end gap-0.5 text-xs font-medium text-primary">
              {{ t('orders.detail') }}
              <XIcon icon="lucide:chevron-right" :size="13" />
            </span>
          </NuxtLink>
        </div>

        <UiPager
          :page="page"
          :total-pages="totalPages"
          class="mt-5"
          @update:page="goToPage"
        />
      </template>
    </template>
  </LayoutSectionContainer>
</template>
