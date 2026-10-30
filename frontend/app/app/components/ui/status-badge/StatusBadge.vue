<script setup lang="ts">
import { cn } from '@/lib/utils'

/**
 * 状态徽标。各页原来各写一份 STATUS_TAG_CLASS（orders / orders/[id] / refunds /
 * messages / coupons 五份分叉），这里收成一处，配色一律走语义 token：
 * 浅底 + 同色系深墨，亮暗两态都过 WCAG AA。
 */
const props = withDefaults(defineProps<{
  tone?: 'neutral' | 'primary' | 'success' | 'warning' | 'danger'
  label?: string
  size?: 'sm' | 'md'
  dot?: boolean
  class?: string
}>(), {
  tone: 'neutral',
  size: 'md',
})

const toneClass: Record<NonNullable<typeof props.tone>, string> = {
  neutral: 'bg-secondary text-secondary-foreground',
  primary: 'bg-primary-soft text-primary-soft-foreground',
  success: 'bg-success-soft text-success-soft-foreground',
  warning: 'bg-warning-soft text-warning-soft-foreground',
  danger: 'bg-destructive-soft text-destructive-soft-foreground',
}

const dotClass: Record<NonNullable<typeof props.tone>, string> = {
  neutral: 'bg-muted-foreground',
  primary: 'bg-primary',
  success: 'bg-success',
  warning: 'bg-warning',
  danger: 'bg-destructive',
}
</script>

<template>
  <span
    :class="cn(
      'inline-flex items-center gap-1 rounded-full font-medium leading-none whitespace-nowrap',
      props.size === 'sm' ? 'px-1.5 py-0.5 text-[10px]' : 'px-2 py-1 text-[11px]',
      toneClass[props.tone],
      props.class,
    )"
  >
    <span
      v-if="props.dot"
      :class="cn('size-1.5 shrink-0 rounded-full', dotClass[props.tone])"
      aria-hidden="true"
    />
    <slot>{{ props.label }}</slot>
  </span>
</template>
