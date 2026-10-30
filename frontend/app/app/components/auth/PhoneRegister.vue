<script setup lang="ts">
const { t } = useI18n()

const phone = ref('')
const verificationCode = ref('')
const codeSent = ref(false)
const codeCountdown = ref(0)

let codeTimer: ReturnType<typeof setInterval> | null = null

watch([codeSent, codeCountdown], () => {
  if (codeSent.value && codeCountdown.value > 0) {
    codeTimer = setInterval(() => {
      codeCountdown.value--
      if (codeCountdown.value <= 0) {
        codeSent.value = false
        if (codeTimer) clearInterval(codeTimer)
      }
    }, 1000)
  }
})

onUnmounted(() => {
  if (codeTimer) clearInterval(codeTimer)
})

function handleSendCode() {
  if (!phone.value) return
  codeSent.value = true
  codeCountdown.value = 60
}

function handleRegister() {
  if (!phone.value || !verificationCode.value) return
  // TODO: 手机号+短信验证码注册。需后端 BFF 补 SMS 验证码下发/校验端点，
  // 并接入短信网关。当前 app BFF 无对应端点。
  console.log('手机号注册/登录:', { phone: phone.value, verificationCode: verificationCode.value })
}
</script>

<template>
  <div class="space-y-4">
    <div class="space-y-1.5">
      <UiLabel for="register-phone-number" class="text-foreground">
        {{ t('authentication.register.phone') }}
      </UiLabel>
      <UiInput
        id="register-phone-number"
        v-model="phone"
        type="tel"
        :placeholder="t('authentication.register.input_phone')"
        autocomplete="tel"
        autocapitalize="none"
        autocorrect="off"
        spellcheck="false"
      />
    </div>

    <div class="space-y-1.5">
      <UiLabel for="register-phone-code" class="text-foreground">
        {{ t('authentication.register.code') }}
      </UiLabel>
      <div class="flex gap-2">
        <UiInput
          id="register-phone-code"
          v-model="verificationCode"
          type="text"
          inputmode="numeric"
          class="min-w-0 flex-1"
          :placeholder="t('authentication.register.input_code')"
          maxlength="6"
          autocomplete="one-time-code"
          autocapitalize="none"
          autocorrect="off"
          spellcheck="false"
        />
        <UiButton
          type="button"
          variant="outline"
          class="shrink-0 px-4"
          :disabled="codeSent"
          @click="handleSendCode"
        >
          {{ codeSent ? `${codeCountdown}s` : t('authentication.register.send_code') }}
        </UiButton>
      </div>
    </div>

    <UiButton type="button" size="lg" class="w-full" @click="handleRegister">
      {{ t('authentication.register.register_or_login') }}
    </UiButton>
  </div>
</template>
