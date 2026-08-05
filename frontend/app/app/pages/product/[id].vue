<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { toast } from 'vue-sonner';
import { cn } from '@/lib/utils';
import { XIcon } from '@/plugins/xicon';
import {
  useGetProduct,
  useListProductAttributes,
  useListProductAttributeValues,
  useListSkus,
  fetchListSkuAttributeCombinationsStore,
  fetchListSkuPricesStore,
  useListCarts,
  useCreateCart,
  useCreateCartItem,
} from '@/api/composables';
import { getCurrentLocale } from '@/utils/locale';
import { useAccessStore } from '@/stores/modules/core/access.state';
import { useUserStore } from '@/stores/modules/core/user.state';
import { queryClient } from '@/plugins/vue-query';

const route = useRoute();
const { t } = useI18n();
const localePath = useLocalePath();

const productId = computed(() => Number(route.params.id));
const currentLocale = computed(() => getCurrentLocale());

// ---------- 登录态 / 当前用户 ----------
const accessStore = useAccessStore();
const userStore = useUserStore();
const isLogin = computed(() => {
  const token = accessStore.accessToken;
  return !!token?.value && !accessStore.loginExpired;
});
const currentUserId = computed(() => userStore.user?.id ?? 0);
const currentTenantId = computed(() => userStore.tenantId ?? 0);

// ---------- 类型 ----------
type ProductTranslation = {
  name?: string;
  shortDescription?: string;
  longDescription?: string;
  languageCode?: string;
};
type ProductEntity = {
  id?: number;
  imageUrl?: string;
  translations?: ProductTranslation[];
};
type AttributeTranslation = {
  name?: string;
  languageCode?: string;
};
type AttributeEntity = {
  id?: number;
  translations?: AttributeTranslation[];
};
type AttributeValueTranslation = {
  displayName?: string;
  languageCode?: string;
};
type AttributeValueEntity = {
  id?: number;
  attributeId?: number;
  translations?: AttributeValueTranslation[];
};
type SkuEntity = {
  id?: number;
  productId?: number;
  stockQty?: number;
};
type SkuPriceEntity = {
  skuId?: number;
  currency?: string;
  amount?: string;
};
type SkuCombinationEntity = {
  skuId?: number;
  attributeId?: number;
  attributeValueId?: number;
};

function pickTranslation<T extends { languageCode?: string }>(
  translations: T[] | undefined,
): T | undefined {
  if (!translations || translations.length === 0) return undefined;
  const match = translations.find((tr) => tr.languageCode === currentLocale.value);
  return match ?? translations[0];
}

const DISPLAY_CURRENCY = 'CNY';

// ---------- 商品详情 ----------
const productQuery = useGetProduct(productId.value, {
  enabled: productId.value > 0,
});
const product = computed<ProductEntity | undefined>(() => {
  return productQuery.data?.value as ProductEntity | undefined;
});
const productLoading = computed(() => productQuery.isLoading.value);
const productTranslation = computed(() => pickTranslation(product.value?.translations));

useHead({ title: () => productTranslation.value?.name || '' });

// ---------- SKU 列表 ----------
const skusQuery = useListSkus(
  computed(() => ({
    page: 1,
    pageSize: 100,
    noPaging: false,
    query: JSON.stringify({ productId: productId.value }),
  })),
);
const skus = computed<SkuEntity[]>(() => {
  const items = (skusQuery.data?.value as any)?.items ?? [];
  return (items as SkuEntity[]) ?? [];
});

// ---------- 属性 & 属性值 ----------
const attributesQuery = useListProductAttributes({
  page: 1,
  pageSize: 100,
  noPaging: false,
});
const attributes = computed<AttributeEntity[]>(() => {
  const items = (attributesQuery.data?.value as any)?.items ?? [];
  return (items as AttributeEntity[]) ?? [];
});

const attributeValuesQuery = useListProductAttributeValues({
  page: 1,
  pageSize: 200,
  noPaging: false,
});
const allAttributeValues = computed<AttributeValueEntity[]>(() => {
  const items = (attributeValuesQuery.data?.value as any)?.items ?? [];
  return (items as AttributeValueEntity[]) ?? [];
});

const valuesByAttribute = computed(() => {
  const map = new Map<number, AttributeEntity[]>();
  for (const v of allAttributeValues.value) {
    if (v.attributeId === undefined) continue;
    const arr = map.get(v.attributeId) ?? [];
    arr.push(v);
    map.set(v.attributeId, arr);
  }
  return map;
});

// ---------- 默认选中每个属性的第一个值 ----------
// 当属性与属性值数据加载完成后，为每个尚无选中值的属性设默认选中其第一个值，
// 使页面初始即有完整选中态，便于用户直观看到当前 SKU 与价格。
watch(
  [attributes, valuesByAttribute],
  ([attrs, valsMap]) => {
    if (!attrs || attrs.length === 0 || !valsMap || valsMap.size === 0) return;
    for (const attr of attrs) {
      if (attr.id === undefined) continue;
      if (selections[attr.id] !== undefined) continue;
      const vals = valsMap.get(attr.id);
      if (!vals || vals.length === 0) continue;
      const first = vals[0];
      if (first?.id !== undefined) {
        selections[attr.id] = String(first.id);
      }
    }
  },
  { immediate: true },
);

// ---------- 每个 SKU 的属性组合 & 价格（通过 fetch*Store 在 watch 中拉取） ----------
const skuCombinationsMap = reactive<Record<number, Map<number, number>>>({});
const skuPricesMap = reactive<Record<number, SkuPriceEntity[]>>({});

watch(
  skus,
  async (list) => {
    for (const k of Object.keys(skuCombinationsMap)) delete skuCombinationsMap[Number(k)];
    for (const k of Object.keys(skuPricesMap)) delete skuPricesMap[Number(k)];
    if (!list || list.length === 0) return;
    await Promise.all(
      list.map(async (sku) => {
        if (sku.id === undefined) return;
        try {
          const combosResp: any = await fetchListSkuAttributeCombinationsStore({
            page: 1,
            pageSize: 100,
            noPaging: false,
            query: JSON.stringify({ skuId: sku.id }),
          });
          const combos = (combosResp?.items ?? []) as SkuCombinationEntity[];
          const map = new Map<number, number>();
          for (const c of combos) {
            if (c.attributeId !== undefined && c.attributeValueId !== undefined) {
              map.set(c.attributeId, c.attributeValueId);
            }
          }
          skuCombinationsMap[sku.id] = map;
        } catch (e) {
          // ignore
        }
        try {
          const pricesResp: any = await fetchListSkuPricesStore({
            page: 1,
            pageSize: 100,
            noPaging: false,
            query: JSON.stringify({ skuId: sku.id }),
          });
          const prices = (pricesResp?.items ?? []) as SkuPriceEntity[];
          skuPricesMap[sku.id] = prices;
        } catch (e) {
          // ignore
        }
      }),
    );
  },
  { immediate: true },
);

// ---------- 用户选择 ----------
// 注意：reka-ui Select 使用字符串值，因此选择以字符串存储，匹配时转为数字。
const selections = reactive<Record<number, string | undefined>>({});
const allAttributesSelected = computed(() => {
  // 无属性商品（单 SKU）：无可选属性，视为已满足选择前提。
  if (attributes.value.length === 0) return true;
  for (const attr of attributes.value) {
    if (attr.id === undefined) continue;
    if (selections[attr.id] === undefined) return false;
  }
  return true;
});

// ---------- 匹配 SKU ----------
const matchedSku = computed<SkuEntity | undefined>(() => {
  if (!allAttributesSelected.value) return undefined;

  // 无属性商品（单 SKU）：无属性组合，若恰好存在一个 SKU 则直接匹配。
  if (attributes.value.length === 0) {
    if (skus.value.length === 1) {
      const only = skus.value[0];
      if (only && only.id !== undefined) return only;
    }
    return undefined;
  }

  const selectionEntries: { attrId: number; valId: number }[] = [];
  for (const [attrIdStr, valIdStr] of Object.entries(selections)) {
    if (valIdStr === undefined) continue;
    const attrId = Number(attrIdStr);
    const valId = Number(valIdStr);
    if (Number.isNaN(attrId) || Number.isNaN(valId)) continue;
    selectionEntries.push({ attrId, valId });
  }
  if (selectionEntries.length === 0) return undefined;
  for (const sku of skus.value) {
    if (sku.id === undefined) continue;
    const combos = skuCombinationsMap[sku.id];
    if (!combos || combos.size === 0) continue;
    if (combos.size !== selectionEntries.length) continue;
    let matches = true;
    for (const { attrId, valId } of selectionEntries) {
      if (combos.get(attrId) !== valId) {
        matches = false;
        break;
      }
    }
    if (matches) return sku;
  }
  return undefined;
});

const matchedSkuPrice = computed<SkuPriceEntity | undefined>(() => {
  const sku = matchedSku.value;
  if (!sku || sku.id === undefined) return undefined;
  const prices = skuPricesMap[sku.id];
  if (!prices || prices.length === 0) return undefined;
  return prices.find((p) => p.currency === DISPLAY_CURRENCY) ?? prices[0];
});

const isOutOfStock = computed(() => {
  const sku = matchedSku.value;
  if (!sku) return false;
  return (sku.stockQty ?? 0) <= 0;
});

const canAddToCart = computed(() => {
  return matchedSku.value !== undefined && !isOutOfStock.value;
});

const displayPrice = computed(() => {
  const price = matchedSkuPrice.value;
  if (!price || !price.amount) return null;
  return price.amount;
});

// ---------- 当前用户的购物车（与 cart.vue 同样的按 userId 过滤逻辑） ----------
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

const createCartMutation = useCreateCart({
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['listCarts'] });
  },
  onError: (err: any) => {
    toast.error(err?.message || t('cart.errors.updateFailed'));
  },
});
const createCartItemMutation = useCreateCartItem({
  onSuccess: () => {
    // 购物车内容已变，刷新购物车项与购物车列表缓存
    queryClient.invalidateQueries({ queryKey: ['listCartItems'] });
    queryClient.invalidateQueries({ queryKey: ['listCarts'] });
    toast.success(t('mall.product.addToCart'));
  },
  onError: (err: any) => {
    toast.error(err?.message || t('cart.errors.updateFailed'));
  },
});

// ---------- 数量与库存 ----------
const quantity = ref(1);

// 当前匹配 SKU 的库存上限（无匹配 SKU 时回落 1，此时数量框禁用）
const matchedSkuStock = computed(() => {
  const sku = matchedSku.value;
  if (!sku || sku.id === undefined) return 1;
  const s = sku.stockQty ?? 1;
  return s < 1 ? 1 : s;
});

function incrementQty() {
  if (quantity.value < matchedSkuStock.value) {
    quantity.value++;
  }
}

function decrementQty() {
  if (quantity.value > 1) {
    quantity.value--;
  }
}

// ---------- 添加购物车 ----------
async function addToCartWithQty(qty: number) {
  if (!canAddToCart.value) return;
  if (!isLogin.value) {
    toast.error(t('authentication.login.please_login'));
    navigateTo(localePath('/login'));
    return;
  }
  const sku = matchedSku.value;
  if (sku?.id === undefined) return;

  let targetCartId = cart.value?.id;

  // 新用户尚无购物车：先创建
  if (targetCartId === undefined) {
    try {
      const created: any = await createCartMutation.mutateAsync({
        userId: currentUserId.value,
        tenantId: currentTenantId.value,
      } as any);
      targetCartId = created?.id;
    } catch {
      return;
    }
  }

  if (targetCartId === undefined) {
    toast.error(t('cart.errors.updateFailed'));
    return;
  }

  await createCartItemMutation.mutateAsync({
    cartId: targetCartId,
    skuId: sku.id,
    quantity: qty,
    tenantId: currentTenantId.value,
  } as any);
}

// ---------- 立即购买：加购后跳转结算页 ----------
async function buyNow() {
  if (!canAddToCart.value) return;
  await addToCartWithQty(quantity.value);
  navigateTo(localePath('/checkout'));
}

// ---------- SKU 卡片选择器：键盘方向键在当前组内切换 ----------
function onRadioKeydown(e: KeyboardEvent, attrId: number) {
  const vals = valuesByAttribute.value.get(attrId) ?? [];
  if (vals.length === 0) return;
  const cur = selections[attrId];
  let idx = vals.findIndex((v) => v.id !== undefined && String(v.id) === cur);
  if (idx < 0) idx = -1;
  let next = idx;
  if (e.key === 'ArrowRight' || e.key === 'ArrowDown') {
    next = (idx + 1) % vals.length;
  } else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') {
    next = (idx - 1 + vals.length) % vals.length;
  } else {
    return;
  }
  e.preventDefault();
  const nv = vals[next];
  if (nv?.id !== undefined) {
    selections[attrId] = String(nv.id);
  }
}
</script>

<template>
  <!-- 麵包屑導航（壓縮的左對齊標題區，替代原 PageHero） -->
  <LayoutSectionContainer class="!py-4">
    <nav class="flex items-center gap-2 text-sm text-muted-foreground">
      <NuxtLink :to="localePath('/')" class="hover:text-foreground">
        {{ t('mall.category.breadcrumbHome') }}
      </NuxtLink>
      <span>/</span>
      <span class="text-foreground">
        {{ productTranslation?.name || t('mall.product.notFound') }}
      </span>
    </nav>
  </LayoutSectionContainer>

  <LayoutSectionContainer>
    <UiSkeleton v-if="productLoading" class="h-8 w-1/2" />

    <div v-else-if="productTranslation" class="grid gap-8 md:grid-cols-12">
      <!-- 左侧：商品图片（1:1 正方形） -->
      <div class="md:col-span-5 rounded-2xl border border-border bg-card overflow-hidden">
        <UiCarousel class="relative">
          <UiCarouselContent>
            <UiCarouselItem>
              <UiImage
                :src="product.imageUrl"
                :alt="productTranslation?.name || ''"
                class="aspect-square w-full rounded-none object-cover"
              />
            </UiCarouselItem>
          </UiCarouselContent>
          <UiCarouselPrevious />
          <UiCarouselNext />
        </UiCarousel>
      </div>

      <!-- 右侧：商品信息 & SKU 选择 -->
      <div class="md:col-span-7 flex flex-col gap-6">
        <h1 class="text-3xl font-bold text-foreground">
          {{ productTranslation.name }}
        </h1>
        <p class="text-sm text-muted-foreground">
          {{ productTranslation.shortDescription }}
        </p>

        <!-- 价格区（高亮，无删除线，无原价） -->
        <div
          v-if="displayPrice !== null"
          class="rounded-xl border border-primary/40 bg-primary/5 p-5 dark:border-green-500/40 dark:bg-green-500/5"
        >
          <p class="text-xs text-muted-foreground">{{ t('mall.product.price') }}</p>
          <p class="mt-1 text-3xl font-extrabold text-primary dark:text-green-400">
            {{ t('mall.product.currencyCny') }}{{ displayPrice }}
          </p>
        </div>

        <!-- SKU 卡片选择器（aria-radiogroup，保留键盘可访问性） -->
        <div v-if="attributes.length > 0" class="flex flex-col gap-4">
          <p class="text-sm text-muted-foreground">{{ t('mall.product.selectSku') }}</p>
          <div v-for="attr in attributes" :key="attr.id" class="flex flex-col gap-2">
            <label class="text-xs font-medium text-foreground">
              {{ pickTranslation(attr.translations)?.name || '—' }}
            </label>
            <div
              role="radiogroup"
              :aria-label="pickTranslation(attr.translations)?.name || ''"
              class="flex flex-wrap gap-2"
              @keydown="onRadioKeydown($event, attr.id!)"
            >
              <button
                v-for="val in valuesByAttribute.get(attr.id!) ?? []"
                :key="val.id"
                type="button"
                role="radio"
                :aria-checked="selections[attr.id!] === String(val.id)"
                :class="cn(
                  'rounded-lg border px-4 py-2 text-sm transition-colors',
                  selections[attr.id!] === String(val.id)
                    ? 'border-primary bg-primary/10 text-primary dark:border-green-500 dark:bg-green-500/10 dark:text-green-400'
                    : 'border-border text-foreground hover:border-primary/40 dark:text-zinc-400',
                )"
                @click="selections[attr.id!] = String(val.id)"
              >
                {{ pickTranslation(val.translations)?.displayName || '—' }}
              </button>
            </div>
          </div>
        </div>

        <p
          v-if="allAttributesSelected && !matchedSku"
          class="text-sm text-destructive"
        >
          {{ t('mall.product.skuUnavailable') }}
        </p>
        <p
          v-else-if="matchedSku && isOutOfStock"
          class="text-sm text-destructive"
        >
          {{ t('mall.product.outOfStock') }}
        </p>

        <!-- 数量框 + 双按钮 -->
        <div class="mt-2 flex items-center gap-3">
          <div class="flex items-center rounded-xl border border-border bg-card">
            <button
              type="button"
              class="flex h-10 w-10 items-center justify-center text-foreground disabled:opacity-40"
              :disabled="quantity <= 1"
              @click="decrementQty"
              aria-label="decrease quantity"
            >
              <XIcon icon="carbon:subtract" :size="16" />
            </button>
            <input
              type="number"
              class="h-10 w-12 bg-transparent text-center text-sm text-foreground outline-none [appearance:textfield] [&::-webkit-inner-spin-button]:appearance-none"
              v-model.number="quantity"
              :min="1"
              :max="matchedSkuStock"
              readonly
            />
            <button
              type="button"
              class="flex h-10 w-10 items-center justify-center text-foreground disabled:opacity-40"
              :disabled="quantity >= matchedSkuStock"
              @click="incrementQty"
              aria-label="increase quantity"
            >
              <XIcon icon="carbon:add" :size="16" />
            </button>
          </div>

          <UiButton
            variant="outline"
            class="flex-1"
            size="lg"
            :disabled="!canAddToCart"
            @click="addToCartWithQty(quantity)"
          >
            <XIcon icon="carbon:shopping-cart" :size="16" class="mr-2" />
            {{ t('mall.product.addToCart') }}
          </UiButton>

          <UiButton
            class="flex-1"
            size="lg"
            :disabled="!canAddToCart"
            @click="buyNow"
          >
            {{ t('mall.product.buyNow') }}
          </UiButton>
        </div>
      </div>
    </div>

    <!-- 商品详情描述（通栏居中，限宽 max-w-4xl，增加内边距提升阅读呼吸感） -->
    <div
      v-if="productTranslation?.longDescription"
      class="mx-auto mt-12 w-full max-w-4xl rounded-2xl border border-border bg-card p-8 md:p-12"
    >
      <h2 class="mb-4 text-xl font-bold text-foreground">
        {{ t('mall.product.description') }}
      </h2>
      <ClientOnly>
        <ContentViewer
          :content="productTranslation.longDescription"
          type="markdown"
        />
        <template #fallback>
          <UiSkeleton class="h-64 w-full" />
        </template>
      </ClientOnly>
    </div>

    <div
      v-if="!productLoading && !productTranslation"
      class="rounded-2xl border border-border bg-card p-12 text-center text-muted-foreground"
    >
      {{ t('mall.product.notFound') }}
    </div>
  </LayoutSectionContainer>
</template>
