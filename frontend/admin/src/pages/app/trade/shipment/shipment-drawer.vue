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
      <ElDivider content-position="left">{{ $t("common.section.basic") }}</ElDivider>

      <ElFormItem :label="$t('pages.mall.shipment.create.orderId')" prop="orderId">
        <ElSelect
          v-model="formData.orderId"
          :placeholder="$t('pages.mall.shipment.create.orderIdPlaceholder')"
          clearable
          filterable
          style="width: 100%"
        >
          <ElOption
            v-for="item in orderOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.shipment.create.carrier')" prop="carrier">
        <ElInput
          v-model="formData.carrier"
          :placeholder="$t('pages.mall.shipment.create.carrierPlaceholder')"
          clearable
        />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.shipment.create.trackingNumber')" prop="trackingNumber">
        <ElInput
          v-model="formData.trackingNumber"
          :placeholder="$t('pages.mall.shipment.create.trackingNumberPlaceholder')"
          clearable
        />
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

import { useCreateShipment, fetchListOrders } from "@/api/composables";
import { $t } from "@/core/i18n";
import { DRAWER_WIDTH } from "@/constants";
import ProModal from "@/components/Pro/ProModal/index.vue";
import { PaginationQuery } from "@/core/transport/rest";

const emit = defineEmits<{
  success: [];
}>();

const { mutateAsync: createShipment } = useCreateShipment();

const visible = ref(false);
const submitLoading = ref(false);
const formRef = ref<FormInstance>();

const formData = reactive<{ orderId: number | undefined; carrier: string; trackingNumber: string }>({
  orderId: undefined,
  carrier: "",
  trackingNumber: "",
});

const formRules: FormRules = {
  orderId: [{ required: true, message: $t("common.placeholder.select"), trigger: "change" }],
  carrier: [{ required: true, message: $t("common.placeholder.input"), trigger: "blur" }],
  trackingNumber: [{ required: true, message: $t("common.placeholder.input"), trigger: "blur" }],
};

const orderOptions = ref<{ value: number; label: string }[]>([]);

const title = computed(() => $t("pages.mall.shipment.create.title"));

// 加载待发货（PAID）订单作为可关联选项
async function loadOptions() {
  try {
    const resp = await fetchListOrders(
      new PaginationQuery({
        paging: { page: 1, pageSize: 200 },
        formValues: { status: "PAID" },
      })
    );
    orderOptions.value = ((resp.items || []) as any[]).map((o) => ({
      value: o.id,
      label: `#${o.id}`,
    }));
  } catch {
    orderOptions.value = [];
  }
}

function open() {
  visible.value = true;
  resetForm();
  loadOptions();
}

function handleClose() {
  visible.value = false;
  resetForm();
}

function resetForm() {
  formData.orderId = undefined;
  formData.carrier = "";
  formData.trackingNumber = "";
  formRef.value?.clearValidate();
}

async function handleSubmit() {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
    submitLoading.value = true;

    await createShipment({
      data: {
        orderId: formData.orderId,
        carrier: formData.carrier,
        trackingNumber: formData.trackingNumber,
      },
    });
    ElMessage.success($t("common.notification.createSuccess"));

    emit("success");
    handleClose();
  } catch (error) {
    if (error !== false) {
      ElMessage.error($t("common.notification.createFailed"));
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
