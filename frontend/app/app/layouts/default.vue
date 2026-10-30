<script setup lang="ts">
import { usePreferences } from '@/core/preferences/use-preferences'

const { isDark } = usePreferences()
const route = useRoute()

// 检查是否是认证页面
const isAuthPage = computed(() => {
  return route.path.includes('/register') || route.path.includes('/login')
})
</script>

<template>
  <div class="flex min-h-screen w-full flex-col">
    <LayoutNavigationProgress />
    <template v-if="!isAuthPage">
      <LayoutHeader />
    </template>
    <main
      :key="$i18n.locale"
      :class="[
        'flex w-full flex-1 flex-col bg-background',
        // 移动端为底部 Tab 栏（详情页为吸底购买栏）让出空间，避免内容被遮挡
        !isAuthPage && 'pt-(--layout-header-height) pb-[calc(var(--layout-tabbar-height)_+_env(safe-area-inset-bottom))] md:pb-0 min-h-screen',
      ]"
    >
      <slot />
    </main>
    <template v-if="!isAuthPage">
      <LayoutFooter />
    </template>
    <LayoutBackToTop v-if="!isAuthPage" />
    <!-- AI 导购悬浮助手（登录后可用，未登录点击跳登录页） -->
    <AssistantChat v-if="!isAuthPage" />
    <!-- 移动端底部 Tab 导航（认证页不渲染） -->
    <LayoutTabBar v-if="!isAuthPage" />
  </div>
</template>
