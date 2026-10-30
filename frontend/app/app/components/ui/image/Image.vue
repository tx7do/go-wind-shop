<script setup lang="ts">
import { useMediaQuery } from '@vueuse/core'

import { usePreferences } from '@/core/preferences/use-preferences'

const props = withDefaults(defineProps<{
  src?: string
  alt?: string
  fallback?: string
}>(), {
  alt: '',
})

/**
 * 兜底占位图跟随应用主题。
 * SVG 作为 <img> 资源加载时读不到页面上的 .dark class 和 CSS 变量，
 * prefers-color-scheme 媒体查询也只跟系统、不跟手动切换，
 * 所以准备亮/暗两份文件，在 JS 里按 theme.mode 挑选；
 * auto 模式经 useMediaQuery 响应系统切换。
 * 传入自定义 fallback 时始终使用它，不做主题区分。
 */
const { themePreferences } = usePreferences()
const systemDark = useMediaQuery('(prefers-color-scheme: dark)')
const resolvedFallback = computed(() => {
  if (props.fallback) return props.fallback
  const { mode } = themePreferences.value
  const dark = mode === 'dark' || (mode === 'auto' && systemDark.value)
  return dark ? '/placeholder-dark.svg' : '/placeholder.svg'
})

const errored = ref(false)
const imgSrc = computed(() =>
  !props.src || errored.value ? resolvedFallback.value : props.src,
)

function handleError() {
  errored.value = true
}

watch(() => props.src, () => {
  errored.value = false
})
</script>

<template>
  <img :src="imgSrc" :alt="alt" class="ui-image" @error="handleError" />
</template>
