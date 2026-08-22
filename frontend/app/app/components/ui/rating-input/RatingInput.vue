<script setup lang="ts">
import { ref, computed } from 'vue';
import { XIcon } from '@/plugins/xicon';

const props = defineProps<{
  modelValue: number;
}>();
const emit = defineEmits<{
  (e: 'update:modelValue', value: number): void;
}>();

const hovered = ref(0);

const active = computed(() => hovered.value || props.modelValue);

function select(n: number) {
  emit('update:modelValue', n);
}
</script>

<template>
  <div
    class="flex items-center gap-1"
    @mouseleave="hovered = 0"
  >
    <button
      v-for="i in 5"
      :key="i"
      type="button"
      :aria-label="`Rating ${i} of 5`"
      class="p-0.5 transition-colors"
      @mouseenter="hovered = i"
      @click="select(i)"
    >
      <XIcon
        icon="lucide:star"
        :size="20"
        :class="i <= active ? 'fill-current text-amber-400' : 'text-muted-foreground'"
      />
    </button>
  </div>
</template>
