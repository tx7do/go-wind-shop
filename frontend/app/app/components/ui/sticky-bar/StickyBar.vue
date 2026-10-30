<script setup lang="ts">
import { cn } from '@/lib/utils'

/**
 * 吸底操作条（购物车结算 / 商品详情购买 / 结算下单）。
 *
 * 原来这些条用硬写的 bottom-[calc(3.5rem+…)] 去避让 TabBar，而 TabBar 高度写在
 * TabBar.vue 的 h-14 里、布局内边距又写成 3.75rem，三个数各抄一份、互相对不上。
 * 现在统一读 --layout-tabbar-height（main.css 单一来源），md 以上 TabBar 隐藏，
 * 只留安全区内边距。
 */
const props = withDefaults(defineProps<{
  /** 是否需要让位给移动端 TabBar */
  aboveTabBar?: boolean
  class?: string
}>(), {
  aboveTabBar: true,
})
</script>

<template>
  <div
    :class="cn(
      'fixed inset-x-0 z-40 border-t border-border bg-card',
      props.aboveTabBar
        ? 'bottom-[calc(var(--layout-tabbar-height)_+_env(safe-area-inset-bottom))] md:bottom-[env(safe-area-inset-bottom)]'
        : 'bottom-[env(safe-area-inset-bottom)]',
      props.class,
    )"
  >
    <div class="mx-auto flex w-full max-w-[1200px] items-center gap-3 px-4 py-2.5 md:px-8">
      <slot />
    </div>
  </div>
</template>
