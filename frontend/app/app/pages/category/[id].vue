<script setup lang="ts">
import { computed, watch } from 'vue';
import { useGetCategory, useListProducts } from '@/api/composables';
import { getCurrentLocale } from '@/utils/locale';

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

// 该分类下的商品（按 categoryId 过滤）
const productsQuery = useListProducts(
  computed(() => ({
    page: 1,
    pageSize: 48,
    noPaging: false,
    query: JSON.stringify({ categoryId: categoryId.value }),
  })),
);
const products = computed<ProductEntity[]>(() => {
  const items = (productsQuery.data?.value as any)?.items ?? [];
  return (items as ProductEntity[]) ?? [];
});
const productsLoading = computed(() => productsQuery.isLoading.value);
</script>

<template>
  <LayoutPageHero
    :title="categoryTranslation?.name || t('mall.category.products')"
    :subtitle="categoryTranslation?.name"
    :description="categoryTranslation?.description"
    icon="carbon:category"
    size="md"
  />

  <LayoutSectionContainer>
    <h2 class="mb-6 text-2xl font-bold text-foreground">
      {{ t('mall.category.products') }}
    </h2>

    <UiCategoryListSkeleton v-if="productsLoading" :count="6" />

    <div
      v-else-if="products.length > 0"
      class="grid grid-cols-[repeat(auto-fill,minmax(260px,1fr))] gap-6"
    >
      <NuxtLink
        v-for="product in products"
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

    <div
      v-else
      class="rounded-2xl border border-border bg-card p-12 text-center text-muted-foreground"
    >
      {{ t('mall.category.empty') }}
    </div>
  </LayoutSectionContainer>
</template>
