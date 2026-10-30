<script setup lang="ts">
import { XIcon } from '@/plugins/xicon'

/**
 * 未登录闸门。原来 10 个页面各抄一份同样的 locked 图标 + 登录按钮，
 * user.vue 那一份还漏了外框。统一到这里。
 */
const props = defineProps<{
  description?: string
}>()

const { t } = useI18n()
const localePath = useLocalePath()
</script>

<template>
  <div class="flex flex-col items-center gap-5 rounded-lg border border-border bg-card p-10 text-center shadow-card">
    <XIcon icon="carbon:locked" :size="40" class="text-muted-foreground" />
    <p class="text-sm text-muted-foreground">
      {{ props.description || t('authentication.login.please_login') }}
    </p>
    <UiButton @click="navigateTo(localePath('/login'))">
      {{ t('navbar.user.login') }}
    </UiButton>
  </div>
</template>
