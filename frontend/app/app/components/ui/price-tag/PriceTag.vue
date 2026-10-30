<script setup lang="ts">
import { cn } from '@/lib/utils'
import { formatFen, splitYuan } from '@/utils/money'

const props = withDefaults(defineProps<{
  /** 价格，单位「分」。为空/0 时整块不渲染——服务端与客户端首屏因此保持一致，不会 hydration 抖动 */
  amount?: number | string | null
  /** 划线原价，单位「分」 */
  oldAmount?: number | string | null
  size?: 'sm' | 'md' | 'lg'
  class?: string
}>(), {
  size: 'md',
})

const has = computed(() => Number(props.amount ?? 0) > 0)
const parts = computed(() => splitYuan(props.amount))
const oldText = computed(() => {
  const fen = Number(props.oldAmount ?? 0)
  if (!Number.isFinite(fen) || fen <= 0 || fen <= Number(props.amount ?? 0)) return null
  return formatFen(fen).replace(/\.00$/, '')
})

const intClass = computed(() =>
  props.size === 'lg' ? 'text-[26px]' : props.size === 'sm' ? 'text-[15px]' : 'text-[17px]',
)
</script>

<template>
  <span v-if="has" :class="cn('inline-flex items-baseline gap-0.5 font-bold tabular-nums text-foreground', props.class)">
    <span :class="cn('font-semibold', props.size === 'lg' ? 'text-sm' : 'text-[11px]')">¥</span>
    <span :class="cn('leading-none tracking-tight', intClass)">{{ parts.int }}</span>
    <span v-if="parts.dec" :class="props.size === 'lg' ? 'text-base' : 'text-xs'">{{ parts.dec }}</span>
    <span
      v-if="oldText"
      class="ml-1.5 text-[11px] font-normal text-muted-foreground line-through"
    >¥{{ oldText }}</span>
  </span>
</template>
