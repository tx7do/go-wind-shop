<script setup lang="ts">
import { cn } from '@/lib/utils'
import { XIcon } from '@/plugins/xicon'
import { getCurrentLocale } from '@/utils/locale'

const { t } = useI18n()
const localePath = useLocalePath()

const props = defineProps<{
  categories: Array<{ id?: number; translations?: Array<{ name?: string; languageCode?: string }> }>
  class?: string
}>()

const currentLocale = computed(() => getCurrentLocale())

function pickTranslation<T extends { languageCode?: string }>(
  translations: T[] | undefined,
): T | undefined {
  if (!translations || translations.length === 0) return undefined
  return translations.find((tr) => tr.languageCode === currentLocale.value) ?? translations[0]
}

// 固定圖標集：按分類 id 取模映射，純視覺用途，不做語義對應
// 與 CategoryQuickNav 保持一致，確保同一分類在金剛區/側欄圖標相同
const ICON_SET = [
  'carbon:shopping-bag',
  'carbon:product',
  'carbon:category',
  'carbon:document',
  'carbon:settings',
  'carbon:add',
  'carbon:shopping-cart',
  'carbon:document-unknown',
] as const

function iconFor(id?: number): string {
  const n = id ?? 0
  return ICON_SET[Math.abs(n) % ICON_SET.length]
}
</script>

<template>
  <LayoutSectionContainer :class="cn('hidden md:block', props.class)">
    <div class="rounded-2xl border border-border bg-card p-3">
      <h2 class="mb-3 px-3 pt-2 text-xs font-bold uppercase tracking-wider text-muted-foreground">
        {{ t('mall.home.categories') }}
      </h2>
      <ul class="space-y-3">
        <li v-for="cat in categories" :key="cat.id">
          <NuxtLink
            :to="localePath('/category/' + cat.id)"
            class="group flex flex-row items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium text-foreground transition-colors hover:bg-muted/50"
          >
            <span class="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
              <XIcon :icon="iconFor(cat.id)" :size="18" />
            </span>
            <span class="min-w-0 flex-1 whitespace-nowrap">
              {{ pickTranslation(cat.translations)?.name || '—' }}
            </span>
            <span class="text-muted-foreground transition-transform group-hover:translate-x-1">
              <XIcon icon="carbon:chevron-right" :size="16" />
            </span>
          </NuxtLink>
        </li>
      </ul>
    </div>
  </LayoutSectionContainer>
</template>
