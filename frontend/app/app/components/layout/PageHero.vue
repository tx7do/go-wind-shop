<script setup lang="ts">
import { XIcon } from '@/plugins/xicon'
import { cn } from '@/lib/utils'

/**
 * 页面抬头。
 *
 * 原来是一座 280–380px 高、居中、叠了两层斜纹的巨型 hero——在商城里任何一页都
 * 比内容本身还占地方。收成一条带标题的浅渐变带：内容左对齐、随 SectionContainer
 * 同宽、只占该占的高度。
 */
const props = defineProps<{
  title: string
  subtitle?: string
  description?: string
  icon?: string
  size?: 'sm' | 'md' | 'lg'
  class?: string
}>()

const padding = computed(() =>
  props.size === 'sm' ? 'py-5' : props.size === 'lg' ? 'py-9' : 'py-7',
)
const titleClass = computed(() =>
  props.size === 'lg'
    ? 'text-2xl md:text-3xl'
    : props.size === 'sm'
      ? 'text-lg md:text-xl'
      : 'text-xl md:text-2xl',
)
</script>

<template>
  <section
    :class="cn(
      'hero-band w-full border-b border-border',
      padding,
      props.class,
    )"
  >
    <div class="mx-auto flex w-full max-w-(--layout-max-width) items-start gap-4 px-4 md:px-8">
      <span
        v-if="icon"
        class="mt-0.5 flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-primary-soft text-primary-soft-foreground"
      >
        <XIcon :icon="icon" :size="22" />
      </span>

      <div class="min-w-0 flex-1">
        <p
          v-if="subtitle"
          class="text-xs font-semibold uppercase tracking-wider text-primary"
        >
          {{ subtitle }}
        </p>

        <h1 :class="cn('font-bold leading-tight text-foreground', titleClass)">
          {{ title }}
        </h1>

        <p
          v-if="description"
          class="mt-1 max-w-2xl text-[13px] leading-relaxed text-muted-foreground md:text-sm"
        >
          {{ description }}
        </p>

        <div v-if="$slots.default" class="mt-4 flex flex-wrap items-center gap-3">
          <slot />
        </div>
      </div>
    </div>
  </section>
</template>
