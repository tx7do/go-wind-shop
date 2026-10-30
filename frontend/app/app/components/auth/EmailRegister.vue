<script setup lang="ts">
const { t } = useI18n()

const email = ref('')
const visibleEnter = ref(false)

const isValidEmail = computed(() => {
  if (!email.value) return false
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.value)
})

function handleNext() {
  if (!isValidEmail.value) return
  visibleEnter.value = true
}
</script>

<template>
  <div>
    <div v-if="!visibleEnter" class="space-y-4">
      <div class="space-y-1.5">
        <UiLabel for="register-email-address" class="text-foreground">
          {{ t('authentication.register.email') }}
        </UiLabel>
        <UiInput
          id="register-email-address"
          v-model="email"
          type="text"
          :placeholder="t('authentication.register.input_email')"
          :aria-invalid="Boolean(email && !isValidEmail)"
          :class="email && !isValidEmail ? 'border-destructive' : ''"
          autocomplete="email"
          autocapitalize="none"
          autocorrect="off"
          spellcheck="false"
          @keydown.enter="handleNext"
        />
        <span v-if="email && !isValidEmail" class="block text-xs text-destructive">
          {{ t('authentication.register.invalid_email') }}
        </span>
      </div>

      <UiButton type="button" size="lg" class="w-full" :disabled="!isValidEmail" @click="handleNext">
        {{ t('authentication.register.next_step') }}
      </UiButton>
    </div>

    <AuthEmailRegisterEnter v-else :email="email" />
  </div>
</template>
