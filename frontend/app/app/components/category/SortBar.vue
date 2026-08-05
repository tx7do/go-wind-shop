<script setup lang="ts">
import { cn } from '@/lib/utils'

const { t } = useI18n()

const props = defineProps<{
  modelValue: 'featured' | 'latest'
  class?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: 'featured' | 'latest'): void
}>()

type SortKey = 'featured' | 'latest'

const options: Array<{ key: SortKey; labelKey: string }> = [
  { key: 'featured', labelKey: 'mall.category.sortFeatured' },
  { key: 'latest', labelKey: 'mall.category.sortLatest' },
]
</script>

<template>
  <div
    :class="cn(
      'flex items-center gap-1 rounded-xl border border-border bg-card p-1',
      props.class,
    )"
  >
    <span class="px-3 text-xs font-medium text-muted-foreground">
      {{ t('mall.category.sortBy') }}
    </span>
    <button
      v-for="opt in options"
      :key="opt.key"
      type="button"
      :class="cn(
        'rounded-lg px-3 py-1.5 text-sm font-medium transition-colors',
        props.modelValue === opt.key
          ? 'bg-primary/10 text-primary'
          : 'text-muted-foreground hover:bg-muted/50 hover:text-foreground',
      )"
      @click="emit('update:modelValue', opt.key)"
    >
      {{ t(opt.labelKey) }}
    </button>
  </div>
</template>
