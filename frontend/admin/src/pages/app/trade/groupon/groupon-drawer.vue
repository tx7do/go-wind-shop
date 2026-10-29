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
      label-width="150px"
      class="drawer-form"
    >
      <ElDivider content-position="left">{{ $t("common.section.basic") }}</ElDivider>

      <ElFormItem :label="$t('pages.mall.groupon.name')" prop="name">
        <ElInput v-model="formData.name" :placeholder="$t('common.placeholder.input')" />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.groupon.skuId')" prop="skuId">
        <ElInputNumber
          v-model="formData.skuId"
          :min="1"
          :step="1"
          :step-strictly="true"
          :precision="0"
          controls-position="right"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.groupon.groupPrice')" prop="groupPriceYuan">
        <ElInputNumber
          v-model="formData.groupPriceYuan"
          :min="0.01"
          :precision="2"
          controls-position="right"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.groupon.groupSize')" prop="groupSize">
        <ElInputNumber
          v-model="formData.groupSize"
          :min="2"
          :step="1"
          :step-strictly="true"
          :precision="0"
          controls-position="right"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.groupon.groupValidHours')" prop="groupValidHours">
        <ElInputNumber
          v-model="formData.groupValidHours"
          :min="1"
          :step="1"
          :step-strictly="true"
          :precision="0"
          controls-position="right"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.seckill.validRange')" prop="validRange">
        <ElDatePicker
          v-model="formData.validRange"
          type="datetimerange"
          value-format="YYYY-MM-DDTHH:mm:ss[Z]"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.groupon.status')" prop="status">
        <ElSelect v-model="formData.status">
          <ElOption :label="$t('pages.mall.groupon.statusLabel.active')" value="ACTIVE" />
          <ElOption :label="$t('pages.mall.groupon.statusLabel.inactive')" value="INACTIVE" />
        </ElSelect>
      </ElFormItem>
    </ElForm>

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

import { useCreateGrouponActivity, useUpdateGrouponActivity } from "@/api/composables";
import { $t } from "@/core/i18n";
import { DRAWER_WIDTH } from "@/constants";
import ProModal from "@/components/Pro/ProModal/index.vue";
import { useDrawerCloseGuard } from "@/components/Pro/composables/use-drawer-close-guard";

const emit = defineEmits<{ success: [] }>();

const { mutateAsync: createGroupon } = useCreateGrouponActivity();
const { mutateAsync: updateGroupon } = useUpdateGrouponActivity();

const visible = ref(false);
const submitLoading = ref(false);
const isCreate = ref(true);
const currentId = ref<number>();
const formRef = ref<FormInstance>();

const formData = reactive<{
  name: string;
  skuId: number;
  groupPriceYuan: number;
  groupSize: number;
  groupValidHours: number;
  validRange: [string, string] | null;
  status: string;
}>({
  name: "",
  skuId: 1,
  groupPriceYuan: 1,
  groupSize: 2,
  groupValidHours: 24,
  validRange: null,
  status: "ACTIVE",
});

const formRules: FormRules = {
  name: [{ required: true, message: $t("common.validate.required"), trigger: "blur" }],
  skuId: [{ required: true, message: $t("common.validate.required"), trigger: "blur" }],
};

const title = computed(() =>
  isCreate.value ? $t("pages.mall.groupon.button.create") : $t("pages.mall.groupon.button.update")
);

const guard = useDrawerCloseGuard(() => ({ ...formData }));

function resetForm() {
  formData.name = "";
  formData.skuId = 1;
  formData.groupPriceYuan = 1;
  formData.groupSize = 2;
  formData.groupValidHours = 24;
  formData.validRange = null;
  formData.status = "ACTIVE";
  formRef.value?.clearValidate();
}

function open(row?: any) {
  visible.value = true;
  if (row) {
    isCreate.value = false;
    currentId.value = row.id;
    formData.name = row.name ?? "";
    formData.skuId = row.skuId ?? 1;
    formData.groupPriceYuan = (row.groupPrice ?? 0) / 100;
    formData.groupSize = row.groupSize ?? 2;
    formData.groupValidHours = row.groupValidHours ?? 24;
    formData.validRange =
      row.validFrom && row.validUntil ? [row.validFrom, row.validUntil] : null;
    formData.status = row.status ?? "ACTIVE";
    guard.syncBaseline();
  } else {
    isCreate.value = true;
    currentId.value = undefined;
    resetForm();
  }
  guard.syncBaseline();
}

function handleClose() {
  visible.value = false;
  resetForm();
}

async function handleSubmit() {
  if (!formRef.value) return;
  try {
    await formRef.value.validate();
    submitLoading.value = true;

    const values: Record<string, any> = {
      name: formData.name,
      skuId: formData.skuId,
      groupPrice: Math.round(formData.groupPriceYuan * 100),
      groupSize: formData.groupSize,
      groupValidHours: formData.groupValidHours,
      status: formData.status,
    };
    if (formData.validRange && formData.validRange.length === 2) {
      values.validFrom = formData.validRange[0];
      values.validUntil = formData.validRange[1];
    }

    if (isCreate.value) {
      await createGroupon(values);
      ElMessage.success($t("common.notification.createSuccess"));
    } else {
      await updateGroupon({ id: currentId.value!, values });
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

defineExpose({ open });
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
