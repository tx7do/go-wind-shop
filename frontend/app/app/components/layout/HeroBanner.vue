<script setup lang="ts">
import { cn } from '@/lib/utils'
import { XIcon } from '@/plugins/xicon'

const { t } = useI18n()
const localePath = useLocalePath()

const props = defineProps<{
  class?: string
}>()
</script>

<template>
  <!-- 不再自带 SectionContainer：首页把它放进桌面端栅格里，自带容器会变成
       列里再套一层 max-w-[1200px] + px-8 py-12，把内容挤扁。 -->
  <div
    :class="cn(
      'hero-banner relative overflow-hidden rounded-xl border border-border px-5 py-6 md:px-10 md:py-12',
      props.class,
    )"
  >
    <span class="inline-flex items-center gap-1.5 rounded-full bg-primary-soft px-3 py-1 text-xs font-semibold text-primary-soft-foreground">
      <XIcon icon="carbon:gift" :size="14" />
      {{ t('mall.home.promoBanner.title') }}
    </span>
    <h2 class="mt-3 max-w-xl text-xl font-extrabold leading-tight tracking-tight text-foreground md:text-3xl">
      {{ t('mall.home.promoBanner.subtitle') }}
    </h2>
    <p class="mt-2 max-w-md text-[13px] leading-relaxed text-muted-foreground md:text-sm">
      {{ t('mall.home.heroDescription') }}
    </p>
    <NuxtLink :to="localePath('/category')">
      <UiButton size="lg" class="mt-5">
        {{ t('mall.home.promoBanner.cta') }}
        <XIcon icon="lucide:arrow-right" :size="16" />
      </UiButton>
    </NuxtLink>
  </div>
</template>

