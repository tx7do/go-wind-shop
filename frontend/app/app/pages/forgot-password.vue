<script setup lang="ts">
import { computed, onUnmounted, reactive, ref } from 'vue';
import { toast } from 'vue-sonner';
import { useSendResetCode, useResetPassword } from '@/api/composables';

definePageMeta({ layout: 'auth' });

const { t } = useI18n();
const localePath = useLocalePath();

useHead({ title: t('authentication.forgotPassword.title') });

// 两步流程：1=发送验证码，2=设置新密码
const step = ref<1 | 2>(1);

const email = ref('');
const code = ref('');
const newPassword = ref('');
const confirmPassword = ref('');

// 发送验证码后的脱敏邮箱预览 + 倒计时
const emailPreview = ref('');
const resendSeconds = ref(0);
let resendTimer: ReturnType<typeof setInterval> | null = null;

// 组件卸载时清理倒计时，避免离开页面后 interval 继续运行（资源泄漏）。
onUnmounted(() => {
  if (resendTimer) {
    clearInterval(resendTimer);
    resendTimer = null;
  }
});

function startCountdown(seconds: number) {
  resendSeconds.value = seconds;
  if (resendTimer) clearInterval(resendTimer);
  resendTimer = setInterval(() => {
    resendSeconds.value -= 1;
    if (resendSeconds.value <= 0 && resendTimer) {
      clearInterval(resendTimer);
      resendTimer = null;
    }
  }, 1000);
}

const emailValid = computed(() => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email.value.trim()));
const codeValid = computed(() => code.value.trim().length > 0);
const passwordValid = computed(() => newPassword.value.length >= 6);
const passwordMatch = computed(() => newPassword.value === confirmPassword.value);

const sendMutation = useSendResetCode({
  onSuccess: (resp: any) => {
    emailPreview.value = resp?.emailPreview ?? email.value;
    toast.success(t('authentication.forgotPassword.sentTo', { email: emailPreview.value }));
    step.value = 2;
    // expiresIn 由后端返回（秒），默认 600（10 分钟）
    startCountdown(resp?.expiresIn ?? 600);
  },
  onError: (err: any) => toast.error(err?.message || t('authentication.forgotPassword.sendFailed')),
});

const resetMutation = useResetPassword({
  onSuccess: () => {
    toast.success(t('authentication.forgotPassword.resetSuccess'));
    navigateTo(localePath('/login'));
  },
  onError: (err: any) => {
    // 429：验证码尝试次数耗尽，后端已作废该验证码。
    // 引导用户回到步骤 1 重新获取验证码，而非停留在已失效的步骤 2。
    if (err?.response?.status === 429) {
      toast.error(t('authentication.forgotPassword.tooManyAttempts'));
      step.value = 1;
      // 验证码已失效，清除已填写的验证码字段。
      code.value = '';
      return;
    }
    toast.error(err?.message || t('authentication.forgotPassword.resetFailed'));
  },
});

function handleSendCode() {
  if (!emailValid.value) {
    toast.error(t('authentication.forgotPassword.unavailableTip'));
    return;
  }
  sendMutation.mutate(email.value.trim());
}

function handleReset() {
  if (!codeValid.value) return;
  if (!passwordValid.value) return;
  if (!passwordMatch.value) {
    toast.error(t('authentication.forgotPassword.passwordMismatch'));
    return;
  }
  resetMutation.mutate({
    email: email.value.trim(),
    code: code.value.trim(),
    newPassword: newPassword.value,
  });
}

const anyPending = computed(() => sendMutation.isPending.value || resetMutation.isPending.value);
</script>

<template>
  <div class="mx-auto w-full max-w-md">
    <div class="rounded-2xl border border-border bg-card p-8 shadow-sm">
      <h1 class="text-xl font-bold text-foreground">
        {{ t('authentication.forgotPassword.title') }}
      </h1>
      <p class="mt-2 text-sm text-muted-foreground">
        {{ t('authentication.forgotPassword.forgot_password_description') }}
      </p>

      <!-- 步骤 1：发送验证码 -->
      <div v-if="step === 1" class="mt-6 flex flex-col gap-4">
        <div class="flex flex-col gap-2">
          <UiLabel class="text-xs text-foreground">{{ t('authentication.forgotPassword.email') }}</UiLabel>
          <UiInput
            v-model="email"
            type="email"
            autocomplete="email"
            :placeholder="t('authentication.login.placeholder_email')"
            @keyup.enter="handleSendCode"
          />
        </div>
        <UiButton class="w-full" :disabled="!emailValid || anyPending" @click="handleSendCode">
          {{ t('authentication.forgotPassword.sendCode') }}
        </UiButton>
      </div>

      <!-- 步骤 2：设置新密码 -->
      <div v-else class="mt-6 flex flex-col gap-4">
        <p class="rounded-md bg-muted/60 px-3 py-2 text-xs text-muted-foreground">
          {{ t('authentication.forgotPassword.sentTo', { email: emailPreview }) }}
        </p>

        <div class="flex flex-col gap-2">
          <UiLabel class="text-xs text-foreground">{{ t('authentication.forgotPassword.code') }}</UiLabel>
          <UiInput v-model="code" :placeholder="t('authentication.forgotPassword.codePlaceholder')" />
        </div>

        <div class="flex flex-col gap-2">
          <UiLabel class="text-xs text-foreground">{{ t('authentication.forgotPassword.newPassword') }}</UiLabel>
          <UiInput
            v-model="newPassword"
            type="password"
            autocomplete="new-password"
            :placeholder="t('authentication.forgotPassword.newPasswordPlaceholder')"
          />
        </div>

        <div class="flex flex-col gap-2">
          <UiLabel class="text-xs text-foreground">{{ t('authentication.forgotPassword.confirmPassword') }}</UiLabel>
          <UiInput
            v-model="confirmPassword"
            type="password"
            autocomplete="new-password"
            :placeholder="t('authentication.forgotPassword.confirmPasswordPlaceholder')"
          />
        </div>

        <UiButton
          class="w-full"
          :disabled="!codeValid || !passwordValid || !passwordMatch || anyPending"
          @click="handleReset"
        >
          {{ t('authentication.forgotPassword.resetNow') }}
        </UiButton>

        <!-- 重发验证码 -->
        <button
          type="button"
          class="text-center text-xs text-muted-foreground transition-colors hover:text-primary disabled:opacity-50"
          :disabled="resendSeconds > 0 || anyPending"
          @click="handleSendCode"
        >
          <span v-if="resendSeconds > 0">
            {{ t('authentication.forgotPassword.resendIn', { s: resendSeconds }) }}
          </span>
          <span v-else>{{ t('authentication.forgotPassword.resend') }}</span>
        </button>
      </div>

      <div class="mt-6 text-center">
        <button
          class="cursor-pointer border-none bg-transparent text-sm text-primary transition-colors hover:text-primary/80 hover:underline"
          @click="navigateTo(localePath('/login'))"
        >
          {{ t('authentication.forgotPassword.back_to_login') }}
        </button>
      </div>
    </div>
  </div>
</template>
