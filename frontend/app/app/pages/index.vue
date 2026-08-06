<script setup lang="ts">
import { computed } from 'vue';
import { useListCategories, useListProducts } from '@/api/composables';
import { getCurrentLocale } from '@/utils/locale';
import { XIcon } from '@/plugins/xicon';
import { PaginationQuery } from '@/core/transport/rest';

const { t } = useI18n();
const localePath = useLocalePath();

useHead({ title: t('mall.home.title') });

const currentLocale = computed(() => getCurrentLocale());

type CategoryTranslation = {
  name?: string;
  description?: string;
  languageCode?: string;
};
type CategoryEntity = {
  id?: number;
  imageUrl?: string;
  translations?: CategoryTranslation[];
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

const categoriesQuery = useListCategories({
  page: 1,
  pageSize: 24,
  noPaging: false,
});
const categories = computed<CategoryEntity[]>(() => {
  const items = (categoriesQuery.data?.value as any)?.items ?? [];
  return (items as CategoryEntity[]) ?? [];
});
const categoriesLoading = computed(() => categoriesQuery.isLoading.value);
const categoriesError = computed(() => categoriesQuery.isError.value);

const featuredQuery = useListProducts(
  new PaginationQuery({
    paging: { page: 1, pageSize: 12 },
    orderBy: ['sort_order'],
  }).toRawParams(),
);
const featured = computed<ProductEntity[]>(() => {
  const items = (featuredQuery.data?.value as any)?.items ?? [];
  return (items as ProductEntity[]) ?? [];
});
const featuredLoading = computed(() => featuredQuery.isLoading.value);
const featuredError = computed(() => featuredQuery.isError.value);
</script>

<template>
  <!-- 首屏複合物件：PC 端左側垂直分類菜單 + 右側大寬屏 Banner（md 以上渲染） -->
  <div class="hidden md:grid md:grid-cols-5 gap-6 px-8 mx-auto max-w-[1200px]">
    <LayoutCategorySideMenu
      v-if="categoriesLoading || categories.length > 0"
      :categories="categories"
      class="md:col-span-1"
    />
    <LayoutHeroBanner class="md:col-span-4" />
  </div>

  <!-- 分类导航：移动端金刚区 -->
  <LayoutCategoryQuickNav
    v-if="categoriesLoading || categories.length > 0"
    :categories="categories"
    class="md:hidden"
  />

  <!-- 精选商品 -->
  <LayoutSectionContainer>
    <h2 class="mb-6 text-2xl font-bold text-foreground">
      {{ t('mall.home.featured') }}
    </h2>

    <div v-if="featuredLoading" class="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-4">
      <UiPostCardSkeleton v-for="_, i in 8" :key="i" />
    </div>

    <UiAppEmpty
      v-else-if="featuredError"
      variant="error"
    >
      <template #action>
        <UiButton variant="outline" size="sm" @click="featuredQuery.refetch()">
          {{ t('ui.button.retry') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <div
      v-else-if="featured.length > 0"
      class="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-4"
    >
      <NuxtLink
        v-for="product in featured"
        :key="product.id"
        :to="localePath('/product/' + product.id)"
        class="group block overflow-hidden rounded-xl border border-border bg-card transition-colors hover:border-primary/60"
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

    <p v-else class="text-muted-foreground">{{ t('mall.loading') }}</p>
  </LayoutSectionContainer>
</template>
