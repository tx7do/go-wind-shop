<script setup lang="ts">
import { Toaster } from "vue-sonner"
import "vue-sonner/style.css"
import { usePreferences } from "@/core/preferences/use-preferences"

// 同步当前主题给 vue-sonner：rich-colors 模式下 success/error toast 的内置配色
// 需要明确 theme，否则暗色背景下默认浅色样式对比度不足
const { isDark } = usePreferences()
const toasterTheme = computed(() => (isDark.value ? "dark" : "light"))
</script>

<template>
  <Toaster
    :theme="toasterTheme"
    rich-colors
    position="top-right"
    :toast-options="{
      classes: {
        toast: 'group toast group-[.toaster]:bg-background group-[.toaster]:text-foreground group-[.toaster]:border-border group-[.toaster]:shadow-lg',
        description: 'group-[.toast]:text-muted-foreground',
        actionButton: 'group-[.toast]:bg-primary group-[.toast]:text-primary-foreground',
        cancelButton: 'group-[.toast]:bg-muted group-[.toast]:text-muted-foreground',
      },
    }"
  />
</template>
