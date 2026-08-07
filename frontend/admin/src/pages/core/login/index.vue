<template>
  <div class="login-layout">
    <!-- 顶部 Logo 区域 -->
    <div class="login-header">
      <div class="header-left">
        <el-image :src="logo" class="header-logo" fit="contain" />
        <span class="header-title">{{ t("core.login.headerTitle") }}</span>
      </div>
      <div class="header-right">
        <ThemeSwitch class="header-icon" />
        <LangSelect class="header-icon" size="text-20px" />
      </div>
    </div>

    <!-- 主内容区 -->
    <div class="login-content">
      <!-- 左侧品牌展示 -->
      <div class="login-brand">
        <div class="brand-content">
          <div class="brand-illustration">
            <SloganIcon class="slogan-icon" />
          </div>
          <div class="brand-info">
            <h2 class="brand-title">{{ t("core.login.brandTitle") }}</h2>
            <p class="brand-desc">{{ t("core.login.brandDesc") }}</p>
          </div>
        </div>
      </div>

      <!-- 右侧登录表单 -->
      <div class="login-form-wrapper">
        <div class="login-form-container">
          <div class="form-header">
            <h2 class="form-title">
              {{ t("core.login.welcomeTitle") }}
              <span class="wave">👋</span>
            </h2>
            <p class="form-subtitle">{{ t("core.login.welcomeSubtitle") }}</p>
          </div>

          <transition name="fade-slide" mode="out-in">
            <component :is="formComponents[component]" class="auth-panel__form" />
          </transition>
        </div>

        <!-- 版权信息 - 放在右侧面板最底部 -->
        <div class="form-copyright">
          <el-text size="small">{{ t("core.login.copyright") }}</el-text>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import logo from "@/assets/images/logo.png";
import ThemeSwitch from "@/components/ThemeSwitch/index.vue";
import SloganIcon from "./icons/slogan.vue";

const { t } = useI18n();

type LayoutMap = "login";

const component = ref<LayoutMap>("login");

const formComponents = {
  login: defineAsyncComponent(() => import("./components/Login.vue")),
};
</script>

<style lang="scss" scoped>
.login-layout {
  display: flex;
  flex-direction: column;
  width: 100%;
  min-height: 100vh;
  background-color: #f5f7ff;

  html:not(.dark) & {
    background-color: #f5f7ff;
  }
}

// 顶部 Header
.login-header {
  position: absolute;
  top: 0;
  right: 0;
  left: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 40px;

  .header-left {
    display: flex;
    gap: 10px;
    align-items: center;

    .header-logo {
      width: 24px;
      height: 24px;
    }

    .header-title {
      font-size: 14px;
      font-weight: 600;
      color: #1a1d28;

      html:not(.dark) & {
        color: #1a1d28;
      }
    }
  }

  .header-right {
    display: flex;
    gap: 12px;
    align-items: center;

    .header-icon {
      cursor: pointer;
      transition: opacity 0.3s ease;

      &:hover {
        opacity: 0.7;
      }
    }
  }
}

// 主内容区
.login-content {
  display: flex;
  flex: 1;
  min-height: 0;
}

// 左侧品牌展示
.login-brand {
  position: relative;
  display: flex;
  flex: 1;
  align-items: center;
  justify-content: center;
  padding: 40px;
  overflow: hidden;
  background: radial-gradient(ellipse at center, #e8f0ff 0%, #f5f7ff 70%);

  html:not(.dark) & {
    background: radial-gradient(ellipse at center, #e8f0ff 0%, #f5f7ff 70%);
  }

  &::before {
    position: absolute;
    top: 30%;
    left: 20%;
    width: 250px;
    height: 250px;
    content: "";
    background: radial-gradient(circle, rgba(64, 158, 255, 0.08) 0%, transparent 70%);
    border-radius: 50%;
    filter: blur(60px);
  }

  &::after {
    position: absolute;
    right: 15%;
    bottom: 25%;
    width: 180px;
    height: 180px;
    content: "";
    background: radial-gradient(circle, rgba(64, 158, 255, 0.06) 0%, transparent 70%);
    border-radius: 50%;
    filter: blur(50px);
  }

  .brand-content {
    position: relative;
    z-index: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
  }

  .brand-illustration {
    margin-bottom: 32px;

    .slogan-icon {
      width: 280px;
      height: 280px;
      filter: drop-shadow(0 8px 32px rgba(64, 158, 255, 0.3));

      html:not(.dark) & {
        filter: drop-shadow(0 8px 32px rgba(64, 158, 255, 0.2));
      }
    }
  }

  .brand-info {
    .brand-title {
      margin: 0 0 8px 0;
      font-size: 20px;
      font-weight: 600;
      color: #1a1d28;
    }

    .brand-desc {
      margin: 0;
      font-size: 13px;
      color: #6b7280;
    }
  }
}

// 右侧登录表单
.login-form-wrapper {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  width: 520px;
  min-width: 480px;
  padding: 60px 40px;
  // 用 EP 语义背景变量，浅色为白、暗色自动跟随，避免硬编码 #141927/#ffffff 两套
  background-color: var(--el-bg-color);

  .login-form-container {
    display: flex;
    flex: 1;
    flex-direction: column;
    justify-content: center;
    width: 100%;
    max-width: 380px;

    .form-header {
      margin-bottom: 24px;

      .form-title {
        margin: 0 0 8px 0;
        font-size: 22px;
        font-weight: 600;
        color: var(--el-text-color-primary);

        .wave {
          display: inline-block;
          transform-origin: 70% 70%;
          animation: wave 2.5s infinite;
        }
      }

      .form-subtitle {
        margin: 0;
        font-size: 13px;
        color: var(--el-text-color-secondary);
      }
    }

    .form-section-title {
      margin: 24px 0 20px 0;
      font-size: 15px;
      font-weight: 600;
      color: var(--el-text-color-primary);
      text-align: center;
    }

    .form-footer {
      padding-top: 16px;
      margin-top: 24px;
      text-align: center;
      border-top: 1px solid var(--el-border-color-lighter);

      > .el-text {
        display: block;
        color: var(--el-text-color-secondary);
      }
    }
  }
}

@keyframes wave {
  0% {
    transform: rotate(0deg);
  }
  10% {
    transform: rotate(14deg);
  }
  20% {
    transform: rotate(-8deg);
  }
  30% {
    transform: rotate(14deg);
  }
  40% {
    transform: rotate(-4deg);
  }
  50% {
    transform: rotate(10deg);
  }
  60% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(0deg);
  }
}

// 响应式
@media (max-width: 768px) {
  .login-content {
    flex-direction: column;
  }

  .login-brand {
    display: none;
  }

  .login-form-wrapper {
    width: 100%;
    min-width: auto;
    padding: 40px 20px;
  }
}

// 覆盖表单样式
// 统一使用 EP 语义变量：浅色/暗色自动跟随主题，不再维护两套硬编码色板。
// 主色用 var(--el-color-primary)，用户在偏好设置改主题色后这里会同步生效。
.auth-panel__form {
  :deep(.el-form-item) {
    margin-bottom: 16px;
  }

  :deep(.el-input__wrapper) {
    background-color: var(--el-input-bg-color, var(--el-fill-color-blank)) !important;
    border: 1px solid var(--el-border-color) !important;
    box-shadow: none !important;
    transition: all 0.2s ease;

    &:hover {
      border-color: var(--el-color-primary) !important;
    }

    &.is-focus {
      border-color: var(--el-color-primary) !important;
      box-shadow: 0 0 0 1px var(--el-color-primary) inset !important;
    }

    .el-input__inner {
      font-weight: 400;
      color: var(--el-input-text-color, var(--el-text-color-regular)) !important;

      &::placeholder {
        color: var(--el-text-color-placeholder) !important;
      }
    }
  }

  // 输入框前缀图标颜色
  :deep(.el-input__prefix) {
    .el-icon {
      color: var(--el-text-color-placeholder) !important;
    }
  }

  // 输入框后缀图标颜色
  :deep(.el-input__suffix) {
    .el-icon {
      color: var(--el-text-color-placeholder) !important;
    }
  }

  :deep(.el-checkbox__label) {
    font-weight: 400;
    color: var(--el-text-color-regular) !important;
  }

  :deep(.el-checkbox__inner) {
    background-color: var(--el-checkbox-bg-color, var(--el-fill-color-blank)) !important;
    border-color: var(--el-border-color) !important;
  }

  :deep(.el-checkbox.is-checked .el-checkbox__inner) {
    background-color: var(--el-color-primary) !important;
    border-color: var(--el-color-primary) !important;
  }

  :deep(.el-link) {
    font-weight: 500;
    color: var(--el-color-primary) !important;
    text-decoration: underline;
    text-decoration-color: var(--el-color-primary-light-5);
    text-underline-offset: 2px;

    &:hover {
      color: var(--el-color-primary-light-3) !important;
      text-decoration-color: var(--el-color-primary-light-3);
    }
  }
}

// 版权信息 - 固定在右侧面板底部
.form-copyright {
  position: absolute;
  right: 0;
  bottom: 20px;
  left: 0;
  text-align: center;

  :deep(.el-text) {
    font-weight: 300;
    color: var(--el-text-color-placeholder);
  }
}
</style>
