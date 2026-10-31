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

      <ElFormItem :label="$t('pages.mall.userCoupon.status')" prop="status">
        <ElSelect v-model="formData.status" :placeholder="$t('common.placeholder.select')">
          <ElOption :label="$t('pages.mall.userCoupon.statusLabel.UNUSED')" value="UNUSED" />
          <ElOption :label="$t('pages.mall.userCoupon.statusLabel.USED')" value="USED" />
          <ElOption :label="$t('pages.mall.userCoupon.statusLabel.EXPIRED')" value="EXPIRED" />
        </ElSelect>
        <!-- 发券默认"未使用"；选"已核销/已过期"用于线下补录场景 -->
        <div class="form-tip">{{ $t("pages.mall.userCoupon.statusTip") }}</div>
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

import { useCreateUserCoupon } from "@/api/composables";
import { $t } from "@/core/i18n";
import { DRAWER_WIDTH } from "@/constants";
import ProModal from "@/components/Pro/ProModal/index.vue";
import { useDrawerCloseGuard } from "@/components/Pro/composables/use-drawer-close-guard";

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

// 关闭守卫：遮罩/ESC/X/取消统一，脏数据先确认（docs/form_drawer_close_guard.md）
const guard = useDrawerCloseGuard(() => ({ ...formData }));

function open() {
  visible.value = true;
  resetForm();

  // 填充完成后记录基线快照（回填不算脏）
  guard.syncBaseline();
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

.form-tip {
  margin-top: 4px;
  font-size: 12px;
  line-height: 1.5;
  color: var(--el-text-color-secondary);
}

.drawer-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}
</style>
