<script setup lang="ts">
import { XIcon } from '@/plugins/xicon'
import { cn } from '@/lib/utils'

const { t } = useI18n()
const route = useRoute()
const localePath = useLocalePath()

// 账户中心侧栏导航：走真实路由跳转（NuxtLink），激活态对比当前路径。
// 文案复用现有 i18n key（menu.* / mall.* / addresses.title 等），无需新增。
const navItems = [
  { to: '/user', icon: 'carbon:user', labelKey: 'menu.my_profile' },
  { to: '/orders', icon: 'carbon:document', labelKey: 'mall.orders.title' },
  { to: '/messages', icon: 'lucide:message-square', labelKey: 'messages.title' },
  { to: '/refunds', icon: 'lucide:rotate-ccw', labelKey: 'refunds.title' },
  { to: '/addresses', icon: 'carbon:location', labelKey: 'addresses.title' },
  { to: '/wishlist', icon: 'lucide:heart', labelKey: 'mall.wishlist.title' },
  { to: '/coupons', icon: 'lucide:ticket', labelKey: 'mall.coupons.title' },
  { to: '/settings', icon: 'carbon:settings', labelKey: 'menu.my_account_security' },
] as const

const isActive = (path: string) => {
  const resolved = localePath(path as any)
  return route.path === resolved
}
</script>

<template>
  <div class="flex min-h-screen w-full flex-col">
    <LayoutNavigationProgress />
    <LayoutHeader />
    <div :class="['flex w-full flex-1 flex-row bg-background pt-(--layout-header-height) min-h-screen max-md:flex-col']">
      <aside class="w-60 shrink-0 border-r border-border max-md:hidden">
        <nav class="sticky top-(--layout-header-height) space-y-1 p-3">
          <NuxtLink
            v-for="item in navItems"
            :key="item.to"
            :to="localePath(item.to as any)"
            :class="cn(
              'flex items-center gap-3 rounded-lg border px-3 py-2.5 text-sm font-medium transition-all',
              isActive(item.to)
                ? 'border-primary/20 bg-primary/5 text-primary'
                : 'border-transparent text-muted-foreground hover:bg-muted hover:text-foreground',
            )"
          >
            <div
              :class="cn(
                'flex h-8 w-8 items-center justify-center rounded-md',
                isActive(item.to)
                  ? 'bg-primary/10 text-primary'
                  : 'bg-muted text-muted-foreground',
              )"
            >
              <XIcon :icon="item.icon" :size="18" />
            </div>
            <span>{{ t(item.labelKey as any) }}</span>
          </NuxtLink>
        </nav>
      </aside>

      <!-- 移动端导航：顶部下拉选择 -->
      <div class="border-b border-border p-3 md:hidden">
        <select
          class="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm font-medium text-foreground outline-none"
          :value="route.path"
          @change="(e) => navigateTo((e.target as HTMLSelectElement).value)"
        >
          <option
            v-for="item in navItems"
            :key="item.to"
            :value="localePath(item.to as any)"
          >
            {{ t(item.labelKey as any) }}
          </option>
        </select>
      </div>

      <main :key="$i18n.locale" class="flex w-full flex-1 flex-col min-w-0">
        <slot />
      </main>
    </div>
    <LayoutFooter />
    <LayoutBackToTop />
  </div>
</template>
