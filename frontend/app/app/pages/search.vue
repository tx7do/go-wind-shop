<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useListProducts } from '@/api/composables';
import { getCurrentLocale } from '@/utils/locale';
import { PaginationQuery } from '@/core/transport/rest';
import { XIcon } from '@/plugins/xicon';

const route = useRoute();
const { t } = useI18n();
const localePath = useLocalePath();

// 搜索关键字：来自 URL ?q=，同时维护本地输入框用于二次搜索。
const keyword = computed(() => (route.query.q as string)?.trim() ?? '');
const searchInput = ref(keyword.value);

// URL 上 q 变化时同步到输入框（如浏览器前进/后退）。
watch(keyword, (v) => {
  searchInput.value = v;
});

useHead({ title: () => `${t('search.title')}: ${keyword.value || '—'}` });

const currentLocale = computed(() => getCurrentLocale());

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

// 拉取在售商品（按 sort_order 排序）。
// 注：后端商品 List 目前不支持按名称模糊搜索（product_repo 未接入 translation
// 的 ContainsFold 查询），故搜索在此 MVP 阶段为前端过滤。pageSize 取较大值以
// 覆盖常见商品规模；后续若后端补齐 name 搜索能力，可将过滤下推至 query。
const productsQuery = useListProducts(
  new PaginationQuery({
    paging: { page: 1, pageSize: 200 },
    orderBy: ['sort_order'],
  }),
);
const allProducts = computed<ProductEntity[]>(() => {
  const items = (productsQuery.data?.value as any)?.items ?? [];
  return (items as ProductEntity[]) ?? [];
});
const loading = computed(() => productsQuery.isLoading.value);

// 按当前语言的商品名做大小写不敏感包含匹配。
function productName(p: ProductEntity): string {
  return pickTranslation(p.translations)?.name ?? '';
}

const results = computed<ProductEntity[]>(() => {
  const q = keyword.value.toLowerCase();
  if (!q) return [];
  return allProducts.value.filter((p) => productName(p).toLowerCase().includes(q));
});

// 顶部搜索框提交：跳转到 /search?q=xxx（交由路由驱动结果）
function submitSearch() {
  const q = searchInput.value.trim();
  navigateTo(localePath('/search') + (q ? `?q=${encodeURIComponent(q)}` : ''));
}
</script>

<template>
  <LayoutPageHero
    :title="t('search.title')"
    :description="keyword ? t('search.resultFor', { q: keyword }) : t('search.placeholder')"
    icon="carbon:search"
    size="sm"
  />

  <LayoutSectionContainer>
    <!-- 搜索框 -->
    <div class="mx-auto mb-6 flex w-full max-w-xl items-center gap-2">
      <UiInput
        v-model="searchInput"
        class="h-11"
        :placeholder="t('search.placeholder')"
        @keyup.enter="submitSearch"
      />
      <UiButton class="h-11 shrink-0" @click="submitSearch">
        <XIcon icon="carbon:search" :size="16" />
        {{ t('search.submitSearch') }}
      </UiButton>
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-4">
      <UiPostCardSkeleton v-for="_, i in 8" :key="i" />
    </div>

    <!-- 无关键字 -->
    <div
      v-else-if="!keyword"
      class="flex flex-col items-center gap-4 rounded-2xl border border-border bg-card p-12 text-center"
    >
      <XIcon icon="carbon:search" :size="48" class="text-muted-foreground" />
      <p class="text-sm text-muted-foreground">{{ t('search.placeholder') }}</p>
    </div>

    <!-- 有结果 -->
    <div v-else-if="results.length > 0">
      <p class="mb-4 text-sm text-muted-foreground">
        {{ t('search.resultCount', { count: results.length }) }}
      </p>
      <div class="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-4">
        <NuxtLink
          v-for="product in results"
          :key="product.id"
          :to="localePath('/product/' + product.id)"
          class="group block overflow-hidden rounded-xl border border-border bg-card shadow-sm transition-colors hover:border-primary/60 dark:shadow-none"
        >
          <UiImage
            :src="product.imageUrl"
            :alt="productName(product)"
            class="aspect-square w-full rounded-none object-cover"
          />
          <div class="p-3">
            <h3 class="line-clamp-2 text-sm font-semibold text-foreground dark:text-slate-200">
              {{ productName(product) || '—' }}
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
    </div>

    <!-- 无结果 -->
    <div
      v-else
      class="flex flex-col items-center gap-4 rounded-2xl border border-border bg-card p-12 text-center"
    >
      <XIcon icon="carbon:document" :size="48" class="text-muted-foreground" />
      <p class="text-sm text-muted-foreground">{{ t('search.empty', { q: keyword }) }}</p>
      <UiButton variant="outline" @click="navigateTo(localePath('/'))">
        {{ t('cart.continueShopping') }}
      </UiButton>
    </div>
  </LayoutSectionContainer>
</template>
