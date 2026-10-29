<script setup lang="ts">
import { XIcon } from '@/plugins/xicon'
import { useListCategories } from '@/api/composables'
import { categoryIconFor } from '@/utils/category-icon'
import { getCurrentLocale } from '@/utils/locale'

const { t } = useI18n()
const localePath = useLocalePath()

useHead({ title: t('mall.category.title') })

const currentLocale = computed(() => getCurrentLocale())

type CategoryTranslation = {
  name?: string
  languageCode?: string
}
type CategoryEntity = {
  id?: number
  imageUrl?: string
  translations?: CategoryTranslation[]
}

function pickTranslation<T extends { languageCode?: string }>(
  translations: T[] | undefined,
): T | undefined {
  if (!translations || translations.length === 0) return undefined
  return translations.find((tr) => tr.languageCode === currentLocale.value) ?? translations[0]
}

const categoriesQuery = useListCategories({
  page: 1,
  pageSize: 100,
  noPaging: false,
})
const categories = computed<CategoryEntity[]>(() => {
  const items = (categoriesQuery.data?.value as any)?.items ?? []
  return (items as CategoryEntity[]) ?? []
})
const loading = computed(() => categoriesQuery.isLoading.value)
const isError = computed(() => categoriesQuery.isError.value)
</script>

<template>
  <LayoutPageHero
    :title="t('mall.category.title')"
    icon="carbon:category"
    size="sm"
  />

  <LayoutSectionContainer>
    <div v-if="loading" class="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
      <UiCategoryCardSkeleton v-for="_, i in 8" :key="i" />
    </div>

    <UiAppEmpty v-else-if="isError" variant="error">
      <template #action>
        <UiButton variant="outline" size="sm" @click="categoriesQuery.refetch()">
          {{ t('ui.button.retry') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <UiAppEmpty v-else-if="categories.length === 0" variant="noData">
      <template #action>
        <UiButton variant="outline" @click="navigateTo(localePath('/'))">
          {{ t('cart.continueShopping') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <div v-else class="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-4">
      <NuxtLink
        v-for="cat in categories"
        :key="cat.id"
        :to="localePath('/category/' + cat.id)"
        class="group flex flex-col items-center gap-3 rounded-2xl border border-border bg-card p-6 transition-colors hover:border-primary/60"
      >
        <UiImage
          v-if="cat.imageUrl"
          :src="cat.imageUrl"
          :alt="pickTranslation(cat.translations)?.name || ''"
          class="h-16 w-16 rounded-xl object-cover"
        />
        <span
          v-else
          class="flex h-16 w-16 items-center justify-center rounded-xl bg-primary/10 text-primary transition-colors group-hover:bg-primary group-hover:text-primary-foreground"
        >
          <XIcon :icon="categoryIconFor(cat.id)" :size="28" />
        </span>
        <span class="line-clamp-1 text-sm font-medium text-foreground dark:text-slate-200">
          {{ pickTranslation(cat.translations)?.name || '—' }}
        </span>
      </NuxtLink>
    </div>
  </LayoutSectionContainer>
</template>
