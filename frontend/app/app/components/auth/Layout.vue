<script setup lang="ts">
defineProps<{
  title: string
  subtitle: string
  activeTab?: string
  tabs?: { key: string; label: string }[]
}>()

const emit = defineEmits<{
  'update:activeTab': [tab: string]
}>()

const { t } = useI18n()
const localePath = useLocalePath()

const features = ['feature_projects', 'feature_isolation', 'feature_permissions', 'feature_analytics']

const tabClass = 'flex-1 cursor-pointer rounded-lg px-2 py-2 text-sm font-medium transition-colors'
const tabIdle = 'text-muted-foreground hover:bg-card hover:text-foreground'
const tabActive = 'bg-card font-semibold text-primary-soft-foreground shadow-sm'
</script>

<template>
  <div class="hero-band flex min-h-dvh w-full flex-col bg-background">
    <LayoutControlPanel />

    <!-- 移动端品牌条：品牌故事压缩成一条，不和表单抢首屏 -->
    <div class="flex items-center gap-3 border-b border-border px-5 py-4 pr-32 lg:hidden">
      <LayoutBrandLogo class="h-9 w-auto shrink-0" />
      <div class="min-w-0">
        <p class="truncate text-base font-bold leading-tight text-foreground">
          {{ t('authentication.login.brand_title') }}
        </p>
        <p class="truncate text-xs text-muted-foreground">
          {{ t('authentication.login.brand_subtitle') }}
        </p>
      </div>
    </div>

    <div class="mx-auto flex w-full max-w-5xl flex-1 flex-col justify-center px-4 py-8 md:px-8 md:py-12 lg:grid lg:grid-cols-2 lg:items-center lg:gap-16">
      <!-- 桌面品牌故事（<lg 收起，不与表单抢宽度） -->
      <div class="hidden lg:flex lg:flex-col">
        <LayoutBrandLogo class="h-14 w-auto self-start" />

        <h1 class="mt-7 text-3xl font-bold leading-tight tracking-tight text-foreground">
          {{ t('authentication.login.brand_title') }}
        </h1>
        <p class="mt-2 text-base text-muted-foreground">
          {{ t('authentication.login.brand_subtitle') }}
        </p>

        <ul class="mt-8 space-y-3">
          <li
            v-for="key in features"
            :key="key"
            class="flex items-center gap-3 text-sm text-foreground"
          >
            <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary-soft text-[11px] font-bold leading-none text-primary-soft-foreground">
              ✓
            </span>
            <span>{{ t(`authentication.login.${key}`) }}</span>
          </li>
        </ul>
      </div>

      <!-- 表单卡片 -->
      <div class="mx-auto w-full max-w-[440px] lg:mx-0 lg:ml-auto">
        <div class="rounded-xl border border-border bg-card p-6 shadow-card md:p-8">
          <header>
            <h2 class="text-xl font-bold leading-tight text-foreground md:text-2xl">
              {{ title }}
            </h2>
            <p class="mt-1.5 text-sm leading-relaxed text-muted-foreground">
              {{ subtitle }}
            </p>
          </header>

          <div v-if="tabs && tabs.length" class="mt-6 flex gap-1 rounded-xl bg-secondary p-1">
            <button
              v-for="tab in tabs"
              :key="tab.key"
              type="button"
              :class="[tabClass, activeTab === tab.key ? tabActive : tabIdle]"
              :aria-pressed="activeTab === tab.key"
              @click="emit('update:activeTab', tab.key)"
            >
              {{ tab.label }}
            </button>
          </div>

          <div class="mt-6">
            <slot />
          </div>

          <div v-if="$slots.switchLink" class="mt-5 text-center text-sm text-muted-foreground">
            <slot name="switchLink" />
          </div>

          <div class="mt-4 text-center">
            <UiButton variant="link" size="sm" class="h-auto p-0 text-xs text-muted-foreground" @click="navigateTo(localePath('/'))">
              ← {{ t('authentication.login.back_home') }}
            </UiButton>
          </div>

          <div class="mt-3 border-t border-border pt-4 text-center text-xs leading-relaxed text-muted-foreground">
            <span>{{ t('authentication.login.terms_prefix') }}</span>
            <UiButton variant="link" size="sm" class="mx-1 h-auto p-0 text-xs" @click="navigateTo(localePath('/terms'))">
              {{ t('authentication.login.terms_of_service') }}
            </UiButton>
            <span>{{ t('authentication.login.terms_and') }}</span>
            <UiButton variant="link" size="sm" class="ml-1 h-auto p-0 text-xs" @click="navigateTo(localePath('/privacy'))">
              {{ t('authentication.login.privacy_policy') }}
            </UiButton>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
