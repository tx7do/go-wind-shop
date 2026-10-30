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
import { translationName } from '@/utils/translation';

const { t } = useI18n();
const localePath = useLocalePath();

useHead({ title: t('mall.wishlist.title') });

const accessStore = useAccessStore();

const isLogin = computed(() => {
  const token = accessStore.accessToken;
  return !!token?.value && !accessStore.loginExpired;
});

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
          productInfoMap[productId] = {
            name: translationName(product?.translations),
            imageUrl: product?.imageUrl ?? '',
          };
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
    <UiAuthGate v-if="!isLogin" />

    <!-- 加载中 -->
    <UiListSkeleton
      v-else-if="wishlistLoading"
      variant="card"
      :rows="8"
      :columns="4"
    />

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
    <template v-else>
      <!-- 空态 -->
      <UiAppEmpty
        v-if="wishlistItems.length === 0"
        variant="noData"
        :description="t('mall.wishlist.empty')"
      >
        <template #action>
          <UiButton variant="outline" @click="navigateTo(localePath('/'))">
            {{ t('navbar.top.home') }}
          </UiButton>
        </template>
      </UiAppEmpty>

      <!-- 收藏列表：移动 2 列 / 桌面 4 列 -->
      <div v-else class="flex flex-col gap-4">
        <div class="grid grid-cols-2 gap-3 md:grid-cols-3 md:gap-4 lg:grid-cols-4">
          <div
            v-for="item in wishlistItems"
            :key="item.id"
            class="relative"
          >
            <UiProductCard
              :product-id="item.productId"
              :name="productInfoMap[item.productId ?? -1]?.name"
              :image="productInfoMap[item.productId ?? -1]?.imageUrl"
            />
            <button
              type="button"
              class="absolute right-2 top-2 flex h-8 w-8 items-center justify-center rounded-full border border-border bg-card text-muted-foreground shadow-card transition-colors hover:text-destructive"
              :aria-label="t('common.delete')"
              :title="t('common.delete')"
              @click.stop="handleRemove(item.id)"
            >
              <XIcon icon="lucide:trash-2" :size="15" />
            </button>
          </div>
        </div>

        <!-- 分页 -->
        <UiPager
          :page="page"
          :total-pages="totalPages"
          @update:page="goToPage"
        />
      </div>
    </template>
  </LayoutSectionContainer>
</template>
