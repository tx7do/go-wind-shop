<script setup lang="ts">
import { computed } from 'vue';
import { XIcon } from '@/plugins/xicon';
import { useGetOrder, useListOrderItems } from '@/api/composables';
import { useAccessStore } from '@/stores/modules/core/access.state';
import { useUserStore } from '@/stores/modules/core/user.state';

const route = useRoute();
const { t } = useI18n();
const localePath = useLocalePath();

useHead({ title: t('mall.orderDetail.title') });

const accessStore = useAccessStore();
const userStore = useUserStore();

const isLogin = computed(() => {
  const token = accessStore.accessToken;
  return !!token?.value && !accessStore.loginExpired;
});

const orderId = computed(() => Number(route.params.id));

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
  recipientName?: string;
  recipientPhone?: string;
  shippingAddress?: string;
  createdAt?: string;
};

type OrderItemEntity = {
  id?: number;
  orderId?: number;
  skuId?: number;
  quantity?: number;
  unitPrice?: number;
  subtotal?: number;
};

const orderQuery = useGetOrder(orderId.value, {
  enabled: computed(() => isLogin.value && orderId.value > 0),
});
const order = computed<OrderEntity | undefined>(() => {
  return orderQuery.data?.value as OrderEntity | undefined;
});
const orderLoading = computed(() => orderQuery.isLoading.value);

const orderItemsQuery = useListOrderItems(
  computed(() => ({
    page: 1,
    pageSize: 100,
    noPaging: false,
    query:
      orderId.value > 0 ? JSON.stringify({ orderId: orderId.value }) : undefined,
  })),
);
const orderItems = computed<OrderItemEntity[]>(() => {
  const items = (orderItemsQuery.data?.value as any)?.items ?? [];
  return (items as OrderItemEntity[]) ?? [];
});
const itemsLoading = computed(() => orderItemsQuery.isLoading.value);

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
  return t('mall.' + STATUS_LABEL_KEY[s ?? 'STATUS_UNSPECIFIED']);
}
function statusTagClass(s: OrderStatus | undefined): string {
  return STATUS_TAG_CLASS[s ?? 'STATUS_UNSPECIFIED'];
}

function formatCreatedAt(ts: string | undefined): string {
  if (!ts) return '—';
  try {
    return new Date(ts).toLocaleString();
  } catch {
    return '—';
  }
}

function displayAmount(v: number | undefined, currency?: string): string {
  const prefix = currency === 'CNY' ? t('mall.product.currencyCny') : '';
  return prefix + (v ?? 0);
}

const orderTotalLabel = computed(() => displayAmount(order.value?.totalAmount, order.value?.currency));
</script>

<template>
  <LayoutPageHero
    :title="t('mall.orderDetail.title')"
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

    <!-- 加载中 / 不存在 -->
    <div v-else-if="orderLoading" class="rounded-2xl border border-border bg-card p-8">
      <UiSkeleton class="mb-4 h-24 w-full" />
      <UiSkeleton class="h-24 w-full" />
    </div>

    <div
      v-else-if="!order"
      class="rounded-2xl border border-border bg-card p-16 text-center text-muted-foreground"
    >
      {{ t('orderDetail.notFound') }}
    </div>

    <!-- 订单详情 -->
    <div v-else class="flex flex-col gap-6">
      <!-- 状态摘要 -->
      <div class="rounded-2xl border border-border bg-card p-6">
        <div class="flex flex-wrap items-center justify-between gap-4">
          <div>
            <p class="text-xs text-muted-foreground">{{ t('orderDetail.orderId') }}</p>
            <p class="mt-1 text-lg font-bold text-foreground">#{{ order.id ?? '—' }}</p>
          </div>
          <div class="text-right">
            <p class="text-xs text-muted-foreground">{{ t('orderDetail.status') }}</p>
            <span
              class="mt-1 inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-medium"
              :class="statusTagClass(order.status)"
            >
              {{ statusLabel(order.status) }}
            </span>
          </div>
          <div class="text-right">
            <p class="text-xs text-muted-foreground">{{ t('orderDetail.createdAt') }}</p>
            <p class="mt-1 text-sm tabular-nums text-foreground">{{ formatCreatedAt(order.createdAt) }}</p>
          </div>
          <div class="text-right">
            <p class="text-xs text-muted-foreground">{{ t('orderDetail.total') }}</p>
            <p class="mt-1 text-lg font-bold text-primary">{{ orderTotalLabel }}</p>
          </div>
        </div>
      </div>

      <!-- 收货信息 -->
      <div class="rounded-2xl border border-border bg-card p-6">
        <h2 class="mb-4 text-base font-bold text-foreground">{{ t('orderDetail.recipientTitle') }}</h2>
        <div class="grid gap-4 sm:grid-cols-3">
          <div class="rounded-md border border-border bg-background/40 p-3">
            <p class="text-[10px] uppercase tracking-wide text-muted-foreground">
              {{ t('orderDetail.recipientName') }}
            </p>
            <p class="mt-1 text-sm text-foreground">{{ order.recipientName || '—' }}</p>
          </div>
          <div class="rounded-md border border-border bg-background/40 p-3">
            <p class="text-[10px] uppercase tracking-wide text-muted-foreground">
              {{ t('orderDetail.recipientPhone') }}
            </p>
            <p class="mt-1 text-sm text-foreground">{{ order.recipientPhone || '—' }}</p>
          </div>
          <div class="rounded-md border border-border bg-background/40 p-3">
            <p class="text-[10px] uppercase tracking-wide text-muted-foreground">
              {{ t('orderDetail.shippingAddress') }}
            </p>
            <p class="mt-1 text-sm text-foreground">{{ order.shippingAddress || '—' }}</p>
          </div>
        </div>
      </div>

      <!-- 订单项 -->
      <div class="overflow-hidden rounded-2xl border border-border bg-card">
        <div class="border-b border-border bg-muted/40 px-6 py-3 text-xs font-medium uppercase tracking-wide text-muted-foreground">
          <div class="grid grid-cols-[1fr_100px_100px_120px] items-center gap-4">
            <span>{{ t('orderDetail.table.item') }}</span>
            <span class="text-right">{{ t('orderDetail.table.unitPrice') }}</span>
            <span class="text-right">{{ t('orderDetail.table.quantity') }}</span>
            <span class="text-right">{{ t('orderDetail.table.subtotal') }}</span>
          </div>
        </div>

        <UiSkeleton v-if="itemsLoading" class="m-6 h-12 w-full" />

        <div
          v-else-if="orderItems.length === 0"
          class="px-6 py-10 text-center text-sm text-muted-foreground"
        >
          {{ t('orderDetail.noItems') }}
        </div>

        <div
          v-for="item in orderItems"
          :key="item.id"
          class="border-b border-border px-6 py-4 last:border-b-0"
        >
          <div class="grid grid-cols-[1fr_100px_100px_120px] items-center gap-4">
            <div class="flex items-center gap-3">
              <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded bg-primary/5 text-lg">
                📦
              </div>
              <span class="line-clamp-1 text-xs text-foreground">
                {{ t('orderDetail.skuId') }}#{{ item.skuId ?? '—' }}
              </span>
            </div>
            <span class="text-right text-xs tabular-nums text-muted-foreground">
              {{ displayAmount(item.unitPrice, order.currency) }}
            </span>
            <span class="text-right text-xs tabular-nums text-muted-foreground">
              ×{{ item.quantity ?? 0 }}
            </span>
            <span class="text-right text-xs tabular-nums text-muted-foreground">
              {{ displayAmount(item.subtotal, order.currency) }}
            </span>
          </div>
        </div>
      </div>

      <div class="flex justify-center">
        <UiButton variant="outline" @click="navigateTo(localePath('/orders'))">
          {{ t('orderDetail.backToList') }}
        </UiButton>
      </div>
    </div>
  </LayoutSectionContainer>
</template>
