<script setup lang="ts">
import { XIcon } from '@/plugins/xicon'
import { useAccessStore } from '@/stores/modules/core/access.state'
import { useUserStore } from '@/stores/modules/core/user.state'
import { useListCarts, useListCartItems } from '@/api/composables'

const { t } = useI18n()
const localePath = useLocalePath()
const route = useRoute()

const accessStore = useAccessStore()
const userStore = useUserStore()

const isLogin = computed(() => {
  const token = accessStore.accessToken
  return !!token?.value && !accessStore.loginExpired
})
const currentUserId = computed(() => userStore.user?.id ?? 0)

// 商品详情页移动端由吸底购买栏接管底部，Tab 栏让位
const isProductDetail = computed(() => /\/product\/\d+/.test(route.path))

// 购物车角标：与 Header 同参数查询，vue-query 按 key 共享缓存，不会双发请求
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

// 去掉 locale 前缀后做前缀匹配，保证亮暗两态与多语言路由都正确
const currentPath = computed(() =>
  route.path.replace(/^\/(zh-CN|en-US)/, '') || '/',
)

const tabs = computed(() => [
  { to: localePath('/'), icon: 'carbon:home', label: t('mall.tab.home'), match: (p: string) => p === '/' },
  { to: localePath('/category'), icon: 'carbon:category', label: t('mall.tab.category'), match: (p: string) => p.startsWith('/category') },
  { to: localePath('/cart'), icon: 'carbon:shopping-cart', label: t('mall.tab.cart'), match: (p: string) => p.startsWith('/cart'), badge: cartCount.value },
  { to: localePath('/user'), icon: 'carbon:user', label: t('mall.tab.mine'), match: (p: string) => p.startsWith('/user') },
])
</script>

<template>
  <nav
    v-if="!isProductDetail"
    class="fixed bottom-0 left-0 right-0 z-50 flex md:hidden border-t border-border bg-card/95 backdrop-blur-md dark:bg-card/90"
    :style="{ paddingBottom: 'env(safe-area-inset-bottom)' }"
    aria-label="Bottom navigation"
  >
    <NuxtLink
      v-for="tab in tabs"
      :key="tab.label"
      :to="tab.to"
      class="relative flex h-14 flex-1 flex-col items-center justify-center gap-0.5 transition-colors"
      :class="tab.match(currentPath) ? 'text-primary' : 'text-muted-foreground'"
      :aria-current="tab.match(currentPath) ? 'page' : undefined"
    >
      <span class="relative">
        <XIcon :icon="tab.icon" :size="22" />
        <span
          v-if="tab.badge && tab.badge > 0"
          class="absolute -right-2.5 -top-1.5 flex h-4 min-w-4 items-center justify-center rounded-full bg-primary px-1 text-[9px] font-bold text-primary-foreground"
        >
          {{ tab.badge > 99 ? '99+' : tab.badge }}
        </span>
      </span>
      <span class="text-[10px] leading-none">{{ tab.label }}</span>
    </NuxtLink>
  </nav>
</template>
