<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <!--
      低库存管理台（只读）。
      数据来源：SKU List RPC，经 stockQty__lte 阈值过滤（后端生成 stock_qty <= N 谓词）。
      默认阈值 10，与后端 StockAlertService.StockAlertThreshold 一致——
      该周期任务扫描同阈值低库存 SKU 并发站内预警，此页是其可视化落点，供运营定位补货。
    -->
    <el-alert
      :title="$t('pages.mall.lowStock.alert')"
      type="warning"
      :closable="false"
      show-icon
      class="mb-4"
    />

    <ProPage ref="pageRef" :config="pageConfig" />
  </div>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";

import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";

import { fetchListSkus } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t } from "@/core/i18n";

const pageRef = ref();

const pageConfig = computed<ProPageConfig>(() => ({
  skeleton: true,
  search: {
    grid: true,
    fields: [
      {
        // 阈值过滤：formValues 透传为后端 query JSON，
        // 经 stockQty__lte → stock_qty LTE 谓词。默认 10 与后端告警阈值对齐。
        type: "input-number",
        label: $t("pages.mall.lowStock.threshold"),
        field: "stockQty__lte",
        initialValue: 10,
        attrs: { min: 0, controlsPosition: "right" },
      },
    ],
  },

  table: {
    listAction: async (query: any) => {
      const { page, pageSize, ...queryParams } = query;
      const result = await fetchListSkus(
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
        prop: "productId",
        label: $t("pages.mall.sku.productId"),
        minWidth: 100,
      },
      {
        prop: "skuCode",
        label: $t("pages.mall.sku.skuCode"),
        minWidth: 160,
        fixed: "left",
        align: "left",
      },
      {
        prop: "stockQty",
        label: $t("pages.mall.sku.stockQty"),
        minWidth: 100,
        align: "right",
      },
      {
        prop: "createdAt",
        label: $t("pages.mall.sku.createdAt"),
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
