<script setup lang="ts">
import { cn } from '@/lib/utils'

const props = defineProps<{
  class?: string
  maxWidth?: 'default' | 'wide' | 'narrow'
  noPadding?: boolean
}>()

const widthClass = computed(() =>
  props.maxWidth === 'wide' ? 'max-w-300'
    : props.maxWidth === 'narrow' ? 'max-w-3xl'
    : 'max-w-(--layout-max-width)'
)
</script>

<template>
  <div :class="cn('w-full mx-auto px-4 md:px-8', !noPadding && 'py-6 md:py-10', widthClass, props.class)">
    <slot />
  </div>
</template>
