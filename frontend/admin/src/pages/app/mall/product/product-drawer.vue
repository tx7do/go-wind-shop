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
      <!-- 基本信息 -->
      <ElDivider content-position="left">{{ $t("common.section.basic") }}</ElDivider>

      <ElFormItem :label="$t('pages.mall.product.status')" prop="status">
        <ElSelect
          v-model="formData.status"
          :placeholder="$t('common.placeholder.select')"
          clearable
          style="width: 100%"
        >
          <ElOption label="Draft" value="PRODUCT_STATUS_DRAFT" />
          <ElOption :label="$t('common.status.active')" value="PRODUCT_STATUS_ACTIVE" />
          <ElOption :label="$t('common.status.inactive')" value="PRODUCT_STATUS_INACTIVE" />
        </ElSelect>
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.product.category')" prop="categoryId">
        <ElSelect
          v-model="formData.categoryId"
          :placeholder="$t('common.placeholder.select')"
          clearable
          style="width: 100%"
        >
          <ElOption
            v-for="item in categoryOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.product.brand')" prop="brandId">
        <ElSelect
          v-model="formData.brandId"
          :placeholder="$t('common.placeholder.select')"
          clearable
          style="width: 100%"
        >
          <ElOption
            v-for="item in brandOptions"
            :key="item.value"
            :label="item.label"
            :value="item.value"
          />
        </ElSelect>
      </ElFormItem>

      <ElFormItem :label="$t('common.table.sortOrder')" prop="sortOrder">
        <ElInputNumber
          v-model="formData.sortOrder"
          :min="1"
          :max="9999"
          controls-position="right"
          :placeholder="$t('common.placeholder.input')"
          style="width: 100%"
        />
      </ElFormItem>

      <ElFormItem :label="$t('pages.mall.product.imageUrl')" prop="imageUrl">
        <ElInput
          v-model="formData.imageUrl"
          :placeholder="$t('common.placeholder.input')"
          clearable
        />
      </ElFormItem>
    </ElForm>

    <!-- 多语言翻译 -->
    <TranslationTabs
      v-if="visible && formData.translations"
      v-model="formData.translations"
      :fields="translationFields"
    />

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

import {
  useCreateProduct,
  useUpdateProduct,
  fetchGetProduct,
  fetchListCategories,
  fetchListBrands,
} from "@/api/composables";
import { $t, useI18n } from "@/core/i18n";
import { DRAWER_WIDTH } from "@/constants";
import ProModal from "@/components/Pro/ProModal/index.vue";
import TranslationTabs from "@/components/Pro/TranslationTabs/index.vue";
import type { TranslationFieldConfig } from "@/components/Pro/TranslationTabs/types";
import { PaginationQuery } from "@/core/transport/rest";

const emit = defineEmits<{
  success: [];
}>();

const { mutateAsync: createProduct } = useCreateProduct();
const { mutateAsync: updateProduct } = useUpdateProduct();
const { locale } = useI18n();

const visible = ref(false);
const submitLoading = ref(false);
const isCreate = ref(true);
const currentId = ref<number>();
const formRef = ref<FormInstance>();

const categoryOptions = ref<{ value: any; label: string }[]>([]);
const brandOptions = ref<{ value: any; label: string }[]>([]);

const formData = reactive<{
  status: string;
  categoryId: any;
  brandId: any;
  sortOrder: number;
  imageUrl: string;
  translations: any[];
}>({
  status: "PRODUCT_STATUS_DRAFT",
  categoryId: undefined,
  brandId: undefined,
  sortOrder: 1,
  imageUrl: "",
  translations: [],
});

const formRules: FormRules = {
  status: [{ required: true, message: $t("common.validation.selectRequired"), trigger: "change" }],
};

const translationFields: TranslationFieldConfig[] = [
  { prop: "name", label: $t("pages.mall.product.name"), type: "input" },
  { prop: "slug", label: $t("pages.mall.product.slug"), type: "input" },
  { prop: "shortDescription", label: $t("pages.mall.product.shortDescription"), type: "textarea" },
  { prop: "longDescription", label: $t("pages.mall.product.longDescription"), type: "textarea" },
];

const title = computed(() =>
  isCreate.value ? $t("pages.mall.product.button.create") : $t("pages.mall.product.button.update")
);

// 加载类目/品牌下拉选项
async function loadOptions() {
  try {
    const [catResp, brandResp] = await Promise.all([
      fetchListCategories(new PaginationQuery({ paging: { page: 1, pageSize: 999 } })),
      fetchListBrands(new PaginationQuery({ paging: { page: 1, pageSize: 999 } })),
    ]);
    categoryOptions.value = ((catResp.items || []) as any[]).map((c) => ({
      value: c.id,
      label: getTranslationField(c, "name") || `#${c.id}`,
    }));
    brandOptions.value = ((brandResp.items || []) as any[]).map((b) => ({
      value: b.id,
      label: getTranslationField(b, "name") || `#${b.id}`,
    }));
  } catch (e) {
    categoryOptions.value = [];
    brandOptions.value = [];
  }
}

// 按 UI 当前语言从 translations 中取对应字段的字面值。
// List 响应含全部语言的 translations，须匹配 languageCode 才能取到
// 当前界面语言的文案；未命中返回空串。
function getTranslationField(row: any, field: string): string {
  const translations = Array.isArray(row?.translations) ? row.translations : [];
  const matched = translations.find(
    (t: any) => t?.languageCode === locale.value
  );
  return matched?.[field] ?? "";
}

function open(row?: any) {
  visible.value = true;
  loadOptions();

  if (row) {
    isCreate.value = false;
    currentId.value = row.id;
    fetchGetProduct({ id: row.id })
      .then((resp: any) => {
        Object.assign(formData, resp);
        if (!Array.isArray(formData.translations)) {
          formData.translations = [];
        }
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
  formData.status = "PRODUCT_STATUS_DRAFT";
  formData.categoryId = undefined;
  formData.brandId = undefined;
  formData.sortOrder = 1;
  formData.imageUrl = "";
  formData.translations = [];
  formRef.value?.clearValidate();
}

async function handleSubmit() {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
    submitLoading.value = true;

    const values = { ...formData };

    if (isCreate.value) {
      await createProduct(values);
      ElMessage.success($t("common.notification.createSuccess"));
    } else {
      await updateProduct({ id: currentId.value!, values });
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
