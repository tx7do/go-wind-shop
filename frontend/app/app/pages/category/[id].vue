<script setup lang="ts">
import { computed, watch } from 'vue';
import { useGetCategory, useListProducts } from '@/api/composables';
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
    :description="categoryTranslation?.description"
    icon="carbon:category"
    size="md"
  />

  <LayoutSectionContainer>
    <h2 class="mb-6 text-2xl font-bold text-foreground">
      {{ t('mall.category.products') }}
    </h2>

    <div v-if="productsLoading" class="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-4">
      <UiPostCardSkeleton v-for="_, i in 6" :key="i" />
    </div>

    <div
      v-else-if="products.length > 0"
      class="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-4"
    >
      <NuxtLink
        v-for="product in products"
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
  </LayoutSectionContainer>
</template>
