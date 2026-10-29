<template>
  <ProModal v-model:visible="visible" :title="title" :config="{ component: 'drawer', drawer: { size: DRAWER_WIDTH } }" :close-guard="guard.allowClose">
    <ElForm ref="formRef" :model="formData" :rules="formRules" label-width="150px" class="drawer-form">
      <ElFormItem :label="$t('pages.mall.memberTier.name')" prop="name">
        <ElInput v-model="formData.name" />
      </ElFormItem>
      <ElFormItem :label="$t('pages.mall.memberTier.level')" prop="level">
        <ElInputNumber v-model="formData.level" :min="1" :step="1" :step-strictly="true" :precision="0" style="width:100%" />
      </ElFormItem>
      <ElFormItem :label="$t('pages.mall.memberTier.minSpend')" prop="minSpendYuan">
        <ElInputNumber v-model="formData.minSpendYuan" :min="0" :precision="0" style="width:100%" />
      </ElFormItem>
      <ElFormItem :label="$t('pages.mall.memberTier.pointsMultiplier')" prop="pointsMultiplier">
        <ElInputNumber v-model="formData.pointsMultiplier" :min="1" :step="1" :step-strictly="true" :precision="0" style="width:100%" />
      </ElFormItem>
      <ElFormItem :label="$t('pages.mall.memberTier.status')" prop="status">
        <ElSelect v-model="formData.status">
          <ElOption label="生效" value="ACTIVE" />
          <ElOption label="停用" value="INACTIVE" />
        </ElSelect>
      </ElFormItem>
    </ElForm>
    <template #footer>
      <div class="drawer-footer">
        <ElButton @click="guard.requestClose(handleClose)">{{ $t("common.button.cancel") }}</ElButton>
        <ElButton type="primary" :loading="submitLoading" @click="handleSubmit">{{ $t("common.button.confirm") }}</ElButton>
      </div>
    </template>
  </ProModal>
</template>

<script lang="ts" setup>
import { computed, reactive, ref, watch } from "vue";
import { ElMessage, type FormInstance, type FormRules } from "element-plus";
import { useCreateMemberTier, useUpdateMemberTier } from "@/api/composables";
import { $t } from "@/core/i18n";
import { DRAWER_WIDTH } from "@/constants";
import ProModal from "@/components/Pro/ProModal/index.vue";
import { useDrawerCloseGuard } from "@/components/Pro/composables/use-drawer-close-guard";

const emit = defineEmits<{ success: [] }>();
const { mutateAsync: createTier } = useCreateMemberTier();
const { mutateAsync: updateTier } = useUpdateMemberTier();
const visible = ref(false);
const submitLoading = ref(false);
const isCreate = ref(true);
const currentId = ref<number>();
const formRef = ref<FormInstance>();

const formData = reactive({ name: "", level: 1, minSpendYuan: 0, pointsMultiplier: 1, status: "ACTIVE" });
const formRules: FormRules = {
  name: [{ required: true, message: $t("common.validate.required"), trigger: "blur" }],
  level: [{ required: true, message: $t("common.validate.required"), trigger: "blur" }],
};

const title = computed(() => isCreate.value ? $t("pages.mall.memberTier.button.create") : $t("pages.mall.memberTier.button.update"));
const guard = useDrawerCloseGuard(() => ({ ...formData }));

function resetForm() {
  formData.name = ""; formData.level = 1; formData.minSpendYuan = 0; formData.pointsMultiplier = 1; formData.status = "ACTIVE";
  formRef.value?.clearValidate();
}

function open(row?: any) {
  visible.value = true;
  if (row) {
    isCreate.value = false; currentId.value = row.id;
    formData.name = row.name ?? ""; formData.level = row.level ?? 1;
    formData.minSpendYuan = (row.minSpend ?? 0) / 100; formData.pointsMultiplier = row.pointsMultiplier ?? 1;
    formData.status = row.status ?? "ACTIVE";
    guard.syncBaseline();
  } else {
    isCreate.value = true; currentId.value = undefined; resetForm();
  }
  guard.syncBaseline();
}

function handleClose() { visible.value = false; resetForm(); }

async function handleSubmit() {
  if (!formRef.value) return;
  try {
    await formRef.value.validate();
    submitLoading.value = true;
    const values = { name: formData.name, level: formData.level, minSpend: Math.round(formData.minSpendYuan * 100), pointsMultiplier: formData.pointsMultiplier, status: formData.status };
    if (isCreate.value) { await createTier(values); ElMessage.success($t("common.notification.createSuccess")); }
    else { await updateTier({ id: currentId.value!, values }); ElMessage.success($t("common.notification.updateSuccess")); }
    emit("success"); handleClose();
  } catch (error) {
    if (error !== false) ElMessage.error(isCreate.value ? $t("common.notification.createFailed") : $t("common.notification.updateFailed"));
  } finally { submitLoading.value = false; }
}

watch(visible, (val) => { if (!val) resetForm(); });
defineExpose({ open });
</script>
