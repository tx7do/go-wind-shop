<script setup lang="ts">
import { XIcon } from '@/plugins/xicon'

const { t } = useI18n()

const handleGoogleRegister = () => {
  // TODO: Google OAuth 注册。需后端补 OAuth 回调端点 + 第三方身份映射。
  console.log('Google 注册')
}

const handleGithubRegister = () => {
  // TODO: GitHub OAuth 注册。需后端补 OAuth 回调端点 + 第三方身份映射。
  console.log('GitHub 注册')
}
</script>

<template>
  <div class="space-y-3">
    <button
      type="button"
      class="flex w-full cursor-pointer items-center justify-center gap-3 rounded-lg border border-border bg-card px-4 py-3 text-sm font-medium text-foreground transition-all hover:border-primary hover:bg-primary/5 active:scale-[0.98]"
      @click="handleGoogleRegister"
    >
      <XIcon icon="logos:google-icon" :size="18" />
      {{ t('authentication.login.social_google') }}
    </button>

    <button
      type="button"
      class="flex w-full cursor-pointer items-center justify-center gap-3 rounded-lg border border-border bg-card px-4 py-3 text-sm font-medium text-foreground transition-all hover:border-primary hover:bg-primary/5 active:scale-[0.98]"
      @click="handleGithubRegister"
    >
      <XIcon icon="logos:github-icon" :size="18" />
      {{ t('authentication.login.social_github') }}
    </button>
  </div>
</template>
