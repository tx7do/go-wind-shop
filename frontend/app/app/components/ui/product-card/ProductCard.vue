<script setup lang="ts">
import { cn } from '@/lib/utils'

/**
 * 商品卡。原来首页 / 分类页 / 搜索页各有一份分叉副本
 * （shadow-sm、dark:shadow-none、key 取 productId 还是 id 都不一样），这里统一：
 * 图 → 两行标题 → 价格 → 活动角标。
 *
 * 标题两行占位用 min-h 固定，价格才不会因为标题一行的卡片错位。
 */
const props = withDefaults(defineProps<{
  productId?: number
  /** 覆盖默认跳转（默认跳 /product/:id） */
  to?: string
  name?: string
  image?: string
  /** 单位「分」，来自 useProductPrices；无价格时不渲染价格行 */
  priceFen?: number | string | null
  oldPriceFen?: number | string | null
  badgeLabel?: string
  badgeTone?: 'neutral' | 'primary' | 'success' | 'warning' | 'danger'
  class?: string
}>(), {
  badgeTone: 'primary',
})

const localePath = useLocalePath()

const href = computed(() =>
  props.to ?? (props.productId !== undefined ? localePath(`/product/${props.productId}`) : undefined),
)
</script>

<template>
  <NuxtLink
    v-if="href"
    :to="href"
    :class="cn(
      'group flex flex-col overflow-hidden rounded-lg border border-border bg-card shadow-card transition-colors hover:border-primary/60',
      props.class,
    )"
  >
    <UiImage
      :src="props.image"
      :alt="props.name || ''"
      class="aspect-square w-full shrink-0 rounded-none object-cover"
    />
    <div class="flex flex-1 flex-col gap-1.5 p-3">
      <h3 class="line-clamp-2 min-h-[2.7em] text-[13px] font-semibold leading-[1.35] text-foreground">
        {{ props.name || '—' }}
      </h3>
      <div class="mt-auto flex flex-wrap items-center gap-x-2 gap-y-1">
        <UiPriceTag :amount="props.priceFen" :old-amount="props.oldPriceFen" />
        <UiStatusBadge
          v-if="props.badgeLabel"
          :label="props.badgeLabel"
          :tone="props.badgeTone"
          size="sm"
        />
      </div>
    </div>
  </NuxtLink>
</template>
