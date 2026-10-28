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

      <ElFormItem :label="$t('pages.mall.stockAdjust.skuId')">
        <div class="readonly-content">{{ formData.skuId ?? '—' }}</div>
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.stockAdjust.currentStock')">
        <div class="readonly-content">{{ formData.stockQty ?? '—' }}</div>
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.stockAdjust.delta')" prop="delta">
        <ElInputNumber v-model="formData.delta" :min="-100000" :max="100000" :controls="false" class="!w-full" />
        <div class="form-tip">{{ $t('pages.mall.stockAdjust.deltaTip') }}</div>
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.stockAdjust.remark')" prop="remark">
        <ElInput v-model="formData.remark" type="textarea" :rows="2" :maxlength="200" show-word-limit />
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

import { useAdjustSkuStock } from "@/api/composables";
import { $t } from "@/core/i18n";
import { DRAWER_WIDTH } from "@/constants";
import ProModal from "@/components/Pro/ProModal/index.vue";

const emit = defineEmits<{
  success: [];
}>();

const { mutateAsync: adjustStock } = useAdjustSkuStock();

const visible = ref(false);
const submitLoading = ref(false);
const formRef = ref<FormInstance>();

const formData = reactive<{
  skuId: number | undefined;
  stockQty: number | undefined;
  delta: number;
  remark: string;
}>({
  skuId: undefined,
  stockQty: undefined,
  delta: 0,
  remark: "",
});

const formRules: FormRules = {
  delta: [
    { required: true, message: $t("common.validate.required"), trigger: "blur" },
    {
      validator: (_rule: any, value: number, callback: (err?: Error) => void) => {
        if (!value || value === 0) {
          callback(new Error($t("pages.mall.stockAdjust.deltaZero")));
        } else {
          callback();
        }
      },
      trigger: "blur",
    },
  ],
};

const title = computed(() => $t("pages.mall.stockAdjust.title"));

function open(row?: any) {
  if (!row) return;
  visible.value = true;
  formData.skuId = row.id;
  formData.stockQty = row.stockQty ?? 0;
  formData.delta = 0;
  formData.remark = "";
}

function handleClose() {
  visible.value = false;
  resetForm();
}

function resetForm() {
  formData.skuId = undefined;
  formData.stockQty = undefined;
  formData.delta = 0;
  formData.remark = "";
  formRef.value?.clearValidate();
}

async function handleSubmit() {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
    submitLoading.value = true;

    await adjustStock({
      id: formData.skuId!,
      values: { delta: formData.delta, remark: formData.remark },
    });
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
  padding: 8px 12px;
  background: var(--el-fill-color-light);
  border-radius: 4px;
  font-size: 13px;
  color: var(--el-text-color-regular);
}

.form-tip {
  width: 100%;
  margin-top: 4px;
  font-size: 12px;
  line-height: 1.4;
  color: var(--el-text-color-secondary);
}

.drawer-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>
