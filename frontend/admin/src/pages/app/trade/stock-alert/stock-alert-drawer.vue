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
      label-width="140px"
      class="drawer-form"
    >
      <ElDivider content-position="left">{{ $t("common.section.basic") }}</ElDivider>

      <ElFormItem :label="$t('pages.mall.stockAlert.skuId')">
        <div class="readonly-content">{{ formData.skuId ?? '—' }}</div>
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.stockAlert.stockQtyAtTrigger')">
        <div class="readonly-content">{{ formData.stockQtyAtTrigger ?? '—' }}</div>
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.stockAlert.threshold')">
        <div class="readonly-content">{{ formData.threshold ?? '—' }}</div>
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.stockAlert.createdAt')">
        <div class="readonly-content">{{ formData.createdAt ?? '—' }}</div>
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.stockAlert.status')" prop="status">
        <ElSelect v-model="formData.status" :placeholder="$t('common.placeholder.select')">
          <ElOption :label="$t('pages.mall.stockAlert.statusLabel.RESOLVED')" value="RESOLVED" />
        </ElSelect>
      </ElFormItem>
    </ElForm>

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

import { useUpdateStockAlert, fetchGetStockAlert } from "@/api/composables";
import { $t } from "@/core/i18n";
import { DRAWER_WIDTH } from "@/constants";
import ProModal from "@/components/Pro/ProModal/index.vue";

const emit = defineEmits<{
  success: [];
}>();

const { mutateAsync: updateStockAlert } = useUpdateStockAlert();

const visible = ref(false);
const submitLoading = ref(false);
const currentId = ref<number>();
const formRef = ref<FormInstance>();

const formData = reactive<{
  skuId: number | string;
  stockQtyAtTrigger: number | string;
  threshold: number | string;
  createdAt: string;
  status: string;
}>({
  skuId: "",
  stockQtyAtTrigger: "",
  threshold: "",
  createdAt: "",
  status: "RESOLVED",
});

const formRules: FormRules = {
  status: [{ required: true, message: $t("common.validate.required"), trigger: "change" }],
};

const title = computed(() => $t("pages.mall.stockAlert.button.update"));

function open(row?: any) {
  if (!row) {
    return;
  }
  visible.value = true;
  currentId.value = row.id;
  fetchGetStockAlert({ id: row.id })
    .then((resp: any) => {
      formData.skuId = resp.skuId ?? "";
      formData.stockQtyAtTrigger = resp.stockQtyAtTrigger ?? "";
      formData.threshold = resp.threshold ?? "";
      formData.createdAt = resp.createdAt ?? "";
      formData.status = "RESOLVED";
    })
    .catch(() => {
      ElMessage.error($t("common.message.error"));
    });
}

function handleClose() {
  visible.value = false;
  resetForm();
}

function resetForm() {
  formData.skuId = "";
  formData.stockQtyAtTrigger = "";
  formData.threshold = "";
  formData.createdAt = "";
  formData.status = "RESOLVED";
  formRef.value?.clearValidate();
}

async function handleSubmit() {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
    submitLoading.value = true;

    await updateStockAlert({ id: currentId.value!, values: { status: formData.status } });
    ElMessage.success($t("common.notification.updateSuccess"));

    emit("success");
    handleClose();
  } catch (error) {
    if (error !== false) {
      ElMessage.error($t("common.notification.updateFailed"));
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

.readonly-content {
  width: 100%;
  max-height: 200px;
  overflow-y: auto;
  padding: 8px 12px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  font-size: 13px;
  color: var(--el-text-color-regular);
  white-space: pre-wrap;
  word-break: break-word;
}

.drawer-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>
