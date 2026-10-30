<script setup lang="ts">
import { cn } from '@/lib/utils'

/**
 * 章节标题。A 稿的层级：标题 16px（桌面 18px）粗体 + 可选说明 + 右侧「查看全部」。
 * 各页原来从 text-2xl 到 text-base 五种写法都有，这里统一一档。
 */
const props = defineProps<{
  title?: string
  description?: string
  moreLabel?: string
  moreTo?: string
  class?: string
}>()

const localePath = useLocalePath()
</script>

<template>
  <div :class="cn('mb-3 flex items-baseline justify-between gap-3', props.class)">
    <div class="min-w-0">
      <h2 class="truncate text-base font-bold text-foreground md:text-lg">
        <slot name="title">{{ props.title }}</slot>
      </h2>
      <p v-if="props.description" class="mt-1 text-xs text-muted-foreground">
        {{ props.description }}
      </p>
    </div>
    <NuxtLink
      v-if="props.moreLabel && props.moreTo"
      :to="localePath(props.moreTo)"
      class="shrink-0 text-xs text-primary"
    >
      {{ props.moreLabel }}
    </NuxtLink>
  </div>
</template>
