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

      <ElFormItem :label="$t('pages.mall.couponTemplate.discountType')" prop="discountType">
        <ElSelect v-model="formData.discountType" :placeholder="$t('common.placeholder.select')">
          <ElOption label="固定金额" value="FIXED_AMOUNT" />
          <ElOption label="百分比" value="PERCENTAGE" />
        </ElSelect>
      </ElFormItem>

      <ElFormItem
        v-if="formData.discountType === 'FIXED_AMOUNT'"
        :label="$t('pages.mall.couponTemplate.discountValue')"
        prop="discountValue"
      >
        <ElInputNumber
          v-model="formData.discountValue"
          :min="0"
          controls-position="right"
          :placeholder="$t('common.placeholder.input')"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem
        v-if="formData.discountType === 'PERCENTAGE'"
        :label="$t('pages.mall.couponTemplate.discountPercentage')"
        prop="discountPercentage"
      >
        <ElInputNumber
          v-model="formData.discountPercentage"
          :min="0"
          :max="100"
          controls-position="right"
          :placeholder="$t('common.placeholder.input')"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.couponTemplate.validFrom')" prop="validFrom">
        <ElDatePicker
          v-model="formData.validFrom"
          type="datetime"
          :placeholder="$t('common.placeholder.select')"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.couponTemplate.validUntil')" prop="validUntil">
        <ElDatePicker
          v-model="formData.validUntil"
          type="datetime"
          :placeholder="$t('common.placeholder.select')"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.couponTemplate.maxRedemptions')" prop="maxRedemptions">
        <ElInputNumber
          v-model="formData.maxRedemptions"
          :min="0"
          controls-position="right"
          :placeholder="$t('common.placeholder.input')"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.couponTemplate.status')" prop="status">
        <ElSelect v-model="formData.status" :placeholder="$t('common.placeholder.select')">
          <ElOption label="生效" value="ACTIVE" />
          <ElOption label="停用" value="INACTIVE" />
        </ElSelect>
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.couponTemplate.currency')" prop="currency">
        <ElInput v-model="formData.currency" :placeholder="$t('common.placeholder.input')" />
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

import { useCreateCouponTemplate, useUpdateCouponTemplate, fetchGetCouponTemplate } from "@/api/composables";
import { $t } from "@/core/i18n";
import { DRAWER_WIDTH } from "@/constants";
import ProModal from "@/components/Pro/ProModal/index.vue";

const emit = defineEmits<{
  success: [];
}>();

const { mutateAsync: createCouponTemplate } = useCreateCouponTemplate();
const { mutateAsync: updateCouponTemplate } = useUpdateCouponTemplate();

const visible = ref(false);
const submitLoading = ref(false);
const isCreate = ref(true);
const currentId = ref<number>();
const formRef = ref<FormInstance>();

const formData = reactive<{
  discountType: string;
  discountValue: number;
  discountPercentage: number;
  validFrom: any;
  validUntil: any;
  maxRedemptions: number;
  status: string;
  currency: string;
}>({
  discountType: "FIXED_AMOUNT",
  discountValue: 0,
  discountPercentage: 0,
  validFrom: null,
  validUntil: null,
  maxRedemptions: 0,
  status: "ACTIVE",
  currency: "CNY",
});

const formRules: FormRules = {
  discountType: [{ required: true, message: $t("common.validate.required"), trigger: "change" }],
  status: [{ required: true, message: $t("common.validate.required"), trigger: "change" }],
  currency: [{ required: true, message: $t("common.validate.required"), trigger: "blur" }],
};

const title = computed(() =>
  isCreate.value ? $t("pages.mall.couponTemplate.button.create") : $t("pages.mall.couponTemplate.button.update")
);

function open(row?: any) {
  visible.value = true;

  if (row) {
    isCreate.value = false;
    currentId.value = row.id;
    fetchGetCouponTemplate({ id: row.id })
      .then((resp: any) => {
        Object.assign(formData, resp);
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
  formData.discountType = "FIXED_AMOUNT";
  formData.discountValue = 0;
  formData.discountPercentage = 0;
  formData.validFrom = null;
  formData.validUntil = null;
  formData.maxRedemptions = 0;
  formData.status = "ACTIVE";
  formData.currency = "CNY";
  formRef.value?.clearValidate();
}

async function handleSubmit() {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
    submitLoading.value = true;

    const values = { ...formData };

    if (isCreate.value) {
      await createCouponTemplate(values);
      ElMessage.success($t("common.notification.createSuccess"));
    } else {
      await updateCouponTemplate({ id: currentId.value!, values });
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
