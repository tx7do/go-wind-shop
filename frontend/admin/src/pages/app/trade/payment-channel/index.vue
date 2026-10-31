<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <ProPage ref="pageRef" :config="pageConfig" @add="handleAdd" @edit="handleEdit">
    </ProPage>

    <PaymentChannelDrawer ref="drawerRef" @success="handleSuccess" />
  </div>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";

import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";
import PaymentChannelDrawer from "./payment-channel-drawer.vue";

import { fetchListPaymentChannelConfigs, useDeletePaymentChannelConfig } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t } from "@/core/i18n";

const { mutateAsync: deleteChannel } = useDeletePaymentChannelConfig();

const pageRef = ref();
const drawerRef = ref();

const CHANNEL_LABEL_KEY: Record<string, string> = {
  PAYMENT_CHANNEL_MOCK: "mock",
  PAYMENT_CHANNEL_WECHAT_NATIVE: "wechatNative",
  PAYMENT_CHANNEL_WECHAT_H5: "wechatH5",
  PAYMENT_CHANNEL_ALIPAY_PAGE: "alipayPage",
};

// bool 列 → tag 文案 + 语义色映射（labelMap 走字符串键）
// 启用=success/停用=info；沙箱=warning（提醒非生产）/生产=primary；凭据缺失=danger 提示风险
// @ts-ignore - intentional incomplete record mirroring existing convention
const enabledTagTypeMap: Record<string, "success" | "primary" | "warning" | "danger" | "info"> = {
  true: "success",
  false: "info",
};
const enabledLabelMap: Record<string, string> = {
  true: $t("pages.mall.paymentChannel.enabledOn"),
  false: $t("pages.mall.paymentChannel.enabledOff"),
};
// @ts-ignore - intentional incomplete record mirroring existing convention
const sandboxTagTypeMap: Record<string, "success" | "primary" | "warning" | "danger" | "info"> = {
  true: "warning",
  false: "primary",
};
const sandboxLabelMap: Record<string, string> = {
  true: $t("pages.mall.paymentChannel.sandboxOn"),
  false: $t("pages.mall.paymentChannel.sandboxOff"),
};
// @ts-ignore - intentional incomplete record mirroring existing convention
const credentialsTagTypeMap: Record<
  string,
  "success" | "primary" | "warning" | "danger" | "info"
> = {
  true: "success",
  false: "danger",
};
const credentialsLabelMap: Record<string, string> = {
  true: $t("pages.mall.paymentChannel.credentialsConfigured"),
  false: $t("pages.mall.paymentChannel.credentialsMissing"),
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
      const result = await fetchListPaymentChannelConfigs(
        new PaginationQuery({
          paging: { page: page || 1, pageSize: pageSize || 10 },
          formValues: queryParams,
        })
      );
      return { items: result.items || [], total: result.total || 0 };
    },
    deleteAction: async (ids: string) => {
      await deleteChannel({ id: ids as any });
    },
    toolbar: [],
    toolbarRight: ["add"],
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
        prop: "name",
        label: $t("pages.mall.paymentChannel.name"),
        minWidth: 140,
        align: "left",
      },
      {
        prop: "channel",
        label: $t("pages.mall.paymentChannel.channel"),
        minWidth: 150,
        align: "left",
        formatter: (row: any) =>
          $t(
            `pages.mall.paymentChannel.channelType.${CHANNEL_LABEL_KEY[row.channel] ?? "mock"}`
          ),
      },
      {
        prop: "appId",
        label: $t("pages.mall.paymentChannel.appId"),
        minWidth: 150,
        align: "left",
      },
      {
        prop: "sandbox",
        label: $t("pages.mall.paymentChannel.sandbox"),
        minWidth: 90,
        align: "left",
        cellType: "tag",
        tagTypeMap: sandboxTagTypeMap,
        labelMap: sandboxLabelMap,
      },
      {
        prop: "enabled",
        label: $t("pages.mall.paymentChannel.enabled"),
        minWidth: 90,
        align: "left",
        cellType: "tag",
        tagTypeMap: enabledTagTypeMap,
        labelMap: enabledLabelMap,
      },
      {
        prop: "credentialsConfigured",
        label: $t("pages.mall.paymentChannel.credentials"),
        minWidth: 110,
        align: "left",
        cellType: "tag",
        tagTypeMap: credentialsTagTypeMap,
        labelMap: credentialsLabelMap,
      },
      {
        prop: "updatedAt",
        label: $t("pages.mall.paymentChannel.updatedAt"),
        minWidth: 160,
        cellType: "date",
        dateFormat: "YYYY-MM-DD HH:mm:ss",
      },
      {
        prop: "action",
        label: $t("common.table.action"),
        fixed: "right",
        width: 160,
        cellType: "tool",
        buttons: [
          { name: "edit", label: $t("common.button.edit"), icon: "lucide:pen-line" },
          {
            name: "delete",
            label: $t("common.button.delete"),
            icon: "lucide:trash-2",
            attrs: { type: "danger" },
          },
        ],
      },
    ],
  },
}));

function handleAdd() {
  drawerRef.value?.open();
}

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
