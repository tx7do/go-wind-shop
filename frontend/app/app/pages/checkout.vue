<script setup lang="ts">
import { computed, reactive } from 'vue';
import { toast } from 'vue-sonner';
import { XIcon } from '@/plugins/xicon';
import {
  useListCarts,
  useListCartItems,
  useCreateOrder,
  useCreatePaymentTransaction,
} from '@/api/composables';
import { useAccessStore } from '@/stores/modules/core/access.state';
import { useUserStore } from '@/stores/modules/core/user.state';
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

  const orderData: orderservicev1_Order = {
    userId: currentUserId.value,
    tenantId: currentTenantId.value,
    recipientName: form.recipientName,
    recipientPhone: form.recipientPhone,
    shippingAddress: form.shippingAddress,
    currency: 'CNY',
    totalAmount: totalAmount.value,
  } as orderservicev1_Order;

  try {
    await orderMutation.mutateAsync(orderData);
  } catch {
    return;
  }

  // 支付占位：触发后端 stub，将订单标记为 PAID
  const paymentData: paymentservicev1_PaymentTransaction = {
    userId: currentUserId.value,
    tenantId: currentTenantId.value,
    amount: totalAmount.value,
    currency: 'CNY',
    paymentMethod: 'BALANCE' as paymentservicev1_PaymentMethod,
    businessType: 'BUSINESS_TYPE_CONSUME' as paymentservicev1_BusinessType,
  } as paymentservicev1_PaymentTransaction;

  try {
    await paymentMutation.mutateAsync(paymentData);
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
    icon="carbon:document-unknown"
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

        <UiSkeleton v-if="itemsLoading" class="mb-4 h-40 w-full" />

        <template v-else>
          <div v-if="cartItems.length === 0" class="py-8 text-center text-sm text-muted-foreground">
            {{ t('cart.empty') }}
          </div>

          <ul v-else class="flex flex-col gap-3">
            <li
              v-for="item in cartItems"
              :key="item.id"
              class="flex items-center gap-3 rounded-md border border-border bg-background/40 p-3"
            >
              <div class="flex h-10 w-10 shrink-0 items-center justify-center rounded bg-primary/5 text-lg">
                📦
              </div>
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
