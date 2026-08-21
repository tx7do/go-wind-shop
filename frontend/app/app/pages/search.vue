<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useSearchProducts } from '@/api/composables';
import { getCurrentLocale } from '@/utils/locale';
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

// 商品搜索走 Elasticsearch 全文检索。
// language 由 getCurrentLocale() 强制注入，服务端硬编码 status=ACTIVE。
// ES 结果只回传 productId/language/name/imageUrl（最小字段集）。
const PAGE_SIZE = 48;
const page = ref(1);

const searchQuery = useSearchProducts(
  computed(() => ({
    query: keyword.value,
    language: getCurrentLocale(),
    page: page.value - 1, // ES 分页 0-based
    pageSize: PAGE_SIZE,
  })),
  { enabled: computed(() => !!keyword.value) },
);

const loading = computed(() => searchQuery.isPending.value);
const loadError = computed(() => searchQuery.isError.value);

const totalCount = computed(() => (searchQuery.data?.value as any)?.total ?? 0);
const totalPages = computed(() => Math.max(1, Math.ceil(totalCount.value / PAGE_SIZE)));

// 关键字变化时回到第 1 页，避免停留在越界页码。
watch(keyword, () => {
  page.value = 1;
});

function goToPage(p: number) {
  const clamped = Math.min(Math.max(1, p), totalPages.value);
  if (clamped !== page.value) {
    page.value = clamped;
  }
}

type SearchResultItem = {
  productId?: number;
  name?: string;
  imageUrl?: string;
};

// ES 搜索结果直接提供 productId/name/imageUrl，无需翻译提取。
const results = computed<SearchResultItem[]>(() => {
  const items = (searchQuery.data?.value as any)?.items ?? [];
  return (items as SearchResultItem[]) ?? [];
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

    <!-- 错误态：搜索请求失败（与"无结果"区分并提供重试） -->
    <UiAppEmpty
      v-else-if="loadError"
      variant="error"
    >
      <template #action>
        <UiButton variant="outline" size="sm" @click="searchQuery.refetch()">
          {{ t('ui.button.retry') }}
        </UiButton>
      </template>
    </UiAppEmpty>

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
        {{ t('search.resultCount', { count: totalCount }) }}
      </p>
      <div class="grid grid-cols-2 gap-4 md:grid-cols-3 lg:grid-cols-4">
        <NuxtLink
          v-for="product in results"
          :key="product.productId"
          :to="localePath('/product/' + product.productId)"
          class="group block overflow-hidden rounded-xl border border-border bg-card shadow-sm transition-colors hover:border-primary/60 dark:shadow-none"
        >
          <UiImage
            :src="product.imageUrl"
            :alt="product.name"
            class="aspect-square w-full rounded-none object-cover"
          />
          <div class="p-3">
            <h3 class="line-clamp-2 text-sm font-semibold text-foreground dark:text-slate-200">
              {{ product.name || '—' }}
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
  </LayoutSectionContainer>
</template>
