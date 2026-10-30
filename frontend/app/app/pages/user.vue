<script setup lang="ts">
definePageMeta({
  layout: 'account',
  middleware: 'auth',
})
import { XIcon } from '@/plugins/xicon'
import { useUserStore } from '@/stores/modules/core/user.state'
import { useAccessStore } from '@/stores/modules/core/access.state'
import { useGetMe } from '@/api/composables/user-profile'
import { useGetMyPointsBalance, useListMyPointsLedgers, useCheckIn, useGetCheckInStatus } from '@/api/composables'

import { toast } from 'vue-sonner'

import { ref } from 'vue'

const { t } = useI18n()

useHead({ title: t('menu.my_profile') })
const localePath = useLocalePath()
const userStore = useUserStore()
const accessStore = useAccessStore()

const isLogin = computed(() => {
  const token = accessStore.accessToken
  return !!token?.value && !accessStore.loginExpired
})

const { data: me, isPending, isError, refetch } = useGetMe({ enabled: isLogin })

const userInfo = computed(() => me.value || userStore.user)
const displayName = computed(
  () => userInfo.value?.realname || userInfo.value?.nickname || userInfo.value?.username || '—',
)

// 账户中心快捷入口：与左侧栏同一组路由，桌面 6 列、移动 3 列
const quickLinks = [
  { to: '/orders', icon: 'carbon:document', labelKey: 'mall.orders.title' },
  { to: '/messages', icon: 'lucide:message-square', labelKey: 'messages.title' },
  { to: '/addresses', icon: 'carbon:location', labelKey: 'addresses.title' },
  { to: '/wishlist', icon: 'lucide:heart', labelKey: 'mall.wishlist.title' },
  { to: '/coupons', icon: 'lucide:ticket', labelKey: 'mall.coupons.title' },
  { to: '/settings', icon: 'carbon:settings', labelKey: 'menu.my_account_security' },
] as const

// ---------- 会员积分 ----------
const pointsBalanceQuery = useGetMyPointsBalance({}, { enabled: isLogin } as any)
// balance 经 protobuf JSON 是字符串，先 Number() 再参与比较/拼接
const pointsBalance = computed(() => Number(pointsBalanceQuery.data.value?.balance ?? 0))
const ledgersQuery = useListMyPointsLedgers({ page: 1, pageSize: 5, noPaging: false }, { enabled: isLogin } as any)
const pointsLedgers = computed(() => (((ledgersQuery.data?.value as any)?.items ?? []) as Array<{
  id?: number; delta?: number | string; reason?: string
}>).map((l) => ({ ...l, deltaNum: Number(l.delta ?? 0) })))
const checkInStatusQuery = useGetCheckInStatus({ enabled: isLogin } as any)
const checkedInToday = computed(() => checkInStatusQuery.data.value?.checkedInToday ?? false)
const consecutiveDays = computed(() => checkInStatusQuery.data.value?.consecutiveDays ?? 0)
const checkInMutation = useCheckIn()
const checkInLoading = ref(false)

async function handleCheckIn() {
  if (checkedInToday.value || checkInLoading.value) return
  checkInLoading.value = true
  try {
    const r = await checkIn()
    if (r?.alreadyCheckedIn) {
      toast.info(t('mall.points.alreadyCheckedIn'))
    } else {
      toast.success(t('mall.points.checkInSuccess', { awarded: r?.awarded ?? 0 }))
    }
    checkInStatusQuery.refetch()
    pointsBalanceQuery.refetch()
    ledgersQuery.refetch()
  } catch {
    toast.error(t('mall.points.checkInFailed'))
  } finally {
    checkInLoading.value = false
  }
}
</script>

<template>
  <LayoutPageHero
    :title="t('menu.my_profile')"
    icon="carbon:user"
    size="sm"
  />

  <LayoutSectionContainer max-width="narrow">
    <UiAuthGate v-if="!isLogin" />

    <UiListSkeleton v-else-if="isPending" :rows="4" :header="false" />

    <UiAppEmpty v-else-if="isError" variant="error">
      <template #action>
        <UiButton variant="outline" size="sm" @click="refetch()">
          {{ t('ui.button.retry') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <div v-else-if="userInfo" class="flex flex-col gap-4">
      <!-- 身份卡 -->
      <section class="rounded-lg border border-border bg-card p-5 shadow-card md:p-6">
        <div class="flex items-center gap-4">
          <UiAvatar class="h-16 w-16 border border-border bg-muted">
            <UiImage
              v-if="userInfo.avatar"
              :src="userInfo.avatar"
              :alt="displayName"
              class="h-full w-full object-cover"
            />
            <span v-else class="flex h-full w-full items-center justify-center text-muted-foreground">
              <XIcon icon="carbon:user" :size="30" />
            </span>
          </UiAvatar>
          <div class="min-w-0 flex-1">
            <h2 class="truncate text-lg font-bold text-foreground">{{ displayName }}</h2>
            <p class="mt-0.5 truncate text-sm text-muted-foreground">{{ userInfo.email || userInfo.username }}</p>
          </div>
          <UiStatusBadge
            v-if="pointsBalance > 0"
            tone="primary"
            size="sm"
            :label="`${t('mall.points.title')} ${pointsBalance}`"
            class="shrink-0"
          />
        </div>

        <!-- 我的订单：按状态快捷入口（带 ?status= 预选订单列表筛选项） -->
        <div class="mt-5 border-t border-border pt-4">
          <div class="mb-2 flex items-center justify-between gap-2">
            <span class="text-xs font-medium text-muted-foreground">{{ t('orders.title') }}</span>
            <NuxtLink :to="localePath('/orders')" class="text-xs font-medium text-primary hover:underline">
              {{ t('orders.filter.all') }}
            </NuxtLink>
          </div>
          <div class="flex flex-wrap items-center gap-2">
            <NuxtLink
              :to="localePath('/orders') + '?status=PENDING_PAYMENT'"
              class="rounded-full bg-muted px-3 py-1 text-[11px] font-medium text-foreground transition-colors hover:bg-primary hover:text-primary-foreground"
            >
              {{ t('orders.filter.pending_payment') }}
            </NuxtLink>
            <NuxtLink
              :to="localePath('/orders') + '?status=FULFILLED'"
              class="rounded-full bg-muted px-3 py-1 text-[11px] font-medium text-foreground transition-colors hover:bg-primary hover:text-primary-foreground"
            >
              {{ t('orders.filter.fulfilled') }}
            </NuxtLink>
          </div>
        </div>

        <!-- 账户快捷入口 -->
        <div class="mt-5 grid grid-cols-3 gap-2 border-t border-border pt-4 md:grid-cols-6">
          <NuxtLink
            v-for="link in quickLinks"
            :key="link.to"
            :to="localePath(link.to as any)"
            class="flex min-w-0 flex-col items-center gap-2 rounded-lg px-1 py-3 text-center transition-colors hover:bg-muted"
          >
            <span class="flex h-9 w-9 items-center justify-center rounded-full bg-primary-soft text-primary-soft-foreground">
              <XIcon :icon="link.icon" :size="18" />
            </span>
            <span class="w-full truncate text-[11px] font-medium text-foreground">{{ t(link.labelKey as any) }}</span>
          </NuxtLink>
        </div>
      </section>

      <!-- 会员积分：余额 + 最近台账 -->
      <section class="rounded-lg border border-border bg-card p-5 shadow-card md:p-6">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div class="min-w-0">
            <p class="text-xs font-medium text-muted-foreground">{{ t('mall.points.title') }}</p>
            <p class="mt-1 text-3xl font-extrabold tabular-nums text-primary">{{ pointsBalance }}</p>
            <p v-if="consecutiveDays > 0" class="mt-1 text-[11px] text-muted-foreground">
              {{ t('mall.points.consecutiveDays', { days: consecutiveDays }) }}
            </p>
          </div>
          <UiButton
            :variant="checkedInToday ? 'outline' : 'default'"
            :disabled="checkedInToday || checkInLoading"
            @click="handleCheckIn"
          >
            <XIcon v-if="checkedInToday" icon="lucide:check" :size="16" />
            {{ checkedInToday ? t('mall.points.checkedIn') : t('mall.points.checkIn') }}
          </UiButton>
        </div>

        <ul v-if="pointsLedgers.length > 0" class="mt-4 divide-y divide-border border-t border-border">
          <li
            v-for="l in pointsLedgers"
            :key="l.id"
            class="flex items-center justify-between gap-3 py-2.5 text-xs"
          >
            <span class="min-w-0 truncate text-muted-foreground">{{ l.reason || t('mall.points.reason') }}</span>
            <span
              class="shrink-0 font-semibold tabular-nums"
              :class="l.deltaNum >= 0 ? 'text-success' : 'text-destructive'"
            >
              {{ l.deltaNum >= 0 ? '+' : '' }}{{ l.deltaNum }}
            </span>
          </li>
        </ul>
      </section>
    </div>

    <UiAppEmpty v-else variant="noData" :description="t('common.noData')" />
  </LayoutSectionContainer>
</template>
