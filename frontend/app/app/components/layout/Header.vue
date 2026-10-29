<script setup lang="ts">
import { XIcon } from '@/plugins/xicon'
import { usePreferences } from '@/core/preferences/use-preferences'
import { useAccessStore } from '@/stores/modules/core/access.state'
import { useAuthStore } from '@/stores/modules/app/auth.state'
import { useUserStore } from '@/stores/modules/core/user.state'
import { useListCarts, useListCartItems, useListUserInbox } from '@/api/composables'

const { t } = useI18n()
const { themePreferences: themePref, setTheme: setThemeMode } = usePreferences()
const currentMode = computed(() => themePref.value.mode)

const localePath = useLocalePath()
const switchLocalePath = useSwitchLocalePath()

const changeLocale = (code: 'zh-CN' | 'en-US') => {
  navigateTo(switchLocalePath(code))
}
const accessStore = useAccessStore()
const authStore = useAuthStore()
const userStore = useUserStore()

const isLogin = computed(() => {
  const token = accessStore.accessToken
  return !!token?.value && !accessStore.loginExpired
})

const currentUserId = computed(() => userStore.user?.id ?? 0)

// 购物车数量徽标：查询当前用户的购物车项，总数用于红点显示。
// enabled: isLogin——购物车接口要求登录，游客发请求会吃 401 并触发重认证跳转，
// 导致匿名访客被踢出公开页面（游客必须能逛商城）。
const cartsQuery = useListCarts(
  computed(() => ({
    page: 1,
    pageSize: 1,
    noPaging: false,
    query: JSON.stringify({ userId: currentUserId.value }),
  })),
  { enabled: isLogin },
)
const cartId = computed(() => {
  const items = ((cartsQuery.data?.value as any)?.items ?? []) as Array<{ id?: number }>
  return items[0]?.id
})
const cartItemsQuery = useListCartItems(
  computed(() => ({
    page: 1,
    pageSize: 100,
    noPaging: false,
    query: cartId.value === undefined ? undefined : JSON.stringify({ cartId: cartId.value }),
  })),
  { enabled: isLogin },
)
const cartCount = computed(() => {
  const items = (cartItemsQuery.data?.value as any)?.items ?? []
  return (items as Array<{ quantity?: number }>).reduce(
    (acc, it) => acc + (it.quantity ?? 0),
    0,
  )
})

// 消息未读徽标：拉取本人站内信收件箱（SENT/RECEIVED 视为未读，与 messages 页
// 同口径），60s 轮询。仅登录后启用；未登录不请求。
const inboxQuery = useListUserInbox(
  computed(() => ({
    page: 1,
    pageSize: 50,
    noPaging: false,
    sorting: [{ field: 'id', direction: 'DESC' }],
  })),
  { enabled: isLogin, refetchInterval: 60_000 },
)
const unreadCount = computed(() => {
  const items = ((inboxQuery.data?.value as any)?.items ?? []) as Array<{
    status?: string
  }>
  return items.filter((m) => m.status === 'SENT' || m.status === 'RECEIVED')
    .length
})

const handleClickLogo = () => navigateTo(localePath('/'))
const handleClickSettings = () => navigateTo(localePath('/user'))
const handleClickLogin = () => navigateTo(localePath('/login'))
const handleClickRegister = () => navigateTo(localePath('/register'))
const handleClickCart = () => navigateTo(localePath('/cart'))
const handleClickLogout = async () => {
  if (isLogin.value) await authStore.logout()
}
</script>

<template>
  <header class="fixed top-0 left-0 right-0 z-1000 flex justify-center bg-background/80 backdrop-blur-md border-b border-border/50 dark:border-border/30 dark:bg-background/60">
    <div class="flex h-(--layout-header-height) w-full max-w-300 items-center gap-6 px-4 max-md:gap-3 max-md:px-3">
      <!-- Logo -->
      <button
        type="button"
        class="flex shrink-0 items-center gap-2 rounded-lg p-1 transition-colors hover:bg-primary/10"
        aria-label="Go to homepage"
        @click="handleClickLogo"
      >
        <img src="/logo.svg" alt="GoWind" width="36" height="36" class="h-9 w-9 shrink-0 max-md:h-8 max-md:w-8" />
        <span class="text-lg font-bold text-primary whitespace-nowrap max-md:hidden">
          {{ t('app.title') }}
        </span>
      </button>

      <!-- 全局搜索（桌面端；移动端首页顶部有搜索入口） -->
      <LayoutSearchBar />

      <!-- 功能按钮区 -->
      <div class="flex shrink-0 items-center gap-1">
        <!-- 消息中心入口（带未读徽标，仅登录后展示） -->
        <NuxtLink
          v-if="isLogin"
          :to="localePath('/messages')"
          class="relative flex h-9 w-9 items-center justify-center rounded-lg text-foreground transition-colors hover:bg-primary/10"
          :aria-label="t('menu.my_messages')"
        >
          <!-- GoWind 风铃图标：铃身内嵌风线，与品牌标志同一笔触 -->
          <svg
            viewBox="0 0 24 24"
            width="18"
            height="18"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            aria-hidden="true"
          >
            <path d="M6 8a6 6 0 0 1 12 0c0 7 3 9 3 9H3s3-2 3-9" />
            <path d="M10.3 21a1.94 1.94 0 0 0 3.4 0" />
            <path d="M9.5 13h4.6a1.9 1.9 0 1 0-1.8-2.5" />
          </svg>
          <span
            v-if="unreadCount > 0"
            class="absolute -right-0.5 -top-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-destructive px-1 text-[9px] font-bold text-destructive-foreground"
          >
            {{ unreadCount > 99 ? '99+' : unreadCount }}
          </span>
        </NuxtLink>

        <!-- 购物车入口（带数量徽标） -->
        <NuxtLink
          :to="localePath('/cart')"
          class="relative flex h-9 w-9 items-center justify-center rounded-lg text-foreground transition-colors hover:bg-primary/10"
          aria-label="Shopping cart"
        >
          <XIcon icon="carbon:shopping-cart" width="18" height="18" />
          <span
            v-if="cartCount > 0"
            class="absolute -right-0.5 -top-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-primary px-1 text-[9px] font-bold text-primary-foreground"
          >
            {{ cartCount > 99 ? '99+' : cartCount }}
          </span>
        </NuxtLink>

        <!-- 用户菜单（移动端收进底部 Tab「我的」） -->
        <UiDropdownMenu :modal="false">
          <UiDropdownMenuTrigger as-child>
            <UiButton variant="ghost" size="icon" aria-label="User menu" class="max-md:hidden">
              <XIcon icon="lucide:user" width="16" height="16" />
            </UiButton>
          </UiDropdownMenuTrigger>
          <UiDropdownMenuContent align="end">
            <template v-if="isLogin">
              <UiDropdownMenuItem @click="handleClickSettings">
                <XIcon icon="lucide:user" width="16" height="16" />
                {{ t('menu.my_profile') }}
              </UiDropdownMenuItem>
              <UiDropdownMenuSeparator />
              <UiDropdownMenuItem class="text-destructive" @click="handleClickLogout">
                <XIcon icon="lucide:log-out" width="16" height="16" />
                {{ t('menu.logout') }}
              </UiDropdownMenuItem>
            </template>
            <template v-else>
              <UiDropdownMenuItem @click="handleClickLogin">
                <XIcon icon="lucide:user" width="16" height="16" />
                {{ t('navbar.user.login') }}
              </UiDropdownMenuItem>
              <UiDropdownMenuItem @click="handleClickRegister">
                <XIcon icon="lucide:user" width="16" height="16" />
                {{ t('navbar.user.register') }}
              </UiDropdownMenuItem>
            </template>
          </UiDropdownMenuContent>
        </UiDropdownMenu>

        <!-- 语言菜单（移动端收进 设置 页） -->
        <UiDropdownMenu :modal="false">
          <UiDropdownMenuTrigger as-child>
            <UiButton variant="ghost" size="icon" aria-label="Language" class="max-md:hidden">
              <XIcon icon="lucide:globe" width="16" height="16" />
            </UiButton>
          </UiDropdownMenuTrigger>
          <UiDropdownMenuContent align="end">
            <UiDropdownMenuItem @click="changeLocale('zh-CN')">
              {{ t('navbar.language.zh-CN') }}
            </UiDropdownMenuItem>
            <UiDropdownMenuItem @click="changeLocale('en-US')">
              {{ t('navbar.language.en-US') }}
            </UiDropdownMenuItem>
          </UiDropdownMenuContent>
        </UiDropdownMenu>

        <!-- 主题菜单（移动端收进 设置 页） -->
        <UiDropdownMenu :modal="false">
          <UiDropdownMenuTrigger as-child>
            <UiButton variant="ghost" size="icon" aria-label="Toggle theme" class="max-md:hidden">
              <XIcon                 :icon="currentMode === 'dark' ? 'lucide:moon' : currentMode === 'light' ? 'lucide:sun' : 'lucide:monitor'"
                width="16" height="16"
                class="theme-icon-animate"
              />
            </UiButton>
          </UiDropdownMenuTrigger>
          <UiDropdownMenuContent align="end">
            <UiDropdownMenuItem @click="setThemeMode('dark')">
              <XIcon icon="lucide:moon" width="16" height="16" />
              {{ t('navbar.theme.dark') }}
            </UiDropdownMenuItem>
            <UiDropdownMenuItem @click="setThemeMode('light')">
              <XIcon icon="lucide:sun" width="16" height="16" />
              {{ t('navbar.theme.light') }}
            </UiDropdownMenuItem>
            <UiDropdownMenuItem @click="setThemeMode('auto')">
              <XIcon icon="lucide:monitor" width="16" height="16" />
              {{ t('navbar.theme.system') }}
            </UiDropdownMenuItem>
          </UiDropdownMenuContent>
        </UiDropdownMenu>
      </div>
    </div>
  </header>
</template>
