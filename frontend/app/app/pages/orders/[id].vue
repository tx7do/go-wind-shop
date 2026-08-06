<script setup lang="ts">
import { computed } from 'vue';
import { toast } from 'vue-sonner';
import { XIcon } from '@/plugins/xicon';
import {
  useGetOrder,
  useListOrderItems,
  useUpdateOrder,
  useCreatePaymentTransaction,
  useListPaymentTransactions,
  useCreatePaymentRefund,
  useListPaymentRefunds,
} from '@/api/composables';
import { queryClient } from '@/plugins/vue-query';
import { useAccessStore } from '@/stores/modules/core/access.state';
import { useUserStore } from '@/stores/modules/core/user.state';
import type {
  orderservicev1_Order_Status,
  paymentservicev1_PaymentTransaction,
  paymentservicev1_PaymentMethod,
  paymentservicev1_BusinessType,
} from '@/api/generated/app/service/v1';

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
    return new Date(ts).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' });
  } catch {
    return '—';
  }
}

// PII 脱敏：手机号中间四位、地址只保留前段
function maskPhone(phone: string | undefined): string {
  if (!phone) return '—';
  const s = String(phone);
  if (s.length < 7) return s.replace(/\d/g, '*');
  return s.slice(0, 3) + '****' + s.slice(-4);
}
function maskAddress(addr: string | undefined): string {
  if (!addr) return '—';
  return addr.length > 12 ? addr.slice(0, 12) + '…' : addr;
}

function displayAmount(v: number | undefined, currency?: string): string {
  const prefix = currency === 'CNY' ? t('mall.product.currencyCny') : '';
  return prefix + (v ?? 0);
}

const orderTotalLabel = computed(() => displayAmount(order.value?.totalAmount, order.value?.currency));

// ---------- 订单操作 ----------
// 状态机：PENDING_PAYMENT → PAID → FULFILLED → CLOSED；PENDING_PAYMENT → CANCELLED
// 买家可执行：去支付 / 取消（待付款）；申请退款（已付款）；确认收货（已履约）
const orderIdValue = computed(() => order.value?.id ?? 0);

// 该订单的支付流水（用于「申请退款」拿到 transactionId/amount/currency，
// 以及判断订单是否已有成功支付可退款）。
// enabled 守卫：订单未加载完（id=0）时不发起请求，避免无过滤条件拉全租户流水。
const transactionsQuery = useListPaymentTransactions(
  computed(() => ({
    page: 1,
    pageSize: 20,
    noPaging: false,
    query:
      orderIdValue.value > 0 ? JSON.stringify({ orderId: orderIdValue.value }) : undefined,
  })),
  { enabled: computed(() => isLogin.value && orderIdValue.value > 0) },
);
type TxnEntity = {
  id?: number;
  amount?: number;
  currency?: string;
  status?: 'PENDING' | 'SUCCEEDED' | 'FAILED' | 'REFUNDED' | 'STATUS_UNSPECIFIED';
};
const transactions = computed<TxnEntity[]>(() => {
  const items = (transactionsQuery.data?.value as any)?.items ?? [];
  return (items as TxnEntity[]) ?? [];
});
// 取已成功的支付流水作为退款来源（一笔订单正常只有一笔 SUCCEEDED）。
const succeededTxn = computed<TxnEntity | undefined>(() =>
  transactions.value.find((t) => t.status === 'SUCCEEDED'),
);

// 已存在的退款单（按 transactionId 查）。若已有一笔 PENDING/SUCCEEDED 退款，
// 则隐藏「申请退款」按钮，防止重复申请造成多笔退款。
const refundQuery = useListPaymentRefunds(
  computed(() => ({
    page: 1,
    pageSize: 10,
    noPaging: false,
    query: succeededTxn.value?.id
      ? JSON.stringify({ transactionId: succeededTxn.value.id })
      : undefined,
  })),
  { enabled: computed(() => isLogin.value && !!succeededTxn.value?.id) },
);
type RefundEntity = {
  id?: number;
  status?: 'PENDING' | 'SUCCEEDED' | 'FAILED' | 'STATUS_UNSPECIFIED';
};
const existingRefunds = computed<RefundEntity[]>(() => {
  const items = (refundQuery.data?.value as any)?.items ?? [];
  return (items as RefundEntity[]) ?? [];
});
// 是否存在未完结的退款（PENDING 处理中 / SUCCEEDED 已到账）。
const hasActiveRefund = computed(() =>
  existingRefunds.value.some((r) => r.status === 'PENDING' || r.status === 'SUCCEEDED'),
);

const updateOrderMutation = useUpdateOrder();
const payMutation = useCreatePaymentTransaction();
const refundMutation = useCreatePaymentRefund();

const anyActionPending = computed(
  () =>
    updateOrderMutation.isPending.value ||
    payMutation.isPending.value ||
    refundMutation.isPending.value,
);

function refreshOrderAndList() {
  queryClient.invalidateQueries({ queryKey: ['getOrder'] });
  queryClient.invalidateQueries({ queryKey: ['listOrders'] });
  queryClient.invalidateQueries({ queryKey: ['listPaymentTransactions'] });
  queryClient.invalidateQueries({ queryKey: ['listPaymentRefunds'] });
}

// 取消订单：PENDING_PAYMENT → CANCELLED（乐观锁 expectedStatus）
async function handleCancel() {
  if (!order.value?.id) return;
  if (anyActionPending.value) return;
  if (!window.confirm(t('orderDetail.confirm.cancel'))) return;
  try {
    await updateOrderMutation.mutateAsync({
      id: order.value.id,
      values: { status: 'CANCELLED' },
      expectedStatus: ['PENDING_PAYMENT'] as orderservicev1_Order_Status[],
    });
    toast.success(t('orderDetail.result.cancelled'));
    refreshOrderAndList();
  } catch (err: any) {
    toast.error(err?.message || t('orderDetail.errors.cancelFailed'));
  }
}

// 确认收货：FULFILLED → CLOSED
async function handleConfirmReceipt() {
  if (!order.value?.id) return;
  if (anyActionPending.value) return;
  if (!window.confirm(t('orderDetail.confirm.confirmReceipt'))) return;
  try {
    await updateOrderMutation.mutateAsync({
      id: order.value.id,
      values: { status: 'CLOSED' },
      expectedStatus: ['FULFILLED'] as orderservicev1_Order_Status[],
    });
    toast.success(t('orderDetail.result.confirmed'));
    refreshOrderAndList();
  } catch (err: any) {
    toast.error(err?.message || t('orderDetail.errors.confirmFailed'));
  }
}

// 立即支付：订单处于待付款时，创建一笔支付流水（余额支付，与结算页一致）
async function handlePayNow() {
  const o = order.value;
  if (!o?.id) return;
  if (anyActionPending.value) return;
  const cryptoObj = (globalThis as unknown as { crypto?: { randomUUID?: () => string } }).crypto;
  const paymentIdempotencyKey = cryptoObj?.randomUUID?.() ?? '';
  if (!paymentIdempotencyKey) {
    toast.error(t('orderDetail.errors.payFailed'));
    return;
  }
  const paymentData: paymentservicev1_PaymentTransaction = {
    userId: o.userId,
    tenantId: o.tenantId,
    orderId: o.id,
    amount: o.totalAmount ?? 0,
    currency: o.currency || 'CNY',
    paymentMethod: 'BALANCE' as paymentservicev1_PaymentMethod,
    businessType: 'BUSINESS_TYPE_CONSUME' as paymentservicev1_BusinessType,
    idempotencyKey: paymentIdempotencyKey,
    businessRefId: o.businessRefId,
  } as paymentservicev1_PaymentTransaction;
  try {
    await payMutation.mutateAsync(paymentData);
    toast.success(t('orderDetail.result.paid'));
    refreshOrderAndList();
  } catch (err: any) {
    toast.error(err?.message || t('orderDetail.errors.payFailed'));
  }
}

// 申请退款：基于该订单的成功支付流水创建退款单
async function handleRequestRefund() {
  const o = order.value;
  const txn = succeededTxn.value;
  if (!o?.id) return;
  if (anyActionPending.value) return;
  if (!txn?.id) {
    toast.error(t('orderDetail.errors.noTransaction'));
    return;
  }
  if (!window.confirm(t('orderDetail.confirm.requestRefund'))) return;
  const cryptoObj = (globalThis as unknown as { crypto?: { randomUUID?: () => string } }).crypto;
  const refundIdempotencyKey = cryptoObj?.randomUUID?.() ?? '';
  if (!refundIdempotencyKey) {
    toast.error(t('orderDetail.errors.refundFailed'));
    return;
  }
  try {
    await refundMutation.mutateAsync({
      userId: o.userId,
      tenantId: o.tenantId,
      transactionId: txn.id,
      amount: txn.amount ?? o.totalAmount ?? 0,
      currency: txn.currency ?? o.currency ?? 'CNY',
      idempotencyKey: refundIdempotencyKey,
      businessRefId: o.businessRefId,
    } as any);
    toast.success(t('orderDetail.result.refundRequested'));
    refreshOrderAndList();
  } catch (err: any) {
    toast.error(err?.message || t('orderDetail.errors.refundFailed'));
  }
}

// 各操作按钮是否可见（仅依据订单当前状态，终态不显示任何操作）。
// canRefund 额外校验：若已存在未完结退款（PENDING/SUCCEEDED），则隐藏按钮，
// 防止对同一笔支付流水重复发起退款造成多笔退款。
const canPay = computed(() => order.value?.status === 'PENDING_PAYMENT');
const canCancel = computed(() => order.value?.status === 'PENDING_PAYMENT');
const canRefund = computed(
  () => order.value?.status === 'PAID' && !!succeededTxn.value?.id && !hasActiveRefund.value,
);
const canConfirm = computed(() => order.value?.status === 'FULFILLED');
const hasAnyAction = computed(
  () => canPay.value || canCancel.value || canRefund.value || canConfirm.value,
);
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
    <div v-else-if="orderLoading" class="rounded-2xl border border-border bg-card p-6">
      <div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
        <UiSkeleton v-for="i in 4" :key="i" class="h-16 w-full rounded-md" />
      </div>
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
            <p class="mt-1 text-sm tabular-nums text-foreground">{{ maskPhone(order.recipientPhone) }}</p>
          </div>
          <div class="rounded-md border border-border bg-background/40 p-3">
            <p class="text-[10px] uppercase tracking-wide text-muted-foreground">
              {{ t('orderDetail.shippingAddress') }}
            </p>
            <p class="mt-1 text-sm text-foreground">{{ maskAddress(order.shippingAddress) }}</p>
          </div>
        </div>
      </div>

      <!-- 订单项 -->
      <div class="overflow-hidden rounded-2xl border border-border bg-card">
        <div class="border-b border-border bg-muted/40 px-6 py-3 text-xs font-medium uppercase tracking-wide text-muted-foreground">
          <div class="grid grid-cols-[1fr_90px_70px_90px] items-center gap-4">
            <span>{{ t('orderDetail.table.item') }}</span>
            <span class="text-right">{{ t('orderDetail.table.unitPrice') }}</span>
            <span class="text-right">{{ t('orderDetail.table.quantity') }}</span>
            <span class="text-right">{{ t('orderDetail.table.subtotal') }}</span>
          </div>
        </div>

        <div v-if="itemsLoading" class="flex flex-col">
          <div v-for="i in 3" :key="i" class="border-b border-border px-6 py-4 last:border-b-0">
            <div class="flex items-center gap-3">
              <UiSkeleton class="h-10 w-10 shrink-0 rounded" />
              <UiSkeleton class="h-4 flex-1" />
              <UiSkeleton class="h-4 w-12" />
              <UiSkeleton class="h-4 w-8" />
              <UiSkeleton class="h-4 w-12" />
            </div>
          </div>
        </div>

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
          <div class="grid grid-cols-[1fr_90px_70px_90px] items-center gap-4">
            <div class="flex items-center gap-3">
              <UiProductPlaceholder
                :seed="item.skuId ?? 0"
                class="h-10 w-10 shrink-0 rounded text-muted-foreground"
              />
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

      <!-- 订单操作 -->
      <div
        v-if="hasAnyAction"
        class="flex flex-wrap items-center gap-3 rounded-2xl border border-border bg-card p-6"
      >
        <span class="text-sm font-medium text-foreground">{{ t('orderDetail.actions.title') }}</span>
        <div class="flex flex-wrap items-center gap-3">
          <UiButton
            v-if="canPay"
            :disabled="anyActionPending"
            @click="handlePayNow"
          >
            {{ t('orderDetail.actions.payNow') }}
          </UiButton>
          <UiButton
            v-if="canConfirm"
            :disabled="anyActionPending"
            @click="handleConfirmReceipt"
          >
            {{ t('orderDetail.actions.confirmReceipt') }}
          </UiButton>
          <UiButton
            v-if="canRefund"
            variant="outline"
            :disabled="anyActionPending"
            @click="handleRequestRefund"
          >
            {{ t('orderDetail.actions.requestRefund') }}
          </UiButton>
          <UiButton
            v-if="canCancel"
            variant="ghost"
            class="text-destructive hover:text-destructive"
            :disabled="anyActionPending"
            @click="handleCancel"
          >
            {{ t('orderDetail.actions.cancel') }}
          </UiButton>
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
