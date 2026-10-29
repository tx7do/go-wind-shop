<template>
  <ProModal
    v-model:visible="visible"
    :title="title"
    :config="{ component: 'drawer', drawer: { size: '860px', closeOnClickModal: false } }"
  >
    <div class="detail-drawer">
      <!-- 订单基本信息 -->
      <ElDivider content-position="left">{{ $t("common.section.basic") }}</ElDivider>
      <ElDescriptions :column="2" border>
        <ElDescriptionsItem :label="$t('pages.mall.order.id')">
          {{ data?.id ?? "-" }}
        </ElDescriptionsItem>
        <ElDescriptionsItem :label="$t('pages.mall.order.userId')">
          {{ data?.userId ?? "-" }}
        </ElDescriptionsItem>
        <ElDescriptionsItem :label="$t('pages.mall.order.totalAmount')">
          {{ fenToYuan(data?.totalAmount) }}
        </ElDescriptionsItem>
        <ElDescriptionsItem :label="$t('pages.mall.order.originalAmount')">
          {{ fenToYuan(data?.originalAmount) }}
        </ElDescriptionsItem>
        <ElDescriptionsItem :label="$t('pages.mall.order.discountAmount')">
          {{ fenToYuan(data?.discountAmount) }}
        </ElDescriptionsItem>
        <ElDescriptionsItem :label="$t('pages.mall.order.currency')">
          {{ data?.currency || "-" }}
        </ElDescriptionsItem>
        <ElDescriptionsItem :label="$t('pages.mall.order.status')">
          <ElTag v-if="data?.status" size="small" effect="dark" round>
            {{ orderStatusLabelMap[data.status] ?? data.status }}
          </ElTag>
          <span v-else>-</span>
        </ElDescriptionsItem>
        <ElDescriptionsItem :label="$t('pages.mall.order.createdAt')">
          {{ formatDateTime(data?.createdAt ?? "") }}
        </ElDescriptionsItem>
        <ElDescriptionsItem :label="$t('pages.mall.order.updatedAt')">
          {{ formatDateTime(data?.updatedAt ?? "") }}
        </ElDescriptionsItem>
        <ElDescriptionsItem :label="$t('pages.mall.order.recipientName')">
          {{ data?.recipientName || "-" }}
        </ElDescriptionsItem>
        <ElDescriptionsItem :label="$t('pages.mall.order.recipientPhone')">
          {{ data?.recipientPhone || "-" }}
        </ElDescriptionsItem>
        <ElDescriptionsItem :label="$t('pages.mall.order.shippingAddress')" :span="2">
          {{ data?.shippingAddress || "-" }}
        </ElDescriptionsItem>
      </ElDescriptions>

      <!-- 订单项明细 -->
      <ElDivider content-position="left">{{ $t("pages.mall.order.orderItemsTitle") }}</ElDivider>
      <ElTable :data="orderItems" border stripe size="small">
        <ElTableColumn type="index" :label="$t('common.table.seq')" width="60" align="center" />
        <ElTableColumn prop="id" :label="$t('pages.mall.order.orderItems.id')" min-width="80" />
        <ElTableColumn
          prop="skuId"
          :label="$t('pages.mall.order.orderItems.skuId')"
          min-width="80"
        />
        <ElTableColumn
          prop="quantity"
          :label="$t('pages.mall.order.orderItems.quantity')"
          min-width="80"
          align="right"
        />
        <ElTableColumn
          prop="unitPrice"
          :label="$t('pages.mall.order.orderItems.unitPrice')"
          min-width="100"
          align="right"
        >
          <template #default="scope">{{ fenToYuan(scope.row.unitPrice) }}</template>
        </ElTableColumn>
        <ElTableColumn
          prop="subtotal"
          :label="$t('pages.mall.order.orderItems.subtotal')"
          min-width="100"
          align="right"
        >
          <template #default="scope">{{ fenToYuan(scope.row.subtotal) }}</template>
        </ElTableColumn>
        <ElTableColumn
          prop="skuSnapshot"
          :label="$t('pages.mall.order.orderItems.skuSnapshot')"
          min-width="160"
        >
          <template #default="scope">{{ formatSnapshot(scope.row.skuSnapshot) }}</template>
        </ElTableColumn>
        <ElTableColumn
          prop="createdAt"
          :label="$t('pages.mall.order.orderItems.createdAt')"
          min-width="160"
        >
          <template #default="scope">
            {{ formatDateTime(scope.row.createdAt ?? "") }}
          </template>
        </ElTableColumn>
      </ElTable>
    </div>
    <template #footer>
      <div class="drawer-footer">
        <ElButton @click="close">{{ $t("common.button.close") }}</ElButton>
      </div>
    </template>
  </ProModal>
</template>

<script lang="ts" setup>
import { ref } from "vue";
import {
  ElButton,
  ElDivider,
  ElDescriptions,
  ElDescriptionsItem,
  ElTag,
  ElTable,
  ElTableColumn,
} from "element-plus";

import ProModal from "@/components/Pro/ProModal/index.vue";
import { formatDateTime } from "@/utils";
import { $t } from "@/core/i18n";
import { fetchGetOrder, fetchListOrderItems } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import type {
  orderservicev1_Order,
  orderservicev1_Order_Status,
} from "@/api/generated/admin/service/v1";

const visible = ref(false);
const title = ref("");
const data = ref<orderservicev1_Order>();
const orderItems = ref<any[]>([]);

// 金额字段以分存储，展示统一转元
function fenToYuan(v: any): string {
  return v === null || v === undefined || v === "" ? "-" : (Number(v) / 100).toFixed(2);
}

// SKU 快照为 JSON 字符串，格式化为可读的 k: v 串；解析失败按原文展示
function formatSnapshot(raw: any): string {
  if (raw === null || raw === undefined || raw === "") return "-";
  try {
    const obj = typeof raw === "string" ? JSON.parse(raw) : raw;
    if (obj && typeof obj === "object") {
      return Object.entries(obj)
        .map(([k, v]) => `${k}: ${v}`)
        .join("；");
    }
    return String(raw);
  } catch {
    return String(raw);
  }
}

// 订单状态 -> 显示文案映射
const orderStatusLabelMap: Record<string, string> = {
  STATUS_UNSPECIFIED: $t("pages.mall.order.statusLabel.STATUS_UNSPECIFIED"),
  PENDING_PAYMENT: $t("pages.mall.order.statusLabel.PENDING_PAYMENT"),
  PAID: $t("pages.mall.order.statusLabel.PAID"),
  CANCELLED: $t("pages.mall.order.statusLabel.CANCELLED"),
  FULFILLED: $t("pages.mall.order.statusLabel.FULFILLED"),
  CLOSED: $t("pages.mall.order.statusLabel.CLOSED"),
};

async function open(id: number) {
  title.value = $t("pages.mall.order.detailTitle");
  visible.value = true;
  data.value = await fetchGetOrder({ id });
  try {
    const result = await fetchListOrderItems(
      new PaginationQuery({
        paging: { page: 1, pageSize: 100 },
        formValues: { orderId: id },
      })
    );
    orderItems.value = result.items || [];
  } catch {
    orderItems.value = [];
  }
}

function close() {
  visible.value = false;
  data.value = undefined;
  orderItems.value = [];
}

defineExpose({ open, close });
</script>

<style lang="scss" scoped>
.detail-drawer {
  padding-right: 10px;
}

.drawer-footer {
  display: flex;
  justify-content: flex-end;
}
</style>
