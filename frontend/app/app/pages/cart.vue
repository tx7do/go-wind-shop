<script setup lang="ts">
import { computed } from 'vue';
import { toast } from 'vue-sonner';
import { XIcon } from '@/plugins/xicon';
import {
  useListCarts,
  useListCartItems,
  useUpdateCartItem,
  useDeleteCartItem,
} from '@/api/composables';
import { useAccessStore } from '@/stores/modules/core/access.state';
import { useUserStore } from '@/stores/modules/core/user.state';

const { t } = useI18n();
const localePath = useLocalePath();

useHead({ title: t('mall.cart.title') });

const accessStore = useAccessStore();
const userStore = useUserStore();

const isLogin = computed(() => {
  const token = accessStore.accessToken;
  return !!token?.value && !accessStore.loginExpired;
});

const currentUserId = computed(() => userStore.user?.id ?? 0);

// 当前用户的购物车（按 userId 过滤）
type CartEntity = { id?: number; userId?: number };

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

// 该购物车下的购物车项（按 cartId 过滤）
type CartItemEntity = {
  id?: number;
  cartId?: number;
  skuId?: number;
  quantity?: number;
};

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

// 价格占位：列表接口返回的 Product 无价格字段，单价/总价需拉 SKU 价格子表聚合，
// 待后端补齐后替换为真实价格计算。当前用占位符保持结构完整。
const UNIT_PRICE_PLACEHOLDER = '—';
const totalAmount = computed(() => 0);
const totalLabel = computed(() => `${t('mall.product.currencyCny')}${totalAmount.value}`);

// ---------- 数量调整 / 移除 ----------
const updateMutation = useUpdateCartItem({
  onSuccess: () => {
    cartItemsQuery.refetch();
  },
  onError: (err: any) => {
    toast.error(err?.message || t('cart.errors.updateFailed'));
  },
});
const deleteMutation = useDeleteCartItem({
  onSuccess: () => {
    cartItemsQuery.refetch();
  },
  onError: (err: any) => {
    toast.error(err?.message || t('cart.errors.removeFailed'));
  },
});

function increaseQty(item: CartItemEntity) {
  if (item.id === undefined) return;
  const next = (item.quantity ?? 0) + 1;
  updateMutation.mutate({ id: item.id, values: { quantity: next } });
}
function decreaseQty(item: CartItemEntity) {
  if (item.id === undefined) return;
  const next = (item.quantity ?? 0) - 1;
  if (next <= 0) {
    deleteMutation.mutate(item.id);
    return;
  }
  updateMutation.mutate({ id: item.id, values: { quantity: next } });
}
function removeItem(item: CartItemEntity) {
  if (item.id === undefined) return;
  deleteMutation.mutate(item.id);
}

function goCheckout() {
  if (cartItems.value.length === 0) {
    toast.error(t('cart.errors.empty'));
    return;
  }
  navigateTo(localePath('/checkout'));
}
</script>

<template>
  <LayoutPageHero
    :title="t('mall.cart.title')"
    :description="t('mall.cart.subtitle')"
    icon="carbon:shopping-cart"
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
    <div v-else-if="itemsLoading" class="rounded-2xl border border-border bg-card overflow-hidden">
      <div class="border-b border-border bg-muted/40 px-6 py-3">
        <UiSkeleton class="h-4 w-32" />
      </div>
      <div v-for="i in 3" :key="i" class="border-b border-border px-6 py-4 last:border-b-0">
        <div class="flex items-center gap-4">
          <UiSkeleton class="h-16 w-16 shrink-0 rounded-md" />
          <UiSkeleton class="h-4 flex-1" />
          <UiSkeleton class="h-8 w-16" />
          <UiSkeleton class="h-8 w-8" />
        </div>
      </div>
    </div>

    <!-- 空购物车 -->
    <div
      v-else-if="cartItems.length === 0"
      class="flex flex-col items-center gap-4 rounded-2xl border border-border bg-card p-16 text-center"
    >
      <XIcon icon="carbon:shopping-cart" :size="48" class="text-muted-foreground" />
      <p class="text-lg text-muted-foreground">{{ t('cart.empty') }}</p>
      <UiButton variant="outline" @click="navigateTo(localePath('/'))">
        {{ t('cart.continueShopping') }}
      </UiButton>
    </div>

    <!-- 购物车列表 -->
    <div v-else class="flex flex-col gap-6">
      <div class="overflow-x-auto rounded-2xl border border-border bg-card">
        <div class="min-w-[640px]">
        <div class="border-b border-border bg-muted/40 px-6 py-3 text-xs font-medium uppercase tracking-wide text-muted-foreground">
          <div class="grid grid-cols-[1fr_120px_160px_48px] items-center gap-4">
            <span>{{ t('cart.table.item') }}</span>
            <span class="text-center">{{ t('cart.table.unitPrice') }}</span>
            <span class="text-center">{{ t('cart.table.quantity') }}</span>
            <span class="text-right">{{ t('cart.table.action') }}</span>
          </div>
        </div>

        <div
          v-for="item in cartItems"
          :key="item.id"
          class="border-b border-border px-6 py-4 last:border-b-0"
        >
          <div class="grid grid-cols-[1fr_120px_160px_48px] items-center gap-4">
            <!-- 商品占位 -->
            <div class="flex items-center gap-4">
              <UiProductPlaceholder
                :seed="item.skuId ?? 0"
                class="h-16 w-16 shrink-0 rounded-md text-muted-foreground"
              />
              <div class="min-w-0">
                <p class="line-clamp-1 text-sm font-medium text-foreground">
                  {{ t('cart.skuId') }}#{{ item.skuId ?? '—' }}
                </p>
                <p class="text-xs text-muted-foreground">{{ t('cart.lineItemNote') }}</p>
              </div>
            </div>

            <!-- 单价占位 -->
            <p class="text-center text-sm text-muted-foreground">
              {{ t('mall.product.currencyCny') }}{{ UNIT_PRICE_PLACEHOLDER }}
            </p>

            <!-- 数量步进 -->
            <div class="flex items-center justify-center gap-2">
              <UiButton
                size="sm"
                variant="outline"
                :disabled="updateMutation.isPending.value"
                @click="decreaseQty(item)"
              >
                <XIcon icon="carbon:subtract" :size="14" />
              </UiButton>
              <span class="w-10 text-center text-sm tabular-nums text-foreground">
                {{ item.quantity ?? 0 }}
              </span>
              <UiButton
                size="sm"
                variant="outline"
                :disabled="updateMutation.isPending.value"
                @click="increaseQty(item)"
              >
                <XIcon icon="carbon:add" :size="14" />
              </UiButton>
            </div>

            <!-- 移除 -->
            <div class="flex justify-end">
              <UiButton
                size="sm"
                variant="ghost"
                :disabled="deleteMutation.isPending.value"
                @click="removeItem(item)"
              >
                <XIcon icon="carbon:trash-can" :size="16" />
              </UiButton>
            </div>
          </div>
        </div>
        </div>
      </div>

      <!-- 结算栏 -->
      <div class="sticky bottom-0 rounded-2xl border border-border bg-card px-6 py-4">
        <div class="flex items-center justify-between">
          <div>
            <p class="text-xs text-muted-foreground">{{ t('cart.total') }}</p>
            <p class="text-2xl font-bold text-primary">{{ totalLabel }}</p>
            <p class="mt-1 text-[10px] text-muted-foreground">{{ t('cart.totalNote') }}</p>
          </div>
          <UiButton size="lg" @click="goCheckout">
            {{ t('cart.checkout') }}
          </UiButton>
        </div>
      </div>
    </div>
  </LayoutSectionContainer>
</template>
