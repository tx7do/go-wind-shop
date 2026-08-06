<script setup lang="ts">
import { computed } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<{
  variant?: 'default' | 'error' | 'noData'
  inContainer?: boolean
  description?: string
  class?: string
}>()

const { t } = useI18n()

// 变体仅决定图标语义：error 用警告色，default/noData 用中性色。
// 不为 error 渲染“假数据”的中性 inbox 图标，避免把错误态误读为空态。
const iconName = computed(() =>
  props.variant === 'error' ? 'lucide:triangle-alert' : 'lucide:inbox',
)
const titleKey = computed(() =>
  props.variant === 'error' ? 'ui.error.title' : '',
)
const descKey = computed(() =>
  props.variant === 'error' ? 'ui.error.description' : '',
)
</script>

<template>
  <div :class="cn('flex w-full flex-col items-center justify-center gap-3 py-12 px-5', props.inContainer && 'my-20', props.class)">
    <div :class="cn('opacity-50', props.variant === 'error' && 'text-destructive')">
      <slot name="image">
        <XIcon :icon="iconName" width="64" height="64" class="text-muted-foreground" />
      </slot>
    </div>
    <template v-if="props.variant === 'error'">
      <p v-if="titleKey" class="text-base font-semibold text-foreground">{{ t(titleKey) }}</p>
      <p v-if="descKey" class="max-w-md text-center text-sm text-muted-foreground">{{ t(descKey) }}</p>
    </template>
    <template v-else>
      <p v-if="description" class="text-sm text-muted-foreground">{{ description }}</p>
      <p v-else-if="$slots.default" class="text-sm text-muted-foreground">
        <slot />
      </p>
    </template>
    <div v-if="$slots.action" class="mt-2">
      <slot name="action" />
    </div>
  </div>
</template>
