<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <!-- 属性值是属性的子资源：不带 attrId 进入时给出引导（否则列表恒空、新增无从归属） -->
    <ElAlert
      v-if="!attributeId"
      :title="$t('pages.mall.productAttributeValue.noAttributeContext')"
      type="info"
      show-icon
      :closable="false"
      class="attr-context-alert"
    >
      <template #default>
        <ElButton
          text
          type="primary"
          size="small"
          @click="router.push('/mall/product-attributes')"
        >
          {{ $t("pages.mall.productAttributeValue.goAttributes") }}
        </ElButton>
      </template>
    </ElAlert>
    <ElAlert
      v-else
      :title="attributeContextTitle"
      type="success"
      show-icon
      :closable="false"
      class="attr-context-alert"
    />

    <ProPage ref="pageRef" :config="pageConfig" @add="handleAdd" @edit="handleEdit">
      <template #displayName="scope: any">
        <span>{{ getTranslationField(scope.row, "displayName") }}</span>
      </template>
    </ProPage>

    <!-- 新增/编辑抽屉 -->
    <ProductAttributeValueDrawer
      ref="drawerRef"
      :attribute-id="attributeId"
      @success="handleSuccess"
    />
  </div>
</template>

<script lang="ts" setup>
import { ref, computed, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import { ElMessage } from "element-plus";

import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";
import ProductAttributeValueDrawer from "./product-attribute-value-drawer.vue";

import { fetchListProductAttributeValues, useDeleteProductAttributeValue, fetchGetProductAttribute } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t, useI18n } from "@/core/i18n";

const { mutateAsync: deleteProductAttributeValue } = useDeleteProductAttributeValue();
const route = useRoute();
const router = useRouter();
const { locale } = useI18n();

const pageRef = ref();
const drawerRef = ref();

const attributeId = computed(() => {
  const v = route.query.attrId;
  const n = Array.isArray(v) ? v[0] : v;
  const parsed = Number(n);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : undefined;
});

// 当前属性上下文展示：带 attrId 进入时标明正在管理哪个属性的值
const attributeName = ref("");
const attributeContextTitle = computed(() => {
  const name = attributeName.value || `#${attributeId.value}`;
  return $t("pages.mall.productAttributeValue.currentAttribute", { name });
});

async function loadAttributeName() {
  attributeName.value = "";
  if (!attributeId.value) return;
  try {
    const resp: any = await fetchGetProductAttribute({ id: attributeId.value });
    attributeName.value = getTranslationField(resp, "name");
  } catch {
    attributeName.value = "";
  }
}

// 当 attrId 变化时重新加载列表与上下文名；immediate 覆盖整页直载（刷新/直链）场景。
// 注意：回调须声明在 attributeName 等状态之后，immediate 首次执行发生在 setup 期间。
watch(
  attributeId,
  () => {
    pageRef.value?.refresh();
    loadAttributeName();
  },
  { immediate: true }
);

// 按 UI 当前语言从 translations 中取对应字段的字面值。
// List 响应含全部语言的 translations，须匹配 languageCode 才能取到
// 当前界面语言的文案；未命中则显示 "-"。
function getTranslationField(row: any, field: string): string {
  const translations = Array.isArray(row?.translations) ? row.translations : [];
  const matched = translations.find(
    (t: any) => t?.languageCode === locale.value
  );
  return matched?.[field] ?? "-";
}

const pageConfig = computed<ProPageConfig>(() => ({
  skeleton: true,
  search: {
    grid: true,
    fields: [],
  },

  table: {
    listAction: async (query: any) => {
      const { page, pageSize } = query;
      if (!attributeId.value) {
        return { items: [], total: 0 };
      }
      const result = await fetchListProductAttributeValues(
        new PaginationQuery({
          paging: { page: page || 1, pageSize: pageSize || 10 },
          formValues: { attribute_id: attributeId.value },
        })
      );
      return { items: result.items || [], total: result.total || 0 };
    },
    deleteAction: async (ids: string) => {
      await deleteProductAttributeValue({ id: ids as any });
    },
    toolbar: [],
    // 无属性上下文时置灰新增（抽屉创建依赖 attributeId 归属），配合顶部引导条
    toolbarRight: attributeId.value
      ? ["add"]
      : [
          {
            name: "add",
            label: "common.button.add",
            attrs: { type: "primary", plain: true },
            disabled: true,
          },
        ],
    defaultToolbar: ["refresh", "exports", "filter"],
    tableAttrs: { border: true, stripe: false },
    columns: [
      { type: "index", label: $t("common.table.seq"), width: 60 },
      {
        prop: "displayName",
        label: $t("pages.mall.productAttributeValue.displayName"),
        minWidth: 160,
        fixed: "left",
        align: "left",
        slotName: "displayName",
      },
      {
        prop: "createdAt",
        label: $t("pages.mall.productAttributeValue.createdAt"),
        minWidth: 160,
        cellType: "date",
        dateFormat: "YYYY-MM-DD HH:mm:ss",
      },
      {
        prop: "action",
        label: $t("common.table.action"),
        fixed: "right",
        width: 160,
        cellType: "tool",
        buttons: [
          { name: "edit", label: $t("common.button.edit"), icon: "lucide:pen-line" },
          {
            name: "delete",
            label: $t("common.button.delete"),
            icon: "lucide:trash-2",
            attrs: { type: "danger" },
          },
        ],
      },
    ],
  },
}));

function handleAdd() {
  if (!attributeId.value) {
    ElMessage.warning($t("pages.mall.productAttributeValue.noAttributeContext"));
    return;
  }
  drawerRef.value?.open();
}

function handleEdit(row: any) {
  drawerRef.value?.open(row);
}

function handleSuccess() {
  pageRef.value?.refresh();
}
</script>

<style lang="scss" scoped>
.app-container {
  padding: 20px;
  width: 100%;
  min-width: 0;
  flex-shrink: 0;
}

.attr-context-alert {
  margin-bottom: 12px;
}
</style>
