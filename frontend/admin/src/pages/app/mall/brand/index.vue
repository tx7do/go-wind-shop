<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <ProPage ref="pageRef" :config="pageConfig" @add="handleAdd" @edit="handleEdit">
      <template #name="scope: any">
        <span>{{ getTranslationField(scope.row, "name") }}</span>
      </template>
      <template #slug="scope: any">
        <span>{{ getTranslationField(scope.row, "slug") }}</span>
      </template>
    </ProPage>

    <!-- 新增/编辑抽屉 -->
    <BrandDrawer ref="drawerRef" @success="handleSuccess" />
  </div>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";

import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";
import BrandDrawer from "./brand-drawer.vue";

import { fetchListBrands, useDeleteBrand } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t, useI18n } from "@/core/i18n";

const { mutateAsync: deleteBrand } = useDeleteBrand();
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

const pageConfig = computed<ProPageConfig>(() => ({
  skeleton: true,
  search: {
    grid: true,
    fields: [
      {
        type: "input",
        label: $t("pages.mall.brand.name"),
        field: "name",
        attrs: { placeholder: $t("common.placeholder.input"), clearable: true },
      },
    ],
  },

  table: {
    listAction: async (query: any) => {
      const { page, pageSize, ...queryParams } = query;
      const result = await fetchListBrands(
        new PaginationQuery({
          paging: { page: page || 1, pageSize: pageSize || 10 },
          formValues: queryParams,
        })
      );
      return { items: result.items || [], total: result.total || 0 };
    },
    deleteAction: async (ids: string) => {
      await deleteBrand({ id: ids as any });
    },
    toolbar: [],
    toolbarRight: ["add"],
    defaultToolbar: ["refresh", "exports", "filter"],
    tableAttrs: { border: true, stripe: false },
    columns: [
      { type: "index", label: $t("common.table.seq"), width: 60 },
      {
        prop: "name",
        label: $t("pages.mall.brand.name"),
        minWidth: 160,
        fixed: "left",
        align: "left",
        slotName: "name",
      },
      {
        prop: "slug",
        label: $t("pages.mall.brand.slug"),
        minWidth: 120,
        align: "left",
        slotName: "slug",
      },
      {
        prop: "createdAt",
        label: $t("pages.mall.brand.createdAt"),
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
