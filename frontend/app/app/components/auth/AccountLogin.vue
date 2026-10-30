<script setup lang="ts">
import { useAuthStore } from '@/stores/modules/app/auth.state'
import { XIcon } from '@/plugins/xicon'

const { t } = useI18n()
const localePath = useLocalePath()
const authStore = useAuthStore()

const username = ref('')
const password = ref('')
const rememberMe = ref(false)
const loading = ref(false)
const errorMsg = ref('')
const passwordVisible = ref(false)

const handleLogin = async () => {
  errorMsg.value = ''
  if (!username.value || !password.value) return

  loading.value = true
  try {
    await authStore.authLogin({
      username: username.value,
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
      <UiLabel for="login-username" class="text-foreground">
        {{ t('authentication.login.username') }}
      </UiLabel>
      <UiInput
        id="login-username"
        v-model="username"
        type="text"
        :placeholder="t('authentication.register.input_username')"
        autocomplete="username"
        autocapitalize="none"
        autocorrect="off"
        spellcheck="false"
      />
    </div>

    <div class="space-y-1.5">
      <UiLabel for="login-password" class="text-foreground">
        {{ t('authentication.login.password') }}
      </UiLabel>
      <div class="relative">
        <UiInput
          id="login-password"
          v-model="password"
          :type="passwordVisible ? 'text' : 'password'"
          :placeholder="t('authentication.login.placeholder_password')"
          autocomplete="current-password"
          class="pr-11"
          @keydown.enter="handleLogin"
        />
        <button
          type="button"
          class="absolute inset-y-0 right-0 flex w-11 cursor-pointer items-center justify-center border-none bg-transparent text-muted-foreground transition-colors hover:text-foreground"
          :aria-label="t('authentication.login.password')"
          :aria-pressed="passwordVisible"
          @click="passwordVisible = !passwordVisible"
        >
          <XIcon :icon="passwordVisible ? 'lucide:eye-off' : 'lucide:eye'" :size="16" />
        </button>
      </div>
    </div>

    <div class="flex flex-wrap items-center justify-between gap-2">
      <UiLabel for="login-remember" class="flex cursor-pointer items-center gap-2 font-normal text-muted-foreground">
        <input
          id="login-remember"
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
