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

      <ElFormItem :label="$t('pages.mall.seckill.name')" prop="name">
        <ElInput v-model="formData.name" :placeholder="$t('common.placeholder.input')" />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.seckill.skuId')" prop="skuId">
        <ElInputNumber
          v-model="formData.skuId"
          :min="1"
          :step="1"
          :step-strictly="true"
          :precision="0"
          controls-position="right"
          :placeholder="$t('common.placeholder.input')"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.seckill.seckillPrice')" prop="seckillPriceYuan">
        <!-- 元输入，提交时转分（后端为最小货币单位） -->
        <ElInputNumber
          v-model="formData.seckillPriceYuan"
          :min="0.01"
          :precision="2"
          controls-position="right"
          :placeholder="$t('common.placeholder.input')"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.seckill.totalStock')" prop="totalStock">
        <ElInputNumber
          v-model="formData.totalStock"
          :min="1"
          :step="1"
          :step-strictly="true"
          :precision="0"
          controls-position="right"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.seckill.perUserLimit')" prop="perUserLimit">
        <ElInputNumber
          v-model="formData.perUserLimit"
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
          :start-placeholder="$t('common.placeholder.input')"
          :end-placeholder="$t('common.placeholder.input')"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.seckill.status')" prop="status">
        <ElSelect v-model="formData.status" :placeholder="$t('common.placeholder.select')">
          <ElOption :label="$t('pages.mall.seckill.statusLabel.active')" value="ACTIVE" />
          <ElOption :label="$t('pages.mall.seckill.statusLabel.inactive')" value="INACTIVE" />
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

import {
  useCreateSeckillActivity,
  useUpdateSeckillActivity,
} from "@/api/composables";
import { $t } from "@/core/i18n";
import { DRAWER_WIDTH } from "@/constants";
import ProModal from "@/components/Pro/ProModal/index.vue";
import { useDrawerCloseGuard } from "@/components/Pro/composables/use-drawer-close-guard";

const emit = defineEmits<{
  success: [];
}>();

const { mutateAsync: createSeckill } = useCreateSeckillActivity();
const { mutateAsync: updateSeckill } = useUpdateSeckillActivity();

const visible = ref(false);
const submitLoading = ref(false);
const isCreate = ref(true);
const currentId = ref<number>();
const formRef = ref<FormInstance>();

// 编辑回填基线（时间窗口字符串数组）
const formData = reactive<{
  name: string;
  skuId: number;
  seckillPriceYuan: number;
  totalStock: number;
  perUserLimit: number;
  validRange: [string, string] | null;
  status: string;
}>({
  name: "",
  skuId: 1,
  seckillPriceYuan: 1,
  totalStock: 100,
  perUserLimit: 1,
  validRange: null,
  status: "ACTIVE",
});

const formRules: FormRules = {
  name: [{ required: true, message: $t("common.validate.required"), trigger: "blur" }],
  skuId: [{ required: true, message: $t("common.validate.required"), trigger: "blur" }],
  status: [{ required: true, message: $t("common.validate.required"), trigger: "change" }],
};

const title = computed(() =>
  isCreate.value ? $t("pages.mall.seckill.button.create") : $t("pages.mall.seckill.button.update")
);

// 关闭守卫：遮罩/ESC/X/取消统一，脏数据先确认（docs/form_drawer_close_guard.md）
const guard = useDrawerCloseGuard(() => ({ ...formData }));

function resetForm() {
  formData.name = "";
  formData.skuId = 1;
  formData.seckillPriceYuan = 1;
  formData.totalStock = 100;
  formData.perUserLimit = 1;
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
    formData.seckillPriceYuan = (row.seckillPrice ?? 0) / 100;
    formData.totalStock = row.totalStock ?? 100;
    formData.perUserLimit = row.perUserLimit ?? 1;
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
      seckillPrice: Math.round(formData.seckillPriceYuan * 100),
      totalStock: formData.totalStock,
      perUserLimit: formData.perUserLimit,
      status: formData.status,
    };
    if (formData.validRange && formData.validRange.length === 2) {
      values.validFrom = formData.validRange[0];
      values.validUntil = formData.validRange[1];
    }

    if (isCreate.value) {
      await createSeckill(values);
      ElMessage.success($t("common.notification.createSuccess"));
    } else {
      await updateSeckill({ id: currentId.value!, values });
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
