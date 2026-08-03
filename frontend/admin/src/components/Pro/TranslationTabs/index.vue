<template>
  <div class="translation-tabs">
    <ElDivider content-position="left">{{ $t("common.section.basic") }}</ElDivider>
    <ElTabs v-model="activeName" type="border-card">
      <ElTabPane
        v-for="(lang, index) in languages"
        :key="lang.languageCode"
        :label="lang.nativeName || lang.languageName || lang.languageCode"
        :name="String(index)"
      >
        <ElForm label-width="120px" class="translation-form">
          <ElFormItem
            v-for="field in fields"
            :key="field.prop"
            :label="field.label"
          >
            <ElInput
              v-if="field.type === 'input'"
              v-model="getOrCreateEntry(lang.languageCode!)![field.prop]"
              :placeholder="$t('common.placeholder.input')"
              clearable
            />
            <ElInput
              v-else-if="field.type === 'textarea'"
              v-model="getOrCreateEntry(lang.languageCode!)![field.prop]"
              type="textarea"
              :rows="4"
              :placeholder="$t('common.placeholder.input')"
            />
            <ElInputNumber
              v-else-if="field.type === 'input-number'"
              v-model="getOrCreateEntry(lang.languageCode!)![field.prop]"
              style="width: 100%"
            />
          </ElFormItem>
        </ElForm>
      </ElTabPane>
    </ElTabs>
  </div>
</template>

<script lang="ts" setup>
import { ref, watchEffect } from "vue";
import { fetchListLanguages } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t } from "@/core/i18n";
import type { TranslationFieldConfig } from "./types";

defineOptions({ name: "TranslationTabs" });

const props = defineProps<{
  modelValue: any[];
  fields: TranslationFieldConfig[];
}>();

const emit = defineEmits<{
  (e: "update:modelValue", v: any[]): void;
}>();

const languages = ref<any[]>([]);
const activeName = ref("0");

// 加载启用的语言列表
async function loadLanguages() {
  try {
    const resp = await fetchListLanguages(new PaginationQuery({ paging: { page: 1, pageSize: 999 } }));
    languages.value = (resp.items || []).filter((l) => l.isEnabled === true);
    if (languages.value.length > 0) {
      activeName.value = "0";
    }
  } catch (e) {
    languages.value = [];
  }
}

loadLanguages();

// 根据 languageCode 查找或创建对应的翻译条目（确保 v-model 绑定生效）
function getOrCreateEntry(languageCode: string): Record<string, any> | undefined {
  if (!props.modelValue) {
    emit("update:modelValue", []);
    return undefined;
  }
  let entry = props.modelValue.find((t: any) => t && t.languageCode === languageCode);
  if (!entry) {
    entry = { languageCode };
    const next = [...props.modelValue, entry];
    emit("update:modelValue", next);
  }
  return entry as Record<string, any>;
}

// 当语言列表变化时，确保每个语言都有对应的翻译条目
watchEffect(() => {
  if (!languages.value.length || !props.modelValue) return;
  let changed = false;
  const next = [...props.modelValue];
  for (const lang of languages.value) {
    const code = lang.languageCode;
    if (!code) continue;
    const exists = next.find((t: any) => t && t.languageCode === code);
    if (!exists) {
      next.push({ languageCode: code });
      changed = true;
    }
  }
  if (changed) {
    emit("update:modelValue", next);
  }
});
</script>

<style lang="scss" scoped>
.translation-tabs {
  margin-top: 12px;
}

.translation-form {
  padding: 0 8px;
}
</style>
