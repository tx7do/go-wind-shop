<script setup lang="ts">
import { useRegister } from '@/api/composables/auth'
import { XIcon } from '@/plugins/xicon'

const { t } = useI18n()
const localePath = useLocalePath()
const registerMutation = useRegister()

const username = ref('')
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const errorMsg = ref('')
const successMsg = ref('')
const passwordVisible = ref(false)
const confirmVisible = ref(false)

const isValidUsername = computed(() => {
  if (!username.value) return false
  return /^[a-zA-Z0-9_]{3,20}$/.test(username.value)
})

const isValidPassword = computed(() => password.value.length >= 6)

const isPasswordMatch = computed(() =>
  password.value === confirmPassword.value && confirmPassword.value.length > 0
)

const isFormValid = computed(() =>
  isValidUsername.value && isValidPassword.value && isPasswordMatch.value
)

async function handleRegister() {
  errorMsg.value = ''
  successMsg.value = ''
  if (!isFormValid.value) return

  loading.value = true
  try {
    await registerMutation.mutateAsync({
      username: username.value,
      password: password.value,
    })
    successMsg.value = t('authentication.register.success')
    setTimeout(() => {
      navigateTo(localePath('/login'))
    }, 1500)
  } catch (e: any) {
    errorMsg.value = e?.message || t('authentication.register.register_failed')
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
    <div
      v-if="successMsg"
      role="status"
      class="rounded-lg bg-success-soft px-3.5 py-2.5 text-sm text-success-soft-foreground"
    >
      {{ successMsg }}
    </div>

    <div class="space-y-1.5">
      <UiLabel for="register-account-username" class="text-foreground">
        {{ t('authentication.register.username') }}
      </UiLabel>
      <UiInput
        id="register-account-username"
        v-model="username"
        type="text"
        :placeholder="t('authentication.register.input_username')"
        :aria-invalid="Boolean(username && !isValidUsername)"
        :class="username && !isValidUsername ? 'border-destructive' : ''"
        autocomplete="username"
        autocapitalize="none"
        autocorrect="off"
        spellcheck="false"
      />
      <span v-if="username && !isValidUsername" class="block text-xs text-destructive">
        {{ t('authentication.register.invalid_username') }}
      </span>
    </div>

    <div class="space-y-1.5">
      <UiLabel for="register-account-password" class="text-foreground">
        {{ t('authentication.register.password') }}
      </UiLabel>
      <div class="relative">
        <UiInput
          id="register-account-password"
          v-model="password"
          :type="passwordVisible ? 'text' : 'password'"
          :placeholder="t('authentication.register.input_password')"
          :aria-invalid="Boolean(password && !isValidPassword)"
          :class="[password && !isValidPassword ? 'border-destructive' : '', 'pr-11']"
          autocomplete="new-password"
        />
        <button
          type="button"
          :class="eyeBtn"
          :aria-pressed="passwordVisible"
          :aria-label="t('authentication.register.password')"
          @click="passwordVisible = !passwordVisible"
        >
          <XIcon :icon="passwordVisible ? 'lucide:eye-off' : 'lucide:eye'" :size="16" />
        </button>
      </div>
      <span v-if="password && !isValidPassword" class="block text-xs text-destructive">
        {{ t('authentication.register.invalid_password') }}
      </span>
    </div>

    <div class="space-y-1.5">
      <UiLabel for="register-account-confirm-password" class="text-foreground">
        {{ t('authentication.register.confirm_password') }}
      </UiLabel>
      <div class="relative">
        <UiInput
          id="register-account-confirm-password"
          v-model="confirmPassword"
          :type="confirmVisible ? 'text' : 'password'"
          :placeholder="t('authentication.register.input_confirm_password')"
          :aria-invalid="Boolean(confirmPassword && !isPasswordMatch)"
          :class="[confirmPassword && !isPasswordMatch ? 'border-destructive' : '', 'pr-11']"
          autocomplete="new-password"
        />
        <button
          type="button"
          :class="eyeBtn"
          :aria-pressed="confirmVisible"
          :aria-label="t('authentication.register.confirm_password')"
          @click="confirmVisible = !confirmVisible"
        >
          <XIcon :icon="confirmVisible ? 'lucide:eye-off' : 'lucide:eye'" :size="16" />
        </button>
      </div>
      <span v-if="confirmPassword && !isPasswordMatch" class="block text-xs text-destructive">
        {{ t('authentication.register.password_not_match') }}
      </span>
    </div>

    <UiButton
      type="button"
      size="lg"
      class="w-full"
      :disabled="!isFormValid || loading"
      @click="handleRegister"
    >
      <UiSpinner v-if="loading" size="sm" class="text-primary-foreground" />
      {{ loading ? t('authentication.register.registering') : t('authentication.register.register') }}
    </UiButton>
  </div>
</template>
