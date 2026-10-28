<template>
  <ProModal
    v-model:visible="visible"
    :title="title"
    :config="{ component: 'drawer', drawer: { size: DRAWER_WIDTH } }"
    :close-guard="guard.allowClose"
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
    </ElForm>

    <!-- 多语言翻译 -->
    <TranslationTabs
      v-if="visible && formData.translations"
      v-model="formData.translations"
      :fields="translationFields"
    />

    <template #footer>
      <div class="drawer-footer">
        <ElButton @click="guard.requestClose(handleClose)">
          {{ $t("common.button.cancel") }}
        </ElButton>
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
  useCreateProductAttribute,
  useUpdateProductAttribute,
  fetchGetProductAttribute,
} from "@/api/composables";
import { $t } from "@/core/i18n";
import { DRAWER_WIDTH } from "@/constants";
import ProModal from "@/components/Pro/ProModal/index.vue";
import { useDrawerCloseGuard } from "@/components/Pro/composables/use-drawer-close-guard";
import TranslationTabs from "@/components/Pro/TranslationTabs/index.vue";
import type { TranslationFieldConfig } from "@/components/Pro/TranslationTabs/types";

const emit = defineEmits<{
  success: [];
}>();

const { mutateAsync: createProductAttribute } = useCreateProductAttribute();
const { mutateAsync: updateProductAttribute } = useUpdateProductAttribute();

const visible = ref(false);
const submitLoading = ref(false);
const isCreate = ref(true);
const currentId = ref<number>();
const formRef = ref<FormInstance>();

const formData = reactive<{ sortOrder: number; translations: any[] }>({
  sortOrder: 1,
  translations: [],
});

const formRules: FormRules = {};

const translationFields: TranslationFieldConfig[] = [
  { prop: "name", label: $t("pages.mall.productAttribute.name"), type: "input" },
];

const title = computed(() =>
  isCreate.value
    ? $t("pages.mall.productAttribute.button.create")
    : $t("pages.mall.productAttribute.button.update")
);

// 关闭守卫：遮罩/ESC/X/取消统一，脏数据先确认（docs/form_drawer_close_guard.md）
const guard = useDrawerCloseGuard(() => ({ ...formData }));

function open(row?: any) {
  visible.value = true;

  if (row) {
    isCreate.value = false;
    currentId.value = row.id;
    fetchGetProductAttribute({ id: row.id })
      .then((resp: any) => {
        Object.assign(formData, resp);
        if (!Array.isArray(formData.translations)) {
          formData.translations = [];
        }
        guard.syncBaseline();
      })
      .catch(() => {
        ElMessage.error($t("common.message.error"));
      });
  } else {
    isCreate.value = true;
    currentId.value = undefined;
    resetForm();
  }

  // 填充完成后记录基线快照（回填不算脏；编辑异步回填在 then 里重记）
  guard.syncBaseline();
}

function handleClose() {
  visible.value = false;
  resetForm();
}

function resetForm() {
  formData.sortOrder = 1;
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
      await createProductAttribute(values);
      ElMessage.success($t("common.notification.createSuccess"));
    } else {
      await updateProductAttribute({ id: currentId.value!, values });
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
