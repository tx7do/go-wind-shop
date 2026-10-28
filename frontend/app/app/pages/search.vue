<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useSearchProducts, fetchSearchSuggestions, rewriteSearchQuery } from '@/api/composables';
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

// ---------- AI 搜索改写 ----------
// 关键词看起来是自然语言（≥6 字符）时异步请求 LLM 改写为 2-4 个检索关键词；
// 改写成功则以关键词组重新检索，展示提示条（可还原原词）；失败静默沿用原词。
const aiKeywords = ref<string[] | null>(null);
const effectiveQuery = computed(() =>
  aiKeywords.value?.length ? aiKeywords.value.join(' ') : keyword.value,
);

watch(keyword, (v) => {
  aiKeywords.value = null;
  if (!v || v.length < 6) return;
  const asked = v;
  rewriteSearchQuery(v)
    .then((resp) => {
      // 只应用对当前关键字的改写（防竞态：慢响应覆盖新一次搜索）
      if (keyword.value !== asked) return;
      if (resp.rewritten && resp.keywords?.length && resp.keywords.join(' ') !== asked) {
        aiKeywords.value = resp.keywords as string[];
      }
    })
    .catch(() => {});
});

function restoreOriginalQuery() {
  aiKeywords.value = null;
}

const searchQuery = useSearchProducts(
  computed(() => ({
    query: effectiveQuery.value,
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
  showSuggest.value = false;
  navigateTo(localePath('/search') + (q ? `?q=${encodeURIComponent(q)}` : ''));
}

// ---------- 搜索联想（输入联想） ----------
// 输入 ≥ 2 字符时防抖 300ms 拉取建议；点击建议直达商品详情。
type Suggestion = { productId?: number; name?: string; imageUrl?: string };

const suggestions = ref<Suggestion[]>([]);
const showSuggest = ref(false);
const suggestLoading = ref(false);
let suggestTimer: ReturnType<typeof setTimeout> | undefined;

watch(searchInput, (v) => {
  if (suggestTimer) clearTimeout(suggestTimer);
  const q = v.trim();
  if (q.length < 2) {
    suggestions.value = [];
    showSuggest.value = false;
    return;
  }
  suggestTimer = setTimeout(async () => {
    suggestLoading.value = true;
    try {
      const resp = await fetchSearchSuggestions(q);
      suggestions.value = ((resp?.items ?? []) as Suggestion[]);
      showSuggest.value = suggestions.value.length > 0;
    } catch {
      suggestions.value = [];
      showSuggest.value = false;
    } finally {
      suggestLoading.value = false;
    }
  }, 300);
});

function pickSuggestion(s: Suggestion) {
  showSuggest.value = false;
  if (s.productId !== undefined) {
    navigateTo(localePath('/product/' + s.productId));
  }
}

function onSuggestInputBlur() {
  // 延迟关闭以允许点击建议项（mousedown 先于 blur）
  setTimeout(() => {
    showSuggest.value = false;
  }, 150);
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
    <!-- AI 搜索改写提示条 -->
    <div
      v-if="aiKeywords?.length"
      class="mx-auto mb-4 flex w-full max-w-xl flex-wrap items-center gap-2 rounded-xl border border-primary/30 bg-primary/5 px-3 py-2 text-xs"
    >
      <span class="flex items-center gap-1 font-medium text-primary">
        <XIcon icon="lucide:sparkles" width="12" height="12" />
        {{ t('search.aiOptimized') }}:
      </span>
      <span
        v-for="kw in aiKeywords"
        :key="kw"
        class="rounded-full bg-primary/10 px-2 py-0.5 text-primary"
      >{{ kw }}</span>
      <button
        type="button"
        class="ml-auto text-muted-foreground underline-offset-2 hover:text-primary hover:underline"
        @click="restoreOriginalQuery"
      >
        {{ t('search.aiRestore') }}
      </button>
    </div>

    <!-- 搜索框（带输入联想下拉） -->
    <div class="relative mx-auto mb-6 w-full max-w-xl">
      <div class="flex items-center gap-2">
        <UiInput
          v-model="searchInput"
          class="h-11"
          :placeholder="t('search.placeholder')"
          @keyup.enter="submitSearch"
          @blur="onSuggestInputBlur"
        />
        <UiButton class="h-11 shrink-0" @click="submitSearch">
          <XIcon icon="carbon:search" :size="16" />
          {{ t('search.submitSearch') }}
        </UiButton>
      </div>
      <div
        v-if="showSuggest"
        class="absolute left-0 right-14 top-12 z-20 overflow-hidden rounded-xl border border-border bg-card shadow-lg"
      >
        <button
          v-for="s in suggestions"
          :key="s.productId"
          type="button"
          class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm transition-colors hover:bg-primary/5"
          @mousedown.prevent="pickSuggestion(s)"
        >
          <UiImage v-if="s.imageUrl" :src="s.imageUrl" :alt="s.name" class="h-8 w-8 rounded object-cover" />
          <span class="line-clamp-1">{{ s.name }}</span>
        </button>
      </div>
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
