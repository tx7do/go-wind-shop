<script setup lang="ts">
import { computed } from 'vue';
import { XIcon } from '@/plugins/xicon';

const props = defineProps<{
  value: number;
  size?: number;
}>();

const starSize = computed(() => props.size ?? 14);
const clamped = computed(() => Math.max(0, Math.min(5, Math.round(props.value))));
</script>

<template>
  <div class="flex items-center gap-0.5" aria-hidden="true">
    <XIcon
      v-for="i in 5"
      :key="i"
      icon="lucide:star"
      :size="starSize"
      :class="i <= clamped ? 'fill-current text-amber-400' : 'text-muted-foreground'"
    />
  </div>
</template>
