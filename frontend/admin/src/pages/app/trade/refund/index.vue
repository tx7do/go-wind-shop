<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <ProPage ref="pageRef" :config="pageConfig"> </ProPage>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";

import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";

import { fetchListPaymentRefunds } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t } from "@/core/i18n";

const pageRef = ref();

// 退款状态 -> Tag 颜色映射
const refundStatusTagTypeMap: Record<
  string,
  "success" | "primary" | "warning" | "danger" | "info"
> = {
  SUCCEEDED: "success",
  PENDING: "warning",
  FAILED: "danger",
  STATUS_UNSPECIFIED: "info",
};

// 退款状态 -> 显示文案映射
const refundStatusLabelMap: Record<string, string> = {
  SUCCEEDED: $t("pages.mall.refund.statusLabel.SUCCEEDED"),
  PENDING: $t("pages.mall.refund.statusLabel.PENDING"),
  FAILED: $t("pages.mall.refund.statusLabel.FAILED"),
  STATUS_UNSPECIFIED: $t("pages.mall.refund.statusLabel.STATUS_UNSPECIFIED"),
};

const pageConfig = computed<ProPageConfig>(() => ({
  skeleton: true,
  search: {
    grid: true,
    fields: [
      {
        type: "input",
        label: $t("pages.mall.refund.transactionId"),
        field: "transactionId",
        attrs: { placeholder: $t("common.placeholder.input"), clearable: true },
      },
    ],
  },

  table: {
    listAction: async (query: any) => {
      const { page, pageSize, ...queryParams } = query;
      const result = await fetchListPaymentRefunds(
        new PaginationQuery({
          paging: { page: page || 1, pageSize: pageSize || 10 },
          formValues: queryParams,
        })
      );
      return { items: result.items || [], total: result.total || 0 };
    },
    toolbar: [],
    toolbarRight: [],
    defaultToolbar: ["refresh", "exports", "filter"],
    tableAttrs: { border: true, stripe: false },
    columns: [
      { type: "index", label: $t("common.table.seq"), width: 60 },
      {
        prop: "id",
        label: $t("pages.mall.refund.id"),
        minWidth: 80,
      },
      {
        prop: "transactionId",
        label: $t("pages.mall.refund.transactionId"),
        minWidth: 120,
      },
      {
        prop: "amount",
        label: $t("pages.mall.refund.amount"),
        minWidth: 100,
        align: "right",
        cellType: "price",
        pricePrefix: "",
      },
      {
        prop: "currency",
        label: $t("pages.mall.refund.currency"),
        minWidth: 80,
      },
      {
        prop: "status",
        label: $t("pages.mall.refund.status"),
        minWidth: 100,
        cellType: "tag",
        tagTypeMap: refundStatusTagTypeMap,
        labelMap: refundStatusLabelMap,
      },
      {
        prop: "businessRefId",
        label: $t("pages.mall.refund.businessRefId"),
        minWidth: 140,
      },
      {
        prop: "createdAt",
        label: $t("pages.mall.refund.createdAt"),
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
