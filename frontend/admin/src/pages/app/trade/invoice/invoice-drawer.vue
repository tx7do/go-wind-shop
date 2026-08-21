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

      <ElFormItem :label="$t('pages.mall.invoice.orderId')" prop="orderId">
        <ElInputNumber
          v-model="formData.orderId"
          :min="0"
          controls-position="right"
          :placeholder="$t('common.placeholder.input')"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.invoice.userId')" prop="userId">
        <ElInputNumber
          v-model="formData.userId"
          :min="0"
          controls-position="right"
          :placeholder="$t('common.placeholder.input')"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.invoice.invoiceNumber')" prop="invoiceNumber">
        <ElInput v-model="formData.invoiceNumber" :placeholder="$t('common.placeholder.input')" />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.invoice.invoiceType')" prop="invoiceType">
        <ElSelect v-model="formData.invoiceType" :placeholder="$t('common.placeholder.select')">
          <ElOption label="增值税普通发票" value="VAT_GENERAL" />
          <ElOption label="增值税专用发票" value="VAT_SPECIAL" />
          <ElOption label="电子发票" value="ELECTRONIC" />
        </ElSelect>
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.invoice.status')" prop="status">
        <ElSelect v-model="formData.status" :placeholder="$t('common.placeholder.select')">
          <ElOption label="待开" value="PENDING" />
          <ElOption label="已开" value="ISSUED" />
          <ElOption label="已作废" value="CANCELLED" />
        </ElSelect>
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.invoice.buyerName')" prop="buyerName">
        <ElInput v-model="formData.buyerName" :placeholder="$t('common.placeholder.input')" />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.invoice.buyerTaxId')" prop="buyerTaxId">
        <ElInput v-model="formData.buyerTaxId" :placeholder="$t('common.placeholder.input')" />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.invoice.buyerAddress')" prop="buyerAddress">
        <ElInput v-model="formData.buyerAddress" :placeholder="$t('common.placeholder.input')" />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.invoice.buyerPhone')" prop="buyerPhone">
        <ElInput v-model="formData.buyerPhone" :placeholder="$t('common.placeholder.input')" />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.invoice.amount')" prop="amount">
        <ElInputNumber
          v-model="formData.amount"
          :min="0"
          controls-position="right"
          :placeholder="$t('common.placeholder.input')"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.invoice.currency')" prop="currency">
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

import { useCreateInvoice, useUpdateInvoice, fetchGetInvoice } from "@/api/composables";
import { $t } from "@/core/i18n";
import { DRAWER_WIDTH } from "@/constants";
import ProModal from "@/components/Pro/ProModal/index.vue";

const emit = defineEmits<{
  success: [];
}>();

const { mutateAsync: createInvoice } = useCreateInvoice();
const { mutateAsync: updateInvoice } = useUpdateInvoice();

const visible = ref(false);
const submitLoading = ref(false);
const isCreate = ref(true);
const currentId = ref<number>();
const formRef = ref<FormInstance>();

const formData = reactive<{
  orderId: number;
  userId: number;
  invoiceNumber: string;
  invoiceType: string;
  status: string;
  buyerName: string;
  buyerTaxId: string;
  buyerAddress: string;
  buyerPhone: string;
  amount: number;
  currency: string;
}>({
  orderId: 0,
  userId: 0,
  invoiceNumber: "",
  invoiceType: "VAT_GENERAL",
  status: "PENDING",
  buyerName: "",
  buyerTaxId: "",
  buyerAddress: "",
  buyerPhone: "",
  amount: 0,
  currency: "CNY",
});

const formRules: FormRules = {
  invoiceNumber: [{ required: true, message: $t("common.validate.required"), trigger: "blur" }],
  invoiceType: [{ required: true, message: $t("common.validate.required"), trigger: "change" }],
  status: [{ required: true, message: $t("common.validate.required"), trigger: "change" }],
  currency: [{ required: true, message: $t("common.validate.required"), trigger: "blur" }],
};

const title = computed(() =>
  isCreate.value ? $t("pages.mall.invoice.button.create") : $t("pages.mall.invoice.button.update")
);

function open(row?: any) {
  visible.value = true;

  if (row) {
    isCreate.value = false;
    currentId.value = row.id;
    fetchGetInvoice({ id: row.id })
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
  formData.orderId = 0;
  formData.userId = 0;
  formData.invoiceNumber = "";
  formData.invoiceType = "VAT_GENERAL";
  formData.status = "PENDING";
  formData.buyerName = "";
  formData.buyerTaxId = "";
  formData.buyerAddress = "";
  formData.buyerPhone = "";
  formData.amount = 0;
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
      await createInvoice(values);
      ElMessage.success($t("common.notification.createSuccess"));
    } else {
      await updateInvoice({ id: currentId.value!, values });
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
