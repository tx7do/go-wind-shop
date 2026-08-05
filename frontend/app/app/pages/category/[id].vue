<script setup lang="ts">
import { computed, ref } from 'vue';
import { useGetCategory, useListCategories, useListProducts, useListBrands } from '@/api/composables';
import { getCurrentLocale } from '@/utils/locale';
import { XIcon } from '@/plugins/xicon';

const route = useRoute();
const { t } = useI18n();
const localePath = useLocalePath();

const categoryId = computed(() => Number(route.params.id));
const currentLocale = computed(() => getCurrentLocale());

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

function pickTranslation<T extends { languageCode?: string }>(
  translations: T[] | undefined,
): T | undefined {
  if (!translations || translations.length === 0) return undefined;
  const match = translations.find((tr) => tr.languageCode === currentLocale.value);
  return match ?? translations[0];
}

// 类目详情（Get 注入 locale，返回匹配语言的翻译）
const categoryQuery = useGetCategory(categoryId.value, {
  enabled: categoryId.value > 0,
});
const category = computed<CategoryEntity | undefined>(() => {
  return categoryQuery.data?.value as CategoryEntity | undefined;
});
const categoryLoading = computed(() => categoryQuery.isLoading.value);
const categoryTranslation = computed(() =>
  pickTranslation(category.value?.translations),
);

useHead({ title: () => categoryTranslation.value?.name || t('mall.category.products') });

// 同级分类列表（用于 FilterSidebar 分类切换）
const categoriesQuery = useListCategories({
  page: 1,
  pageSize: 24,
  noPaging: false,
});
const categories = computed<CategoryEntity[]>(() => {
  const items = (categoriesQuery.data?.value as any)?.items ?? [];
  return (items as CategoryEntity[]) ?? [];
});

// 品牌列表（用于 FilterSidebar 品牌筛选）
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

// 该分类下的商品（按 categoryId 过滤 + 可选 brandId + 排序）
// sorting 传单对象（修正 generated 客户端第 332 行把 sorting 当单对象序列化的 bug，
// 传数组形式时 .field 为 undefined，排序参数不会发送）
const productsQuery = useListProducts(
  computed(() => {
    const q: Record<string, unknown> = { categoryId: categoryId.value };
    if (selectedBrandId.value !== undefined) {
      q.brandId = selectedBrandId.value;
    }
    const sorting =
      selectedSort.value === 'latest'
        ? { field: 'created_at', direction: 'DESC' }
        : { field: 'sort_order', direction: 'ASC' };
    return {
      page: 1,
      pageSize: 48,
      noPaging: false,
      query: JSON.stringify(q),
      sorting: sorting as any,
    };
  }),
);
const products = computed<ProductEntity[]>(() => {
  const items = (productsQuery.data?.value as any)?.items ?? [];
  return (items as ProductEntity[]) ?? [];
});
const productsLoading = computed(() => productsQuery.isLoading.value);
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
        {{ categoryTranslation?.name || t('mall.category.products') }}
      </span>
    </nav>
  </LayoutSectionContainer>

  <!-- 左右分欄：PC 端 FilterSidebar(1/4) + 商品區(3/4)，移動端單列 -->
  <LayoutSectionContainer class="grid grid-cols-1 gap-6 md:grid-cols-4">
    <CategoryFilterSidebar
      :categories="categories"
      :brands="brands"
      :current-category-id="categoryId"
      :selected-brand-id="selectedBrandId"
      class="md:col-span-1"
      @update:selected-brand-id="selectedBrandId = $event"
    />

    <div class="md:col-span-3">
      <CategorySortBar
        v-model="selectedSort"
        class="mb-4"
      />

      <div v-if="productsLoading" class="grid grid-cols-2 gap-4 md:grid-cols-3">
        <UiPostCardSkeleton v-for="_, i in 6" :key="i" />
      </div>

      <div
        v-else-if="products.length > 0"
        class="grid grid-cols-2 gap-4 md:grid-cols-3"
      >
        <NuxtLink
          v-for="product in products"
          :key="product.id"
          :to="localePath('/product/' + product.id)"
          class="group block overflow-hidden rounded-xl border border-border bg-card shadow-sm transition-colors hover:border-primary/60 dark:shadow-none"
        >
          <UiImage
            :src="product.imageUrl"
            :alt="pickTranslation(product.translations)?.name || ''"
            class="aspect-square w-full rounded-none object-cover"
          />
          <div class="p-3">
            <h3 class="line-clamp-2 text-sm font-semibold text-foreground dark:text-slate-200">
              {{ pickTranslation(product.translations)?.name || '—' }}
            </h3>
            <div class="mt-3 flex items-center justify-between">
              <span class="text-xs text-muted-foreground">{{ t('mall.product.viewDetail') }}</span>
              <span class="flex h-7 w-7 items-center justify-center rounded-full bg-primary/10 text-primary transition-colors group-hover:bg-primary group-hover:text-primary-foreground dark:bg-green-500/20 dark:text-green-400 dark:group-hover:bg-green-500 dark:group-hover:text-green-950">
                <XIcon icon="carbon:shopping-cart" :size="14" />
              </span>
            </div>
          </div>
        </NuxtLink>
      </div>

      <div
        v-else
        class="flex flex-col items-center gap-4 rounded-2xl border border-border bg-card p-12 text-center"
      >
        <XIcon icon="carbon:document" :size="48" class="text-muted-foreground" />
        <p class="text-sm text-muted-foreground">{{ t('mall.category.empty') }}</p>
        <UiButton variant="outline" @click="navigateTo(localePath('/'))">
          {{ t('cart.continueShopping') }}
        </UiButton>
      </div>
    </div>
  </LayoutSectionContainer>
</template>
