<script setup lang="ts">
defineProps<{
  icon: string
  title: string
  description: string
  items: string[]
}>()
</script>

<template>
  <div class="w-full">
    <LayoutPageHero
      :title="title"
      :description="description"
      :icon="icon"
      size="sm"
    />

    <LayoutSectionContainer max-width="narrow">
      <article class="rounded-xl border border-border bg-card p-6 shadow-card md:p-10">
        <ol class="space-y-5 md:space-y-6">
          <li
            v-for="(item, index) in items"
            :key="index"
            class="flex items-start gap-3.5 md:gap-4"
          >
            <span class="mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-primary-soft font-mono text-xs font-semibold tabular-nums text-primary-soft-foreground">
              {{ String(index + 1).padStart(2, '0') }}
            </span>
            <p class="min-w-0 max-w-prose flex-1 text-[15px] leading-7 text-foreground">
              {{ item }}
            </p>
          </li>
        </ol>
      </article>
    </LayoutSectionContainer>
  </div>
</template>
