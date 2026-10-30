<script setup lang="ts">
import { cn } from '@/lib/utils'
import { XIcon } from '@/plugins/xicon'
import { categoryIconFor } from '@/utils/category-icon'
import { pickTranslation } from '@/utils/translation'

const localePath = useLocalePath()

/**
 * 移动端金刚区。
 *
 * 不自带 SectionContainer：首页把它放进版面栅格，自带容器会在列里再套一层
 * max-width + 内边距。标题也交给页面的 UiSectionHeader，全站一处口径。
 */
const props = defineProps<{
  categories: Array<{ id?: number; translations?: Array<{ name?: string; languageCode?: string }> }>
  class?: string
}>()
</script>

<template>
  <div :class="cn('flex gap-3 overflow-x-auto pb-1 no-scrollbar', props.class)">
    <NuxtLink
      v-for="cat in props.categories"
      :key="cat.id"
      :to="localePath('/category/' + cat.id)"
      class="flex w-[68px] shrink-0 flex-col items-center gap-2"
    >
      <span class="flex h-[52px] w-[52px] items-center justify-center rounded-xl bg-primary-soft text-primary-soft-foreground">
        <XIcon :icon="categoryIconFor(cat.id)" :size="24" />
      </span>
      <span class="line-clamp-1 w-full text-center text-xs text-foreground">
        {{ pickTranslation(cat.translations)?.name || '—' }}
      </span>
    </NuxtLink>
  </div>
</template>
