<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <ProPage
      ref="pageRef"
      :config="pageConfig"
      @add="handleAdd"
      @operate="handleOperate"
    ></ProPage>

    <!-- 创建物流单抽屉 -->
    <ShipmentDrawer ref="drawerRef" @success="handleRefresh" />
  </div>
</template>

<script lang="ts" setup>
import { computed, ref } from "vue";
import { ElMessage, ElMessageBox } from "element-plus";

import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";
import ShipmentDrawer from "./shipment-drawer.vue";

import { fetchListShipments, useUpdateShipment } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t } from "@/core/i18n";

const { mutateAsync: updateShipment } = useUpdateShipment();

const pageRef = ref();
const drawerRef = ref();

// 物流单状态 -> Tag 颜色映射
const shipmentStatusTagTypeMap: Record<
  string,
  "success" | "primary" | "warning" | "danger" | "info"
> = {
  PENDING: "warning",
  SHIPPED: "primary",
  DELIVERED: "success",
  STATUS_UNSPECIFIED: "info",
};

// 物流单状态 -> 显示文案映射
const shipmentStatusLabelMap: Record<string, string> = {
  STATUS_UNSPECIFIED: $t("pages.mall.shipment.statusLabel.STATUS_UNSPECIFIED"),
  PENDING: $t("pages.mall.shipment.statusLabel.PENDING"),
  SHIPPED: $t("pages.mall.shipment.statusLabel.SHIPPED"),
  DELIVERED: $t("pages.mall.shipment.statusLabel.DELIVERED"),
};

const pageConfig = computed<ProPageConfig>(() => ({
  skeleton: true,
  search: {
    grid: true,
    fields: [
      {
        type: "input",
        label: $t("pages.mall.shipment.orderId"),
        field: "orderId",
        attrs: { placeholder: $t("common.placeholder.input"), clearable: true },
      },
      {
        type: "input",
        label: $t("pages.mall.shipment.carrier"),
        field: "carrier",
        attrs: { placeholder: $t("common.placeholder.input"), clearable: true },
      },
    ],
  },

  table: {
    listAction: async (query: any) => {
      const { page, pageSize, ...queryParams } = query;
      const result = await fetchListShipments(
        new PaginationQuery({
          paging: { page: page || 1, pageSize: pageSize || 10 },
          formValues: queryParams,
        })
      );
      return { items: result.items || [], total: result.total || 0 };
    },
    toolbar: [],
    toolbarRight: ["add"],
    defaultToolbar: ["refresh", "exports", "filter"],
    tableAttrs: { border: true, stripe: false },
    columns: [
      { type: "index", label: $t("common.table.seq"), width: 60 },
      {
        prop: "id",
        label: $t("pages.mall.shipment.id"),
        minWidth: 80,
      },
      {
        prop: "orderId",
        label: $t("pages.mall.shipment.orderId"),
        minWidth: 100,
      },
      {
        prop: "carrier",
        label: $t("pages.mall.shipment.carrier"),
        minWidth: 100,
      },
      {
        prop: "trackingNumber",
        label: $t("pages.mall.shipment.trackingNumber"),
        minWidth: 140,
      },
      {
        prop: "status",
        label: $t("pages.mall.shipment.status"),
        minWidth: 100,
        cellType: "tag",
        tagTypeMap: shipmentStatusTagTypeMap,
        labelMap: shipmentStatusLabelMap,
      },
      {
        prop: "createdAt",
        label: $t("pages.mall.shipment.createdAt"),
        minWidth: 160,
        cellType: "date",
        dateFormat: "YYYY-MM-DD HH:mm:ss",
      },
      {
        prop: "action",
        label: $t("common.table.action"),
        fixed: "right",
        width: 180,
        cellType: "tool",
        buttons: [
          {
            name: "fulfill",
            label: $t("pages.mall.shipment.fulfill"),
            icon: "lucide:truck",
            attrs: { type: "warning" },
            visible: (row: any) => row.status === "PENDING",
          },
          {
            name: "deliver",
            label: $t("pages.mall.shipment.deliver"),
            icon: "lucide:package-check",
            attrs: { type: "success" },
            visible: (row: any) => row.status === "SHIPPED",
          },
        ],
      },
    ],
  },
}));

async function handleOperate(data: { name: string; row: any }) {
  const { name, row } = data;

  // 状态机前置条件：发货 PENDING→SHIPPED，签收 SHIPPED→DELIVERED
  const transitionMap: Record<
    string,
    { target: "SHIPPED" | "DELIVERED"; expected: ("PENDING" | "SHIPPED")[]; confirmKey: string }
  > = {
    fulfill: { target: "SHIPPED", expected: ["PENDING"], confirmKey: "pages.mall.shipment.confirm.fulfill" },
    deliver: { target: "DELIVERED", expected: ["SHIPPED"], confirmKey: "pages.mall.shipment.confirm.deliver" },
  };

  const action = transitionMap[name];
  if (!action) return;

  try {
    await ElMessageBox.confirm($t(action.confirmKey), $t("common.notification.confirmTitle"), {
      type: "warning",
    });
  } catch {
    return;
  }

  try {
    await updateShipment({
      id: row.id,
      values: { status: action.target },
      expectedStatus: action.expected,
    });
    ElMessage.success($t("common.notification.updateSuccess"));
    pageRef.value?.refresh();
  } catch {
    // 状态机前置条件失败（expected_status 不匹配）后端返回 Conflict，
    // 通常意味着该物流单状态已被其他操作改变，提示刷新。
    ElMessage.error($t("common.notification.updateFailed"));
  }
}

function handleAdd() {
  drawerRef.value?.open();
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
