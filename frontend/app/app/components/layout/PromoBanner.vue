<script setup lang="ts">
import { cn } from '@/lib/utils'
import { XIcon } from '@/plugins/xicon'

const { t } = useI18n()
const localePath = useLocalePath()

const props = defineProps<{
  class?: string
}>()

// useId() 在 SSR 与客户端水合时产出一致值，避免 hydration mismatch
const gradId = `promo-${useId()}`
</script>

<template>
  <LayoutSectionContainer :class="props.class" no-padding>
    <div
      :class="cn(
        'relative overflow-hidden rounded-2xl border border-border',
        'flex flex-col md:flex-row md:items-stretch',
        'min-h-[180px] md:min-h-[280px]',
      )"
    >
      <!-- 装饰背景层 -->
      <div class="absolute inset-0 pointer-events-none" aria-hidden="true">
        <svg
          width="100%"
          height="100%"
          viewBox="0 0 800 280"
          preserveAspectRatio="xMidYMid slice"
          :style="{ position: 'absolute', inset: 0 }"
        >
          <defs>
            <linearGradient :id="`${gradId}-bg`" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" :stop-color="'hsl(var(--primary))'" stop-opacity="0.22" />
              <stop offset="50%" :stop-color="'hsl(var(--primary))'" stop-opacity="0.06" />
              <stop offset="100%" stop-color="transparent" stop-opacity="0" />
            </linearGradient>
            <pattern :id="`${gradId}-grid`" width="40" height="40" patternUnits="userSpaceOnUse" patternTransform="rotate(15)">
              <path d="M0 40 L40 40 M40 0 L40 40" :stroke="'hsl(var(--primary))'" stroke-opacity="0.04" stroke-width="1" fill="none" />
            </pattern>
          </defs>
          <rect width="800" height="280" :fill="`url(#${gradId}-bg)`" />
          <rect width="800" height="280" :fill="`url(#${gradId}-grid)`" />
          <g :stroke="'hsl(var(--primary))'" stroke-opacity="0.10" fill="none" stroke-width="1">
            <circle cx="680" cy="140" r="90" />
            <circle cx="680" cy="140" r="70" />
            <circle cx="680" cy="140" r="50" />
            <circle cx="750" cy="140" r="60" />
          </g>
        </svg>
      </div>

      <!-- 文案层（左栏） -->
      <div class="relative z-10 flex flex-1 flex-col justify-center gap-3 p-8 md:p-12">
        <span class="w-fit rounded-full bg-primary/15 px-3 py-1 text-xs font-medium text-primary">
          {{ t('mall.home.promoBanner.title') }}
        </span>
        <h2 class="max-w-md text-2xl font-bold leading-tight text-foreground md:text-3xl">
          {{ t('mall.home.promoBanner.subtitle') }}
        </h2>
        <div class="mt-2">
          <UiButton
            size="lg"
            @click="navigateTo(localePath('/'))"
          >
            {{ t('mall.home.promoBanner.cta') }}
            <XIcon icon="lucide:arrow-right" width="16" height="16" class="ml-1" />
          </UiButton>
        </div>
      </div>

      <!-- 装饰图形层（右栏，仅桌面） -->
      <div class="relative z-10 hidden w-[42%] items-center justify-center md:flex" aria-hidden="true">
        <svg width="320" height="320" viewBox="0 0 320 320" class="opacity-70">
          <g :stroke="'hsl(var(--primary))'" fill="none" stroke-width="1.5">
            <circle cx="160" cy="160" r="40" stroke-opacity="0.4" />
            <circle cx="160" cy="160" r="80" stroke-opacity="0.25" />
            <circle cx="160" cy="160" r="120" stroke-opacity="0.15" />
            <circle cx="160" cy="160" r="150" stroke-opacity="0.08" />
          </g>
          <g :stroke="'hsl(var(--primary))'" stroke-opacity="0.2" fill="none" stroke-width="1">
            <line x1="160" y1="0" x2="160" y2="320" />
            <line x1="0" y1="160" x2="320" y2="160" />
          </g>
        </svg>
      </div>
    </div>
  </LayoutSectionContainer>
</template>
