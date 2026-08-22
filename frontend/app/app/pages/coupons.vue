<script setup lang="ts">
definePageMeta({
  layout: 'account',
  middleware: 'auth',
});
import { computed, ref } from 'vue';
import { XIcon } from '@/plugins/xicon';
import { useListCouponTemplates, useClaimCouponTemplate } from '@/api/composables';
import { useAccessStore } from '@/stores/modules/core/access.state';
import { queryClient } from '@/plugins/vue-query';

const { t } = useI18n();
const localePath = useLocalePath();

useHead({ title: t('mall.coupons.title') });

const accessStore = useAccessStore();

const isLogin = computed(() => {
  const token = accessStore.accessToken;
  return !!token?.value && !accessStore.loginExpired;
});

// 折扣类型展示文案：FIXED_AMOUNT → 固定金额，PERCENTAGE → 百分比，其他 → —。
function discountTypeLabel(t: any): string {
  switch (t) {
    case 'FIXED_AMOUNT':
      return 'fixed';
    case 'PERCENTAGE':
      return 'percentage';
    default:
      return '—';
  }
}

// 折扣数值展示：FIXED_AMOUNT → discountValue（固定金额）；PERCENTAGE → discountPercentage（百分比）。
// 其余类型不展示数值。
function discountAmount(t: any): string {
  switch (t) {
    case 'FIXED_AMOUNT':
      return '—';
    case 'PERCENTAGE':
      return '—';
    default:
      return '—';
  }
}

const PAGE_SIZE = 12;
const page = ref(1);

// 领券中心列表：匿名可读（rest_server 白名单）。core 侧 ListClaimable 在 repo 层用
// claimable=true AND status=ACTIVE 谓词过滤 + service 层后过滤有效窗口。
const couponsQuery = useListCouponTemplates(
  computed(() => ({
    page: page.value,
    pageSize: PAGE_SIZE,
    noPaging: false,
    sorting: [{ field: 'id', direction: 'DESC' }],
  })),
);

// 乐观已领记录：成功领取的 templateId → true，前端展示"已领取"态。
// 真实限领由 core 事务内 Count 原子兜底，此处仅为 UX。
const claimedSet = ref<Set<number>>(new Set());

const templateItems = computed<any[]>(() => {
  const items = (couponsQuery.data?.value as any)?.items ?? [];
  return (items as any[]) ?? [];
});
const couponsLoading = computed(() => couponsQuery.isPending.value);
const couponsError = computed(() => couponsQuery.isError.value);

const totalCount = computed(() => (couponsQuery.data?.value as any)?.total ?? 0);
const totalPages = computed(() => Math.max(1, Math.ceil(totalCount.value / PAGE_SIZE)));

const claimMutation = useClaimCouponTemplate({
  onSuccess: (_data, variables) => {
    // 成功领取：乐观标记 + 失效本人券列表（让结算页/钱包刷新）。
    claimedSet.value.add(variables.couponTemplateId as number);
    void queryClient.invalidateQueries({ queryKey: ['listUserCoupons'] });
  },
});

async function handleClaim(templateId: number | undefined) {
  if (templateId === undefined) return;
  try {
    await claimMutation.mutateAsync({ couponTemplateId: templateId });
  } catch {
    // 错误态由按钮 disabled + toast 体现
  }
}

function formatValidUntil(ts: string | undefined): string {
  if (!ts) return '—';
  try {
    return new Date(ts).toLocaleDateString(undefined, { dateStyle: 'medium' });
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
    :title="t('mall.coupons.title')"
    :description="t('mall.coupons.subtitle')"
    icon="lucide:ticket"
    size="sm"
  />

  <LayoutSectionContainer>
    <!-- 加载中 -->
    <div v-if="couponsLoading" class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <div
        v-for="i in 6"
        :key="i"
        class="rounded-2xl border border-border bg-card p-6"
      >
        <UiSkeleton class="mb-4 h-6 w-20" />
        <UiSkeleton class="mb-2 h-4 w-full" />
        <UiSkeleton class="h-4 w-24" />
      </div>
    </div>

    <!-- 错误态 -->
    <UiAppEmpty
      v-else-if="couponsError"
      variant="error"
    >
      <template #action>
        <UiButton variant="outline" size="sm" @click="couponsQuery.refetch()">
          {{ t('ui.button.retry') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <!-- 已加载 -->
    <div v-else class="flex flex-col gap-4">
      <!-- 空态 -->
      <UiAppEmpty
        v-if="templateItems.length === 0"
        variant="noData"
        :description="t('mall.coupons.empty')"
      />

      <!-- 可领券卡片网格 -->
      <div v-else class="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <div
          v-for="item in templateItems"
          :key="item.id"
          class="flex flex-col gap-4 rounded-2xl border border-border bg-card p-6"
        >
          <div class="flex items-center gap-3">
            <div class="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10 text-primary">
              <XIcon icon="lucide:ticket" :size="20" />
            </div>
            <div class="min-w-0">
              <p class="text-xs uppercase tracking-wide text-muted-foreground">
                {{ t('mall.coupons.discountType') }}
              </p>
              <p class="truncate text-sm font-medium text-foreground">
                {{ discountTypeLabel(item.discountType) }}
              </p>
            </div>
          </div>

          <dl class="grid grid-cols-2 gap-2 text-sm">
            <div>
              <dt class="text-xs text-muted-foreground">{{ t('mall.coupons.discount') }}</dt>
              <dd class="font-medium text-foreground">{{ discountAmount(item.discountType) }}</dd>
            </div>
            <div>
              <dt class="text-xs text-muted-foreground">{{ t('mall.coupons.validUntil') }}</dt>
              <dd class="font-medium text-foreground">{{ formatValidUntil(item.validUntil) }}</dd>
            </div>
          </dl>

          <UiButton
            class="mt-auto"
            :disabled="claimedSet.has(item.id) || claimMutation.isPending.value"
            @click="handleClaim(item.id)"
          >
            {{ claimedSet.has(item.id) ? t('mall.coupons.claimed') : t('mall.coupons.claim') }}
          </UiButton>
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
