<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import {
  useSearchProducts,
  fetchSearchSuggestions,
  rewriteSearchQuery,
  useProductPrices,
  useProductPromos,
} from '@/api/composables';
import { getCurrentLocale } from '@/utils/locale';
import { XIcon } from '@/plugins/xicon';

const route = useRoute();
const { t } = useI18n();
const localePath = useLocalePath();

// 搜索关键字：来自 URL ?q=（keyword 为等价的旧别名），同时维护本地输入框用于二次搜索。
const keyword = computed(() => {
  const raw = (route.query.q ?? route.query.keyword) as string | undefined;
  return raw?.trim() ?? '';
});
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

// 搜索结果不含价格：用两次批量查询补「起」价与活动价（与首页同一套口径）。
const resultIds = computed(() =>
  results.value.map((r) => r.productId).filter((id): id is number => id !== undefined),
);
const { priceByProductId, skuIdsByProduct } = useProductPrices(resultIds);
const { promoByProductId } = useProductPromos(resultIds, priceByProductId, skuIdsByProduct);

const cards = computed(() =>
  results.value
    .map((r) => {
      if (r.productId === undefined) return undefined;
      const base = priceByProductId.value.get(r.productId);
      const promo = promoByProductId.value.get(r.productId);
      return {
        id: r.productId,
        name: r.name,
        image: r.imageUrl,
        priceFen: promo?.fen ?? base?.fen ?? null,
        oldPriceFen: promo ? base?.fen ?? null : null,
        badgeLabel: promo ? t(promo.labelKey) : undefined,
        badgeTone: promo?.tone,
      };
    })
    .filter((c): c is NonNullable<typeof c> => c !== undefined),
);

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
    icon="carbon:search"
    size="sm"
  >
    <!-- 搜索框（带输入联想下拉） -->
    <div class="w-full max-w-xl">
      <div class="flex items-center gap-2">
        <div class="relative min-w-0 flex-1">
          <XIcon
            icon="lucide:search"
            :size="16"
            class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground"
          />
          <UiInput
            v-model="searchInput"
            class="h-11 rounded-full border-border bg-card pl-9 pr-3"
            :placeholder="t('search.placeholder')"
            @keyup.enter="submitSearch"
            @blur="onSuggestInputBlur"
          />
          <div
            v-if="showSuggest"
            class="absolute inset-x-0 top-[calc(100%+0.25rem)] z-30 overflow-hidden rounded-lg border border-border bg-popover text-popover-foreground shadow-pop"
          >
            <button
              v-for="s in suggestions"
              :key="s.productId"
              type="button"
              class="flex w-full items-center gap-2 px-3 py-2 text-left text-sm transition-colors hover:bg-secondary"
              @mousedown.prevent="pickSuggestion(s)"
            >
              <UiImage
                v-if="s.imageUrl"
                :src="s.imageUrl"
                :alt="s.name"
                class="size-8 shrink-0 rounded object-cover"
              />
              <span class="line-clamp-1">{{ s.name }}</span>
            </button>
          </div>
        </div>
        <UiButton class="h-11 shrink-0 rounded-full" @click="submitSearch">
          <XIcon icon="carbon:search" :size="16" />
          {{ t('search.submitSearch') }}
        </UiButton>
      </div>
    </div>
  </LayoutPageHero>

  <LayoutSectionContainer>
    <!-- AI 搜索改写提示条 -->
    <div
      v-if="aiKeywords?.length"
      class="mb-4 flex flex-wrap items-center gap-2 rounded-lg bg-primary-soft px-3 py-2 text-xs text-primary-soft-foreground"
    >
      <span class="flex items-center gap-1 font-semibold">
        <XIcon icon="lucide:sparkles" :size="12" />
        {{ t('search.aiOptimized') }}
      </span>
      <span
        v-for="kw in aiKeywords"
        :key="kw"
        class="rounded-full bg-card px-2 py-0.5 font-medium"
      >{{ kw }}</span>
      <UiButton
        variant="link"
        size="sm"
        class="ml-auto px-0 text-xs text-primary-soft-foreground underline-offset-2 hover:underline"
        @click="restoreOriginalQuery"
      >
        {{ t('search.aiRestore') }}
      </UiButton>
    </div>

    <!-- 无关键字 -->
    <UiAppEmpty v-if="!keyword" :description="t('search.placeholder')">
      <template #action>
        <UiButton variant="outline" @click="navigateTo(localePath('/'))">
          {{ t('ui.button.view_homepage') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <UiListSkeleton v-else-if="loading" variant="card" :rows="8" :columns="4" />

    <!-- 错误态：搜索请求失败（与"无结果"区分并提供重试） -->
    <UiAppEmpty v-else-if="loadError" variant="error">
      <template #action>
        <UiButton variant="outline" size="sm" @click="searchQuery.refetch()">
          {{ t('ui.button.retry') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <template v-else-if="cards.length > 0">
      <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
        <p class="min-w-0 text-sm font-medium text-foreground">
          {{ t('search.resultFor', { q: keyword }) }}
        </p>
        <UiStatusBadge
          tone="neutral"
          :label="t('search.resultCount', { count: totalCount })"
        />
      </div>

      <div class="grid grid-cols-2 gap-3 md:grid-cols-3 md:gap-4 lg:grid-cols-4">
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

      <UiPager v-model:page="page" :total-pages="totalPages" class="mt-6" />
    </template>

    <!-- 无结果 -->
    <UiAppEmpty v-else :description="t('search.empty', { q: keyword })">
      <template #action>
        <UiButton variant="outline" @click="navigateTo(localePath('/'))">
          {{ t('cart.continueShopping') }}
        </UiButton>
      </template>
    </UiAppEmpty>
  </LayoutSectionContainer>
</template>
