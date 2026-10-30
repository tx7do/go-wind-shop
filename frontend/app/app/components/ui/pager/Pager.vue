<script setup lang="ts">
import { cn } from '@/lib/utils'

/**
 * 上一页/下一页分页条。订单、退款、发票、收藏、分类、搜索六处是同一份三段式模板
 * 加同一份 goToPage 夹取函数，这里收成一处（夹取逻辑在组件内，页面只报页码）。
 */
const props = defineProps<{
  page: number
  totalPages: number
  class?: string
}>()

const emit = defineEmits<{ 'update:page': [number] }>()

const { t } = useI18n()

const visible = computed(() => props.totalPages > 1)

function go(next: number) {
  const clamped = Math.min(Math.max(next, 1), props.totalPages)
  if (clamped !== props.page) emit('update:page', clamped)
}
</script>

<template>
  <nav
    v-if="visible"
    :class="cn('flex items-center justify-center gap-4 py-2', props.class)"
    aria-label="pagination"
  >
    <UiButton variant="outline" size="sm" :disabled="props.page <= 1" @click="go(props.page - 1)">
      {{ t('ui.pagination.previous') }}
    </UiButton>
    <span class="text-xs tabular-nums text-muted-foreground">
      {{ t('ui.pagination.page', { current: props.page, total: props.totalPages }) }}
    </span>
    <UiButton variant="outline" size="sm" :disabled="props.page >= props.totalPages" @click="go(props.page + 1)">
      {{ t('ui.pagination.next') }}
    </UiButton>
  </nav>
</template>
