<script setup lang="ts">
import { computed } from 'vue';
import {
  useListCategories,
  useListProducts,
  useProductPrices,
  useProductPromos,
} from '@/api/composables';
import { XIcon } from '@/plugins/xicon';
import { PaginationQuery } from '@/core/transport/rest';
import { translationName } from '@/utils/translation';

const { t } = useI18n();
const localePath = useLocalePath();

useHead({ title: t('mall.home.title') });

type CategoryEntity = {
  id?: number;
  imageUrl?: string;
  translations?: Array<{ name?: string; languageCode?: string }>;
};
type ProductEntity = {
  id?: number;
  imageUrl?: string;
  translations?: Array<{ name?: string; languageCode?: string }>;
};

const categoriesQuery = useListCategories({
  page: 1,
  pageSize: 24,
  noPaging: false,
});
const categories = computed<CategoryEntity[]>(
  () => ((categoriesQuery.data?.value as any)?.items ?? []) as CategoryEntity[],
);

const featuredQuery = useListProducts(
  new PaginationQuery({
    paging: { page: 1, pageSize: 12 },
    orderBy: ['sort_order'],
  }).toRawParams(),
);
const featured = computed<ProductEntity[]>(
  () => ((featuredQuery.data?.value as any)?.items ?? []) as ProductEntity[],
);
const featuredLoading = computed(() => featuredQuery.isLoading.value);
const featuredError = computed(() => featuredQuery.isError.value);

// 商品实体里没有价格字段（价格在 SkuPrice，按 SKU 存「分」），
// 这里用两次批量查询拿到整屏商品的「起」价与活动价，而不是每商品发两次请求。
const productIds = computed(() =>
  featured.value.map((p) => p.id).filter((id): id is number => id !== undefined),
);
const { priceByProductId, skuIdsByProduct } = useProductPrices(productIds);
const { promoByProductId } = useProductPromos(productIds, priceByProductId, skuIdsByProduct);

const cards = computed(() =>
  featured.value
    .map((p) => {
      if (p.id === undefined) return undefined;
      const base = priceByProductId.value.get(p.id);
      const promo = promoByProductId.value.get(p.id);
      return {
        id: p.id,
        name: translationName(p.translations),
        image: p.imageUrl,
        // 有活动时展示活动价，现价退成划线价——与详情页同一套口径
        priceFen: promo?.fen ?? base?.fen ?? null,
        oldPriceFen: promo ? base?.fen ?? null : null,
        badgeLabel: promo ? t(promo.labelKey) : undefined,
        badgeTone: promo?.tone,
      };
    })
    .filter((c): c is NonNullable<typeof c> => c !== undefined),
);
</script>

<template>
  <LayoutSectionContainer no-padding class="pt-3 md:pt-6">
    <!-- 移动端搜索入口：假输入框样式，点击跳搜索页 -->
    <NuxtLink
      :to="localePath('/search')"
      class="mb-4 flex h-10 items-center gap-2 rounded-full border border-border bg-card px-4 text-sm text-muted-foreground transition-colors active:border-primary/60 md:hidden"
      :aria-label="t('navbar.top.search_placeholder')"
    >
      <XIcon icon="lucide:search" :size="16" />
      <span>{{ t('navbar.top.search_placeholder') }}</span>
    </NuxtLink>

    <div class="md:grid md:grid-cols-5 md:gap-5">
      <LayoutCategorySideMenu
        v-if="categories.length > 0"
        :categories="categories"
        class="hidden self-start md:col-span-1 md:block"
      />

      <LayoutHeroBanner class="md:col-span-4" />

      <LayoutCategoryQuickNav
        v-if="categories.length > 0"
        :categories="categories"
        class="mt-6 md:hidden"
      />
    </div>
  </LayoutSectionContainer>

  <!-- 精选商品 -->
  <LayoutSectionContainer>
    <UiSectionHeader :title="t('mall.home.featured')" />

    <UiListSkeleton v-if="featuredLoading" variant="card" :rows="8" :columns="4" />

    <UiAppEmpty v-else-if="featuredError" variant="error">
      <template #action>
        <UiButton variant="outline" size="sm" @click="featuredQuery.refetch()">
          {{ t('ui.button.retry') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <div
      v-else-if="cards.length > 0"
      class="grid grid-cols-2 gap-3 md:grid-cols-3 md:gap-4 lg:grid-cols-4"
    >
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

    <UiAppEmpty v-else :description="t('common.noData')" />
  </LayoutSectionContainer>
</template>
