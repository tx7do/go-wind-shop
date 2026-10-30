<script setup lang="ts">
const props = defineProps<{
  email: string
}>()

const { t } = useI18n()

const code = ref('')
const isCodeComplete = ref(false)
const codeSent = ref(true)
const codeCountdown = ref(60)

let timer: ReturnType<typeof setInterval> | null = null

function startCountdown() {
  codeSent.value = true
  codeCountdown.value = 60

  timer = setInterval(() => {
    codeCountdown.value--
    if (codeCountdown.value <= 0) {
      if (timer) clearInterval(timer)
      codeSent.value = false
    }
  }, 1000)
}

onMounted(() => {
  startCountdown()
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

function handleComplete(value: string) {
  if (value.length === 6) {
    isCodeComplete.value = true
  }
}

function handleResend() {
  if (codeSent.value) return
  startCountdown()
}

function handleConfirm() {
  if (!isCodeComplete.value) return
  // TODO: 邮箱验证码注册流程。后端 AuthenticationService.RegisterUser 目前不接受/校验
  // 邮箱验证码（仅 username+password 创建账号），故邮箱验证注册链路尚未打通。
  // 待后端补齐邮箱验证码校验能力后，此处改调 useRegister 并附带 email + 已校验的 code。
  console.log('确认注册')
}

const resendBtn = 'cursor-pointer border-none bg-transparent p-0 text-sm font-medium text-primary transition-colors underline-offset-4 hover:underline disabled:cursor-not-allowed disabled:text-muted-foreground disabled:no-underline'
</script>

<template>
  <div class="space-y-4">
    <div class="rounded-lg bg-primary-soft px-4 py-3.5 text-center">
      <p class="text-sm font-medium text-foreground">{{ t('authentication.register.code_sent_title') }}</p>
      <p class="mt-1 break-all text-base font-semibold text-primary-soft-foreground">{{ props.email }}</p>
      <p class="mt-1 text-xs text-muted-foreground">{{ t('authentication.register.code_sent_subtitle') }}</p>
    </div>

    <div class="space-y-1.5">
      <UiLabel for="register-email-code" class="block text-center text-foreground">
        {{ t('authentication.register.code') }}
      </UiLabel>
      <div class="flex justify-center">
        <UiOtpInput
          id="register-email-code"
          v-model="code"
          :length="6"
          :disabled="!codeSent"
          type="number"
          @complete="handleComplete"
        />
      </div>
    </div>

    <div class="text-center">
      <button type="button" :class="resendBtn" :disabled="codeSent" @click="handleResend">
        {{ codeSent ? `${codeCountdown}s ${t('authentication.register.resend_after')}` : t('authentication.register.resend') }}
      </button>
    </div>

    <UiButton type="button" size="lg" class="w-full" :disabled="!isCodeComplete" @click="handleConfirm">
      {{ t('authentication.register.confirm') }}
    </UiButton>
  </div>
</template>
