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

      <ElFormItem :label="$t('pages.mall.userCoupon.userId')" prop="userId">
        <ElInputNumber
          v-model="formData.userId"
          :min="1"
          controls-position="right"
          :placeholder="$t('common.placeholder.input')"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.userCoupon.couponTemplateId')" prop="couponTemplateId">
        <ElInputNumber
          v-model="formData.couponTemplateId"
          :min="1"
          controls-position="right"
          :placeholder="$t('common.placeholder.input')"
          style="width: 100%"
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

import { useCreateUserCoupon } from "@/api/composables";
import { $t } from "@/core/i18n";
import { DRAWER_WIDTH } from "@/constants";
import ProModal from "@/components/Pro/ProModal/index.vue";

const emit = defineEmits<{
  success: [];
}>();

const { mutateAsync: createUserCoupon } = useCreateUserCoupon();

const visible = ref(false);
const submitLoading = ref(false);
const formRef = ref<FormInstance>();

const formData = reactive<{ userId: number; couponTemplateId: number; status: string }>({
  userId: 0,
  couponTemplateId: 0,
  status: "UNUSED",
});

const formRules: FormRules = {
  userId: [{ required: true, message: $t("common.validate.required"), trigger: "blur" }],
  couponTemplateId: [{ required: true, message: $t("common.validate.required"), trigger: "blur" }],
};

const title = computed(() => $t("pages.mall.userCoupon.button.create"));

function open() {
  visible.value = true;
  resetForm();
}

function handleClose() {
  visible.value = false;
  resetForm();
}

function resetForm() {
  formData.userId = 0;
  formData.couponTemplateId = 0;
  formData.status = "UNUSED";
  formRef.value?.clearValidate();
}

async function handleSubmit() {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
    submitLoading.value = true;

    await createUserCoupon({ ...formData });
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
