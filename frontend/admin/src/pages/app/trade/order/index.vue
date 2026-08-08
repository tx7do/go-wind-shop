<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <ProPage ref="pageRef" :config="pageConfig" @operate="handleOperate"></ProPage>

    <!-- 订单详情抽屉（只读） -->
    <OrderDetailDrawer ref="drawerRef" @close="handleRefresh" />
  </div>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";

import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";
import OrderDetailDrawer from "./order-detail-drawer.vue";

import { fetchListOrders, useUpdateOrder } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t } from "@/core/i18n";

const { mutateAsync: updateOrder } = useUpdateOrder();

const pageRef = ref();
const drawerRef = ref();

// 订单状态 -> Tag 颜色映射
const orderStatusTagTypeMap: Record<string, "success" | "primary" | "warning" | "danger" | "info"> =
  {
    PENDING_PAYMENT: "warning",
    PAID: "primary",
    CANCELLED: "info",
    FULFILLED: "success",
    CLOSED: "info",
    STATUS_UNSPECIFIED: "info",
  };

// 订单状态 -> 显示文案映射
const orderStatusLabelMap: Record<string, string> = {
  STATUS_UNSPECIFIED: $t("pages.mall.order.statusLabel.STATUS_UNSPECIFIED"),
  PENDING_PAYMENT: $t("pages.mall.order.statusLabel.PENDING_PAYMENT"),
  PAID: $t("pages.mall.order.statusLabel.PAID"),
  CANCELLED: $t("pages.mall.order.statusLabel.CANCELLED"),
  FULFILLED: $t("pages.mall.order.statusLabel.FULFILLED"),
  CLOSED: $t("pages.mall.order.statusLabel.CLOSED"),
};

const pageConfig = computed<ProPageConfig>(() => ({
  skeleton: true,
  search: {
    grid: true,
    fields: [
      {
        type: "input",
        label: $t("pages.mall.order.userId"),
        field: "userId",
        attrs: { placeholder: $t("common.placeholder.input"), clearable: true },
      },
    ],
  },

  table: {
    listAction: async (query: any) => {
      const { page, pageSize, ...queryParams } = query;
      const result = await fetchListOrders(
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
        label: $t("pages.mall.order.id"),
        minWidth: 80,
      },
      {
        prop: "userId",
        label: $t("pages.mall.order.userId"),
        minWidth: 100,
      },
      {
        prop: "totalAmount",
        label: $t("pages.mall.order.totalAmount"),
        minWidth: 100,
        align: "right",
        cellType: "price",
        pricePrefix: "",
      },
      {
        prop: "currency",
        label: $t("pages.mall.order.currency"),
        minWidth: 80,
      },
      {
        prop: "status",
        label: $t("pages.mall.order.status"),
        minWidth: 100,
        cellType: "tag",
        tagTypeMap: orderStatusTagTypeMap,
        labelMap: orderStatusLabelMap,
      },
      {
        prop: "createdAt",
        label: $t("pages.mall.order.createdAt"),
        minWidth: 160,
        cellType: "date",
        dateFormat: "YYYY-MM-DD HH:mm:ss",
      },
      {
        prop: "action",
        label: $t("common.table.action"),
        fixed: "right",
        width: 220,
        cellType: "tool",
        buttons: [
          { name: "detail", label: $t("common.button.detail"), icon: "lucide:eye" },
          {
            name: "close",
            label: $t("pages.mall.order.close"),
            icon: "lucide:octagon-x",
            attrs: { type: "danger" },
            visible: (row: any) => row.status === "PAID" || row.status === "FULFILLED",
          },
        ],
      },
    ],
  },
}));

async function handleOperate(data: { name: string; row: any }) {
  const { name, row } = data;

  if (name === "detail") {
    drawerRef.value?.open(row.id);
    return;
  }

  if (name === "close") {
    try {
      await ElMessageBox.confirm($t("pages.mall.order.close"), $t("common.notification.confirmTitle"), {
        type: "warning",
      });
    } catch {
      return;
    }

    try {
      await updateOrder({ id: row.id, values: { status: "CLOSED" } });
      ElMessage.success($t("common.notification.updateSuccess"));
      pageRef.value?.refresh();
    } catch {
      ElMessage.error($t("common.notification.updateFailed"));
    }
  }
}

function handleRefresh() {
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
