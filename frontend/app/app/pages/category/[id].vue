<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useGetCategory, useListCategories, useListProducts, useListBrands } from '@/api/composables';
import { getCurrentLocale } from '@/utils/locale';
import { XIcon } from '@/plugins/xicon';
import { PaginationQuery } from '@/core/transport/rest';

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
const categoryQuery = useGetCategory(computed(() => categoryId.value), {
  enabled: computed(() => categoryId.value > 0),
});
const category = computed<CategoryEntity | undefined>(() => {
  return categoryQuery.data?.value as CategoryEntity | undefined;
});
const categoryLoading = computed(() => categoryQuery.isPending.value);
const categoryError = computed(() => categoryQuery.isError.value);
const categoryTranslation = computed(() =>
  pickTranslation(category.value?.translations),
);

useHead({ title: () => categoryTranslation.value?.name || t('mall.category.products') });

// 路由 id 变化（同组件复用 /category/1→/category/2）滚回顶部，避免停留于
// 前一类目页底的滚动位置。
watch(categoryId, () => {
  if (import.meta.client) scrollToTop();
});

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

// 切换分类/品牌/排序时回到第 1 页，避免停留在越界页码。
watch([selectedSort, selectedBrandId, categoryId], () => {
  page.value = 1;
});

function goToPage(p: number) {
  const clamped = Math.min(Math.max(1, p), totalPages.value);
  if (clamped !== page.value) {
    page.value = clamped;
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
        <UiSkeleton v-if="categoryLoading" class="h-4 w-24" />
        <template v-else>
          {{ categoryTranslation?.name || t('mall.category.products') }}
        </template>
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

      <UiAppEmpty
        v-else-if="productsError"
        variant="error"
      >
        <template #action>
          <UiButton variant="outline" size="sm" @click="productsQuery.refetch()">
            {{ t('ui.button.retry') }}
          </UiButton>
        </template>
      </UiAppEmpty>

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

      <UiAppEmpty
        v-else-if="!category"
        variant="noData"
        :description="t('mall.category.notFound')"
      >
        <template #action>
          <UiButton variant="outline" @click="navigateTo(localePath('/'))">
            {{ t('cart.continueShopping') }}
          </UiButton>
        </template>
      </UiAppEmpty>

      <UiAppEmpty
        v-else
        variant="noData"
        :description="t('mall.category.empty')"
      >
        <template #action>
          <UiButton variant="outline" @click="navigateTo(localePath('/'))">
            {{ t('cart.continueShopping') }}
          </UiButton>
        </template>
      </UiAppEmpty>

      <div
        v-if="totalPages > 1"
        class="flex items-center justify-center gap-4 py-2"
      >
        <UiButton
          variant="outline"
          size="sm"
          :disabled="page <= 1"
          @click="goToPage(page - 1)"
        >
          {{ t('ui.pagination.previous') }}
        </UiButton>
        <span class="text-xs tabular-nums text-muted-foreground">
          {{ t('ui.pagination.page', { current: page, total: totalPages }) }}
        </span>
        <UiButton
          variant="outline"
          size="sm"
          :disabled="page >= totalPages"
          @click="goToPage(page + 1)"
        >
          {{ t('ui.pagination.next') }}
        </UiButton>
      </div>
    </div>
  </LayoutSectionContainer>
</template>
