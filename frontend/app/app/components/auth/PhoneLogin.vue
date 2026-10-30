<script setup lang="ts">
const { t } = useI18n()

const phone = ref('')
const verificationCode = ref('')
const codeSent = ref(false)
const countdown = ref(0)

let timer: ReturnType<typeof setTimeout> | null = null

watch(countdown, (val) => {
  if (val > 0) {
    timer = setTimeout(() => countdown.value--, 1000)
  } else if (codeSent.value) {
    codeSent.value = false
  }
})

onUnmounted(() => {
  if (timer) clearTimeout(timer)
})

function handleSendCode() {
  if (!phone.value) return
  codeSent.value = true
  countdown.value = 60
}

function handleLogin() {
  if (!phone.value || !verificationCode.value) return
  // TODO: 手机号+短信验证码登录。需后端 BFF 补 SMS 验证码下发/校验端点，
  // 并接入短信网关。当前 app BFF 无对应端点。
  console.log('登录信息:', { phone: phone.value, verificationCode: verificationCode.value })
}
</script>

<template>
  <div class="space-y-4">
    <div class="space-y-1.5">
      <UiLabel for="login-phone" class="text-foreground">
        {{ t('authentication.register.phone') }}
      </UiLabel>
      <UiInput
        id="login-phone"
        v-model="phone"
        type="tel"
        :placeholder="t('authentication.login.placeholder_phone')"
        autocomplete="tel"
        autocapitalize="none"
        autocorrect="off"
        spellcheck="false"
      />
    </div>

    <div class="space-y-1.5">
      <UiLabel for="login-code" class="text-foreground">
        {{ t('authentication.register.code') }}
      </UiLabel>
      <div class="flex gap-2">
        <UiInput
          id="login-code"
          v-model="verificationCode"
          type="text"
          inputmode="numeric"
          class="min-w-0 flex-1"
          :placeholder="t('authentication.login.placeholder_code')"
          maxlength="6"
          autocomplete="one-time-code"
          autocapitalize="none"
          autocorrect="off"
          spellcheck="false"
          @keydown.enter="handleLogin"
        />
        <UiButton
          type="button"
          variant="outline"
          class="shrink-0 px-4"
          :disabled="codeSent"
          @click="handleSendCode"
        >
          {{ codeSent ? `${countdown}s` : t('authentication.register.send_code') }}
        </UiButton>
      </div>
    </div>

    <UiButton type="button" size="lg" class="w-full" @click="handleLogin">
      {{ t('authentication.login.login') }}
    </UiButton>
  </div>
</template>
