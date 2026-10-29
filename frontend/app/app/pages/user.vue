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

// ---------- 会员积分 ----------
const pointsBalanceQuery = useGetMyPointsBalance({}, { enabled: isLogin } as any)
const pointsBalance = computed(() => pointsBalanceQuery.data.value?.balance ?? 0)
const ledgersQuery = useListMyPointsLedgers({ page: 1, pageSize: 5, noPaging: false }, { enabled: isLogin } as any)
const pointsLedgers = computed(() => ((ledgersQuery.data?.value as any)?.items ?? []) as Array<{
  id?: number; delta?: number; reason?: string
}>)
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
  <div class="w-full">
    <LayoutPageHero
      :title="t('menu.my_profile')"
      icon="carbon:user"
      size="md"
    />

    <section class="w-full py-12 max-md:py-8">
      <div class="mx-auto max-w-3xl px-4">
        <!-- Not logged in -->
        <div v-if="!isLogin" class="flex flex-col items-center gap-6 py-20">
          <XIcon icon="carbon:locked" :size="48" class="text-muted-foreground" />
          <p class="text-lg text-muted-foreground">{{ t('authentication.login.please_login') }}</p>
          <UiButton @click="navigateTo(localePath('/login'))">
            {{ t('navbar.user.login') }}
          </UiButton>
        </div>

        <!-- Loading -->
        <div v-else-if="isPending" class="rounded-2xl border border-border bg-card p-8">
          <div class="flex items-center gap-6">
            <UiSkeleton class="h-20 w-20 rounded-full" />
            <div class="space-y-2">
              <UiSkeleton class="h-6 w-32" />
              <UiSkeleton class="h-4 w-48" />
            </div>
          </div>
        </div>

        <UiAppEmpty
          v-else-if="isError"
          variant="error"
        >
          <template #action>
            <UiButton variant="outline" size="sm" @click="refetch()">
              {{ t('ui.button.retry') }}
            </UiButton>
          </template>
        </UiAppEmpty>

        <!-- User Info -->
        <div v-else-if="userInfo" class="rounded-2xl border border-border bg-card p-8">
          <div class="flex items-center gap-6">
            <div class="flex h-20 w-20 items-center justify-center rounded-full bg-primary/10 text-primary">
              <XIcon icon="carbon:user" :size="40" />
            </div>
            <div>
              <h2 class="text-xl font-bold text-foreground">{{ userInfo.realname || userInfo.nickname || userInfo.username }}</h2>
              <p class="text-sm text-muted-foreground">{{ userInfo.email || userInfo.username }}</p>
            </div>
          </div>

          <!-- 我的订单：按状态快捷入口（带 ?status= 预选订单列表筛选项） -->
          <div class="mt-4 flex flex-wrap items-center gap-2 rounded-xl border border-border bg-background/40 p-4">
            <span class="text-xs font-medium text-muted-foreground">{{ t('orders.title') }}：</span>
            <NuxtLink
              :to="localePath('/orders') + '?status=PENDING_PAYMENT'"
              class="rounded-full bg-muted px-3 py-1 text-[11px] text-foreground transition-colors hover:bg-primary hover:text-primary-foreground"
            >
              {{ t('orders.filter.pending_payment') }}
            </NuxtLink>
            <NuxtLink
              :to="localePath('/orders') + '?status=FULFILLED'"
              class="rounded-full bg-muted px-3 py-1 text-[11px] text-foreground transition-colors hover:bg-primary hover:text-primary-foreground"
            >
              {{ t('orders.filter.fulfilled') }}
            </NuxtLink>
            <NuxtLink
              :to="localePath('/orders')"
              class="rounded-full bg-muted px-3 py-1 text-[11px] text-foreground transition-colors hover:bg-primary hover:text-primary-foreground"
            >
              {{ t('orders.filter.all') }}
            </NuxtLink>
          </div>

          <!-- 会员积分：余额 + 最近台账 -->
          <div class="mt-4 rounded-xl border border-border bg-background/40 p-4">
            <div class="flex items-center justify-between">
              <span class="text-xs font-medium text-muted-foreground">{{ t('mall.points.title') }}</span>
              <div class="flex items-center gap-2">
                <span class="text-2xl font-extrabold text-primary tabular-nums">{{ pointsBalance }}</span>
                <UiButton
                  size="sm"
                  :variant="checkedInToday ? 'outline' : 'default'"
                  :disabled="checkedInToday || checkInLoading"
                  @click="handleCheckIn"
                >
                  {{ checkedInToday ? t('mall.points.checkedIn') : t('mall.points.checkIn') }}
                </UiButton>
              </div>
            </div>
            <p v-if="consecutiveDays > 0" class="mt-1 text-[10px] text-muted-foreground">
              {{ t('mall.points.consecutiveDays', { days: consecutiveDays }) }}
            </p>
            <ul v-if="pointsLedgers.length > 0" class="mt-3 flex flex-col gap-1.5">
              <li
                v-for="l in pointsLedgers"
                :key="l.id"
                class="flex items-center justify-between text-[11px]"
              >
                <span class="text-muted-foreground">{{ l.reason || t('mall.points.reason') }}</span>
                <span
                  class="font-medium tabular-nums"
                  :class="(l.delta ?? 0) >= 0 ? 'text-primary' : 'text-red-500'"
                >
                  {{ (l.delta ?? 0) >= 0 ? '+' : '' }}{{ l.delta }}
                </span>
              </li>
            </ul>
          </div>
        </div>
      </div>
    </section>
  </div>
</template>
