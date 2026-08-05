<script setup lang="ts">
import { cn } from '@/lib/utils'
import { getCurrentLocale } from '@/utils/locale'

const { t } = useI18n()
const localePath = useLocalePath()

const props = defineProps<{
  categories: Array<{ id?: number; translations?: Array<{ name?: string; languageCode?: string }> }>
  brands: Array<{ id?: number; translations?: Array<{ name?: string; languageCode?: string }> }>
  currentCategoryId?: number
  selectedBrandId?: number
  class?: string
}>()

const emit = defineEmits<{
  (e: 'update:selectedBrandId', value: number | undefined): void
}>()

const currentLocale = computed(() => getCurrentLocale())

function pickTranslation<T extends { languageCode?: string }>(
  translations: T[] | undefined,
): T | undefined {
  if (!translations || translations.length === 0) return undefined
  return translations.find((tr) => tr.languageCode === currentLocale.value) ?? translations[0]
}

// 品牌篩選：單選語義。選中未選中的品牌則設為當前，再次點擊已選中品牌則取消。
function onBrandToggle(brandId: number | undefined) {
  if (brandId === undefined) return
  if (props.selectedBrandId === brandId) {
    emit('update:selectedBrandId', undefined)
  } else {
    emit('update:selectedBrandId', brandId)
  }
}
</script>

<template>
  <LayoutSectionContainer :class="cn('hidden md:block', props.class)" no-padding>
    <div class="rounded-2xl border border-border bg-card p-4">
      <h2 class="mb-4 text-sm font-bold text-foreground">
        {{ t('mall.category.filters') }}
      </h2>

      <!-- 分類切換列表 -->
      <div class="mb-6">
        <h3 class="mb-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          {{ t('mall.category.filterCategories') }}
        </h3>
        <ul class="space-y-1">
          <li v-for="cat in categories" :key="cat.id">
            <NuxtLink
              :to="localePath('/category/' + cat.id)"
              :class="cn(
                'flex items-center justify-between rounded-lg px-3 py-2 text-sm transition-colors',
                cat.id === props.currentCategoryId
                  ? 'bg-primary/10 text-primary'
                  : 'text-foreground hover:bg-muted/50',
              )"
            >
              <span class="truncate">
                {{ pickTranslation(cat.translations)?.name || '—' }}
              </span>
              <span v-if="cat.id === props.currentCategoryId" class="text-primary">
                •
              </span>
            </NuxtLink>
          </li>
        </ul>
      </div>

      <!-- 品牌篩選 -->
      <div>
        <h3 class="mb-2 text-xs font-semibold uppercase tracking-wider text-muted-foreground">
          {{ t('mall.category.filterBrands') }}
        </h3>
        <ul class="space-y-1">
          <li v-for="brand in brands" :key="brand.id">
            <label
              :class="cn(
                'flex cursor-pointer items-center gap-2 rounded-lg px-3 py-2 text-sm transition-colors',
                brand.id === props.selectedBrandId
                  ? 'bg-primary/10 text-primary'
                  : 'text-foreground hover:bg-muted/50',
              )"
            >
              <input
                type="checkbox"
                class="h-4 w-4 rounded border-border accent-[hsl(var(--primary))]"
                :checked="brand.id === props.selectedBrandId"
                @change="onBrandToggle(brand.id)"
              />
              <span class="truncate">
                {{ pickTranslation(brand.translations)?.name || '—' }}
              </span>
            </label>
          </li>
        </ul>
      </div>
    </div>
  </LayoutSectionContainer>
</template>
