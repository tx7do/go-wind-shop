<script setup lang="ts">
import { cn } from '@/lib/utils'
import { XIcon } from '@/plugins/xicon'
import { categoryIconFor } from '@/utils/category-icon'
import { pickTranslation } from '@/utils/translation'

const { t } = useI18n()
const localePath = useLocalePath()

/**
 * 桌面端首页左侧垂直分类菜单。
 *
 * 图标改用与金刚区同一个 `categoryIconFor`：原来这里有一份本地 ICON_SET 副本，
 * 与共享那份的第 5、6 项已经分叉，同一个分类在金刚区和侧栏显示的是不同图标。
 */
const props = defineProps<{
  categories: Array<{ id?: number; translations?: Array<{ name?: string; languageCode?: string }> }>
  class?: string
}>()
</script>

<template>
  <nav
    :class="cn('rounded-lg border border-border bg-card p-2', props.class)"
    :aria-label="t('mall.home.categories')"
  >
    <h2 class="px-3 pb-1 pt-2 text-xs font-bold uppercase tracking-wider text-muted-foreground">
      {{ t('mall.home.categories') }}
    </h2>
    <ul>
      <li v-for="cat in props.categories" :key="cat.id">
        <NuxtLink
          :to="localePath('/category/' + cat.id)"
          class="group flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium text-foreground transition-colors hover:bg-secondary"
        >
          <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-primary-soft text-primary-soft-foreground">
            <XIcon :icon="categoryIconFor(cat.id)" :size="18" />
          </span>
          <span class="min-w-0 flex-1 truncate">
            {{ pickTranslation(cat.translations)?.name || '—' }}
          </span>
          <span class="text-muted-foreground transition-transform group-hover:translate-x-0.5">
            <XIcon icon="carbon:chevron-right" :size="16" />
          </span>
        </NuxtLink>
      </li>
    </ul>
  </nav>
</template>
