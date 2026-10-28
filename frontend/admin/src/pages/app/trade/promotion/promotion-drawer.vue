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

      <ElFormItem :label="$t('pages.mall.promotion.name')" prop="name">
        <ElInput v-model="formData.name" :placeholder="$t('common.placeholder.input')" />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.promotion.promoType')" prop="promoType">
        <ElSelect v-model="formData.promoType" :placeholder="$t('common.placeholder.select')">
          <ElOption
            :label="$t('pages.mall.promotion.promoTypeLabel.SPECIAL_PRICE')"
            value="SPECIAL_PRICE"
          />
          <ElOption
            :label="$t('pages.mall.promotion.promoTypeLabel.FULL_REDUCTION')"
            value="FULL_REDUCTION"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem
        v-if="formData.promoType === 'SPECIAL_PRICE'"
        :label="$t('pages.mall.promotion.productId')"
        prop="productId"
      >
        <ElInputNumber
          v-model="formData.productId"
          :min="1"
          :controls="false"
          :placeholder="$t('common.placeholder.input')"
          class="!w-full"
        />
        <div class="form-tip">{{ $t("pages.mall.promotion.productIdTip") }}</div>
      </ElFormItem>

      <ElFormItem
        v-if="formData.promoType === 'SPECIAL_PRICE'"
        :label="$t('pages.mall.promotion.specialPrice')"
        prop="specialPrice"
      >
        <ElInputNumber
          v-model="specialPriceYuan"
          :min="0.01"
          :precision="2"
          :step="1"
          :controls="false"
          class="!w-full"
        />
        <div class="form-tip">{{ $t("pages.mall.promotion.amountUnitTip") }}</div>
      </ElFormItem>

      <template v-if="formData.promoType === 'FULL_REDUCTION'">
        <ElFormItem :label="$t('pages.mall.promotion.thresholdAmount')" prop="thresholdAmount">
          <ElInputNumber
            v-model="thresholdYuan"
            :min="0.01"
            :precision="2"
            :step="1"
            :controls="false"
            class="!w-full"
          />
          <div class="form-tip">{{ $t("pages.mall.promotion.amountUnitTip") }}</div>
        </ElFormItem>

        <ElFormItem :label="$t('pages.mall.promotion.reductionAmount')" prop="reductionAmount">
          <ElInputNumber
            v-model="reductionYuan"
            :min="0.01"
            :precision="2"
            :step="1"
            :controls="false"
            class="!w-full"
          />
        </ElFormItem>
      </template>

      <ElFormItem :label="$t('pages.mall.promotion.currency')" prop="currency">
        <ElSelect v-model="formData.currency" :placeholder="$t('common.placeholder.select')">
          <ElOption label="CNY" value="CNY" />
        </ElSelect>
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.promotion.validity')" prop="validRange">
        <ElDatePicker
          v-model="validRange"
          type="datetimerange"
          value-format="x"
          :start-placeholder="$t('pages.mall.promotion.validFrom')"
          :end-placeholder="$t('pages.mall.promotion.validUntil')"
          class="!w-full"
        />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.promotion.priority')" prop="priority">
        <ElInputNumber
          v-model="formData.priority"
          :min="0"
          :max="9999"
          :controls="false"
          class="!w-full"
        />
        <div class="form-tip">{{ $t("pages.mall.promotion.priorityTip") }}</div>
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.promotion.status')" prop="status">
        <ElSelect v-model="formData.status" :placeholder="$t('common.placeholder.select')">
          <ElOption :label="$t('pages.mall.promotion.statusLabel.ACTIVE')" value="ACTIVE" />
          <ElOption :label="$t('pages.mall.promotion.statusLabel.INACTIVE')" value="INACTIVE" />
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

import { useCreatePromotion, useUpdatePromotion, fetchGetPromotion } from "@/api/composables";
import { $t } from "@/core/i18n";
import { DRAWER_WIDTH } from "@/constants";
import ProModal from "@/components/Pro/ProModal/index.vue";
import { useDrawerCloseGuard } from "@/components/Pro/composables/use-drawer-close-guard";

const emit = defineEmits<{
  success: [];
}>();

const { mutateAsync: createPromotion } = useCreatePromotion();
const { mutateAsync: updatePromotion } = useUpdatePromotion();

const visible = ref(false);
const submitLoading = ref(false);
const isCreate = ref(true);
const currentId = ref<number>();
const formRef = ref<FormInstance>();

const formData = reactive<{
  name: string;
  promoType: string;
  productId: number | undefined;
  specialPrice: number;
  thresholdAmount: number;
  reductionAmount: number;
  priority: number;
  currency: string;
  status: string;
}>({
  name: "",
  promoType: "SPECIAL_PRICE",
  productId: undefined,
  specialPrice: 0,
  thresholdAmount: 0,
  reductionAmount: 0,
  priority: 0,
  currency: "CNY",
  status: "ACTIVE",
});

// 有效窗口（毫秒时间戳数组），与后端 Timestamp 字段互转
const validRange = ref<[number, number] | null>(null);

// 金额展示用元、存储用分（后端 amount 单位为最小货币单位）
const specialPriceYuan = computed({
  get: () => (formData.specialPrice ? formData.specialPrice / 100 : undefined),
  set: (v?: number) => {
    formData.specialPrice = v ? Math.round(v * 100) : 0;
  },
});
const thresholdYuan = computed({
  get: () => (formData.thresholdAmount ? formData.thresholdAmount / 100 : undefined),
  set: (v?: number) => {
    formData.thresholdAmount = v ? Math.round(v * 100) : 0;
  },
});
const reductionYuan = computed({
  get: () => (formData.reductionAmount ? formData.reductionAmount / 100 : undefined),
  set: (v?: number) => {
    formData.reductionAmount = v ? Math.round(v * 100) : 0;
  },
});

const formRules: FormRules = {
  name: [{ required: true, message: $t("common.validate.required"), trigger: "blur" }],
  promoType: [{ required: true, message: $t("common.validate.required"), trigger: "change" }],
  currency: [{ required: true, message: $t("common.validate.required"), trigger: "change" }],
  status: [{ required: true, message: $t("common.validate.required"), trigger: "change" }],
};

const title = computed(() =>
  isCreate.value
    ? $t("pages.mall.promotion.button.create")
    : $t("pages.mall.promotion.button.update")
);

// 关闭守卫：遮罩/ESC/X/取消统一，脏数据先确认（docs/form_drawer_close_guard.md）
// validRange 在 formData 之外（有效窗口 ref），一并纳入脏判定
const guard = useDrawerCloseGuard(() => ({ ...formData, validRange: validRange.value }));

function open(row?: any) {
  visible.value = true;

  if (row) {
    isCreate.value = false;
    currentId.value = row.id;
    fetchGetPromotion({ id: row.id })
      .then((resp: any) => {
        formData.name = resp.name ?? "";
        formData.promoType = resp.promoType ?? "SPECIAL_PRICE";
        formData.productId = resp.productId || undefined;
        formData.specialPrice = Number(resp.specialPrice ?? 0);
        formData.thresholdAmount = Number(resp.thresholdAmount ?? 0);
        formData.reductionAmount = Number(resp.reductionAmount ?? 0);
        formData.priority = resp.priority ?? 0;
        formData.currency = resp.currency ?? "CNY";
        formData.status = resp.status ?? "ACTIVE";
        if (resp.validFrom && resp.validUntil) {
          validRange.value = [
            new Date(resp.validFrom).getTime(),
            new Date(resp.validUntil).getTime(),
          ];
        } else {
          validRange.value = null;
        }
        guard.syncBaseline();
      })
      .catch(() => {
        ElMessage.error($t("common.message.error"));
      });
  } else {
    isCreate.value = true;
    currentId.value = undefined;
  }

  // 填充完成后记录基线快照（回填不算脏；编辑异步回填在 then 里重记）
  guard.syncBaseline();
}

function handleClose() {
  visible.value = false;
  resetForm();
}

function resetForm() {
  formData.name = "";
  formData.promoType = "SPECIAL_PRICE";
  formData.productId = undefined;
  formData.specialPrice = 0;
  formData.thresholdAmount = 0;
  formData.reductionAmount = 0;
  formData.priority = 0;
  formData.currency = "CNY";
  formData.status = "ACTIVE";
  validRange.value = null;
  formRef.value?.clearValidate();
}

async function handleSubmit() {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
    submitLoading.value = true;

    const values: Record<string, any> = {
      name: formData.name,
      promoType: formData.promoType,
      priority: formData.priority,
      currency: formData.currency,
      status: formData.status,
    };
    if (formData.promoType === "SPECIAL_PRICE") {
      values.productId = formData.productId;
      values.specialPrice = formData.specialPrice;
    }
    if (formData.promoType === "FULL_REDUCTION") {
      values.thresholdAmount = formData.thresholdAmount;
      values.reductionAmount = formData.reductionAmount;
      // 店铺级活动清空商品绑定
      values.productId = 0;
    }
    if (validRange.value) {
      values.validFrom = new Date(validRange.value[0]).toISOString();
      values.validUntil = new Date(validRange.value[1]).toISOString();
    }

    if (isCreate.value) {
      await createPromotion(values);
      ElMessage.success($t("common.notification.createSuccess"));
    } else {
      await updatePromotion({ id: currentId.value!, values });
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
