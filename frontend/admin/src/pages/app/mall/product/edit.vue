<template>
  <div class="app-container product-edit-page">
    <!-- 顶栏：标题 + 操作 -->
    <ElCard shadow="never" class="edit-header">
      <div class="edit-header__inner">
        <div class="edit-header__title">
          <ElButton text @click="handleCancel">
            <IconifyIcon icon="lucide:arrow-left" />
          </ElButton>
          <span class="edit-header__text">{{ pageTitle }}</span>
          <ElTag v-if="!isCreate" size="small" type="info">#{{ currentId }}</ElTag>
        </div>
        <div class="edit-header__actions">
          <ElButton @click="handleCancel">{{ $t("common.button.cancel") }}</ElButton>
          <ElButton
            type="primary"
            :loading="submitLoading"
            @click="handleSubmit"
          >
            {{ $t("common.button.confirm") }}
          </ElButton>
        </div>
      </div>
    </ElCard>

    <ElRow :gutter="16" class="edit-body">
      <!-- 左侧：主表单（多语言内容） -->
      <ElCol :xl="17" :lg="16" :md="24">
        <ElCard shadow="never" class="edit-card">
          <template #header>
            <div class="flex w-full items-center justify-between">
              <span>{{ $t("common.section.basic") }}</span>
              <ElButton
                size="small"
                type="primary"
                plain
                :loading="aiCopyLoading"
                @click="handleAiGenerateCopy"
              >
                <IconifyIcon icon="lucide:sparkles" class="mr-1 h-4 w-4" />
                {{ $t("pages.mall.product.aiGenerate") }}
              </ElButton>
            </div>
          </template>
          <ElForm
            ref="formRef"
            :model="formData"
            :rules="formRules"
            label-width="120px"
          >
            <ElFormItem :label="$t('pages.mall.product.name')" prop="name">
              <ElInput
                v-model="formData.name"
                :placeholder="$t('common.placeholder.input')"
                maxlength="200"
              />
            </ElFormItem>

            <ElFormItem :label="$t('pages.mall.product.slug')" prop="slug">
              <ElInput
                v-model="formData.slug"
                :placeholder="$t('common.placeholder.input')"
                maxlength="200"
              />
            </ElFormItem>

            <ElFormItem
              :label="$t('pages.mall.product.shortDescription')"
              prop="shortDescription"
            >
              <ElInput
                v-model="formData.shortDescription"
                type="textarea"
                :rows="3"
                :placeholder="$t('common.placeholder.input')"
                maxlength="500"
                show-word-limit
              />
            </ElFormItem>

            <ElFormItem
              :label="$t('pages.mall.product.longDescription')"
              prop="longDescription"
            >
              <ElInput
                v-model="formData.longDescription"
                type="textarea"
                :rows="8"
                :placeholder="$t('common.placeholder.input')"
              />
            </ElFormItem>
          </ElForm>

          <!-- 多语言翻译 -->
          <TranslationTabs
            v-if="formData.translations"
            v-model="formData.translations"
            :fields="translationFields"
          />
        </ElCard>
      </ElCol>

      <!-- 右栏：设置面板 -->
      <ElCol :xl="7" :lg="8" :md="24">
        <ElCard shadow="never" class="edit-card">
          <template #header>
            <span>{{ $t("pages.mall.product.settings") }}</span>
          </template>
          <ElForm
            :model="formData"
            :rules="formRules"
            label-width="100px"
            label-position="top"
          >
            <ElFormItem :label="$t('pages.mall.product.status')" prop="status">
              <ElSelect
                v-model="formData.status"
                :placeholder="$t('common.placeholder.select')"
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
                filterable
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
                filterable
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
        </ElCard>
      </ElCol>
    </ElRow>
  </div>
</template>

<script lang="ts" setup>
import { computed, reactive, ref, onMounted } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from "element-plus";

import {
  useCreateProduct,
  useUpdateProduct,
  fetchGetProduct,
  fetchListCategories,
  fetchListBrands,
} from "@/api/composables";
import { $t, useI18n } from "@/core/i18n";
import TranslationTabs from "@/components/Pro/TranslationTabs/index.vue";
import type { TranslationFieldConfig } from "@/components/Pro/TranslationTabs/types";
import { PaginationQuery } from "@/core/transport/rest";
import { generateAiText } from "@/api/composables/ai-chat";
import { Icon as IconifyIcon } from "@iconify/vue";

defineOptions({ name: "MallProductEdit" });

const route = useRoute();
const router = useRouter();
const { locale } = useI18n();

const { mutateAsync: createProduct } = useCreateProduct();
const { mutateAsync: updateProduct } = useUpdateProduct();

// 路由约定：/mall/products/create 新建；/mall/products/:id/edit 编辑
const routeId = route.params.id as string | undefined;
const isCreate = !routeId;
const currentId = routeId ? Number(routeId) : undefined;

const pageTitle = computed(() =>
  isCreate
    ? $t("pages.mall.product.button.create")
    : $t("pages.mall.product.button.update")
);

const submitLoading = ref(false);
const formRef = ref<FormInstance>();

const categoryOptions = ref<{ value: any; label: string }[]>([]);
const brandOptions = ref<{ value: any; label: string }[]>([]);

const formData = reactive<{
  name: string;
  slug: string;
  shortDescription: string;
  longDescription: string;
  status: string;
  categoryId: any;
  brandId: any;
  sortOrder: number;
  imageUrl: string;
  translations: any[];
}>({
  name: "",
  slug: "",
  shortDescription: "",
  longDescription: "",
  status: "PRODUCT_STATUS_DRAFT",
  categoryId: undefined,
  brandId: undefined,
  sortOrder: 1,
  imageUrl: "",
  translations: [],
});

const formRules: FormRules = {
  name: [{ required: true, message: $t("common.validation.required"), trigger: "blur" }],
  status: [{ required: true, message: $t("common.validation.selectRequired"), trigger: "change" }],
};

// ── AI 文案生成：据名称/分类/品牌生成简述与详情，填入表单 ──
const aiCopyLoading = ref(false);

async function handleAiGenerateCopy() {
  const name = formData.name.trim();
  const shortDescription = formData.shortDescription.trim();
  const longDescription = formData.longDescription.trim();
  if (!name && !shortDescription) {
    ElMessage.warning($t("pages.mall.product.aiNeedName"));
    return;
  }
  if ((shortDescription || longDescription) &&
      !(await ElMessageBox.confirm($t("pages.mall.product.aiOverwriteConfirm"), { type: "warning" }).catch(() => false))) {
    return;
  }

  const categoryLabel = categoryOptions.value.find((c) => c.value === formData.categoryId)?.label ?? "";
  const brandLabel = brandOptions.value.find((b) => b.value === formData.brandId)?.label ?? "";

  aiCopyLoading.value = true;
  try {
    const resp: any = await generateAiText({
      system:
        "你是电商文案专家。根据商品信息输出 JSON 对象：" +
        '{"shortDescription": "60字左右的中文卖点简述", "longDescription": "300字左右的中文商品详情，可用短段落"}。' +
        "只输出 JSON 本身，不要 Markdown 代码块标记或任何解释。",
      prompt:
        `商品名称：${name || "（未命名，参考下方提示）"}
` +
        (categoryLabel ? `分类：${categoryLabel}
` : "") +
        (brandLabel ? `品牌：${brandLabel}
` : "") +
        (shortDescription ? `已有简述（可参考扩展）：${shortDescription}
` : ""),
      maxTokens: 1200,
    });

    const text = String(resp?.text ?? "");
    const begin = text.indexOf("{");
    const end = text.lastIndexOf("}");
    if (!text || begin < 0 || end <= begin) {
      throw new Error("invalid ai output");
    }
    const parsed = JSON.parse(text.slice(begin, end + 1));
    if (parsed.shortDescription) formData.shortDescription = String(parsed.shortDescription);
    if (parsed.longDescription) formData.longDescription = String(parsed.longDescription);
    ElMessage.success($t("pages.mall.product.aiGenerateSuccess"));
  } catch {
    ElMessage.error($t("pages.mall.product.aiGenerateFailed"));
  } finally {
    aiCopyLoading.value = false;
  }
}

const translationFields: TranslationFieldConfig[] = [
  { prop: "name", label: $t("pages.mall.product.name"), type: "input" },
  { prop: "slug", label: $t("pages.mall.product.slug"), type: "input" },
  { prop: "shortDescription", label: $t("pages.mall.product.shortDescription"), type: "textarea" },
  { prop: "longDescription", label: $t("pages.mall.product.longDescription"), type: "textarea" },
];

// 加载类目/品牌下拉选项（label 按 UI 当前语言取翻译，未命中回落 #id）
function getTranslationField(row: any, field: string): string {
  const translations = Array.isArray(row?.translations) ? row.translations : [];
  const matched = translations.find((t: any) => t?.languageCode === locale.value);
  return matched?.[field] ?? "";
}

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
  } catch {
    categoryOptions.value = [];
    brandOptions.value = [];
  }
}

// 编辑模式：拉取商品详情回填
async function loadProduct() {
  if (isCreate || !currentId) return;
  try {
    const resp: any = await fetchGetProduct({ id: currentId });
    formData.name = getTranslationField(resp, "name");
    formData.slug = getTranslationField(resp, "slug") || resp.slug || "";
    formData.shortDescription = getTranslationField(resp, "shortDescription");
    formData.longDescription = getTranslationField(resp, "longDescription");
    formData.status = resp.status ?? "PRODUCT_STATUS_DRAFT";
    formData.categoryId = resp.categoryId;
    formData.brandId = resp.brandId;
    formData.sortOrder = resp.sortOrder ?? 1;
    formData.imageUrl = resp.imageUrl ?? "";
    formData.translations = Array.isArray(resp.translations) ? resp.translations : [];
  } catch {
    ElMessage.error($t("common.message.error"));
    handleCancel();
  }
}

function handleCancel() {
  router.push("/mall/products");
}

async function handleSubmit() {
  if (!formRef.value) return;

  try {
    await formRef.value.validate();
    submitLoading.value = true;

    // 主表单的 name/slug/描述 属当前语言——同步回 translations 对应语言条目后，
    // 只提交 Product 消息认识的字段（protojson 严格模式拒绝未知字段）。
    syncCurrentLocaleIntoTranslations();
    const { name: _n, slug: _s, shortDescription: _sd, longDescription: _ld, ...values } = {
      ...formData,
    };

    if (isCreate) {
      await createProduct(values);
      ElMessage.success($t("common.notification.createSuccess"));
    } else {
      await updateProduct({ id: currentId!, values });
      ElMessage.success($t("common.notification.updateSuccess"));
    }

    router.push("/mall/products");
  } catch (error) {
    if (error !== false) {
      ElMessage.error(
        isCreate
          ? $t("common.notification.createFailed")
          : $t("common.notification.updateFailed")
      );
    }
  } finally {
    submitLoading.value = false;
  }
}


// 主表单（当前语言）文本同步回 translations 对应语言条目，提交时随 translations 一起发送。
function syncCurrentLocaleIntoTranslations() {
  if (!Array.isArray(formData.translations)) {
    formData.translations = [];
  }
  let entry = formData.translations.find((t: any) => t?.languageCode === locale.value);
  if (!entry) {
    entry = { languageCode: locale.value };
    formData.translations.push(entry);
  }
  entry.name = formData.name;
  entry.slug = formData.slug;
  entry.shortDescription = formData.shortDescription;
  entry.longDescription = formData.longDescription;
}

onMounted(() => {
  loadOptions();
  loadProduct();
});
</script>

<style lang="scss" scoped>
.product-edit-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.edit-header {
  .edit-header__inner {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .edit-header__title {
    display: flex;
    align-items: center;
    gap: 8px;

    .edit-header__text {
      font-size: 16px;
      font-weight: 600;
    }
  }

  .edit-header__actions {
    display: flex;
    gap: 10px;
  }
}

.edit-body {
  flex: 1;
}

.edit-card {
  margin-bottom: 16px;
}
</style>
