<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue';
import { toast } from 'vue-sonner';
import {
  useGetProduct,
  useListProductAttributes,
  useListProductAttributeValues,
  useListSkus,
  fetchListSkuAttributeCombinationsStore,
  fetchListSkuPricesStore,
} from '@/api/composables';
import { getCurrentLocale } from '@/utils/locale';

const route = useRoute();
const { t } = useI18n();
const localePath = useLocalePath();

const productId = computed(() => Number(route.params.id));
const currentLocale = computed(() => getCurrentLocale());

// ---------- 类型 ----------
type ProductTranslation = {
  name?: string;
  shortDescription?: string;
  longDescription?: string;
  languageCode?: string;
};
type ProductEntity = {
  id?: number;
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
  const map = new Map<number, AttributeValueEntity[]>();
  for (const v of allAttributeValues.value) {
    if (v.attributeId === undefined) continue;
    const arr = map.get(v.attributeId) ?? [];
    arr.push(v);
    map.set(v.attributeId, arr);
  }
  return map;
});

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
  for (const attr of attributes.value) {
    if (attr.id === undefined) continue;
    if (selections[attr.id] === undefined) return false;
  }
  return attributes.value.length > 0;
});

// ---------- 匹配 SKU ----------
const matchedSku = computed<SkuEntity | undefined>(() => {
  if (!allAttributesSelected.value) return undefined;
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

// ---------- 添加购物车（占位） ----------
function addToCart() {
  if (!canAddToCart.value) return;
  toast.success(t('mall.product.addToCart'));
}
</script>

<template>
  <LayoutPageHero
    :title="productTranslation?.name || ''"
    :description="productTranslation?.shortDescription"
    icon="carbon:product"
    size="sm"
  />

  <LayoutSectionContainer>
    <UiSkeleton v-if="productLoading" class="h-8 w-1/2" />

    <div v-else-if="productTranslation" class="grid gap-10 md:grid-cols-2">
      <!-- 图片轮播（占位） -->
      <div class="rounded-2xl border border-border bg-card overflow-hidden">
        <UiCarousel class="relative">
          <UiCarouselContent>
            <UiCarouselItem>
              <div class="flex h-96 items-center justify-center bg-primary/5 text-7xl">
                📦
              </div>
            </UiCarouselItem>
          </UiCarouselContent>
          <UiCarouselPrevious />
          <UiCarouselNext />
        </UiCarousel>
      </div>

      <!-- 信息 & SKU 选择 -->
      <div class="flex flex-col gap-6">
        <h1 class="text-3xl font-bold text-foreground">
          {{ productTranslation.name }}
        </h1>
        <p class="text-sm text-muted-foreground">
          {{ productTranslation.shortDescription }}
        </p>

        <div
          v-if="displayPrice !== null"
          class="rounded-xl border border-border bg-primary/5 p-4"
        >
          <p class="text-xs text-muted-foreground">{{ t('mall.product.price') }}</p>
          <p class="mt-1 text-2xl font-bold text-primary">
            {{ t('mall.product.currencyCny') }}{{ displayPrice }}
          </p>
        </div>

        <!-- SKU 选择器 -->
        <div v-if="attributes.length > 0" class="flex flex-col gap-4">
          <p class="text-sm text-muted-foreground">{{ t('mall.product.selectSku') }}</p>
          <div v-for="attr in attributes" :key="attr.id" class="flex flex-col gap-2">
            <label class="text-xs font-medium text-foreground">
              {{ pickTranslation(attr.translations)?.name || '—' }}
            </label>
            <UiSelect v-model="selection[attr.id!]">
              <UiSelectTrigger>
                <UiSelectValue
                  :placeholder="t('mall.product.selectAttribute')"
                />
              </UiSelectTrigger>
              <UiSelectContent>
                <UiSelectItem
                  v-for="val in valuesByAttribute.get(attr.id!) ?? []"
                  :key="val.id"
                  :value="String(val.id)"
                >
                  {{ pickTranslation(val.translations)?.displayName || '—' }}
                </UiSelectItem>
              </UiSelectContent>
            </UiSelect>
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

        <UiButton
          class="mt-2"
          size="lg"
          :disabled="!canAddToCart"
          @click="addToCart"
        >
          {{
            isOutOfStock
              ? t('mall.product.outOfStock')
              : t('mall.product.addToCart')
          }}
        </UiButton>
      </div>
    </div>

    <!-- 商品详情描述 -->
    <div
      v-if="productTranslation?.longDescription"
      class="mt-12 rounded-2xl border border-border bg-card p-8"
    >
      <h2 class="mb-4 text-xl font-bold text-foreground">
        {{ t('mall.product.description') }}
      </h2>
      <article class="prose prose-sm max-w-none text-muted-foreground">
        {{ productTranslation.longDescription }}
      </article>
    </div>

    <div
      v-if="!productLoading && !productTranslation"
      class="rounded-2xl border border-border bg-card p-12 text-center text-muted-foreground"
    >
      {{ t('mall.product.notFound') }}
    </div>
  </LayoutSectionContainer>
</template>
