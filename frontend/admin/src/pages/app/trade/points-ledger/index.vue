<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <ProPage ref="pageRef" :config="pageConfig"> </ProPage>
  </div>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";

import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";

import { fetchListPointsLedgers } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t } from "@/core/i18n";

const pageRef = ref();

const TYPE_KEY: Record<string, string> = {
  EARN: "earn",
  DEDUCT: "deduct",
};

const pageConfig = computed<ProPageConfig>(() => ({
  skeleton: true,
  search: {
    grid: true,
    fields: [],
  },

  table: {
    listAction: async (query: any) => {
      const { page, pageSize, ...queryParams } = query;
      const result = await fetchListPointsLedgers(
        new PaginationQuery({
          paging: { page: page || 1, pageSize: pageSize || 10 },
          formValues: queryParams,
        })
      );
      return { items: result.items || [], total: result.total || 0 };
    },
    deleteAction: async () => {},
    toolbar: [],
    defaultToolbar: ["refresh", "filter"],
    tableAttrs: { border: true, stripe: false },
    columns: [
      { type: "index", label: $t("common.table.seq"), width: 60 },
      {
        prop: "id",
        label: "ID",
        minWidth: 70,
        align: "left",
      },
      {
        prop: "userId",
        label: $t("pages.mall.points.userId"),
        minWidth: 90,
        align: "left",
      },
      {
        prop: "type",
        label: $t("pages.mall.points.type"),
        minWidth: 100,
        align: "left",
        formatter: (row: any) =>
          $t(`pages.mall.points.typeLabel.${TYPE_KEY[row.type] ?? "earn"}`),
      },
      {
        prop: "delta",
        label: $t("pages.mall.points.delta"),
        minWidth: 100,
        align: "left",
        formatter: (row: any) => `${(row.delta ?? 0) >= 0 ? "+" : ""}${row.delta ?? 0}`,
      },
      {
        prop: "balanceAfter",
        label: $t("pages.mall.points.balanceAfter"),
        minWidth: 110,
        align: "left",
      },
      {
        prop: "reason",
        label: $t("pages.mall.points.reason"),
        minWidth: 140,
        align: "left",
      },
      {
        prop: "refOrderId",
        label: $t("pages.mall.points.refOrderId"),
        minWidth: 100,
        align: "left",
      },
      {
        prop: "createdAt",
        label: $t("pages.mall.points.createdAt"),
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
