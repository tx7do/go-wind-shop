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
    <UiButton type="button" variant="outline" size="lg" class="w-full gap-3" @click="handleGoogleRegister">
      <XIcon icon="mdi:google" :size="18" />
      <span>{{ t('authentication.login.social_google') }}</span>
    </UiButton>

    <UiButton type="button" variant="outline" size="lg" class="w-full gap-3" @click="handleGithubRegister">
      <XIcon icon="mdi:github" :size="20" />
      <span>{{ t('authentication.login.social_github') }}</span>
    </UiButton>
  </div>
</template>
