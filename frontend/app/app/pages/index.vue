<script setup lang="ts">
import { computed } from 'vue';
import { useListCategories, useListProducts } from '@/api/composables';
import { getCurrentLocale } from '@/utils/locale';
import { XIcon } from '@/plugins/xicon';

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
  <LayoutPromoBanner />

  <!-- 分类导航：移动端金刚区 -->
  <LayoutCategoryQuickNav
    v-if="categoriesLoading || categories.length > 0"
    :categories="categories"
    class="md:hidden"
  />

  <!-- 分类导航：PC 端卡片网格 -->
  <LayoutSectionContainer class="hidden md:block">
    <h2 class="mb-6 text-2xl font-bold text-foreground">
      {{ t('mall.home.categories') }}
    </h2>

    <UiCategoryListSkeleton v-if="categoriesLoading" :count="8" />

    <div
      v-else-if="categories.length > 0"
      class="grid grid-cols-2 gap-4 md:grid-cols-4 lg:grid-cols-5"
    >
      <NuxtLink
        v-for="cat in categories"
        :key="cat.id"
        :to="localePath('/category/' + cat.id)"
        class="group block overflow-hidden rounded-xl border border-border bg-card transition-colors hover:border-primary/60"
      >
        <UiImage
          :src="cat.imageUrl"
          :alt="pickTranslation(cat.translations)?.name || ''"
          class="aspect-square w-full rounded-none object-cover"
        />
        <div class="p-3">
          <h3 class="line-clamp-1 text-sm font-medium text-foreground">
            {{ pickTranslation(cat.translations)?.name || '—' }}
          </h3>
        </div>
      </NuxtLink>
    </div>

    <p v-else class="text-muted-foreground">{{ t('mall.loading') }}</p>
  </LayoutSectionContainer>

  <!-- 精选商品 -->
  <LayoutSectionContainer>
    <h2 class="mb-6 text-2xl font-bold text-foreground">
      {{ t('mall.home.featured') }}
    </h2>

    <div v-if="featuredLoading" class="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-4">
      <UiPostCardSkeleton v-for="_, i in 8" :key="i" />
    </div>

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
          class="aspect-[3/4] w-full rounded-none object-cover"
        />
        <div class="p-3">
          <h3 class="line-clamp-2 text-sm font-medium text-foreground">
            {{ pickTranslation(product.translations)?.name || '—' }}
          </h3>
          <div class="mt-3 flex items-center justify-between">
            <span class="text-xs text-muted-foreground">{{ t('mall.product.viewDetail') }}</span>
            <span class="flex h-7 w-7 items-center justify-center rounded-full bg-primary/10 text-primary transition-colors group-hover:bg-primary group-hover:text-primary-foreground">
              <XIcon icon="carbon:shopping-cart" :size="14" />
            </span>
          </div>
        </div>
      </NuxtLink>
    </div>

    <p v-else class="text-muted-foreground">{{ t('mall.loading') }}</p>
  </LayoutSectionContainer>
</template>
