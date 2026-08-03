<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <ProPage ref="pageRef" :config="pageConfig"> </ProPage>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";

import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";

import { fetchListPaymentTransactions } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t } from "@/core/i18n";

const pageRef = ref();

// 支付流水状态 -> Tag 颜色映射
const paymentStatusTagTypeMap: Record<
  string,
  "success" | "primary" | "warning" | "danger" | "info"
> = {
  SUCCEEDED: "success",
  PENDING: "warning",
  FAILED: "danger",
  REFUNDED: "info",
  STATUS_UNSPECIFIED: "info",
};

// 支付流水状态 -> 显示文案映射
const paymentStatusLabelMap: Record<string, string> = {
  SUCCEEDED: $t("pages.mall.paymentTransaction.statusLabel.SUCCEEDED"),
  PENDING: $t("pages.mall.paymentTransaction.statusLabel.PENDING"),
  FAILED: $t("pages.mall.paymentTransaction.statusLabel.FAILED"),
  REFUNDED: $t("pages.mall.paymentTransaction.statusLabel.REFUNDED"),
  STATUS_UNSPECIFIED: $t("pages.mall.paymentTransaction.statusLabel.STATUS_UNSPECIFIED"),
};

const pageConfig = computed<ProPageConfig>(() => ({
  skeleton: true,
  search: {
    grid: true,
    fields: [
      {
        type: "input",
        label: $t("pages.mall.paymentTransaction.orderId"),
        field: "orderId",
        attrs: { placeholder: $t("common.placeholder.input"), clearable: true },
      },
      {
        type: "input",
        label: $t("pages.mall.paymentTransaction.userId"),
        field: "userId",
        attrs: { placeholder: $t("common.placeholder.input"), clearable: true },
      },
    ],
  },

  table: {
    listAction: async (query: any) => {
      const { page, pageSize, ...queryParams } = query;
      const result = await fetchListPaymentTransactions(
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
        label: $t("pages.mall.paymentTransaction.id"),
        minWidth: 80,
      },
      {
        prop: "orderId",
        label: $t("pages.mall.paymentTransaction.orderId"),
        minWidth: 100,
      },
      {
        prop: "userId",
        label: $t("pages.mall.paymentTransaction.userId"),
        minWidth: 100,
      },
      {
        prop: "amount",
        label: $t("pages.mall.paymentTransaction.amount"),
        minWidth: 100,
        align: "right",
        cellType: "price",
        pricePrefix: "",
      },
      {
        prop: "currency",
        label: $t("pages.mall.paymentTransaction.currency"),
        minWidth: 80,
      },
      {
        prop: "status",
        label: $t("pages.mall.paymentTransaction.status"),
        minWidth: 100,
        cellType: "tag",
        tagTypeMap: paymentStatusTagTypeMap,
        labelMap: paymentStatusLabelMap,
      },
      {
        prop: "paymentMethod",
        label: $t("pages.mall.paymentTransaction.paymentMethod"),
        minWidth: 120,
      },
      {
        prop: "businessType",
        label: $t("pages.mall.paymentTransaction.businessType"),
        minWidth: 120,
      },
      {
        prop: "createdAt",
        label: $t("pages.mall.paymentTransaction.createdAt"),
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
