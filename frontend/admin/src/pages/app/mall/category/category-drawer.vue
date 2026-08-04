<template>
  <ProModal
    v-model:visible="visible"
    :title="title"
    :config="{ component: 'drawer', drawer: { size: DRAWER_WIDTH, closeOnClickModal: false } }"
  >
    <ElForm
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-width="120px"
      class="drawer-form"
    >
      <!-- 基本信息 -->
      <ElDivider content-position="left">{{ $t("common.section.basic") }}</ElDivider>

      <ElFormItem :label="$t('common.table.sortOrder')" prop="sortOrder">
        <ElInputNumber
          v-model="formData.sortOrder"
          :min="1"
          :max="9999"
          controls-position="right"
          :placeholder="$t('common.placeholder.input')"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.category.parent')" prop="parentId">
        <ElTreeSelect
          v-model="formData.parentId"
          :data="parentTreeData"
          :props="{ label: 'label', children: 'children' }"
          :placeholder="$t('common.placeholder.select')"
          clearable
          check-strictly
          :render-after-expand="false"
          style="width: 100%"
        />
      </ElFormItem>
    </ElForm>

    <!-- 多语言翻译 -->
    <TranslationTabs
      v-if="visible && formData.translations"
      v-model="formData.translations"
      :fields="translationFields"
    />

    <template #footer>
      <div class="drawer-footer">
        <ElButton @click="handleClose">{{ $t("common.button.cancel") }}</ElButton>
        <ElButton type="primary" :loading="submitLoading" @click="handleSubmit">
          {{ $t("common.button.confirm") }}
        </ElButton>
      </div>
    </template>
  </ProModal>
</template>

<script lang="ts" setup>
import { computed, reactive, ref, watch } from "vue";
import { ElMessage, type FormInstance, type FormRules } from "element-plus";

import {
  useCreateCategory,
  useUpdateCategory,
  fetchGetCategory,
  fetchListCategories,
} from "@/api/composables";
import { $t, useI18n } from "@/core/i18n";
import { DRAWER_WIDTH } from "@/constants";
import ProModal from "@/components/Pro/ProModal/index.vue";
import TranslationTabs from "@/components/Pro/TranslationTabs/index.vue";
import type { TranslationFieldConfig } from "@/components/Pro/TranslationTabs/types";
import { PaginationQuery } from "@/core/transport/rest";

const emit = defineEmits<{
  success: [];
}>();

const { mutateAsync: createCategory } = useCreateCategory();
const { mutateAsync: updateCategory } = useUpdateCategory();
const { locale } = useI18n();

const visible = ref(false);
const submitLoading = ref(false);
const isCreate = ref(true);
const currentId = ref<number>();
const formRef = ref<FormInstance>();

const parentTreeData = ref<any[]>([]);

const formData = reactive<{ sortOrder: number; parentId: any; translations: any[] }>({
  sortOrder: 1,
  parentId: undefined,
  translations: [],
});

const formRules: FormRules = {};

const translationFields: TranslationFieldConfig[] = [
  { prop: "name", label: $t("pages.mall.category.name"), type: "input" },
  { prop: "slug", label: $t("pages.mall.category.slug"), type: "input" },
  { prop: "description", label: $t("pages.mall.category.description"), type: "textarea" },
  { prop: "fullPath", label: $t("pages.mall.category.fullPath"), type: "input" },
];

const title = computed(() =>
  isCreate.value ? $t("pages.mall.category.button.create") : $t("pages.mall.category.button.update")
);

// 构建父级类目树（用于 TreeSelect 选择）
async function loadParentTree() {
  try {
    const resp = await fetchListCategories(
      new PaginationQuery({ paging: { page: 1, pageSize: 999 } })
    );
    const items: any[] = resp.items || [];
    // 构建树结构
    const map = new Map<number, any>();
    const roots: any[] = [];
    items.forEach((it: any) => {
      map.set(it.id, {
        value: it.id,
        label: getTranslationField(it, "name") || `#${it.id}`,
        children: [],
      });
    });
    items.forEach((it: any) => {
      const pid = it.parentId;
      const node = map.get(it.id);
      if (pid && map.has(pid)) {
        map.get(pid).children.push(node);
      } else if (node) {
        roots.push(node);
      }
    });
    parentTreeData.value = roots;
  } catch (e) {
    parentTreeData.value = [];
  }
}

// 按 UI 当前语言从 translations 中取对应字段的字面值。
// List 响应含全部语言的 translations，须匹配 languageCode 才能取到
// 当前界面语言的文案；未命中返回空串。
function getTranslationField(row: any, field: string): string {
  const translations = Array.isArray(row?.translations) ? row.translations : [];
  const matched = translations.find(
    (t: any) => t?.languageCode === locale.value
  );
  return matched?.[field] ?? "";
}

function open(row?: any) {
  visible.value = true;
  loadParentTree();

  if (row) {
    isCreate.value = false;
    currentId.value = row.id;
    // 编辑模式：获取完整 DTO（包含 translations）
    fetchGetCategory({ id: row.id })
      .then((resp: any) => {
        Object.assign(formData, resp);
        if (!Array.isArray(formData.translations)) {
          formData.translations = [];
        }
      })
      .catch(() => {
        ElMessage.error($t("common.message.error"));
      });
  } else {
    isCreate.value = true;
    currentId.value = undefined;
    resetForm();
  }
}

function handleClose() {
  visible.value = false;
  resetForm();
}

function resetForm() {
  formData.sortOrder = 1;
  formData.parentId = undefined;
  formData.translations = [];
  formRef.value?.clearValidate();
}

async function handleSubmit() {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
    submitLoading.value = true;

    const values = { ...formData };

    if (isCreate.value) {
      await createCategory(values);
      ElMessage.success($t("common.notification.createSuccess"));
    } else {
      await updateCategory({ id: currentId.value!, values });
      ElMessage.success($t("common.notification.updateSuccess"));
    }

    emit("success");
    handleClose();
  } catch (error) {
    if (error !== false) {
      ElMessage.error(
        isCreate.value
          ? $t("common.notification.createFailed")
          : $t("common.notification.updateFailed")
      );
    }
  } finally {
    submitLoading.value = false;
  }
}

watch(visible, (val) => {
  if (!val) resetForm();
});

defineExpose({
  open,
});
</script>

<style lang="scss" scoped>
.drawer-form {
  padding-right: 10px;
}

.drawer-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>
