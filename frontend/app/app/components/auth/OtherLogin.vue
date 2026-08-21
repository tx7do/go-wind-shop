<script setup lang="ts">
import { XIcon } from '@/plugins/xicon'

const { t } = useI18n()

const handleGoogleLogin = () => {
  // TODO: Google OAuth 登录。需后端补 OAuth 回调端点 + 第三方身份映射。
  console.log('Google 登录')
}

const handleGithubLogin = () => {
  // TODO: GitHub OAuth 登录。需后端补 OAuth 回调端点 + 第三方身份映射。
  console.log('GitHub 登录')
}
</script>

<template>
  <div class="space-y-3">
    <button
      class="flex w-full cursor-pointer items-center justify-center gap-3 rounded-lg border border-border bg-card px-4 py-3 text-sm font-medium text-foreground transition-all hover:border-primary hover:bg-primary/5 active:scale-[0.98]"
      @click="handleGithubLogin"
    >
      <XIcon icon="mdi:github" :size="20" />
      <span>{{ t('authentication.login.social_github') }}</span>
    </button>

    <button
      class="flex w-full cursor-pointer items-center justify-center gap-3 rounded-lg border border-border bg-card px-4 py-3 text-sm font-medium text-foreground transition-all hover:border-primary hover:bg-primary/5 active:scale-[0.98]"
      @click="handleGoogleLogin"
    >
      <XIcon icon="mdi:google" :size="20" />
      <span>{{ t('authentication.login.social_google') }}</span>
    </button>
  </div>
</template>
