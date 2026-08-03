<script setup lang="ts">
import { computed } from 'vue';
import { useListCategories, useListProducts } from '@/api/composables';
import { getCurrentLocale } from '@/utils/locale';

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
  translations?: CategoryTranslation[];
};

type ProductTranslation = {
  name?: string;
  languageCode?: string;
};
type ProductEntity = {
  id?: number;
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

const featuredQuery = useListProducts({
  page: 1,
  pageSize: 12,
  noPaging: false,
  sorting: [{ field: 'sort_order', direction: 'asc' }],
});
const featured = computed<ProductEntity[]>(() => {
  const items = (featuredQuery.data?.value as any)?.items ?? [];
  return (items as ProductEntity[]) ?? [];
});
const featuredLoading = computed(() => featuredQuery.isLoading.value);
</script>

<template>
  <LayoutPageHero
    :title="t('mall.home.title')"
    :subtitle="t('mall.home.subtitle')"
    :description="t('mall.home.heroDescription')"
    icon="carbon:shopping-bag"
    size="lg"
  />

  <!-- 分类导航 -->
  <LayoutSectionContainer>
    <h2 class="mb-6 text-2xl font-bold text-foreground">
      {{ t('mall.home.categories') }}
    </h2>

    <UiCategoryListSkeleton v-if="categoriesLoading" :count="4" />

    <div
      v-else-if="categories.length > 0"
      class="grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-6"
    >
      <NuxtLink
        v-for="cat in categories"
        :key="cat.id"
        :to="localePath('/category/' + cat.id)"
        class="group block rounded-2xl border border-border bg-card p-6 transition-colors hover:border-primary/60"
      >
        <div class="flex h-32 items-center justify-center rounded-xl bg-primary/5 text-5xl">
          🛍️
        </div>
        <h3 class="mt-4 line-clamp-1 text-base font-semibold text-foreground">
          {{ pickTranslation(cat.translations)?.name || '—' }}
        </h3>
      </NuxtLink>
    </div>

    <p v-else class="text-muted-foreground">{{ t('mall.loading') }}</p>
  </LayoutSectionContainer>

  <!-- 精选商品 -->
  <LayoutSectionContainer>
    <h2 class="mb-6 text-2xl font-bold text-foreground">
      {{ t('mall.home.featured') }}
    </h2>

    <UiCategoryListSkeleton v-if="featuredLoading" :count="4" />

    <div
      v-else-if="featured.length > 0"
      class="grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-6"
    >
      <NuxtLink
        v-for="product in featured"
        :key="product.id"
        :to="localePath('/product/' + product.id)"
        class="group block overflow-hidden rounded-2xl border border-border bg-card transition-colors hover:border-primary/60"
      >
        <div class="flex h-40 items-center justify-center bg-primary/5 text-5xl">
          📦
        </div>
        <div class="p-4">
          <h3 class="line-clamp-2 text-sm font-medium text-foreground">
            {{ pickTranslation(product.translations)?.name || '—' }}
          </h3>
          <p class="mt-3 text-xs text-muted-foreground">
            {{ t('mall.product.viewDetail') }}
          </p>
        </div>
      </NuxtLink>
    </div>

    <p v-else class="text-muted-foreground">{{ t('mall.loading') }}</p>
  </LayoutSectionContainer>
</template>
