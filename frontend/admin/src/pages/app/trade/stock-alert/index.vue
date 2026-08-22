<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <el-alert
      :title="$t('pages.mall.stockAlert.alert')"
      type="warning"
      :closable="false"
      show-icon
      class="mb-4"
    />

    <ProPage ref="pageRef" :config="pageConfig" @edit="handleEdit">
    </ProPage>

    <StockAlertDrawer ref="drawerRef" @success="handleSuccess" />
  </div>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";

import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";
import StockAlertDrawer from "./stock-alert-drawer.vue";

import { fetchListStockAlerts } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t } from "@/core/i18n";

const pageRef = ref();
const drawerRef = ref();

const pageConfig = computed<ProPageConfig>(() => ({
  skeleton: true,
  search: {
    grid: true,
    fields: [],
  },

  table: {
    listAction: async (query: any) => {
      const { page, pageSize, ...queryParams } = query;
      const result = await fetchListStockAlerts(
        new PaginationQuery({
          paging: { page: page || 1, pageSize: pageSize || 10 },
          formValues: queryParams,
        })
      );
      return { items: result.items || [], total: result.total || 0 };
    },
    toolbar: [],
    defaultToolbar: ["refresh", "exports", "filter"],
    tableAttrs: { border: true, stripe: false },
    columns: [
      { type: "index", label: $t("common.table.seq"), width: 60 },
      {
        prop: "id",
        label: "ID",
        minWidth: 80,
        align: "left",
      },
      {
        prop: "skuId",
        label: $t("pages.mall.stockAlert.skuId"),
        minWidth: 120,
        align: "left",
      },
      {
        prop: "stockQtyAtTrigger",
        label: $t("pages.mall.stockAlert.stockQtyAtTrigger"),
        minWidth: 140,
        align: "right",
      },
      {
        prop: "threshold",
        label: $t("pages.mall.stockAlert.threshold"),
        minWidth: 100,
        align: "right",
      },
      {
        prop: "status",
        label: $t("pages.mall.stockAlert.status"),
        minWidth: 100,
        align: "left",
        cellType: "tag",
        tagTypeMap: {
          OPEN: "warning",
          RESOLVED: "success",
        } as Record<string, "success" | "primary" | "warning" | "danger" | "info">,
        labelMap: {
          OPEN: $t("pages.mall.stockAlert.statusLabel.OPEN"),
          RESOLVED: $t("pages.mall.stockAlert.statusLabel.RESOLVED"),
        },
      },
      {
        prop: "createdAt",
        label: $t("pages.mall.stockAlert.createdAt"),
        minWidth: 160,
        cellType: "date",
        dateFormat: "YYYY-MM-DD HH:mm:ss",
      },
      {
        prop: "action",
        label: $t("common.table.action"),
        fixed: "right",
        width: 100,
        cellType: "tool",
        buttons: [
          { name: "edit", label: $t("common.button.edit"), icon: "lucide:pen-line" },
        ],
      },
    ],
  },
}));

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
