<script setup lang="ts">
import { cn } from '@/lib/utils'

const props = defineProps<{
  class?: string
}>()
</script>

<template>
  <div :class="cn('overflow-hidden rounded-lg border border-border bg-card shadow-card', props.class)">
    <UiSkeleton class="aspect-square w-full rounded-none" />
    <div class="space-y-2 p-3">
      <UiSkeleton class="h-3.5 w-11/12" />
      <UiSkeleton class="h-3.5 w-2/3" />
      <UiSkeleton class="mt-1 h-5 w-24" />
    </div>
  </div>
</template>
