<script setup lang="ts">
import { cn } from '@/lib/utils'

const props = withDefaults(defineProps<{
  count?: number
  class?: string
}>(), {
  count: 6,
})
</script>

<template>
  <div :class="cn('grid grid-cols-2 gap-4 md:grid-cols-4 lg:grid-cols-5', props.class)">
    <UiCategoryCardSkeleton v-for="(_, i) in count" :key="i" />
  </div>
</template>
