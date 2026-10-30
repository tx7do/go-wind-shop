<script setup lang="ts">
import { computed, ref } from 'vue'
import { toast } from 'vue-sonner'
import {
  useListCouponTemplates,
  useExchangePointsForCoupon,
  useGetMyPointsBalance,
} from '@/api/composables'
import { useAccessStore } from '@/stores/modules/core/access.state'
import { queryClient } from '@/plugins/vue-query'
import { formatFen } from '@/utils/money'

definePageMeta({
  layout: 'account',
  middleware: 'auth',
})

const { t } = useI18n()
const localePath = useLocalePath()
const accessStore = useAccessStore()

const isLogin = computed(() => {
  const token = accessStore.accessToken
  return !!token?.value && !accessStore.loginExpired
})

type CouponTemplate = {
  id?: number; name?: string; discountType?: string; discountValue?: number | string;
  discountPercentage?: number; validFrom?: string; validUntil?: string;
  pointsPrice?: number | string; maxRedemptions?: number | string; redeemedCount?: number | string;
  currency?: string; status?: string;
}

const couponTemplatesQuery = useListCouponTemplates({ page: 1, pageSize: 50, noPaging: false } as any)
const couponsLoading = computed(() => couponTemplatesQuery.isPending.value)
const couponsError = computed(() => couponTemplatesQuery.isError.value)
const coupons = computed(() => {
  const items = ((couponTemplatesQuery.data?.value as any)?.items ?? []) as CouponTemplate[]
  return items.filter((c) => c.status === 'ACTIVE')
})

// 面额：固定金额券按「分」转元（最多两位），折扣券按百分比
function faceValue(c: CouponTemplate): string {
  return c.discountType === 'FIXED_AMOUNT'
    ? `${t('mall.product.currencyCny')}${formatFen(c.discountValue, 0)}`
    : `${c.discountPercentage ?? 0}%`
}

// uint64 字段在 JSON 里是字符串，比较前一律 Number()
const maxOf = (c: CouponTemplate) => Number(c.maxRedemptions ?? 0)
const redeemedOf = (c: CouponTemplate) => Number(c.redeemedCount ?? 0)
const pointsPriceOf = (c: CouponTemplate) => Number(c.pointsPrice ?? 0)

const isSoldOut = (c: CouponTemplate) => maxOf(c) > 0 && redeemedOf(c) >= maxOf(c)

const pointsBalanceQuery = useGetMyPointsBalance({}, { enabled: isLogin } as any)
const pointsBalance = computed(() => Number(pointsBalanceQuery.data.value?.balance ?? 0))

const exchangeMutation = useExchangePointsForCoupon({
  onSuccess: () => {
    toast.success(t('coupons.exchangeSuccess'))
    queryClient.invalidateQueries({ queryKey: ['myPointsBalance'] })
    queryClient.invalidateQueries({ queryKey: ['listUserCoupons'] })
    queryClient.invalidateQueries({ queryKey: ['listCouponTemplates'] })
  },
  onError: (err: any) => {
    toast.error(err?.message || t('coupons.exchangeFailed'))
  },
})
const exchangingId = ref<number | undefined>(undefined)

async function handleExchange(id: number) {
  if (!isLogin.value) {
    navigateTo(localePath('/login'))
    return
  }
  exchangingId.value = id
  try {
    await exchangeMutation.mutateAsync(Number(id))
  } finally {
    exchangingId.value = undefined
  }
}
</script>

<template>
  <LayoutPageHero
    :title="t('mall.coupons.title')"
    icon="lucide:ticket"
    size="sm"
  >
    <UiStatusBadge
      v-if="isLogin"
      tone="primary"
      :label="`${t('mall.coupons.myPoints')} ${pointsBalance}`"
    />
  </LayoutPageHero>

  <LayoutSectionContainer>
    <UiListSkeleton v-if="couponsLoading" :rows="3" :header="false" />

    <UiAppEmpty
      v-else-if="couponsError"
      variant="error"
    >
      <template #action>
        <UiButton variant="outline" size="sm" @click="couponTemplatesQuery.refetch()">
          {{ t('ui.button.retry') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <UiAppEmpty
      v-else-if="coupons.length === 0"
      variant="noData"
      :description="t('mall.coupons.empty')"
    />

    <div v-else class="grid gap-3 md:grid-cols-2 md:gap-4">
      <article
        v-for="c in coupons"
        :key="c.id"
        class="flex overflow-hidden rounded-lg border border-border bg-card shadow-card"
      >
        <div class="flex w-24 shrink-0 flex-col items-center justify-center gap-1 border-r border-dashed border-border bg-primary-soft p-4 text-primary-soft-foreground md:w-28">
          <span class="text-xl font-extrabold tabular-nums md:text-2xl">{{ faceValue(c) }}</span>
        </div>

        <div class="flex min-w-0 flex-1 flex-col gap-1.5 p-4">
          <div class="flex items-start justify-between gap-2">
            <p class="min-w-0 flex-1 text-sm font-semibold leading-snug text-foreground">
              {{ c.name || t('checkout.couponLabel') }}
            </p>
            <UiStatusBadge
              :tone="isSoldOut(c) ? 'warning' : 'primary'"
              size="sm"
              :label="isSoldOut(c) ? t('mall.coupons.redeemed') : t('mall.coupons.off')"
              class="shrink-0"
            />
          </div>

          <p class="text-[11px] text-muted-foreground">
            {{ t('mall.coupons.validUntil') }} {{ c.validUntil?.slice(0, 10) || '—' }}
          </p>
          <p v-if="redeemedOf(c) > 0 && maxOf(c) > 0" class="text-[11px] text-muted-foreground">
            {{ t('mall.coupons.redeemed') }} {{ redeemedOf(c) }}/{{ maxOf(c) }}
          </p>

          <div v-if="pointsPriceOf(c) > 0" class="mt-auto flex justify-end pt-1">
            <UiButton
              size="sm"
              :disabled="!isLogin || exchangingId === c.id || isSoldOut(c) || pointsBalance < pointsPriceOf(c)"
              @click="handleExchange(c.id!)"
            >
              {{ exchangingId === c.id ? t('mall.coupons.exchanging') : t('mall.coupons.exchange', { points: c.pointsPrice }) }}
            </UiButton>
          </div>
        </div>
      </article>
    </div>
  </LayoutSectionContainer>
</template>
