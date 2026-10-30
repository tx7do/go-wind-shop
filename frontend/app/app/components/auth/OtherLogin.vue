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
    <UiButton type="button" variant="outline" size="lg" class="w-full gap-3" @click="handleGithubLogin">
      <XIcon icon="mdi:github" :size="20" />
      <span>{{ t('authentication.login.social_github') }}</span>
    </UiButton>

    <UiButton type="button" variant="outline" size="lg" class="w-full gap-3" @click="handleGoogleLogin">
      <XIcon icon="mdi:google" :size="18" />
      <span>{{ t('authentication.login.social_google') }}</span>
    </UiButton>
  </div>
</template>
