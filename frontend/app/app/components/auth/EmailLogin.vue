<script setup lang="ts">
import { useAuthStore } from '@/stores/modules/app/auth.state'
import { XIcon } from '@/plugins/xicon'

const { t } = useI18n()
const localePath = useLocalePath()
const authStore = useAuthStore()

const email = ref('')
const password = ref('')
const rememberMe = ref(false)
const loading = ref(false)
const errorMsg = ref('')
const passwordVisible = ref(false)

const handleLogin = async () => {
  errorMsg.value = ''
  if (!email.value || !password.value) return

  loading.value = true
  try {
    await authStore.authLogin({
      email: email.value,
      password: password.value,
    }, async () => {
      await navigateTo(localePath('/'))
    })
  } catch (e: any) {
    errorMsg.value = e?.message || t('authentication.login.login_failed')
  } finally {
    loading.value = false
  }
}

const eyeBtn = 'absolute inset-y-0 right-0 flex w-11 cursor-pointer items-center justify-center border-none bg-transparent text-muted-foreground transition-colors hover:text-foreground'
</script>

<template>
  <div class="space-y-4">
    <div
      v-if="errorMsg"
      role="alert"
      class="rounded-lg bg-destructive-soft px-3.5 py-2.5 text-sm text-destructive-soft-foreground"
    >
      {{ errorMsg }}
    </div>

    <div class="space-y-1.5">
      <UiLabel for="login-email" class="text-foreground">
        {{ t('authentication.register.email') }}
      </UiLabel>
      <UiInput
        id="login-email"
        v-model="email"
        type="email"
        :placeholder="t('authentication.login.placeholder_email')"
        autocomplete="email"
        autocapitalize="none"
        autocorrect="off"
        spellcheck="false"
      />
    </div>

    <div class="space-y-1.5">
      <UiLabel for="login-password-email" class="text-foreground">
        {{ t('authentication.login.password') }}
      </UiLabel>
      <div class="relative">
        <UiInput
          id="login-password-email"
          v-model="password"
          :type="passwordVisible ? 'text' : 'password'"
          :placeholder="t('authentication.login.placeholder_password')"
          autocomplete="current-password"
          class="pr-11"
          @keydown.enter="handleLogin"
        />
        <button
          type="button"
          :class="eyeBtn"
          :aria-pressed="passwordVisible"
          :aria-label="t('authentication.login.password')"
          @click="passwordVisible = !passwordVisible"
        >
          <XIcon :icon="passwordVisible ? 'lucide:eye-off' : 'lucide:eye'" :size="16" />
        </button>
      </div>
    </div>

    <div class="flex flex-wrap items-center justify-between gap-2">
      <UiLabel for="login-email-remember" class="flex cursor-pointer items-center gap-2 font-normal text-muted-foreground">
        <input
          id="login-email-remember"
          v-model="rememberMe"
          type="checkbox"
          class="h-4 w-4 rounded border-border accent-primary"
        />
        <span>{{ t('authentication.login.remember_me') }}</span>
      </UiLabel>
      <UiButton variant="link" size="sm" class="h-auto p-0" @click="navigateTo(localePath('/forgot-password'))">
        {{ t('authentication.login.forgot_password') }}
      </UiButton>
    </div>

    <UiButton type="button" size="lg" class="w-full" :disabled="loading" @click="handleLogin">
      <UiSpinner v-if="loading" size="sm" class="text-primary-foreground" />
      {{ loading ? t('authentication.login.logging_in') : t('authentication.login.login') }}
    </UiButton>
  </div>
</template>
