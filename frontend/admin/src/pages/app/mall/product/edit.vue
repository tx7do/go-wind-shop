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
          <ElButton type="primary" :loading="submitLoading" @click="handleSubmit">
            {{ $t("common.button.confirm") }}
          </ElButton>
        </div>
      </div>
    </ElCard>

    <!-- 步骤条 -->
    <ElCard shadow="never" class="edit-steps">
      <ElSteps :active="currentStep" align-center finish-status="success" process-status="process">
        <ElStep
          v-for="s in steps"
          :key="s.key"
          :title="$t(s.titleKey)"
          class="cursor-pointer"
          @click="handleStepClick(s.index)"
        />
      </ElSteps>
    </ElCard>

    <!-- ── Step 0：基本信息 ──（三步用 v-show 常驻：切步不销毁状态/丢失未保存编辑） -->
    <ElRow v-show="currentStep === 0" :gutter="16" class="edit-body">
      <!-- 左侧：主表单（多语言内容，不含详情富文本——它属第三步） -->
      <ElCol :xl="17" :lg="16" :md="24">
        <ElCard shadow="never" class="edit-card">
          <template #header>
            <div class="flex w-full items-center justify-between">
              <div class="flex items-center gap-3">
                <span>{{ $t("common.section.basic") }}</span>
                <!-- 语言切换器（CMS 范式）：单表单实例，切换即编辑对应语言的翻译；
                       ✓/✗ 标记该语言是否已有翻译（按 name 是否非空判定） -->
                <ElSelect
                  :model-value="editingLocale"
                  size="small"
                  style="width: 180px"
                  @change="handleEditingLocaleChange"
                >
                  <ElOption
                    v-for="lang in languageOptions"
                    :key="lang.code"
                    :value="lang.code"
                    :label="lang.label"
                  >
                    <span class="flex items-center justify-between w-full">
                      <span>{{ lang.label }}</span>
                      <IconifyIcon
                        :icon="lang.hasTranslation ? 'lucide:check' : 'lucide:x'"
                        :class="lang.hasTranslation ? 'text-success' : 'text-warning'"
                        class="ml-2 h-3.5 w-3.5"
                      />
                    </span>
                  </ElOption>
                </ElSelect>
              </div>
            </div>
          </template>
          <ElForm ref="formRef" :model="formData" :rules="formRules" label-width="120px">
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

            <ElFormItem :label="$t('pages.mall.product.shortDescription')" prop="shortDescription">
              <ElInput
                v-model="formData.shortDescription"
                type="textarea"
                :rows="3"
                :placeholder="$t('common.placeholder.input')"
                maxlength="500"
                show-word-limit
              />
            </ElFormItem>
          </ElForm>
        </ElCard>
      </ElCol>

      <!-- 右栏：设置面板 -->
      <ElCol :xl="7" :lg="8" :md="24">
        <ElCard shadow="never" class="edit-card">
          <template #header>
            <span>{{ $t("pages.mall.product.settings") }}</span>
          </template>
          <ElForm :model="formData" :rules="formRules" label-width="100px" label-position="top">
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

    <!-- ── Step 1：销售信息（属性规格 + SKU 明细）── -->
    <StepSalesInfo
      v-show="currentStep === 1"
      ref="salesStepRef"
      :product-id="currentId"
      @loaded="handleSalesLoaded"
    />

    <!-- ── Step 2：图文详情 ── -->
    <ElCard v-show="currentStep === 2" shadow="never" class="edit-card">
      <template #header>
        <div class="flex w-full items-center justify-between">
          <span>{{ $t("pages.mall.product.longDescription") }}</span>
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
      <Editor
        v-model="formData.longDescription"
        :editor-type="EditorType.RICH_TEXT"
        :height="420"
        :placeholder="$t('common.placeholder.input')"
      />
    </ElCard>

    <!-- 底部导航 -->
    <ElCard shadow="never" class="edit-footer">
      <div class="edit-footer__inner">
        <ElButton :disabled="currentStep === 0" @click="currentStep--">
          {{ $t("pages.mall.product.wizard.prev") }}
        </ElButton>
        <ElButton v-if="currentStep < steps.length - 1" type="primary" plain @click="handleNext">
          {{ $t("pages.mall.product.wizard.next") }}
        </ElButton>
        <span class="edit-footer__spacer" />
        <ElButton type="primary" :loading="submitLoading" @click="handleSubmit">
          {{ $t("common.button.confirm") }}
        </ElButton>
      </div>
    </ElCard>
  </div>
</template>

<script lang="ts" setup>
import { computed, reactive, ref, onMounted } from "vue";
import { useRoute, useRouter, onBeforeRouteLeave } from "vue-router";
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from "element-plus";

import {
  useCreateProduct,
  useUpdateProduct,
  fetchGetProduct,
  fetchListProducts,
  fetchListCategories,
  fetchListBrands,
  fetchListLanguages,
} from "@/api/composables";
import { Editor, EditorType } from "@/components/Editor";
import { $t, useI18n } from "@/core/i18n";
import { PaginationQuery } from "@/core/transport/rest";
import { generateAiText } from "@/api/composables/ai-chat";
import { Icon as IconifyIcon } from "@iconify/vue";
import StepSalesInfo from "./steps/StepSalesInfo.vue";

defineOptions({ name: "MallProductEdit" });

const route = useRoute();
const router = useRouter();
const { locale } = useI18n();

const { mutateAsync: createProduct } = useCreateProduct();
const { mutateAsync: updateProduct } = useUpdateProduct();

// 路由约定：/mall/products/create 新建；/mall/products/:id/edit 编辑。
// 新建商品在第一步「下一步/保存」时建档，建档后 currentId 就地更新（向导可继续第二、三步）。
const currentId = ref<number | undefined>(route.params.id ? Number(route.params.id) : undefined);
const isCreate = computed(() => !currentId.value);

const pageTitle = computed(() =>
  isCreate.value ? $t("pages.mall.product.button.create") : $t("pages.mall.product.button.update")
);

// ── 分步向导：基本信息 → 销售信息 → 图文详情 ──
const steps = [
  { key: "basic", titleKey: "pages.mall.product.wizard.basic", index: 0 },
  { key: "sales", titleKey: "pages.mall.product.wizard.sales", index: 1 },
  { key: "rich", titleKey: "pages.mall.product.wizard.rich", index: 2 },
];
const initialStep = Number(route.query.step);
const currentStep = ref(
  Number.isInteger(initialStep) && initialStep >= 0 && initialStep <= 2 ? initialStep : 0
);

const salesStepRef = ref<InstanceType<typeof StepSalesInfo>>();
const submitLoading = ref(false);
const formRef = ref<FormInstance>();

const categoryOptions = ref<{ value: any; label: string }[]>([]);
const brandOptions = ref<{ value: any; label: string }[]>([]);

// ── 语言切换器（CMS 范式）：单一内容表单，editingLocale 决定当前编辑的语言。
//    切语言 = 暂存旧语言表单值进 translations → 回填新语言翻译。
const editingLocale = ref(locale.value);
const languages = ref<any[]>([]);
const languageOptions = computed(() =>
  languages.value.map((l: any) => ({
    code: l.languageCode,
    label: l.nativeName || l.languageName || l.languageCode,
    hasTranslation:
      (formData.translations || []).some(
        (t: any) => t?.languageCode === l.languageCode && !!t?.name
      ) ||
      (isCreate.value && l.languageCode === editingLocale.value),
  }))
);

async function loadLanguages() {
  try {
    const resp = await fetchListLanguages(
      new PaginationQuery({ paging: { page: 1, pageSize: 999 } })
    );
    languages.value = (resp.items || []).filter((l: any) => l.isEnabled === true);
  } catch {
    languages.value = [];
  }
}

// 切语言：先把旧语言的表单值暂存进 translations，再回填新语言翻译。
// 注意不能用 v-model：ElSelect 先行更新 editingLocale 会让下方守卫误判"没切换"，
// 从而跳过暂存/回填（受控写法：:model-value + @change 手动推进）。
function handleEditingLocaleChange(newLocale: string) {
  if (!newLocale || newLocale === editingLocale.value) return;
  const known = languages.value.some((l: any) => l.languageCode === newLocale);
  const target = known ? newLocale : editingLocale.value;
  if (target === editingLocale.value) return;
  syncLocaleIntoTranslations(editingLocale.value);
  editingLocale.value = target;
  loadLocaleIntoForm(target);
}

function loadLocaleIntoForm(code: string) {
  const entry = (formData.translations || []).find((t: any) => t?.languageCode === code);
  formData.name = entry?.name ?? "";
  formData.slug = entry?.slug ?? "";
  formData.shortDescription = entry?.shortDescription ?? "";
  formData.longDescription = entry?.longDescription ?? "";
  formRef.value?.clearValidate();
}

function syncLocaleIntoTranslations(code: string) {
  if (!Array.isArray(formData.translations)) {
    formData.translations = [];
  }
  const empty =
    !formData.name && !formData.slug && !formData.shortDescription && !formData.longDescription;
  let entry = formData.translations.find((t: any) => t?.languageCode === code);
  if (!entry) {
    // 该语言完全没填过内容：不产生空翻译条目（避免提交脏数据）
    if (empty) return;
    entry = { languageCode: code };
    formData.translations.push(entry);
  }
  entry.name = formData.name;
  entry.slug = formData.slug;
  entry.shortDescription = formData.shortDescription;
  entry.longDescription = formData.longDescription;
}

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
  if (
    (shortDescription || longDescription) &&
    !(await ElMessageBox.confirm($t("pages.mall.product.aiOverwriteConfirm"), {
      type: "warning",
    }).catch(() => false))
  ) {
    return;
  }

  const categoryLabel =
    categoryOptions.value.find((c) => c.value === formData.categoryId)?.label ?? "";
  const brandLabel = brandOptions.value.find((b) => b.value === formData.brandId)?.label ?? "";

  aiCopyLoading.value = true;
  try {
    const resp: any = await generateAiText({
      system:
        "你是电商文案专家。根据商品信息输出 JSON 对象：" +
        '{"shortDescription": "60字左右的中文卖点简述", "longDescription": "300字左右的中文商品详情，用简单 HTML 段落书写（只允许 <p> 标签分段，如 "<p>第一段</p><p>第二段</p>"，禁止其他标签与属性）}。' +
        "只输出 JSON 本身，不要 Markdown 代码块标记或任何解释。",
      prompt:
        `商品名称：${name || "（未命名，参考下方提示）"}
` +
        (categoryLabel
          ? `分类：${categoryLabel}
`
          : "") +
        (brandLabel
          ? `品牌：${brandLabel}
`
          : "") +
        (shortDescription
          ? `已有简述（可参考扩展）：${shortDescription}
`
          : ""),
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

// 编辑模式：拉取商品详情回填（translations 全量留存，表单按 editingLocale 取值）
async function loadProduct() {
  if (isCreate.value || !currentId.value) return;
  try {
    const resp: any = await fetchGetProduct({ id: String(currentId.value) });
    formData.translations = Array.isArray(resp.translations) ? resp.translations : [];
    loadLocaleIntoForm(editingLocale.value);
    formData.status = resp.status ?? "PRODUCT_STATUS_DRAFT";
    formData.categoryId = resp.categoryId;
    formData.brandId = resp.brandId;
    formData.sortOrder = resp.sortOrder ?? 1;
    formData.imageUrl = resp.imageUrl ?? "";
    // 回填完成：以当前状态为干净基线
    formBaseline = formSnapshot();
  } catch {
    ElMessage.error($t("common.message.error"));
    handleCancel();
  }
}

// ── 脏数据守卫：回填完成记基线，离开（取消/路由跳转）前比对确认 ──
// 基线 = translations 表单字段 + locale 无关设置字段 + 销售信息（规格/SKU/待删）
let formBaseline = "";

// 销售信息异步回填完成后再重记基线（否则空的 sales 快照 vs 回填后会误报脏）
function handleSalesLoaded() {
  formBaseline = formSnapshot();
}

function formSnapshot(): string {
  const { status, categoryId, brandId, sortOrder, imageUrl, translations } = formData;
  return JSON.stringify({
    basic: { status, categoryId, brandId, sortOrder, imageUrl, translations },
    sales: salesStepRef.value?.snapshot() ?? "",
  });
}

function isFormDirty(): boolean {
  return formBaseline !== "" && formSnapshot() !== formBaseline;
}

async function confirmDiscard(): Promise<boolean> {
  if (!isFormDirty()) return true;
  try {
    await ElMessageBox.confirm($t("pages.mall.product.discardConfirm"), { type: "warning" });
    return true;
  } catch {
    return false;
  }
}

async function handleCancel() {
  if (!(await confirmDiscard())) return;
  router.push("/mall/products");
}

// 路由级守卫：侧边栏/标签页跳转同样拦截（与抽屉四路径守卫同语义）
onBeforeRouteLeave(async () => {
  return await confirmDiscard();
});

// ── 新建商品建档：Create 返回 Empty，用「创建前后 id 差集」反查新商品 ID ──
async function createProductAndGetId(values: Record<string, any>): Promise<number> {
  const beforeQuery = new PaginationQuery({
    paging: { page: 1, pageSize: 50 },
    orderBy: ["-created_at"],
  });
  const before: any = await fetchListProducts(beforeQuery);
  const beforeIds = new Set(((before?.items ?? []) as any[]).map((i) => String(i.id)));

  await createProduct(values);

  const afterQuery = new PaginationQuery({
    paging: { page: 1, pageSize: 10 },
    orderBy: ["-created_at"],
  });
  const after: any = await fetchListProducts(afterQuery);
  const hit = ((after?.items ?? []) as any[]).find((i) => !beforeIds.has(String(i.id)));
  if (!hit?.id) throw new Error("created product not found");
  return Number(hit.id);
}

async function saveProduct(): Promise<boolean> {
  if (!formRef.value) return false;
  try {
    await formRef.value.validate();
  } catch {
    currentStep.value = 0; // 校验失败回落第一步
    return false;
  }

  submitLoading.value = true;
  try {
    // 表单编辑的翻译先落回 translations 对应语言条目，
    // 只提交 Product 消息认识的字段（protojson 严格模式拒绝未知字段）：
    // name/slug/描述属于翻译载荷，不在 Product 本体上提交。
    syncLocaleIntoTranslations(editingLocale.value);
    const values: Record<string, any> = { ...formData };
    delete values.name;
    delete values.slug;
    delete values.shortDescription;
    delete values.longDescription;

    if (isCreate.value) {
      currentId.value = await createProductAndGetId(values);
      ElMessage.success($t("common.notification.createSuccess"));
    } else {
      await updateProduct({ id: currentId.value!, values });
      ElMessage.success($t("common.notification.updateSuccess"));
    }
    formBaseline = formSnapshot();
    return true;
  } catch {
    ElMessage.error(
      isCreate.value
        ? $t("common.notification.createFailed")
        : $t("common.notification.updateFailed")
    );
    return false;
  } finally {
    submitLoading.value = false;
  }
}

// 「下一步」：第一步携带建档（新建商品先落库才有 productId 供销售信息使用）
async function handleNext() {
  if (currentStep.value === 0 && isCreate.value) {
    if (!(await saveProduct())) return;
  }
  currentStep.value++;
}

// 步骤条点跳：未建档时只允许停留在第一步（销售信息/图文详情需要 productId）
async function handleStepClick(index: number) {
  if (index === currentStep.value) return;
  if (index > 0 && isCreate.value) {
    ElMessage.warning($t("pages.mall.product.wizard.needSaveFirst"));
    return;
  }
  currentStep.value = index;
}

// 保存：商品基本信息 → 销售信息（规格/SKU/价格 diff 编排），全成功才跳列表
async function handleSubmit() {
  if (!(await saveProduct())) return;
  if (salesStepRef.value) {
    const salesOk = await salesStepRef.value.persistAll();
    if (!salesOk) {
      currentStep.value = 1;
      return;
    }
  }
  // 已落库：以当前状态为新基线，否则跳转路由时脏守卫会把刚保存的内容误判为未保存改动
  formBaseline = formSnapshot();
  ElMessage.success($t("common.notification.updateSuccess"));
  router.push("/mall/products");
}

onMounted(() => {
  loadOptions();
  loadLanguages();
  loadProduct();
  // 新建模式没有异步回填，直接记基线；编辑模式在 loadProduct 回填后重记
  if (isCreate.value) {
    formBaseline = formSnapshot();
  }
});
</script>

<style lang="scss" scoped>
.product-edit-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

// 基本信息步：左右两栏卡片等高（左栏表单短，拉伸与右栏「设置」卡齐平）
.edit-body {
  display: flex;
  flex-wrap: wrap; // 小屏 md=24 时两栏换行
  flex: 1;

  :deep(.el-col) {
    display: flex;
  }

  .edit-card {
    flex: 1;
    width: 100%;
    margin-bottom: 0;
  }
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

.edit-footer {
  .edit-footer__inner {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .edit-footer__spacer {
    flex: 1;
  }
}
</style>
