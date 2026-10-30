<script setup lang="ts">
import { cn } from '@/lib/utils'

/**
 * 列表骨架。订单/退款/发票/收藏/购物车/结算/地址/消息/我的 9 个页面各自手搓过
 * 一份「表头条 + N 行灰条」，这里收成一处。
 * variant=row 用于表格型列表，card 用于商品网格。
 */
const props = withDefaults(defineProps<{
  rows?: number
  variant?: 'row' | 'card'
  /** 是否带一条表头骨架 */
  header?: boolean
  /** card 变体的列数（与页面真实网格保持一致，避免加载完跳动） */
  columns?: number
  class?: string
}>(), {
  rows: 4,
  variant: 'row',
  header: true,
  columns: 2,
})

const gridClass = computed(() =>
  props.columns === 4
    ? 'grid-cols-2 md:grid-cols-4'
    : props.columns === 3
      ? 'grid-cols-2 md:grid-cols-3'
      : 'grid-cols-2',
)
</script>

<template>
  <div
    v-if="props.variant === 'card'"
    :class="cn('grid gap-3', gridClass, props.class)"
    role="status"
    aria-live="polite"
  >
    <UiPostCardSkeleton v-for="i in props.rows" :key="i" />
  </div>

  <div v-else :class="cn('overflow-hidden rounded-lg border border-border bg-card shadow-card', props.class)" role="status" aria-live="polite">
    <div v-if="props.header" class="border-b border-border bg-muted/40 px-4 py-3">
      <UiSkeleton class="h-3.5 w-40" />
    </div>
    <div
      v-for="i in props.rows"
      :key="i"
      class="flex items-center gap-4 border-b border-border px-4 py-4 last:border-b-0"
    >
      <UiSkeleton class="h-4 w-12 shrink-0" />
      <UiSkeleton class="h-4 flex-1" />
      <UiSkeleton class="h-4 w-20 shrink-0" />
      <UiSkeleton class="h-5 w-16 shrink-0 rounded-full" />
    </div>
  </div>
</template>
