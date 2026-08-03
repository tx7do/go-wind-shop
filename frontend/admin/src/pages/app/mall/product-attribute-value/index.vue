<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <ProPage ref="pageRef" :config="pageConfig" @add="handleAdd" @edit="handleEdit">
      <template #displayName="scope: any">
        <span>{{ getFirstTranslationField(scope.row, "displayName") }}</span>
      </template>
    </ProPage>

    <!-- 新增/编辑抽屉 -->
    <ProductAttributeValueDrawer ref="drawerRef" :attribute-id="attributeId" @success="handleSuccess" />
  </div>
</template>

<script lang="ts" setup>
import { ref, computed, watch } from "vue";
import { useRoute } from "vue-router";

import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";
import ProductAttributeValueDrawer from "./product-attribute-value-drawer.vue";

import { fetchListProductAttributeValues, useDeleteProductAttributeValue } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t } from "@/core/i18n";

const { mutateAsync: deleteProductAttributeValue } = useDeleteProductAttributeValue();
const route = useRoute();

const pageRef = ref();
const drawerRef = ref();

const attributeId = computed(() => {
  const v = route.query.attrId;
  const n = Array.isArray(v) ? v[0] : v;
  const parsed = Number(n);
  return Number.isFinite(parsed) && parsed > 0 ? parsed : undefined;
});

// 当 attrId 变化时重新加载列表
watch(attributeId, () => {
  pageRef.value?.refresh();
});

function getFirstTranslationField(row: any, field: string): string {
  if (Array.isArray(row?.translations) && row.translations.length > 0) {
    return (row.translations[0] as any)?.[field] ?? "-";
  }
  return "-";
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
    toolbarRight: ["add"],
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
  if (!attributeId.value) return;
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
