<script setup lang="ts">
import { computed, reactive, watch } from 'vue';
import { toast } from 'vue-sonner';
import { cn } from '@/lib/utils';
import { XIcon } from '@/plugins/xicon';
import {
  useListCarts,
  useListCartItems,
  useUpdateCartItem,
  useDeleteCartItem,
  fetchListSkuPricesStore,
  fetchListSkuAttributeCombinationsStore,
  useListProductAttributes,
  useListProductAttributeValues,
  fetchGetSkuStore,
  fetchGetProductStore,
} from '@/api/composables';
import { getCurrentLocale } from '@/utils/locale';
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
type CartItemEntity = {
  id?: number;
  cartId?: number;
  skuId?: number;
  quantity?: number;
};

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

// ---------- 属性名 / 属性值 displayName 反查 map ----------
const currentLocale = computed(() => getCurrentLocale());

function pickTranslation<T extends { languageCode?: string }>(
  translations: T[] | undefined,
): T | undefined {
  if (!translations || translations.length === 0) return undefined;
  const match = translations.find((tr) => tr.languageCode === currentLocale.value);
  return match ?? translations[0];
}

const attributesQuery = useListProductAttributes({
  page: 1,
  pageSize: 100,
  noPaging: false,
});
const attributeValuesQuery = useListProductAttributeValues({
  page: 1,
  pageSize: 200,
  noPaging: false,
});

const attributeNameMap = computed(() => {
  const items = ((attributesQuery.data?.value as any)?.items ?? []) as Array<{
    id?: number;
    translations?: Array<{ name?: string; languageCode?: string }>;
  }>;
  const map = new Map<number, string>();
  for (const a of items) {
    if (a.id === undefined) continue;
    const tr = pickTranslation(a.translations);
    if (tr?.name) map.set(a.id, tr.name);
  }
  return map;
});

const attributeValueDisplayNameMap = computed(() => {
  const items = ((attributeValuesQuery.data?.value as any)?.items ?? []) as Array<{
    id?: number;
    attributeId?: number;
    translations?: Array<{ displayName?: string; languageCode?: string }>;
  }>;
  const map = new Map<number, string>();
  for (const v of items) {
    if (v.id === undefined) continue;
    const tr = pickTranslation(v.translations);
    if (tr?.displayName) map.set(v.id, tr.displayName);
  }
  return map;
});

// ---------- 每个 SKU 的价格（通过 fetchListSkuPricesStore 在 watch 中拉取） ----------
// skuPricesMap: skuId → amount（string，CNY，单位：分）
const skuPricesMap = reactive<Record<number, string>>({});

watch(
  cartItems,
  async (items) => {
    for (const k of Object.keys(skuPricesMap)) delete skuPricesMap[Number(k)];
    if (!items || items.length === 0) return;
    await Promise.all(
      items.map(async (item) => {
        const skuId = item.skuId;
        if (skuId === undefined) return;
        try {
          const resp: any = await fetchListSkuPricesStore({
            page: 1,
            pageSize: 10,
            noPaging: false,
            query: JSON.stringify({ skuId }),
          });
          const prices = (resp?.items ?? []) as Array<{
            skuId?: number;
            currency?: string;
            amount?: string;
          }>;
          const cny = prices.find((p) => p.currency === 'CNY') ?? prices[0];
          if (cny?.amount) {
            skuPricesMap[skuId] = cny.amount;
          }
        } catch {
          // ignore
        }
      }),
    );
  },
  { immediate: true },
);

// ---------- 每个 SKU 的规格组合（用于显示规格描述） ----------
const skuCombosMap = reactive<Record<number, Array<{ attrId: number; valId: number }>>>({});

watch(
  cartItems,
  async (items) => {
    for (const k of Object.keys(skuCombosMap)) delete skuCombosMap[Number(k)];
    if (!items || items.length === 0) return;
    await Promise.all(
      items.map(async (item) => {
        const skuId = item.skuId;
        if (skuId === undefined) return;
        try {
          const resp: any = await fetchListSkuAttributeCombinationsStore({
            page: 1,
            pageSize: 100,
            noPaging: false,
            query: JSON.stringify({ skuId }),
          });
          const combos = (resp?.items ?? []) as Array<{
            skuId?: number;
            attributeId?: number;
            attributeValueId?: number;
          }>;
          const list: Array<{ attrId: number; valId: number }> = [];
          for (const c of combos) {
            if (c.attributeId !== undefined && c.attributeValueId !== undefined) {
              list.push({ attrId: c.attributeId, valId: c.attributeValueId });
            }
          }
          if (list.length > 0) skuCombosMap[skuId] = list;
        } catch {
          // ignore
        }
      }),
    );
  },
  { immediate: true },
);

// ---------- 商品名 / 商品图片：skuId → SKU → productId → Product ----------
// productInfoMap: skuId → { name, imageUrl }
const productInfoMap = reactive<Record<number, { name: string; imageUrl: string }>>({});

watch(
  cartItems,
  async (items) => {
    for (const k of Object.keys(productInfoMap)) delete productInfoMap[Number(k)];
    if (!items || items.length === 0) return;
    await Promise.all(
      items.map(async (item) => {
        const skuId = item.skuId;
        if (skuId === undefined) return;
        try {
          const sku: any = await fetchGetSkuStore(skuId);
          const productId = sku?.productId;
          if (productId === undefined) return;
          const product: any = await fetchGetProductStore(productId);
          const imageUrl = product?.imageUrl ?? '';
          const tr = pickTranslation(
            product?.translations as Array<{ name?: string; languageCode?: string }> | undefined,
          );
          const name = tr?.name ?? '';
          productInfoMap[skuId] = { name, imageUrl };
        } catch {
          // ignore
        }
      }),
    );
  },
  { immediate: true },
);

function productName(skuId: number | undefined): string {
  if (skuId === undefined) return '';
  return productInfoMap[skuId]?.name ?? '';
}
function productImageUrl(skuId: number | undefined): string {
  if (skuId === undefined) return '';
  return productInfoMap[skuId]?.imageUrl ?? '';
}

function describeSku(skuId: number | undefined): string {
  if (skuId === undefined) return '';
  const combos = skuCombosMap[skuId];
  if (!combos || combos.length === 0) return '';
  const parts: string[] = [];
  for (const { attrId, valId } of combos) {
    const attrName = attributeNameMap.value.get(attrId);
    const valName = attributeValueDisplayNameMap.value.get(valId);
    if (!attrName || !valName) return '';
    parts.push(`${attrName}: ${valName}`);
  }
  return parts.join('，');
}

// 单价展示（分→元，两位小数）
function unitPriceLabel(skuId: number | undefined): string {
  if (skuId === undefined) return '';
  const amountStr = skuPricesMap[skuId];
  if (!amountStr) return '';
  const cents = parseFloat(amountStr);
  if (Number.isNaN(cents)) return '';
  return (cents / 100).toFixed(2);
}

// ---------- 选中态 ----------
// selectedIds: 当前被勾选的 cart item id 集合
const selectedIds = reactive<Set<number>>(new Set());

const allSelected = computed(() => {
  if (cartItems.value.length === 0) return false;
  return cartItems.value.every((it) => it.id !== undefined && selectedIds.has(it.id));
});

function toggleSelectAll(checked: boolean) {
  selectedIds.clear();
  if (checked) {
    for (const it of cartItems.value) {
      if (it.id !== undefined) selectedIds.add(it.id);
    }
  }
}

function toggleSelectItem(id: number | undefined, checked: boolean) {
  if (id === undefined) return;
  if (checked) {
    selectedIds.add(id);
  } else {
    selectedIds.delete(id);
  }
}

// 已选商品总额：只累加被勾选的 cart item（分→元，两位小数）
const selectedAmount = computed(() => {
  let sum = 0;
  for (const item of cartItems.value) {
    if (item.id === undefined || !selectedIds.has(item.id)) continue;
    const amountStr = skuPricesMap[item.skuId ?? -1];
    if (!amountStr) continue;
    const cents = parseFloat(amountStr);
    if (Number.isNaN(cents)) continue;
    const yuan = cents / 100;
    sum += yuan * (item.quantity ?? 0);
  }
  return sum.toFixed(2);
});
const selectedLabel = computed(() => `${t('mall.product.currencyCny')}${selectedAmount.value}`);

const selectedCount = computed(() => selectedIds.size);

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
  selectedIds.delete(item.id);
}

function goCheckout() {
  if (selectedIds.size === 0) {
    toast.error(t('cart.emptySelected'));
    return;
  }
  navigateTo(localePath('/checkout'));
}
</script>

<template>
  <!-- 交易进度条：1.购物车(当前激活) → 2.核对订单 → 3.线上支付 → 4.完成 -->
  <LayoutSectionContainer class="!py-4">
    <ol class="flex items-center justify-center gap-2 md:gap-4">
      <li
        v-for="(step, idx) in [
          { key: 'cart', label: t('checkout.steps.cart') },
          { key: 'confirm', label: t('checkout.steps.confirm') },
          { key: 'pay', label: t('checkout.steps.pay') },
          { key: 'done', label: t('checkout.steps.done') },
        ]"
        :key="step.key"
        class="flex items-center gap-2 md:gap-4"
      >
        <div class="flex flex-col items-center gap-1">
          <span
            :class="cn(
              'flex h-7 w-7 items-center justify-center rounded-full border text-xs font-bold transition-colors',
              idx === 0
                ? 'border-primary bg-primary/10 text-primary dark:border-green-500 dark:bg-green-500/10 dark:text-green-400'
                : 'border-border text-muted-foreground',
            )"
          >
            {{ idx + 1 }}
          </span>
          <span
            :class="cn(
              'text-[10px] font-medium transition-colors',
              idx === 0
                ? 'text-primary dark:text-green-400'
                : 'text-muted-foreground',
            )"
          >
            {{ step.label }}
          </span>
        </div>
        <span
          v-if="idx < 3"
          :class="cn(
            'h-px w-8 md:w-16 transition-colors',
            idx < 0
              ? 'bg-primary dark:bg-green-500'
              : 'bg-border',
          )"
        ></span>
      </li>
    </ol>
  </LayoutSectionContainer>

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

    <!-- 购物车列表 + 结算看板（左右分栏） -->
    <div v-else class="grid gap-6 md:grid-cols-12">
      <!-- 左侧：商品清单（8 栏） -->
      <div class="md:col-span-8 rounded-2xl border border-border bg-card overflow-hidden">
        <!-- 表头（含全选） -->
        <div class="border-b border-border bg-muted/40 px-6 py-3 text-xs font-medium uppercase tracking-wide text-muted-foreground">
          <div class="grid grid-cols-[32px_1fr_120px_160px_48px] items-center gap-4">
            <span>
              <input
                type="checkbox"
                class="h-4 w-4 rounded border-border accent-[hsl(var(--primary))]"
                :checked="allSelected"
                @change="toggleSelectAll(($event.target as HTMLInputElement).checked)"
              />
            </span>
            <span>{{ t('cart.table.item') }}</span>
            <span class="text-center">{{ t('cart.table.unitPrice') }}</span>
            <span class="text-center">{{ t('cart.table.quantity') }}</span>
            <span class="text-right">{{ t('cart.table.action') }}</span>
          </div>
        </div>

        <!-- 表体 -->
        <div
          v-for="item in cartItems"
          :key="item.id"
          class="border-b border-border px-6 py-4 last:border-b-0"
        >
          <div class="grid grid-cols-[32px_1fr_120px_160px_48px] items-center gap-4">
            <!-- 单选 checkbox -->
            <span>
              <input
                type="checkbox"
                class="h-4 w-4 rounded border-border accent-[hsl(var(--primary))]"
                :checked="item.id !== undefined && selectedIds.has(item.id)"
                @change="toggleSelectItem(item.id, ($event.target as HTMLInputElement).checked)"
              />
            </span>

            <!-- 商品缩略图（1:1）+ 标题 + 规格描述 -->
            <div class="flex items-center gap-4">
              <img
                v-if="productImageUrl(item.skuId)"
                :src="productImageUrl(item.skuId)"
                :alt="productName(item.skuId)"
                class="aspect-square h-16 w-16 shrink-0 rounded-md object-cover"
              />
              <UiProductPlaceholder
                v-else
                :seed="item.skuId ?? 0"
                class="aspect-square h-16 w-16 shrink-0 rounded-md text-muted-foreground"
              />
              <div class="min-w-0">
                <p class="line-clamp-1 text-sm font-semibold text-foreground dark:text-slate-200">
                  {{ productName(item.skuId) || '—' }}
                </p>
                <p
                  v-if="describeSku(item.skuId)"
                  class="mt-1 line-clamp-1 text-xs text-muted-foreground"
                >
                  {{ describeSku(item.skuId) }}
                </p>
              </div>
            </div>

            <!-- 单价 -->
            <p class="text-center text-sm text-muted-foreground">
              {{ t('mall.product.currencyCny') }}{{ unitPriceLabel(item.skuId) || '—' }}
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

      <!-- 右侧：结算看板（4 栏，sticky） -->
      <div class="md:col-span-4 rounded-2xl border border-border bg-card p-6 md:sticky md:top-24 md:self-start">
        <h2 class="mb-4 text-xl font-bold text-foreground">{{ t('cart.checkout') }}</h2>

        <div class="rounded-xl border border-primary/40 bg-primary/5 p-5 dark:border-green-500/40 dark:bg-green-500/5">
          <p class="text-xs text-muted-foreground">{{ t('cart.selectedAmount') }}</p>
          <p class="mt-1 text-3xl font-extrabold text-primary dark:text-green-400">
            {{ selectedLabel }}
          </p>
          <p class="mt-1 text-[10px] text-muted-foreground">
            {{ selectedCount }} {{ t('cart.table.item') }}
          </p>
        </div>

        <UiButton
          class="mt-5 w-full"
          size="lg"
          @click="goCheckout"
        >
          {{ t('cart.goCheckout') }}
        </UiButton>
      </div>
    </div>
  </LayoutSectionContainer>
</template>
