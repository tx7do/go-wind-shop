<script setup lang="ts">
import { XIcon } from '@/plugins/xicon'
import { usePreferences } from '@/core/preferences/use-preferences'

definePageMeta({
  layout: 'account',
  middleware: 'auth',
})

const { t, locale } = useI18n()
const switchLocalePath = useSwitchLocalePath()

useHead({ title: t('settings.preference.title') })

const { themePreferences: themePref, setTheme: setThemeMode } = usePreferences()

// 语言切换：跳转到目标 locale 的对应路径（参考 ControlPanel.vue 实现）
const languageOptions: { key: 'zh-CN' | 'en-US'; labelKey: string }[] = [
  { key: 'zh-CN', labelKey: 'settings.preference.language_chinese' },
  { key: 'en-US', labelKey: 'settings.preference.language_english' },
]

const handleLocaleChange = (v: string) => {
  navigateTo(switchLocalePath(v as 'zh-CN' | 'en-US'))
}
</script>

<template>
  <LayoutPageHero
    :title="t('settings.preference.title')"
    :description="t('settings.preference.subtitle')"
    icon="carbon:settings"
    size="sm"
  />

  <LayoutSectionContainer max-width="narrow">
    <div class="mb-8">
      <h2 class="mb-2 text-lg font-semibold text-foreground">{{ t('settings.preference.theme_settings') }}</h2>
      <p class="mb-4 text-sm text-muted-foreground">{{ t('settings.preference.theme_desc') }}</p>
      <UiSettingRow :label="t('settings.preference.theme')">
        <UiSelect v-model="themePref.mode" @update:model-value="(v: any) => setThemeMode(v)">
          <UiSelectTrigger class="w-[180px] h-8">
            <UiSelectValue />
          </UiSelectTrigger>
          <UiSelectContent>
            <UiSelectItem value="light">{{ t('settings.preference.theme_light') }}</UiSelectItem>
            <UiSelectItem value="dark">{{ t('settings.preference.theme_dark') }}</UiSelectItem>
            <UiSelectItem value="auto">{{ t('settings.preference.theme_auto') }}</UiSelectItem>
          </UiSelectContent>
        </UiSelect>
      </UiSettingRow>
    </div>

    <div class="mb-8">
      <h2 class="mb-2 text-lg font-semibold text-foreground">{{ t('settings.preference.language_settings') }}</h2>
      <p class="mb-4 text-sm text-muted-foreground">{{ t('settings.preference.language_desc') }}</p>
      <UiSettingRow :label="t('settings.preference.language')">
        <UiSelect
          :model-value="locale"
          @update:model-value="(v: any) => handleLocaleChange(v as string)"
        >
          <UiSelectTrigger class="w-[180px] h-8">
            <UiSelectValue />
          </UiSelectTrigger>
          <UiSelectContent>
            <UiSelectItem
              v-for="opt in languageOptions"
              :key="opt.key"
              :value="opt.key"
            >
              {{ t(opt.labelKey as any) }}
            </UiSelectItem>
          </UiSelectContent>
        </UiSelect>
      </UiSettingRow>
    </div>
  </LayoutSectionContainer>
</template>
