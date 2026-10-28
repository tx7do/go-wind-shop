<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <el-alert
      :title="$t('pages.mall.stockOpLog.alert')"
      type="info"
      :closable="false"
      show-icon
      class="mb-4"
    />

    <ProPage ref="pageRef" :config="pageConfig">
    </ProPage>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";

import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";

import { fetchListStockOpLogs } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t } from "@/core/i18n";

const pageRef = ref();

const pageConfig = computed<ProPageConfig>(() => ({
  skeleton: true,
  search: {
    grid: true,
    fields: [
      {
        field: "skuId",
        label: $t("pages.mall.stockOpLog.skuId"),
        component: "el-input-number",
        componentProps: { controls: false, min: 1, style: { width: "100%" } },
      },
      {
        field: "reason",
        label: $t("pages.mall.stockOpLog.reason"),
        component: "el-select",
        componentProps: {
          clearable: true,
          style: { width: "100%" },
        },
        options: [
          { label: $t("pages.mall.stockOpLog.reasonLabel.ORDER_DEDUCT"), value: "ORDER_DEDUCT" },
          { label: $t("pages.mall.stockOpLog.reasonLabel.CANCEL_RESTORE"), value: "CANCEL_RESTORE" },
          { label: $t("pages.mall.stockOpLog.reasonLabel.REFUND_RESTORE"), value: "REFUND_RESTORE" },
          { label: $t("pages.mall.stockOpLog.reasonLabel.MANUAL_ADJUST"), value: "MANUAL_ADJUST" },
        ],
      },
    ],
  },

  table: {
    listAction: async (query: any) => {
      const { page, pageSize, ...queryParams } = query;
      const result = await fetchListStockOpLogs(
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
        label: $t("pages.mall.stockOpLog.skuId"),
        minWidth: 90,
        align: "left",
      },
      {
        prop: "productId",
        label: $t("pages.mall.stockOpLog.productId"),
        minWidth: 90,
        align: "left",
      },
      {
        prop: "delta",
        label: $t("pages.mall.stockOpLog.delta"),
        minWidth: 90,
        align: "right",
        cellType: "tag",
        tagTypeMap: {} as Record<string, string>,
        // 正负着色由 formatter 处理（tagTypeMap 仅静态键）
        formatter: (row: any) => (row.delta > 0 ? `+${row.delta}` : `${row.delta}`),
      },
      {
        prop: "stockBefore",
        label: $t("pages.mall.stockOpLog.stockBefore"),
        minWidth: 100,
        align: "right",
      },
      {
        prop: "stockAfter",
        label: $t("pages.mall.stockOpLog.stockAfter"),
        minWidth: 100,
        align: "right",
      },
      {
        prop: "reason",
        label: $t("pages.mall.stockOpLog.reason"),
        minWidth: 110,
        align: "left",
        cellType: "tag",
        tagTypeMap: {
          ORDER_DEDUCT: "info",
          CANCEL_RESTORE: "warning",
          REFUND_RESTORE: "warning",
          MANUAL_ADJUST: "primary",
        } as Record<string, "success" | "primary" | "warning" | "danger" | "info">,
        labelMap: {
          ORDER_DEDUCT: $t("pages.mall.stockOpLog.reasonLabel.ORDER_DEDUCT"),
          CANCEL_RESTORE: $t("pages.mall.stockOpLog.reasonLabel.CANCEL_RESTORE"),
          REFUND_RESTORE: $t("pages.mall.stockOpLog.reasonLabel.REFUND_RESTORE"),
          MANUAL_ADJUST: $t("pages.mall.stockOpLog.reasonLabel.MANUAL_ADJUST"),
        },
      },
      {
        prop: "refId",
        label: $t("pages.mall.stockOpLog.refId"),
        minWidth: 90,
        align: "left",
      },
      {
        prop: "remark",
        label: $t("pages.mall.stockOpLog.remark"),
        minWidth: 140,
        align: "left",
        showOverflowTooltip: true,
      },
      {
        prop: "createdBy",
        label: $t("pages.mall.stockOpLog.createdBy"),
        minWidth: 90,
        align: "left",
      },
      {
        prop: "createdAt",
        label: $t("pages.mall.stockOpLog.createdAt"),
        minWidth: 160,
        cellType: "date",
        dateFormat: "YYYY-MM-DD HH:mm:ss",
      },
    ],
  },
}));
</script>

<style lang="scss" scoped>
.app-container {
  padding: 20px;
  width: 100%;
  min-width: 0;
  flex-shrink: 0;
}
</style>
