<script setup lang="ts">
import { cn } from '@/lib/utils'
import { XIcon } from '@/plugins/xicon'

const { t } = useI18n()
const localePath = useLocalePath()

const props = defineProps<{
  class?: string
}>()

// useId() 在 SSR 与客户端水合时产出一致值，避免 hydration mismatch
const uid = useId()
const gradId = `hero-${uid}`
const gridId = `hero-grid-${uid}`
const radialId = `hero-radial-${uid}`
</script>

<template>
  <LayoutSectionContainer :class="cn('hidden md:block', props.class)" no-padding>
    <div
      :class="cn(
        'relative h-[380px] overflow-hidden rounded-2xl border border-border',
        'flex items-stretch',
      )"
    >
      <!-- 装饰背景层 -->
      <div class="absolute inset-0 pointer-events-none" aria-hidden="true">
        <svg
          width="100%"
          height="100%"
          viewBox="0 0 1600 380"
          preserveAspectRatio="xMidYMid slice"
          :style="{ position: 'absolute', inset: 0 }"
        >
          <defs>
            <linearGradient :id="gradId" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" :stop-color="'hsl(var(--primary))'" stop-opacity="0.28" />
              <stop offset="45%" :stop-color="'hsl(var(--primary))'" stop-opacity="0.10" />
              <stop offset="100%" stop-color="transparent" stop-opacity="0" />
            </linearGradient>
            <pattern :id="gridId" width="48" height="48" patternUnits="userSpaceOnUse" patternTransform="rotate(18)">
              <path d="M0 48 L48 48 M48 0 L48 48" :stroke="'hsl(var(--primary))'" stroke-opacity="0.05" stroke-width="1" fill="none" />
            </pattern>
            <radialGradient :id="radialId" cx="85%" cy="50%" r="60%">
              <stop offset="0%" :stop-color="'hsl(var(--primary))'" stop-opacity="0.18" />
              <stop offset="60%" :stop-color="'hsl(var(--primary))'" stop-opacity="0.04" />
              <stop offset="100%" stop-color="transparent" stop-opacity="0" />
            </radialGradient>
          </defs>

          <rect width="1600" height="380" :fill="`url(#${gradId})`" />
          <rect width="1600" height="380" :fill="`url(#${gridId})`" />
          <rect width="1600" height="380" :fill="`url(#${radialId})`" />

          <!-- 同心圓裝飾組：右側 -->
          <g :stroke="'hsl(var(--primary))'" fill="none" stroke-width="1.2">
            <circle cx="1380" cy="190" r="160" stroke-opacity="0.18" />
            <circle cx="1380" cy="190" r="120" stroke-opacity="0.13" />
            <circle cx="1380" cy="190" r="80" stroke-opacity="0.08" />
            <circle cx="1380" cy="190" r="40" stroke-opacity="0.05" />
            <circle cx="1500" cy="90"  r="70" stroke-opacity="0.10" />
            <circle cx="1500" cy="290" r="70" stroke-opacity="0.10" />
          </g>

          <!-- 射線裝飾：右側放射 -->
          <g :stroke="'hsl(var(--primary))'" stroke-opacity="0.12" fill="none" stroke-width="1">
            <line x1="1380" y1="190" x2="1380" y2="0" />
            <line x1="1380" y1="190" x2="1380" y2="380" />
            <line x1="1380" y1="190" x2="1100" y2="190" />
            <line x1="1380" y1="190" x2="1600" y2="190" />
            <line x1="1380" y1="190" x2="1200" y2="0" />
            <line x1="1380" y1="190" x2="1560" y2="0" />
            <line x1="1380" y1="190" x2="1200" y2="380" />
            <line x1="1380" y1="190" x2="1560" y2="380" />
          </g>

          <!-- 左側光斑 -->
          <g :stroke="'hsl(var(--primary))'" fill="none" stroke-width="1">
            <circle cx="120" cy="60"  r="30" stroke-opacity="0.10" />
            <circle cx="220" cy="340" r="50" stroke-opacity="0.06" />
            <circle cx="60"  cy="300" r="20" stroke-opacity="0.08" />
          </g>
        </svg>
      </div>

      <!-- 文案層（左欄） -->
      <div class="relative z-10 flex flex-1 flex-col justify-center gap-4 p-12">
        <span class="w-fit rounded-full bg-primary/15 px-3 py-1 text-xs font-medium text-primary">
          {{ t('mall.home.promoBanner.title') }}
        </span>
        <h2 class="max-w-lg text-3xl font-extrabold leading-tight text-foreground md:text-4xl">
          {{ t('mall.home.promoBanner.subtitle') }}
        </h2>
        <div class="mt-3">
          <UiButton
            size="lg"
            @click="navigateTo(localePath('/'))"
          >
            {{ t('mall.home.promoBanner.cta') }}
            <XIcon icon="lucide:arrow-right" width="16" height="16" class="ml-1" />
          </UiButton>
        </div>
      </div>

      <!-- 裝飾圖形層（右欄，僅桌面） -->
      <div class="relative z-10 hidden w-[42%] items-center justify-center md:flex" aria-hidden="true">
        <svg width="380" height="380" viewBox="0 0 380 380" class="opacity-70">
          <g :stroke="'hsl(var(--primary))'" fill="none" stroke-width="1.5">
            <circle cx="190" cy="190" r="50" stroke-opacity="0.45" />
            <circle cx="190" cy="190" r="90" stroke-opacity="0.30" />
            <circle cx="190" cy="190" r="130" stroke-opacity="0.18" />
            <circle cx="190" cy="190" r="170" stroke-opacity="0.10" />
          </g>
          <g :stroke="'hsl(var(--primary))'" stroke-opacity="0.2" fill="none" stroke-width="1">
            <line x1="190" y1="0" x2="190" y2="380" />
            <line x1="0" y1="190" x2="380" y2="190" />
            <line x1="55" y1="55" x2="325" y2="325" />
            <line x1="325" y1="55" x2="55" y2="325" />
          </g>
        </svg>
      </div>
    </div>
  </LayoutSectionContainer>
</template>
