<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <ProPage ref="pageRef" :config="pageConfig" @add="handleAdd" @edit="handleEdit">
      <template #name="scope: any">
        <span>{{ getTranslationField(scope.row, "name") }}</span>
      </template>
      <template #slug="scope: any">
        <span>{{ getTranslationField(scope.row, "slug") }}</span>
      </template>
      <template #status="scope: any">
        <ElTag size="small" :type="statusTagType(scope.row.status)" effect="plain">
          {{ statusLabel(scope.row.status) }}
        </ElTag>
      </template>
    </ProPage>

    <!-- 新增/编辑抽屉 -->
    <ProductDrawer ref="drawerRef" @success="handleSuccess" />
  </div>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";
import { ElTag } from "element-plus";

import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";
import ProductDrawer from "./product-drawer.vue";

import { fetchListProducts, useDeleteProduct } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t, useI18n } from "@/core/i18n";

const { mutateAsync: deleteProduct } = useDeleteProduct();
const { locale } = useI18n();

const pageRef = ref();
const drawerRef = ref();

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

function statusLabel(status: any): string {
  switch (status) {
    case "PRODUCT_STATUS_ACTIVE":
      return $t("common.status.active");
    case "PRODUCT_STATUS_INACTIVE":
      return $t("common.status.inactive");
    case "PRODUCT_STATUS_DRAFT":
      return "Draft";
    default:
      return $t("common.text.unknown");
  }
}

function statusTagType(status: any): "success" | "info" | "warning" {
  switch (status) {
    case "PRODUCT_STATUS_ACTIVE":
      return "success";
    case "PRODUCT_STATUS_INACTIVE":
      return "info";
    case "PRODUCT_STATUS_DRAFT":
      return "warning";
    default:
      return "info";
  }
}

const pageConfig = computed<ProPageConfig>(() => ({
  skeleton: true,
  search: {
    grid: true,
    fields: [
      {
        type: "input",
        label: $t("pages.mall.product.name"),
        field: "name",
        attrs: { placeholder: $t("common.placeholder.input"), clearable: true },
      },
    ],
  },

  table: {
    listAction: async (query: any) => {
      const { page, pageSize, ...queryParams } = query;
      const result = await fetchListProducts(
        new PaginationQuery({
          paging: { page: page || 1, pageSize: pageSize || 10 },
          formValues: queryParams,
        })
      );
      return { items: result.items || [], total: result.total || 0 };
    },
    deleteAction: async (ids: string) => {
      await deleteProduct({ id: ids as any });
    },
    toolbar: [],
    toolbarRight: ["add"],
    defaultToolbar: ["refresh", "exports", "filter"],
    tableAttrs: { border: true, stripe: false },
    columns: [
      { type: "index", label: $t("common.table.seq"), width: 60 },
      {
        prop: "name",
        label: $t("pages.mall.product.name"),
        minWidth: 160,
        fixed: "left",
        align: "left",
        slotName: "name",
      },
      {
        prop: "slug",
        label: $t("pages.mall.product.slug"),
        minWidth: 120,
        align: "left",
        slotName: "slug",
      },
      {
        prop: "status",
        label: $t("pages.mall.product.status"),
        width: 100,
        slotName: "status",
      },
      {
        prop: "createdAt",
        label: $t("pages.mall.product.createdAt"),
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
</style>
