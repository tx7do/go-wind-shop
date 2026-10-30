<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue';
import { toast } from 'vue-sonner';
import { useSendResetCode, useResetPassword } from '@/api/composables';
import { XIcon } from '@/plugins/xicon';

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
const newVisible = ref(false);
const confirmVisible = ref(false);

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
    //
    // 传输层 RequestClient.request 的 catch 会剥离 axios 包装，
    // 仅抛出 error.response.data（kratos 错误体 {code, reason, message}），
    // 其中 code 为 HTTP 状态码（见 errors.New(429, ...) 与 errors.pb.go）。
    // 故此处判 err.code === 429，而非 err.response.status。
    if (err?.code === 429) {
      toast.error(t('authentication.forgotPassword.tooManyAttempts'));
      step.value = 1;
      // 验证码已被后端作废，回到步骤 1 重新获取。此处的步骤 2 输入（验证码、
      // 新密码、确认密码）全部失效，必须清空，避免在新一轮步骤 2 中残留并
      // 被误用提交。
      code.value = '';
      newPassword.value = '';
      confirmPassword.value = '';
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

const stepDotActive = 'bg-primary text-primary-foreground';
const stepDotDone = 'bg-primary-soft text-primary-soft-foreground';
const stepDotIdle = 'border border-border bg-card text-muted-foreground';
const eyeBtn = 'absolute inset-y-0 right-0 flex w-11 cursor-pointer items-center justify-center border-none bg-transparent text-muted-foreground transition-colors hover:text-foreground';
</script>

<template>
  <AuthLayout
    :title="t('authentication.forgotPassword.title')"
    :subtitle="t('authentication.forgotPassword.forgot_password_description')"
  >
    <ol class="mb-5 flex items-center gap-2" :aria-label="t('authentication.forgotPassword.title')">
      <li class="flex items-center gap-2">
        <span
          class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-[11px] font-semibold leading-none"
          :class="step === 1 ? stepDotActive : stepDotDone"
        >
          1
        </span>
        <span
          class="whitespace-nowrap text-xs"
          :class="step === 1 ? 'font-medium text-foreground' : 'text-muted-foreground'"
        >
          {{ t('authentication.forgotPassword.step_send') }}
        </span>
      </li>
      <li class="h-px flex-1 bg-border" aria-hidden="true" />
      <li class="flex items-center gap-2">
        <span
          class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-[11px] font-semibold leading-none"
          :class="step === 2 ? stepDotActive : stepDotIdle"
        >
          2
        </span>
        <span
          class="whitespace-nowrap text-xs"
          :class="step === 2 ? 'font-medium text-foreground' : 'text-muted-foreground'"
        >
          {{ t('authentication.forgotPassword.step_reset') }}
        </span>
      </li>
    </ol>

    <!-- 步骤 1：发送验证码 -->
    <div v-if="step === 1" class="space-y-4">
      <div class="space-y-1.5">
        <UiLabel for="forgot-email" class="text-foreground">
          {{ t('authentication.forgotPassword.email') }}
        </UiLabel>
        <UiInput
          id="forgot-email"
          v-model="email"
          type="email"
          autocomplete="email"
          autocapitalize="none"
          autocorrect="off"
          spellcheck="false"
          :placeholder="t('authentication.login.placeholder_email')"
          @keyup.enter="handleSendCode"
        />
      </div>

      <UiButton
        type="button"
        size="lg"
        class="w-full"
        :disabled="!emailValid || anyPending"
        @click="handleSendCode"
      >
        <UiSpinner v-if="anyPending" size="sm" class="text-primary-foreground" />
        {{ t('authentication.forgotPassword.sendCode') }}
      </UiButton>
    </div>

    <!-- 步骤 2：设置新密码 -->
    <div v-else class="space-y-4">
      <p class="break-all rounded-lg bg-primary-soft px-3.5 py-2.5 text-xs leading-relaxed text-primary-soft-foreground">
        {{ t('authentication.forgotPassword.sentTo', { email: emailPreview }) }}
      </p>

      <div class="space-y-1.5">
        <UiLabel for="forgot-code" class="text-foreground">
          {{ t('authentication.forgotPassword.code') }}
        </UiLabel>
        <div class="flex justify-center">
          <UiOtpInput id="forgot-code" v-model="code" :length="6" :disabled="anyPending" type="number" />
        </div>
      </div>

      <div class="space-y-1.5">
        <UiLabel for="forgot-new-password" class="text-foreground">
          {{ t('authentication.forgotPassword.newPassword') }}
        </UiLabel>
        <div class="relative">
          <UiInput
            id="forgot-new-password"
            v-model="newPassword"
            :type="newVisible ? 'text' : 'password'"
            autocomplete="new-password"
            class="pr-11"
            :placeholder="t('authentication.forgotPassword.newPasswordPlaceholder')"
          />
          <button
            type="button"
            :class="eyeBtn"
            :aria-pressed="newVisible"
            :aria-label="t('authentication.forgotPassword.newPassword')"
            @click="newVisible = !newVisible"
          >
            <XIcon :icon="newVisible ? 'lucide:eye-off' : 'lucide:eye'" :size="16" />
          </button>
        </div>
      </div>

      <div class="space-y-1.5">
        <UiLabel for="forgot-confirm-password" class="text-foreground">
          {{ t('authentication.forgotPassword.confirmPassword') }}
        </UiLabel>
        <div class="relative">
          <UiInput
            id="forgot-confirm-password"
            v-model="confirmPassword"
            :type="confirmVisible ? 'text' : 'password'"
            autocomplete="new-password"
            class="pr-11"
            :placeholder="t('authentication.forgotPassword.confirmPasswordPlaceholder')"
          />
          <button
            type="button"
            :class="eyeBtn"
            :aria-pressed="confirmVisible"
            :aria-label="t('authentication.forgotPassword.confirmPassword')"
            @click="confirmVisible = !confirmVisible"
          >
            <XIcon :icon="confirmVisible ? 'lucide:eye-off' : 'lucide:eye'" :size="16" />
          </button>
        </div>
        <span
          v-if="confirmPassword.length > 0 && !passwordMatch"
          class="block text-xs text-destructive"
        >
          {{ t('authentication.forgotPassword.passwordMismatch') }}
        </span>
      </div>

      <UiButton
        type="button"
        size="lg"
        class="w-full"
        :disabled="!codeValid || !passwordValid || !passwordMatch || anyPending"
        @click="handleReset"
      >
        <UiSpinner v-if="anyPending" size="sm" class="text-primary-foreground" />
        {{ t('authentication.forgotPassword.resetNow') }}
      </UiButton>

      <UiButton
        type="button"
        variant="link"
        size="sm"
        class="mx-auto block h-auto p-0 text-xs font-normal text-muted-foreground hover:text-primary"
        :disabled="resendSeconds > 0 || anyPending"
        @click="handleSendCode"
      >
        <span v-if="resendSeconds > 0">
          {{ t('authentication.forgotPassword.resendIn', { s: resendSeconds }) }}
        </span>
        <span v-else>{{ t('authentication.forgotPassword.resend') }}</span>
      </UiButton>
    </div>

    <template #switchLink>
      <UiButton variant="link" size="sm" class="h-auto p-0" @click="navigateTo(localePath('/login'))">
        {{ t('authentication.forgotPassword.back_to_login') }}
      </UiButton>
    </template>
  </AuthLayout>
</template>
