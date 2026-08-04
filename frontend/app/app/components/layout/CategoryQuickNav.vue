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
  <LayoutSectionContainer :class="props.class">
    <h2 class="mb-4 text-lg font-bold text-foreground">
      {{ t('mall.home.categories') }}
    </h2>

    <div
      class="-mx-4 px-4 overflow-x-auto md:hidden"
      style="-webkit-overflow-scrolling: touch; scrollbar-width: none;"
    >
      <div class="flex gap-3 w-max pb-2">
        <NuxtLink
          v-for="cat in categories"
          :key="cat.id"
          :to="localePath('/category/' + cat.id)"
          class="flex w-[72px] shrink-0 flex-col items-center gap-2 rounded-xl border border-border bg-card p-2 transition-colors hover:border-primary/60"
        >
          <span class="flex h-12 w-12 items-center justify-center rounded-lg bg-primary/10 text-primary">
            <XIcon :icon="iconFor(cat.id)" :size="24" />
          </span>
          <span class="line-clamp-1 w-full text-center text-[11px] text-muted-foreground">
            {{ pickTranslation(cat.translations)?.name || '—' }}
          </span>
        </NuxtLink>
      </div>
    </div>
  </LayoutSectionContainer>
</template>
