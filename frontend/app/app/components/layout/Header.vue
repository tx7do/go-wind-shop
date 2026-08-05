<script setup lang="ts">
import { XIcon } from '@/plugins/xicon'
import { usePreferences } from '@/core/preferences/use-preferences'
import { useAccessStore } from '@/stores/modules/core/access.state'
import { useAuthStore } from '@/stores/modules/app/auth.state'
import { useUserStore } from '@/stores/modules/core/user.state'
import { useListCarts, useListCartItems } from '@/api/composables'

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
const cartsQuery = useListCarts(
  computed(() => ({
    page: 1,
    pageSize: 1,
    noPaging: false,
    query: JSON.stringify({ userId: currentUserId.value }),
  })),
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
)
const cartCount = computed(() => {
  const items = (cartItemsQuery.data?.value as any)?.items ?? []
  return (items as Array<{ quantity?: number }>).reduce(
    (acc, it) => acc + (it.quantity ?? 0),
    0,
  )
})

const handleClickLogo = () => navigateTo(localePath('/'))
const handleClickSettings = () => navigateTo(localePath('/settings'))
const handleClickUserHomepage = () => navigateTo(localePath('/user'))
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
        <img src="/logo.png" alt="Logo" width="36" height="36" class="h-9 w-9 shrink-0 max-md:h-8 max-md:w-8" />
        <span class="text-lg font-bold text-primary whitespace-nowrap max-md:hidden">
          {{ t('app.title') }}
        </span>
      </button>

      <!-- 功能按钮区 -->
      <div class="flex shrink-0 items-center gap-1">
        <!-- 购物车入口（带数量徽标） -->
        <NuxtLink
          :to="localePath('/cart')"
          class="relative flex h-9 w-9 items-center justify-center rounded-lg text-foreground transition-colors hover:bg-primary/10"
          aria-label="Shopping cart"
        >
          <XIcon icon="carbon:shopping-cart" width="18" height="18" />
          <span
            v-if="cartCount > 0"
            class="absolute -right-0.5 -top-0.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-green-500 px-1 text-[9px] font-bold text-white dark:bg-green-400 dark:text-green-950"
          >
            {{ cartCount > 99 ? '99+' : cartCount }}
          </span>
        </NuxtLink>

        <!-- 用户菜单 -->
        <UiDropdownMenu :modal="false">
          <UiDropdownMenuTrigger as-child>
            <UiButton variant="ghost" size="icon" aria-label="User menu">
              <XIcon icon="lucide:user" width="16" height="16" />
            </UiButton>
          </UiDropdownMenuTrigger>
          <UiDropdownMenuContent align="end">
            <template v-if="isLogin">
              <UiDropdownMenuItem @click="handleClickUserHomepage">
                <XIcon icon="lucide:home" width="16" height="16" />
                {{ t('menu.homepage') }}
              </UiDropdownMenuItem>
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

        <!-- 语言菜单 -->
        <UiDropdownMenu :modal="false">
          <UiDropdownMenuTrigger as-child>
            <UiButton variant="ghost" size="icon" aria-label="Language">
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

        <!-- 主题菜单 -->
        <UiDropdownMenu :modal="false">
          <UiDropdownMenuTrigger as-child>
            <UiButton variant="ghost" size="icon" aria-label="Toggle theme">
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
