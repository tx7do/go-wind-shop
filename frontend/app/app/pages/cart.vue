<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { toast } from 'vue-sonner';
import { cn } from '@/lib/utils';
import { formatFen } from '@/utils/money';
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

// 注：user_id 行级隔离由后端 UserPrivacy 策略强制（钉定为当前登录用户），
// 前端 userId 值会被服务端忽略，带与不带等效。enabled 守卫避免未登录/未 hydrate 发请求。
const cartsQuery = useListCarts(
  computed(() => ({
    page: 1,
    pageSize: 1,
    noPaging: false,
    query: JSON.stringify({ userId: currentUserId.value }),
  })),
  { enabled: isLogin },
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
  { enabled: computed(() => isLogin.value && cartId.value !== undefined) },
);
const cartItems = computed<CartItemEntity[]>(() => {
  const items = (cartItemsQuery.data?.value as any)?.items ?? [];
  return (items as CartItemEntity[]) ?? [];
});
const itemsLoading = computed(() => cartItemsQuery.isPending.value);
const itemsError = computed(() => cartItemsQuery.isError.value);

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

// 单价（分）：交给 UiPriceTag 渲染，页面里不再手拼「元」串
function unitPriceFen(skuId: number | undefined): number | null {
  if (skuId === undefined) return null;
  const amountStr = skuPricesMap[skuId];
  if (!amountStr) return null;
  const fen = Number(amountStr);
  return Number.isFinite(fen) && fen > 0 ? fen : null;
}

// ---------- 选中态 ----------
// selectedIds: 当前被勾选的 cart item id 集合
const selectedIds = reactive<Set<number>>(new Set());

// 首次加载到商品后默认全选——购物车有货但结算区显示 ¥0.00 会像出了故障；
// 仅自动全选一次，用户手动取消后不被 refetch 打回去。
const didAutoSelect = ref(false);

watch(
  cartItems,
  (items) => {
    const ids = new Set(
      items.map((it) => it.id).filter((v): v is number => v !== undefined),
    );
    // 清理已不存在（被删）的选中项，避免计数虚高
    for (const id of [...selectedIds]) {
      if (!ids.has(id)) selectedIds.delete(id);
    }
    if (ids.size === 0) {
      didAutoSelect.value = false;
    } else if (!didAutoSelect.value) {
      for (const id of ids) selectedIds.add(id);
      didAutoSelect.value = true;
    }
  },
  { immediate: true },
);

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

// 展示用：元 → 分，交给 formatFen 出「0.00」也不会像 UiPriceTag 那样整块消失
const selectedAmountFen = computed(() => Math.round(Number(selectedAmount.value) * 100));
const selectedAmountText = computed(() => formatFen(selectedAmountFen.value));

const selectedCount = computed(() => selectedIds.size);

const showContent = computed(
  () => isLogin.value && !itemsLoading.value && !itemsError.value && cartItems.value.length > 0,
);

const checkoutSteps = [
  { key: 'cart', labelKey: 'checkout.steps.cart' },
  { key: 'confirm', labelKey: 'checkout.steps.confirm' },
  { key: 'pay', labelKey: 'checkout.steps.pay' },
  { key: 'done', labelKey: 'checkout.steps.done' },
];

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
  <LayoutPageHero
    :title="t('mall.cart.title')"
    :description="t('mall.cart.subtitle')"
    icon="carbon:shopping-cart"
    size="sm"
  >
    <!-- 交易进度：当前停在「购物车」这一步（390px 放不下四个步骤，移动端收起） -->
    <ol class="hidden flex-wrap items-center gap-x-2 gap-y-1 text-[11px] font-medium md:flex">
      <li
        v-for="(step, idx) in checkoutSteps"
        :key="step.key"
        class="flex items-center gap-2"
      >
        <span
          :class="cn(
            'rounded-full px-2 py-0.5',
            idx === 0 ? 'bg-primary-soft text-primary-soft-foreground' : 'text-muted-foreground',
          )"
        >
          {{ idx + 1 }}. {{ t(step.labelKey) }}
        </span>
        <span
          v-if="idx < checkoutSteps.length - 1"
          class="h-px w-3 bg-border"
          aria-hidden="true"
        ></span>
      </li>
    </ol>
  </LayoutPageHero>

  <LayoutSectionContainer>
    <!-- 未登录 -->
    <UiAuthGate v-if="!isLogin" />

    <!-- 加载中 -->
    <UiListSkeleton v-else-if="itemsLoading" :rows="3" />

    <!-- 错误态 -->
    <UiAppEmpty v-else-if="itemsError" variant="error">
      <template #action>
        <UiButton variant="outline" size="sm" @click="cartItemsQuery.refetch()">
          {{ t('ui.button.retry') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <!-- 空购物车 -->
    <UiAppEmpty v-else-if="cartItems.length === 0" :description="t('cart.empty')">
      <template #action>
        <UiButton variant="outline" @click="navigateTo(localePath('/'))">
          {{ t('cart.continueShopping') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <!-- 商品清单 + 结算汇总 -->
    <div v-else class="grid gap-5 pb-24 md:grid-cols-12 md:gap-6 md:pb-0">
      <div class="md:col-span-8">
        <!-- 移动端：全选条 -->
        <div class="mb-3 flex items-center justify-between md:hidden">
          <label class="flex cursor-pointer items-center gap-2 text-xs font-medium text-foreground">
            <span class="relative flex size-[18px] shrink-0 items-center justify-center">
              <input
                type="checkbox"
                class="peer size-[18px] cursor-pointer appearance-none rounded-[5px] border border-border bg-card transition-colors checked:border-primary checked:bg-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                :checked="allSelected"
                :aria-label="t('cart.selectAll')"
                @change="toggleSelectAll(($event.target as HTMLInputElement).checked)"
              />
              <XIcon
                icon="carbon:checkmark"
                :size="12"
                class="pointer-events-none absolute text-primary-foreground opacity-0 peer-checked:opacity-100"
              />
            </span>
            {{ t('cart.selectAll') }}
          </label>
          <span class="text-xs tabular-nums text-muted-foreground">
            {{ selectedCount }} / {{ cartItems.length }}
          </span>
        </div>

        <!-- 移动端：卡片列表 -->
        <ul class="space-y-3 md:hidden">
          <li
            v-for="item in cartItems"
            :key="item.id"
            class="relative rounded-lg border border-border bg-card p-3 shadow-card"
          >
            <UiButton
              variant="ghost"
              size="icon"
              class="absolute right-1 top-1 size-8 text-muted-foreground"
              :disabled="deleteMutation.isPending.value"
              :aria-label="t('cart.table.action')"
              @click="removeItem(item)"
            >
              <XIcon icon="carbon:trash-can" :size="16" />
            </UiButton>

            <div class="flex items-start gap-3">
              <span class="relative mt-1 flex size-[18px] shrink-0 items-center justify-center">
                <input
                  type="checkbox"
                  class="peer size-[18px] cursor-pointer appearance-none rounded-[5px] border border-border bg-card transition-colors checked:border-primary checked:bg-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  :checked="item.id !== undefined && selectedIds.has(item.id)"
                  :aria-label="productName(item.skuId) || t('cart.table.item')"
                  @change="toggleSelectItem(item.id, ($event.target as HTMLInputElement).checked)"
                />
                <XIcon
                  icon="carbon:checkmark"
                  :size="12"
                  class="pointer-events-none absolute text-primary-foreground opacity-0 peer-checked:opacity-100"
                />
              </span>

              <UiImage
                :src="productImageUrl(item.skuId)"
                :alt="productName(item.skuId)"
                class="size-20 shrink-0 rounded-md object-cover"
              />

              <div class="min-w-0 flex-1">
                <p class="line-clamp-2 pr-7 text-sm font-semibold leading-snug text-foreground">
                  {{ productName(item.skuId) || '—' }}
                </p>
                <p
                  v-if="describeSku(item.skuId)"
                  class="mt-1 line-clamp-1 text-xs text-muted-foreground"
                >
                  {{ describeSku(item.skuId) }}
                </p>

                <div class="mt-2 flex flex-wrap items-center justify-between gap-x-2 gap-y-2">
                  <UiPriceTag v-if="unitPriceFen(item.skuId)" :amount="unitPriceFen(item.skuId)" size="sm" />
                  <span v-else class="text-sm text-muted-foreground">—</span>

                  <div class="flex items-center gap-2">
                    <UiButton
                      variant="outline"
                      size="icon"
                      class="size-7 border-border bg-secondary text-foreground"
                      :disabled="updateMutation.isPending.value"
                      :aria-label="t('cart.decreaseQty')"
                      @click="decreaseQty(item)"
                    >
                      <XIcon icon="carbon:subtract" :size="14" />
                    </UiButton>
                    <span class="min-w-6 text-center text-sm font-medium tabular-nums text-foreground">
                      {{ item.quantity ?? 0 }}
                    </span>
                    <UiButton
                      variant="outline"
                      size="icon"
                      class="size-7 border-border bg-secondary text-foreground"
                      :disabled="updateMutation.isPending.value"
                      :aria-label="t('cart.increaseQty')"
                      @click="increaseQty(item)"
                    >
                      <XIcon icon="carbon:add" :size="14" />
                    </UiButton>
                  </div>
                </div>
              </div>
            </div>
          </li>
        </ul>

        <!-- 桌面端：行式列表 -->
        <div class="hidden overflow-hidden rounded-lg border border-border bg-card shadow-card md:block">
          <div
            class="grid grid-cols-[24px_minmax(0,1fr)_120px_132px_40px] items-center gap-4 border-b border-border px-4 py-2.5 text-xs font-medium text-muted-foreground"
          >
            <span class="relative flex size-4 items-center justify-center">
              <input
                type="checkbox"
                class="peer size-4 cursor-pointer appearance-none rounded border border-border bg-card transition-colors checked:border-primary checked:bg-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                :checked="allSelected"
                :aria-label="t('cart.selectAll')"
                @change="toggleSelectAll(($event.target as HTMLInputElement).checked)"
              />
              <XIcon
                icon="carbon:checkmark"
                :size="11"
                class="pointer-events-none absolute text-primary-foreground opacity-0 peer-checked:opacity-100"
              />
            </span>
            <span>{{ t('cart.table.item') }}</span>
            <span class="text-right">{{ t('cart.table.unitPrice') }}</span>
            <span class="text-center">{{ t('cart.table.quantity') }}</span>
            <span class="text-right">{{ t('cart.table.action') }}</span>
          </div>

          <div
            v-for="item in cartItems"
            :key="item.id"
            class="grid grid-cols-[24px_minmax(0,1fr)_120px_132px_40px] items-center gap-4 border-b border-border px-4 py-3 transition-colors last:border-b-0 hover:bg-secondary"
          >
            <span class="relative flex size-4 items-center justify-center">
              <input
                type="checkbox"
                class="peer size-4 cursor-pointer appearance-none rounded border border-border bg-card transition-colors checked:border-primary checked:bg-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                :checked="item.id !== undefined && selectedIds.has(item.id)"
                :aria-label="productName(item.skuId) || t('cart.table.item')"
                @change="toggleSelectItem(item.id, ($event.target as HTMLInputElement).checked)"
              />
              <XIcon
                icon="carbon:checkmark"
                :size="11"
                class="pointer-events-none absolute text-primary-foreground opacity-0 peer-checked:opacity-100"
              />
            </span>

            <div class="flex min-w-0 items-center gap-3">
              <UiImage
                :src="productImageUrl(item.skuId)"
                :alt="productName(item.skuId)"
                class="size-14 shrink-0 rounded-md object-cover"
              />
              <div class="min-w-0">
                <p class="truncate text-sm font-semibold text-foreground">
                  {{ productName(item.skuId) || '—' }}
                </p>
                <p
                  v-if="describeSku(item.skuId)"
                  class="mt-0.5 truncate text-xs text-muted-foreground"
                >
                  {{ describeSku(item.skuId) }}
                </p>
              </div>
            </div>

            <div class="flex justify-end">
              <UiPriceTag v-if="unitPriceFen(item.skuId)" :amount="unitPriceFen(item.skuId)" size="sm" />
              <span v-else class="text-sm text-muted-foreground">—</span>
            </div>

            <div class="flex items-center justify-center gap-2">
              <UiButton
                variant="outline"
                size="icon"
                class="size-8 border-border bg-secondary text-foreground"
                :disabled="updateMutation.isPending.value"
                :aria-label="t('cart.decreaseQty')"
                @click="decreaseQty(item)"
              >
                <XIcon icon="carbon:subtract" :size="14" />
              </UiButton>
              <span class="min-w-7 text-center text-sm font-medium tabular-nums text-foreground">
                {{ item.quantity ?? 0 }}
              </span>
              <UiButton
                variant="outline"
                size="icon"
                class="size-8 border-border bg-secondary text-foreground"
                :disabled="updateMutation.isPending.value"
                :aria-label="t('cart.increaseQty')"
                @click="increaseQty(item)"
              >
                <XIcon icon="carbon:add" :size="14" />
              </UiButton>
            </div>

            <div class="flex justify-end">
              <UiButton
                variant="ghost"
                size="icon"
                class="size-8 text-muted-foreground"
                :disabled="deleteMutation.isPending.value"
                :aria-label="t('cart.table.action')"
                @click="removeItem(item)"
              >
                <XIcon icon="carbon:trash-can" :size="16" />
              </UiButton>
            </div>
          </div>
        </div>
      </div>

      <!-- 桌面端：右侧结算汇总 -->
      <aside class="hidden md:col-span-4 md:block">
        <div class="md:sticky md:top-24">
          <div class="rounded-lg border border-border bg-card p-5 shadow-card">
            <h2 class="text-base font-bold text-foreground">{{ t('cart.checkout') }}</h2>

            <div class="mt-4 rounded-lg bg-primary-soft px-4 py-3.5">
              <p class="text-xs text-primary-soft-foreground">{{ t('cart.selectedAmount') }}</p>
              <p class="mt-1 text-2xl font-extrabold tabular-nums text-primary-soft-foreground">
                {{ t('mall.product.currencyCny') }}{{ selectedAmountText }}
              </p>
              <p class="mt-1 text-[11px] text-primary-soft-foreground">
                {{ selectedCount }} {{ t('cart.table.item') }}
              </p>
            </div>

            <label class="mt-4 flex cursor-pointer items-center gap-2 text-sm text-foreground">
              <span class="relative flex size-[18px] shrink-0 items-center justify-center">
                <input
                  type="checkbox"
                  class="peer size-[18px] cursor-pointer appearance-none rounded-[5px] border border-border bg-card transition-colors checked:border-primary checked:bg-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
                  :checked="allSelected"
                  :aria-label="t('cart.selectAll')"
                  @change="toggleSelectAll(($event.target as HTMLInputElement).checked)"
                />
                <XIcon
                  icon="carbon:checkmark"
                  :size="12"
                  class="pointer-events-none absolute text-primary-foreground opacity-0 peer-checked:opacity-100"
                />
              </span>
              {{ t('cart.selectAll') }}
            </label>

            <UiButton class="mt-4 w-full" size="lg" @click="goCheckout">
              {{ t('cart.goCheckout') }}
            </UiButton>

            <p class="mt-2 text-[11px] leading-relaxed text-muted-foreground">
              {{ t('cart.totalNote') }}
            </p>
          </div>
        </div>
      </aside>
    </div>
  </LayoutSectionContainer>

  <!-- 移动端吸底：合计 + 去结算，避让 TabBar；桌面端落在右侧汇总卡 -->
  <UiStickyBar v-if="showContent" class="md:hidden">
    <label class="flex shrink-0 cursor-pointer items-center gap-2 text-xs font-medium text-foreground">
      <span class="relative flex size-[18px] shrink-0 items-center justify-center">
        <input
          type="checkbox"
          class="peer size-[18px] cursor-pointer appearance-none rounded-[5px] border border-border bg-card transition-colors checked:border-primary checked:bg-primary focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
          :checked="allSelected"
          :aria-label="t('cart.selectAll')"
          @change="toggleSelectAll(($event.target as HTMLInputElement).checked)"
        />
        <XIcon
          icon="carbon:checkmark"
          :size="12"
          class="pointer-events-none absolute text-primary-foreground opacity-0 peer-checked:opacity-100"
        />
      </span>
      {{ t('cart.selectAll') }}
    </label>

    <div class="ml-auto flex min-w-0 flex-col items-end">
      <span class="text-[10px] leading-none text-muted-foreground">{{ t('cart.total') }}</span>
      <p class="mt-0.5 text-lg font-extrabold leading-none tabular-nums text-primary">
        {{ t('mall.product.currencyCny') }}{{ selectedAmountText }}
      </p>
    </div>

    <UiButton class="shrink-0" size="lg" @click="goCheckout">
      {{ t('cart.goCheckout') }}
    </UiButton>
  </UiStickyBar>
</template>
