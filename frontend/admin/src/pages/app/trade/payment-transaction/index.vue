<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <ProPage ref="pageRef" :config="pageConfig"></ProPage>
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

// 支付方式/业务类型 -> 显示文案映射（裸枚举直接展示无翻译）
const paymentMethodLabelMap: Record<string, string> = {
  ALIPAY: $t("pages.mall.paymentTransaction.methodLabel.ALIPAY"),
  WECHAT: $t("pages.mall.paymentTransaction.methodLabel.WECHAT"),
  BALANCE: $t("pages.mall.paymentTransaction.methodLabel.BALANCE"),
  CARD: $t("pages.mall.paymentTransaction.methodLabel.CARD"),
  MOCK: $t("pages.mall.paymentTransaction.methodLabel.MOCK"),
};

const businessTypeLabelMap: Record<string, string> = {
  BUSINESS_TYPE_CONSUME: $t("pages.mall.paymentTransaction.businessTypeLabel.BUSINESS_TYPE_CONSUME"),
  BUSINESS_TYPE_REFUND: $t("pages.mall.paymentTransaction.businessTypeLabel.BUSINESS_TYPE_REFUND"),
  BUSINESS_TYPE_RECHARGE: $t("pages.mall.paymentTransaction.businessTypeLabel.BUSINESS_TYPE_RECHARGE"),
};

// @ts-ignore - intentional incomplete record mirroring existing convention
const paymentMethodTagTypeMap: Record<
  string,
  "success" | "primary" | "warning" | "danger" | "info"
> = {
  ALIPAY: "primary",
  WECHAT: "success",
  BALANCE: "warning",
  CARD: "info",
  MOCK: "info",
};
// @ts-ignore - intentional incomplete record mirroring existing convention
const businessTypeTagTypeMap: Record<
  string,
  "success" | "primary" | "warning" | "danger" | "info"
> = {
  BUSINESS_TYPE_CONSUME: "primary",
  BUSINESS_TYPE_REFUND: "warning",
  BUSINESS_TYPE_RECHARGE: "success",
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
        formatter: (row: any) =>
          row.amount === null || row.amount === undefined || row.amount === ""
            ? "-"
            : (Number(row.amount) / 100).toFixed(2),
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
        cellType: "tag",
        tagTypeMap: paymentMethodTagTypeMap,
        labelMap: paymentMethodLabelMap,
      },
      {
        prop: "businessType",
        label: $t("pages.mall.paymentTransaction.businessType"),
        minWidth: 120,
        cellType: "tag",
        tagTypeMap: businessTypeTagTypeMap,
        labelMap: businessTypeLabelMap,
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
