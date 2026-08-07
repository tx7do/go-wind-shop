<script setup lang="ts">
definePageMeta({
  layout: 'account',
  middleware: 'auth',
})
import { XIcon } from '@/plugins/xicon'
import { useUserStore } from '@/stores/modules/core/user.state'
import { useAccessStore } from '@/stores/modules/core/access.state'
import { useGetMe } from '@/api/composables/user-profile'

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
            <span class="text-xs font-medium text-muted-foreground">{{ t('mall.orders.title') }}：</span>
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
        </div>
      </div>
    </section>
  </div>
</template>
