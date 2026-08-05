<script setup lang="ts">
import { toast } from 'vue-sonner';

definePageMeta({ layout: 'auth' });

const { t } = useI18n();
const localePath = useLocalePath();

useHead({ title: t('authentication.forgotPassword.title') });

const email = ref('');
const sending = ref(false);

// 后端 AuthenticationService 暂未提供密码重置 RPC（无发送重置邮件 / 校验令牌接口）。
// 此页先作为占位，保留邮箱表单与文案骨架；点击发送给出明确提示，避免死链与误操作。
// 后端补齐 ResetPassword / SendResetEmail 后，在此接通即可。
async function handleSend() {
  if (!email.value.trim()) {
    toast.error(t('authentication.forgotPassword.unavailableTip'));
    return;
  }
  sending.value = true;
  // 模拟极短延时以呈现按钮态；功能未上线，统一以提示收尾。
  await new Promise((r) => setTimeout(r, 300));
  sending.value = false;
  toast.info(t('authentication.forgotPassword.unavailableTip'));
}
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

      <div class="mt-6 flex flex-col gap-2">
        <UiLabel class="text-xs text-foreground">
          {{ t('authentication.forgotPassword.email') }}
        </UiLabel>
        <UiInput
          v-model="email"
          type="email"
          autocomplete="email"
          :placeholder="t('authentication.login.placeholder_email')"
        />
      </div>

      <!-- 功能未上线提示条 -->
      <div class="mt-4 rounded-md border border-amber-500/30 bg-amber-500/10 px-3 py-2 text-xs text-amber-600 dark:text-amber-400">
        {{ t('authentication.forgotPassword.unavailableTip') }}
      </div>

      <UiButton class="mt-5 w-full" :disabled="sending" @click="handleSend">
        {{ t('authentication.forgotPassword.send') }}
      </UiButton>

      <div class="mt-4 text-center">
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
