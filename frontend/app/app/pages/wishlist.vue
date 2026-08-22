<script setup lang="ts">
definePageMeta({
  layout: 'account',
  middleware: 'auth',
});
import { computed, ref, watch, reactive } from 'vue';
import { XIcon } from '@/plugins/xicon';
import { useListWishlist, useDeleteWishlist } from '@/api/composables';
import { fetchGetProductStore } from '@/api/composables';
import { useAccessStore } from '@/stores/modules/core/access.state';
import { getCurrentLocale } from '@/utils/locale';

const { t } = useI18n();
const localePath = useLocalePath();

useHead({ title: t('mall.wishlist.title') });

const accessStore = useAccessStore();

const isLogin = computed(() => {
  const token = accessStore.accessToken;
  return !!token?.value && !accessStore.loginExpired;
});

// 当前 locale（用于商品翻译选取）。
const currentLocale = computed(() => getCurrentLocale());

// pickTranslation：从商品 translations 中选取当前 locale 的翻译，无则回退首条。
// 仿 cart.vue 同名函数。
function pickTranslation<T extends { languageCode?: string }>(
  translations: T[] | undefined,
): T | undefined {
  if (!translations || translations.length === 0) return undefined;
  const match = translations.find((tr) => tr.languageCode === currentLocale.value);
  return match ?? translations[0];
}

type WishlistEntity = {
  id?: number;
  productId?: number;
  createdAt?: string;
};

const PAGE_SIZE = 10;
const page = ref(1);

// 收藏列表：按 viewer user_id 行级隔离由后端 UserPrivacy 强制，前端无需带 userId。
const wishlistQuery = useListWishlist(
  computed(() => ({
    page: page.value,
    pageSize: PAGE_SIZE,
    noPaging: false,
    sorting: [{ field: 'id', direction: 'DESC' }],
  })),
  { enabled: isLogin },
);

const wishlistItems = computed<WishlistEntity[]>(() => {
  const items = (wishlistQuery.data?.value as any)?.items ?? [];
  return (items as WishlistEntity[]) ?? [];
});
const wishlistLoading = computed(() => wishlistQuery.isPending.value);
const wishlistError = computed(() => wishlistQuery.isError.value);

const totalCount = computed(() => (wishlistQuery.data?.value as any)?.total ?? 0);
const totalPages = computed(() => Math.max(1, Math.ceil(totalCount.value / PAGE_SIZE)));

// ---------- 商品名 / 商品图片：productId → Product ----------
// productInfoMap: productId → { name, imageUrl }
const productInfoMap = reactive<Record<number, { name: string; imageUrl: string }>>({});

watch(
  wishlistItems,
  async (items) => {
    for (const k of Object.keys(productInfoMap)) delete productInfoMap[Number(k)];
    if (!items || items.length === 0) return;
    await Promise.all(
      items.map(async (item) => {
        const productId = item.productId;
        if (productId === undefined) return;
        try {
          const product: any = await fetchGetProductStore(productId);
          const imageUrl = product?.imageUrl ?? '';
          const tr = pickTranslation(
            product?.translations as Array<{ name?: string; languageCode?: string }> | undefined,
          );
          productInfoMap[productId] = { name: tr?.name ?? '', imageUrl };
        } catch {
          // ignore
        }
      }),
    );
  },
  { immediate: true },
);

const deleteMutation = useDeleteWishlist({
  onSuccess: () => {
    wishlistQuery.refetch();
  },
});

async function handleRemove(id: number | undefined) {
  if (id === undefined) return;
  try {
    await deleteMutation.mutateAsync(id);
  } catch {
    // 忽略——错误态由 query 的 error 体现
  }
}

function formatCreatedAt(ts: string | undefined): string {
  if (!ts) return '—';
  try {
    return new Date(ts).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' });
  } catch {
    return '—';
  }
}

function goToPage(p: number) {
  const clamped = Math.min(Math.max(1, p), totalPages.value);
  if (clamped !== page.value) {
    page.value = clamped;
  }
}
</script>

<template>
  <LayoutPageHero
    :title="t('mall.wishlist.title')"
    :description="t('mall.wishlist.subtitle')"
    icon="lucide:heart"
    size="sm"
  />

  <LayoutSectionContainer>
    <!-- 未登录 -->
    <div
      v-if="!isLogin"
      class="flex flex-col items-center gap-6 rounded-2xl border border-border bg-card p-12 text-center"
    >
      <XIcon icon="carbon:locked" :size="48" class="text-muted-foreground" />
      <p class="text-lg text-muted-foreground">{{ t('authentication.login.please_login') }}</p>
      <UiButton @click="navigateTo(localePath('/login'))">
        {{ t('navbar.user.login') }}
      </UiButton>
    </div>

    <!-- 加载中 -->
    <div v-else-if="wishlistLoading" class="overflow-hidden rounded-2xl border border-border bg-card">
      <div class="border-b border-border bg-muted/40 px-6 py-3">
        <UiSkeleton class="h-4 w-48" />
      </div>
      <div v-for="i in 4" :key="i" class="border-b border-border px-6 py-4 last:border-b-0">
        <div class="flex items-center gap-4">
          <UiSkeleton class="h-4 w-10" />
          <UiSkeleton class="h-5 w-16" />
          <UiSkeleton class="h-4 w-14 ml-auto" />
          <UiSkeleton class="h-4 w-24" />
          <UiSkeleton class="h-4 w-8" />
        </div>
      </div>
    </div>

    <!-- 错误态 -->
    <UiAppEmpty
      v-else-if="wishlistError"
      variant="error"
    >
      <template #action>
        <UiButton variant="outline" size="sm" @click="wishlistQuery.refetch()">
          {{ t('ui.button.retry') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <!-- 已加载 -->
    <div v-else class="flex flex-col gap-4">
      <!-- 空态 -->
      <UiAppEmpty
        v-if="wishlistItems.length === 0"
        variant="noData"
        :description="t('mall.wishlist.empty')"
      />

      <!-- 收藏列表 -->
      <div v-else class="overflow-x-auto rounded-2xl border border-border bg-card">
        <div class="min-w-[700px]">
        <div class="border-b border-border bg-muted/40 px-6 py-3 text-xs font-medium uppercase tracking-wide text-muted-foreground">
          <div class="grid grid-cols-[1fr_140px_80px] items-center gap-4">
            <span>{{ t('mall.wishlist.table.product') }}</span>
            <span class="text-right">{{ t('mall.wishlist.table.createdAt') }}</span>
            <span class="text-right">{{ t('mall.wishlist.table.action') }}</span>
          </div>
        </div>

      <div
        v-for="item in wishlistItems"
        :key="item.id"
        class="block border-b border-border px-6 py-4 transition-colors last:border-b-0 hover:bg-muted/30"
      >
        <div class="grid grid-cols-[1fr_140px_80px] items-center gap-4">
          <span class="flex items-center gap-3 min-w-0">
            <img
              :src="productInfoMap[item.productId ?? -1]?.imageUrl || ''"
              class="h-10 w-10 shrink-0 rounded-md border border-border object-cover"
              alt=""
            />
            <span class="truncate text-sm text-foreground">
              {{ productInfoMap[item.productId ?? -1]?.name ?? '—' }}
            </span>
          </span>
          <span class="text-right text-xs tabular-nums text-muted-foreground">
            {{ formatCreatedAt(item.createdAt) }}
          </span>
          <span class="flex justify-end">
            <UiButton
              variant="outline"
              size="sm"
              @click="handleRemove(item.id)"
            >
              {{ t('mall.wishlist.table.action') }}
            </UiButton>
          </span>
        </div>
      </div>
      </div>
    </div>

      <!-- 分页 -->
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
    </div>
  </LayoutSectionContainer>
</template>
