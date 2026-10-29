<script setup lang="ts">
import { computed, ref } from 'vue'
import { toast } from 'vue-sonner'
import { cn } from '@/lib/utils'
import { XIcon } from '@/plugins/xicon'
import {
  useListCouponTemplates,
  useExchangePointsForCoupon,
  useGetMyPointsBalance,
} from '@/api/composables'
import { useAccessStore } from '@/stores/modules/core/access.state'
import { queryClient } from '@/plugins/vue-query'

const { t } = useI18n()
const localePath = useLocalePath()
const accessStore = useAccessStore()

const isLogin = computed(() => {
  const token = accessStore.accessToken
  return !!token?.value && !accessStore.loginExpired
})

const couponTemplatesQuery = useListCouponTemplates({ page: 1, pageSize: 50, noPaging: false } as any)
const coupons = computed(() => {
  const items = ((couponTemplatesQuery.data?.value as any)?.items ?? []) as Array<{
    id?: number; name?: string; discountType?: string; discountValue?: number;
    discountPercentage?: number; validFrom?: string; validUntil?: string;
    pointsPrice?: number; maxRedemptions?: number; redeemedCount?: number;
    currency?: string; status?: string;
  }>
  return items.filter((c) => c.status === 'ACTIVE')
})

const pointsBalanceQuery = useGetMyPointsBalance({}, { enabled: isLogin } as any)
const pointsBalance = computed(() => pointsBalanceQuery.data.value?.balance ?? 0)

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
    await exchangeMutation.mutateAsync({ couponTemplateId: id })
  } finally {
    exchangingId.value = undefined
  }
}

const STATUS_KEY: Record<string, string> = { ACTIVE: 'active', INACTIVE: 'inactive' }
</script>

<template>
  <LayoutSectionContainer>
    <div class="mb-6 flex items-center justify-between">
      <h1 class="text-2xl font-bold text-foreground">{{ t('coupons.title') }}</h1>
      <div v-if="isLogin" class="rounded-full bg-primary/15 px-4 py-1 text-sm font-bold text-primary tabular-nums">
        {{ t('coupons.myPoints') }}: {{ pointsBalance }}
      </div>
    </div>

    <div v-if="couponTemplatesQuery.isPending" class="flex flex-col gap-4">
      <UiSkeleton v-for="i in 3" :key="i" class="h-32 rounded-xl" />
    </div>

    <UiAppEmpty
      v-else-if="coupons.length === 0"
      variant="noData"
      :description="t('coupons.empty')"
    />

    <div v-else class="grid gap-4 md:grid-cols-2">
      <div
        v-for="c in coupons"
        :key="c.id"
        class="flex overflow-hidden rounded-xl border border-border bg-card"
      >
        <div class="flex w-28 shrink-0 flex-col items-center justify-center bg-primary/10 p-4">
          <template v-if="c.discountType === 'FIXED_AMOUNT'">
            <span class="text-2xl font-extrabold text-primary">{{ t('mall.product.currencyCny') }}{{ ((c.discountValue ?? 0) / 100).toFixed(0) }}</span>
            <span class="text-[10px] text-muted-foreground">{{ t('coupons.off') }}</span>
          </template>
          <template v-else>
            <span class="text-2xl font-extrabold text-primary">{{ c.discountPercentage ?? 0 }}%</span>
            <span class="text-[10px] text-muted-foreground">{{ t('coupons.off') }}</span>
          </template>
        </div>
        <div class="flex flex-1 flex-col justify-between p-4">
          <div>
            <p class="text-sm font-bold text-foreground">{{ c.name || t('coupons.title') }}</p>
            <p class="mt-0.5 text-[10px] text-muted-foreground">
              {{ t('coupons.validUntil') }}: {{ c.validUntil?.slice(0, 10) || '—' }}
            </p>
            <p v-if="(c.redeemedCount ?? 0) > 0 && c.maxRedemptions" class="text-[10px] text-muted-foreground">
              {{ t('coupons.redeemed') }}: {{ c.redeemedCount }}/{{ c.maxRedemptions }}
            </p>
          </div>
          <div class="mt-2 flex items-center justify-end">
            <UiButton
              v-if="(c.pointsPrice ?? 0) > 0"
              size="sm"
              :disabled="!isLogin || exchangingId === c.id || pointsBalance < (c.pointsPrice ?? 0)"
              @click="handleExchange(c.id!)"
            >
              {{ exchangingId === c.id ? t('coupons.exchanging') : t('coupons.exchange', { points: c.pointsPrice }) }}
            </UiButton>
          </div>
        </div>
      </div>
    </div>
  </LayoutSectionContainer>
</template>
