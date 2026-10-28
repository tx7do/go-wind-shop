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
        !isAuthPage && 'pt-(--layout-header-height) min-h-screen',
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
  </div>
</template>
