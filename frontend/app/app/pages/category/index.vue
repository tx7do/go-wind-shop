<script setup lang="ts">
import { XIcon } from '@/plugins/xicon'
import { useListCategories } from '@/api/composables'
import { categoryIconFor } from '@/utils/category-icon'
import { translationName } from '@/utils/translation'

const { t } = useI18n()
const localePath = useLocalePath()

useHead({ title: t('mall.category.title') })

type CategoryEntity = {
  id?: number
  imageUrl?: string
  translations?: Array<{ name?: string; languageCode?: string }>
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
    <UiListSkeleton v-if="loading" variant="card" :rows="8" :columns="4" />

    <UiAppEmpty v-else-if="isError" variant="error">
      <template #action>
        <UiButton variant="outline" size="sm" @click="categoriesQuery.refetch()">
          {{ t('ui.button.retry') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <UiAppEmpty v-else-if="categories.length === 0" :description="t('common.noData')">
      <template #action>
        <UiButton variant="outline" @click="navigateTo(localePath('/'))">
          {{ t('cart.continueShopping') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <div v-else class="grid grid-cols-2 gap-3 md:grid-cols-4 md:gap-4">
      <NuxtLink
        v-for="cat in categories"
        :key="cat.id"
        :to="localePath('/category/' + cat.id)"
        class="group flex flex-col items-center gap-3 rounded-lg border border-border bg-card px-3 py-5 shadow-card transition-colors hover:border-primary/60"
      >
        <span class="flex size-14 shrink-0 items-center justify-center overflow-hidden rounded-xl bg-primary-soft text-primary-soft-foreground transition-colors group-hover:bg-primary group-hover:text-primary-foreground">
          <UiImage
            v-if="cat.imageUrl"
            :src="cat.imageUrl"
            :alt="translationName(cat.translations)"
            class="size-14 object-cover"
          />
          <XIcon v-else :icon="categoryIconFor(cat.id)" :size="26" />
        </span>
        <span class="line-clamp-1 w-full text-center text-sm font-medium text-foreground">
          {{ translationName(cat.translations) }}
        </span>
      </NuxtLink>
    </div>
  </LayoutSectionContainer>
</template>
