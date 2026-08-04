<script setup lang="ts">
import { computed, reactive } from 'vue';
import { toast } from 'vue-sonner';
import { XIcon } from '@/plugins/xicon';
import {
  useListCarts,
  useListCartItems,
  useCreateOrder,
  useCreatePaymentTransaction,
  fetchGetOrderByIdempotencyKey,
} from '@/api/composables';
import { useAccessStore } from '@/stores/modules/core/access.state';
import { useUserStore } from '@/stores/modules/core/user.state';
import { queryClient } from '@/plugins/vue-query';
import type {
  orderservicev1_Order,
  paymentservicev1_PaymentTransaction,
  paymentservicev1_PaymentMethod,
  paymentservicev1_BusinessType,
} from '@/api/generated/app/service/v1';

const { t } = useI18n();
const localePath = useLocalePath();

useHead({ title: t('mall.checkout.title') });

const accessStore = useAccessStore();
const userStore = useUserStore();

const isLogin = computed(() => {
  const token = accessStore.accessToken;
  return !!token?.value && !accessStore.loginExpired;
});
const currentUserId = computed(() => userStore.user?.id ?? 0);
const currentTenantId = computed(() => userStore.tenantId ?? 0);

// ---------- 购物车快照（用于订单摘要） ----------
type CartEntity = { id?: number; userId?: number };
type CartItemEntity = { id?: number; skuId?: number; quantity?: number };

const cartsQuery = useListCarts(
  computed(() => ({
    page: 1,
    pageSize: 1,
    noPaging: false,
    query: JSON.stringify({ userId: currentUserId.value }),
  })),
);
const cart = computed<CartEntity | undefined>(() => {
  const items = ((cartsQuery.data?.value as any)?.items ?? []) as CartEntity[];
  return items[0];
});
const cartId = computed(() => cart.value?.id);

const cartItemsQuery = useListCartItems(
  computed(() => ({
    page: 1,
    pageSize: 100,
    noPaging: false,
    query: cartId.value === undefined ? undefined : JSON.stringify({ cartId: cartId.value }),
  })),
);
const cartItems = computed<CartItemEntity[]>(() => {
  const items = (cartItemsQuery.data?.value as any)?.items ?? [];
  return (items as CartItemEntity[]) ?? [];
});
const itemsLoading = computed(() => cartItemsQuery.isLoading.value);

const totalAmount = computed(() => 0);
const totalLabel = computed(() => `${t('mall.product.currencyCny')}${totalAmount.value}`);

// ---------- 收货表单 ----------
const form = reactive({
  recipientName: '',
  recipientPhone: '',
  shippingAddress: '',
});

const formValid = computed(() => {
  return (
    form.recipientName.trim().length > 0 &&
    form.recipientPhone.trim().length > 0 &&
    form.shippingAddress.trim().length > 0
  );
});

// ---------- 下单 + 支付 ----------
const orderMutation = useCreateOrder({
  onSuccess: () => {
    toast.success(t('checkout.orderCreated'));
  },
  onError: (err: any) => {
    toast.error(err?.message || t('checkout.errors.orderFailed'));
  },
});
const paymentMutation = useCreatePaymentTransaction({
  onError: (err: any) => {
    toast.error(err?.message || t('checkout.errors.paymentFailed'));
  },
});

async function placeOrder() {
  if (!isLogin.value) {
    navigateTo(localePath('/login'));
    return;
  }
  if (!formValid.value) {
    toast.error(t('checkout.errors.invalidForm'));
    return;
  }
  if (cartItems.value.length === 0) {
    toast.error(t('checkout.errors.emptyCart'));
    return;
  }

  // 为本次下单生成一对幂等键与业务单号。
  // idempotency_key：订单与支付各一个，防止重放导致重复下单/重复扣款。
  // business_ref_id：跨域对账键，订单与支付共用同一值，便于后续对账。
  const uuidGen = (
    globalThis as unknown as { crypto?: { randomUUID?: () => string } }
  )?.crypto?.randomUUID;
  const orderIdempotencyKey = uuidGen?.() ?? '';
  const paymentIdempotencyKey = uuidGen?.() ?? '';
  const businessRefId = uuidGen?.() ?? '';
  if (!orderIdempotencyKey || !paymentIdempotencyKey || !businessRefId) {
    toast.error(t('checkout.errors.orderFailed'));
    return;
  }

  const orderData: orderservicev1_Order = {
    userId: currentUserId.value,
    tenantId: currentTenantId.value,
    recipientName: form.recipientName,
    recipientPhone: form.recipientPhone,
    shippingAddress: form.shippingAddress,
    currency: 'CNY',
    totalAmount: 0,
    idempotencyKey: orderIdempotencyKey,
    businessRefId: businessRefId,
  } as orderservicev1_Order;

  try {
    await orderMutation.mutateAsync(orderData);
  } catch {
    return;
  }

  // 订单创建后回查以拿到 orderId 与后端计算的真实 totalAmount。
  // Order.Create 返回 empty，只能通过 (idempotency_key, tenant_id) 反查。
  let createdOrder: orderservicev1_Order | null = null;
  try {
    createdOrder = (await fetchGetOrderByIdempotencyKey(
      orderIdempotencyKey,
      currentTenantId.value,
    )) as orderservicev1_Order;
  } catch {
    toast.error(t('checkout.errors.orderFailed'));
    return;
  }

  const orderId = createdOrder?.id;
  const realAmount = createdOrder?.totalAmount ?? 0;
  if (!orderId || realAmount <= 0) {
    toast.error(t('checkout.errors.orderFailed'));
    return;
  }

  // 余额支付（暂固定为 BALANCE，待后端支持多支付方式后扩展选择 UI）
  const paymentData: paymentservicev1_PaymentTransaction = {
    userId: currentUserId.value,
    tenantId: currentTenantId.value,
    orderId: orderId,
    amount: realAmount,
    currency: 'CNY',
    paymentMethod: 'BALANCE' as paymentservicev1_PaymentMethod,
    businessType: 'BUSINESS_TYPE_CONSUME' as paymentservicev1_BusinessType,
    idempotencyKey: paymentIdempotencyKey,
    businessRefId: businessRefId,
  } as paymentservicev1_PaymentTransaction;

  try {
    await paymentMutation.mutateAsync(paymentData);
    // 订单事务已在后端清空购物车（扣库存+清 cart_item 同事务），
    // 此处仅需刷新前端缓存以反映空车状态。
    queryClient.invalidateQueries({ queryKey: ['listCarts'] });
    queryClient.invalidateQueries({ queryKey: ['listCartItems'] });
    toast.success(t('checkout.paymentSuccess'));
    navigateTo(localePath('/orders'));
  } catch {
    // 错误已由 onError 处理
  }
}
</script>

<template>
  <LayoutPageHero
    :title="t('mall.checkout.title')"
    :description="t('mall.checkout.subtitle')"
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

    <div v-else class="grid gap-6 lg:grid-cols-[1fr_380px]">
      <!-- 收货信息 -->
      <div class="rounded-2xl border border-border bg-card p-6">
        <h2 class="mb-1 text-xl font-bold text-foreground">{{ t('checkout.recipientTitle') }}</h2>
        <p class="mb-6 text-sm text-muted-foreground">{{ t('checkout.recipientDesc') }}</p>

        <div class="flex flex-col gap-5">
          <div class="flex flex-col gap-2">
            <UiLabel class="text-xs text-foreground">{{ t('checkout.recipientName') }}</UiLabel>
            <UiInput
              v-model="form.recipientName"
              type="text"
              autocomplete="name"
              :placeholder="t('checkout.recipientNamePlaceholder')"
            />
          </div>
          <div class="flex flex-col gap-2">
            <UiLabel class="text-xs text-foreground">{{ t('checkout.recipientPhone') }}</UiLabel>
            <UiInput
              v-model="form.recipientPhone"
              type="tel"
              autocomplete="tel"
              :placeholder="t('checkout.recipientPhonePlaceholder')"
            />
          </div>
          <div class="flex flex-col gap-2">
            <UiLabel class="text-xs text-foreground">{{ t('checkout.shippingAddress') }}</UiLabel>
            <UiTextarea
              v-model="form.shippingAddress"
              autocomplete="street-address"
              :placeholder="t('checkout.shippingAddressPlaceholder')"
              class="min-h-[96px]"
            />
          </div>
        </div>
      </div>

      <!-- 订单摘要 -->
      <div class="rounded-2xl border border-border bg-card p-6 lg:sticky lg:top-24 lg:self-start">
        <h2 class="mb-4 text-xl font-bold text-foreground">{{ t('checkout.summary') }}</h2>

        <div v-if="itemsLoading" class="flex flex-col gap-3">
          <div v-for="i in 3" :key="i" class="flex items-center gap-3 rounded-md border border-border bg-background/40 p-3">
            <UiSkeleton class="h-10 w-10 shrink-0 rounded" />
            <UiSkeleton class="h-4 flex-1" />
            <UiSkeleton class="h-4 w-6" />
          </div>
        </div>

        <template v-else>
          <div
            v-if="cartItems.length === 0"
            class="flex flex-col items-center gap-3 py-10 text-center"
          >
            <XIcon icon="carbon:shopping-cart" :size="40" class="text-muted-foreground" />
            <p class="text-sm text-muted-foreground">{{ t('cart.empty') }}</p>
            <UiButton variant="outline" size="sm" @click="navigateTo(localePath('/'))">
              {{ t('cart.continueShopping') }}
            </UiButton>
          </div>

          <ul v-else class="flex flex-col gap-3">
            <li
              v-for="item in cartItems"
              :key="item.id"
              class="flex items-center gap-3 rounded-md border border-border bg-background/40 p-3"
            >
              <UiProductPlaceholder
                :seed="item.skuId ?? 0"
                class="h-10 w-10 shrink-0 rounded text-muted-foreground"
              />
              <div class="min-w-0 flex-1">
                <p class="line-clamp-1 text-xs text-foreground">
                  {{ t('cart.skuId') }}#{{ item.skuId ?? '—' }}
                </p>
                <p class="text-[10px] text-muted-foreground">{{ t('cart.lineItemNote') }}</p>
              </div>
              <span class="text-xs tabular-nums text-muted-foreground">×{{ item.quantity ?? 0 }}</span>
            </li>
          </ul>
        </template>

        <UiSeparator class="my-5" />

        <div class="flex items-center justify-between">
          <span class="text-sm text-muted-foreground">{{ t('cart.total') }}</span>
          <span class="text-xl font-bold text-primary">{{ totalLabel }}</span>
        </div>
        <p class="mt-1 text-[10px] text-muted-foreground">{{ t('cart.totalNote') }}</p>

        <UiButton
          class="mt-5 w-full"
          size="lg"
          :disabled="!formValid || cartItems.length === 0 || orderMutation.isPending.value || paymentMutation.isPending.value"
          @click="placeOrder"
        >
          {{ t('checkout.placeOrder') }}
        </UiButton>
        <p class="mt-3 text-center text-[10px] text-muted-foreground">
          {{ t('checkout.disclaimer') }}
        </p>
      </div>
    </div>
  </LayoutSectionContainer>
</template>
