<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import {
  useGetCategory,
  useListCategories,
  useListProducts,
  useListBrands,
  useProductPrices,
  useProductPromos,
} from '@/api/composables';
import { XIcon } from '@/plugins/xicon';
import { PaginationQuery } from '@/core/transport/rest';
import { pickTranslation, translationName } from '@/utils/translation';

const route = useRoute();
const { t } = useI18n();
const localePath = useLocalePath();

const categoryId = computed(() => Number(route.params.id));

type CategoryTranslation = {
  name?: string;
  description?: string;
  languageCode?: string;
};
type CategoryEntity = {
  id?: number;
  translations?: CategoryTranslation[];
};
type BrandEntity = {
  id?: number;
  translations?: Array<{ name?: string; languageCode?: string }>;
};
type ProductTranslation = {
  name?: string;
  languageCode?: string;
};
type ProductEntity = {
  id?: number;
  imageUrl?: string;
  translations?: ProductTranslation[];
};

// 类目详情（Get 注入 locale，返回匹配语言的翻译）
const categoryQuery = useGetCategory(computed(() => categoryId.value), {
  enabled: computed(() => categoryId.value > 0),
});
const category = computed<CategoryEntity | undefined>(() => {
  return categoryQuery.data?.value as CategoryEntity | undefined;
});
const categoryLoading = computed(() => categoryQuery.isPending.value);
const categoryTranslation = computed(() =>
  pickTranslation(category.value?.translations),
);

useHead({ title: () => categoryTranslation.value?.name || t('mall.category.products') });

// 路由 id 变化（同组件复用 /category/1→/category/2）滚回顶部，避免停留于
// 前一类目页底的滚动位置。
watch(categoryId, () => {
  if (import.meta.client) scrollToTop();
});

// 同级分类列表（用于分类切换）
const categoriesQuery = useListCategories({
  page: 1,
  pageSize: 24,
  noPaging: false,
});
const categories = computed<CategoryEntity[]>(() => {
  const items = (categoriesQuery.data?.value as any)?.items ?? [];
  return (items as CategoryEntity[]) ?? [];
});

// 品牌列表（用于品牌筛选）
const brandsQuery = useListBrands({
  page: 1,
  pageSize: 24,
  noPaging: false,
});
const brands = computed<BrandEntity[]>(() => {
  const items = (brandsQuery.data?.value as any)?.items ?? [];
  return (items as BrandEntity[]) ?? [];
});

// 排序状态：'featured' → sort_order ASC，'latest' → created_at DESC
const selectedSort = ref<'featured' | 'latest'>('featured');
// 品牌筛选状态：undefined 表示不筛选，否则按 brandId 过滤
const selectedBrandId = ref<number | undefined>(undefined);

const sortOptions: Array<{ key: 'featured' | 'latest'; labelKey: string }> = [
  { key: 'featured', labelKey: 'mall.category.sortFeatured' },
  { key: 'latest', labelKey: 'mall.category.sortLatest' },
];

// 品牌筛选单选取值：选中未选中的品牌则设为当前，再次点击已选中的品牌则取消。
function toggleBrand(brandId?: number) {
  if (brandId === undefined) return;
  selectedBrandId.value = selectedBrandId.value === brandId ? undefined : brandId;
}

// 该分类下的商品（按 categoryId 过滤 + 可选 brandId + 排序）
// 排序走 orderBy 字符串数组（后端 ListWithPaging 的 orderByStringConverter 路径，
// 非 sorting 结构化切片——后者被 generated 客户端按单对象序列化，与后端
// []*Sorting 切片不匹配，排序参数会丢失）。
const PAGE_SIZE = 48;
const page = ref(1);
const productsQuery = useListProducts(
  computed(() => {
    const q: Record<string, unknown> = { categoryId: categoryId.value };
    if (selectedBrandId.value !== undefined) {
      q.brandId = selectedBrandId.value;
    }
    // orderBy 用 "-" 前缀表示 DESC，无前缀 ASC（与后端 OrderByStringConverter 约定一致）
    const orderBy =
      selectedSort.value === 'latest' ? ['-created_at'] : ['sort_order'];
    return new PaginationQuery({
      paging: { page: page.value, pageSize: PAGE_SIZE },
      formValues: q,
      orderBy,
    });
  }),
  { enabled: computed(() => categoryId.value > 0) },
);
const products = computed<ProductEntity[]>(() => {
  const items = (productsQuery.data?.value as any)?.items ?? [];
  return (items as ProductEntity[]) ?? [];
});
const productsLoading = computed(() => productsQuery.isLoading.value);
const productsError = computed(() => productsQuery.isError.value);

const totalCount = computed(() => (productsQuery.data?.value as any)?.total ?? 0);
const totalPages = computed(() => Math.max(1, Math.ceil(totalCount.value / PAGE_SIZE)));

// 商品实体里没有价格字段（价格在 SkuPrice，按 SKU 存「分」），
// 两次批量查询拿整屏商品的「起」价与活动价，避免每商品各发两次请求。
const productIds = computed(() =>
  products.value.map((p) => p.id).filter((id): id is number => id !== undefined),
);
const { priceByProductId, skuIdsByProduct } = useProductPrices(productIds);
const { promoByProductId } = useProductPromos(productIds, priceByProductId, skuIdsByProduct);

const cards = computed(() =>
  products.value
    .map((p) => {
      if (p.id === undefined) return undefined;
      const base = priceByProductId.value.get(p.id);
      const promo = promoByProductId.value.get(p.id);
      return {
        id: p.id,
        name: translationName(p.translations),
        image: p.imageUrl,
        priceFen: promo?.fen ?? base?.fen ?? null,
        oldPriceFen: promo ? base?.fen ?? null : null,
        badgeLabel: promo ? t(promo.labelKey) : undefined,
        badgeTone: promo?.tone,
      };
    })
    .filter((c): c is NonNullable<typeof c> => c !== undefined),
);

// 切换分类/品牌/排序时回到第 1 页，避免停留在越界页码。
watch([selectedSort, selectedBrandId, categoryId], () => {
  page.value = 1;
});
</script>

<template>
  <LayoutSectionContainer no-padding class="pt-3 md:pt-5">
    <nav class="flex items-center gap-1.5 text-xs text-muted-foreground md:text-sm">
      <NuxtLink
        :to="localePath('/')"
        class="shrink-0 transition-colors hover:text-foreground"
      >
        {{ t('mall.category.breadcrumbHome') }}
      </NuxtLink>
      <XIcon
        v-if="categoryLoading || category"
        icon="lucide:chevron-right"
        :size="12"
        class="shrink-0 text-muted-foreground"
      />
      <UiSkeleton v-if="categoryLoading" class="h-3.5 w-24" />
      <span v-else-if="category" class="line-clamp-1 min-w-0 font-medium text-foreground">
        {{ translationName(category.translations) }}
      </span>
    </nav>
  </LayoutSectionContainer>

  <LayoutSectionContainer class="md:grid md:grid-cols-[13rem_1fr] md:items-start md:gap-6 md:pt-5">
    <aside
      v-if="(categories.length > 0 || brands.length > 0) && (categoryLoading || category)"
      class="hidden md:block"
    >
      <div class="sticky top-[calc(var(--layout-header-height)+1rem)] rounded-lg border border-border bg-card p-4 shadow-card">
        <h2 class="mb-3 text-sm font-bold text-foreground">
          {{ t('mall.category.filters') }}
        </h2>

        <div v-if="categories.length > 0">
          <p class="mb-2 text-xs font-semibold text-muted-foreground">
            {{ t('mall.category.filterCategories') }}
          </p>
          <ul class="space-y-0.5">
            <li v-for="cat in categories" :key="cat.id">
              <NuxtLink
                :to="localePath('/category/' + cat.id)"
                :class="[
                  'flex items-center rounded-md px-2.5 py-2 text-sm transition-colors',
                  cat.id === categoryId
                    ? 'bg-primary-soft font-medium text-primary-soft-foreground'
                    : 'text-foreground hover:bg-secondary',
                ]"
              >
                <span class="truncate">{{ translationName(cat.translations) }}</span>
              </NuxtLink>
            </li>
          </ul>
        </div>

        <div v-if="brands.length > 0" class="mt-5 border-t border-border pt-4">
          <p class="mb-2 text-xs font-semibold text-muted-foreground">
            {{ t('mall.category.filterBrands') }}
          </p>
          <ul class="space-y-0.5">
            <li v-for="brand in brands" :key="brand.id">
              <UiButton
                type="button"
                variant="ghost"
                :class="[
                  'h-auto w-full justify-start gap-2 rounded-md px-2.5 py-2 text-left text-sm',
                  brand.id === selectedBrandId
                    ? 'bg-primary-soft font-medium text-primary-soft-foreground'
                    : 'text-foreground',
                ]"
                :aria-pressed="brand.id === selectedBrandId"
                @click="toggleBrand(brand.id)"
              >
                <XIcon
                  v-if="brand.id === selectedBrandId"
                  icon="lucide:check"
                  class="shrink-0"
                />
                <span class="truncate">{{ translationName(brand.translations) }}</span>
              </UiButton>
            </li>
          </ul>
          <UiButton
            v-if="selectedBrandId !== undefined"
            variant="link"
            size="sm"
            class="mt-2 px-0 text-xs"
            @click="selectedBrandId = undefined"
          >
            {{ t('common.reset') }}
          </UiButton>
        </div>
      </div>
    </aside>

    <div class="min-w-0">
      <div
        v-if="categoryLoading || category"
        class="mb-4 flex flex-wrap items-start justify-between gap-3"
      >
        <div class="min-w-0">
          <h1 class="line-clamp-1 text-lg font-bold text-foreground md:text-xl">
            <UiSkeleton v-if="categoryLoading" class="h-6 w-32" />
            <template v-else>{{ translationName(category?.translations) }}</template>
          </h1>
          <p class="mt-1 line-clamp-2 text-xs text-muted-foreground md:text-sm">
            {{ categoryTranslation?.description || t('mall.category.products') }}
          </p>
        </div>

        <div class="flex shrink-0 items-center gap-2">
          <span class="text-xs font-medium text-muted-foreground">
            {{ t('mall.category.sortBy') }}
          </span>
          <UiSelect v-model="selectedSort">
            <UiSelectTrigger class="h-8 w-[7.5rem] text-xs md:text-sm">
              <UiSelectValue />
            </UiSelectTrigger>
            <UiSelectContent>
              <UiSelectItem
                v-for="opt in sortOptions"
                :key="opt.key"
                :value="opt.key"
              >
                {{ t(opt.labelKey) }}
              </UiSelectItem>
            </UiSelectContent>
          </UiSelect>
        </div>
      </div>

      <div v-if="categoryLoading || category" class="mb-3 space-y-2 md:hidden">
        <div
          v-if="categories.length > 0"
          class="-mx-4 flex gap-2 overflow-x-auto px-4 pb-1 no-scrollbar"
        >
          <NuxtLink
            v-for="cat in categories"
            :key="cat.id"
            :to="localePath('/category/' + cat.id)"
            :class="[
              'shrink-0 rounded-full border px-3 py-1.5 text-xs font-medium transition-colors',
              cat.id === categoryId
                ? 'border-primary bg-primary-soft text-primary-soft-foreground'
                : 'border-border bg-card text-foreground',
            ]"
          >
            {{ translationName(cat.translations) }}
          </NuxtLink>
        </div>

        <div
          v-if="brands.length > 0"
          class="-mx-4 flex items-center gap-2 overflow-x-auto px-4 pb-1 no-scrollbar"
        >
          <span class="shrink-0 text-xs text-muted-foreground">
            {{ t('mall.category.filterBrands') }}
          </span>
          <UiButton
            v-for="brand in brands"
            :key="brand.id"
            type="button"
            variant="outline"
            size="sm"
            :class="[
              'shrink-0 rounded-full',
              brand.id === selectedBrandId
                ? 'border-primary bg-primary-soft text-primary-soft-foreground'
                : '',
            ]"
            :aria-pressed="brand.id === selectedBrandId"
            @click="toggleBrand(brand.id)"
          >
            {{ translationName(brand.translations) }}
          </UiButton>
          <UiButton
            v-if="selectedBrandId !== undefined"
            variant="link"
            size="sm"
            class="shrink-0 px-0 text-xs"
            @click="selectedBrandId = undefined"
          >
            {{ t('common.reset') }}
          </UiButton>
        </div>
      </div>

      <UiListSkeleton v-if="productsLoading" variant="card" :rows="6" :columns="3" />

      <UiAppEmpty v-else-if="productsError" variant="error">
        <template #action>
          <UiButton variant="outline" size="sm" @click="productsQuery.refetch()">
            {{ t('ui.button.retry') }}
          </UiButton>
        </template>
      </UiAppEmpty>

      <UiAppEmpty v-else-if="!categoryLoading && !category" :description="t('mall.category.notFound')">
        <template #action>
          <UiButton variant="outline" @click="navigateTo(localePath('/'))">
            {{ t('cart.continueShopping') }}
          </UiButton>
        </template>
      </UiAppEmpty>

      <div v-else-if="cards.length > 0" class="grid grid-cols-2 gap-3 md:grid-cols-3 md:gap-4">
        <UiProductCard
          v-for="card in cards"
          :key="card.id"
          :product-id="card.id"
          :name="card.name"
          :image="card.image"
          :price-fen="card.priceFen"
          :old-price-fen="card.oldPriceFen"
          :badge-label="card.badgeLabel"
          :badge-tone="card.badgeTone"
        />
      </div>

      <UiAppEmpty v-else :description="t('mall.category.empty')">
        <template #action>
          <UiButton variant="outline" @click="navigateTo(localePath('/'))">
            {{ t('cart.continueShopping') }}
          </UiButton>
        </template>
      </UiAppEmpty>

      <UiPager v-model:page="page" :total-pages="totalPages" class="mt-6" />
    </div>
  </LayoutSectionContainer>
</template>
