<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <ProPage ref="pageRef" :config="pageConfig" @add="handleAdd" @edit="handleEdit">
    </ProPage>

    <SeckillDrawer ref="drawerRef" @success="handleSuccess" />
  </div>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";

import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";
import SeckillDrawer from "./seckill-drawer.vue";

import { fetchListSeckillActivities, useDeleteSeckillActivity } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t } from "@/core/i18n";

const { mutateAsync: deleteSeckill } = useDeleteSeckillActivity();

const pageRef = ref();
const drawerRef = ref();

const STATUS_KEY: Record<string, string> = {
  ACTIVE: "active",
  INACTIVE: "inactive",
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
      const result = await fetchListSeckillActivities(
        new PaginationQuery({
          paging: { page: page || 1, pageSize: pageSize || 10 },
          formValues: queryParams,
        })
      );
      return { items: result.items || [], total: result.total || 0 };
    },
    deleteAction: async (ids: string) => {
      await deleteSeckill({ id: ids as any });
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
        label: $t("pages.mall.seckill.name"),
        minWidth: 150,
        align: "left",
      },
      {
        prop: "skuId",
        label: $t("pages.mall.seckill.skuId"),
        minWidth: 90,
        align: "left",
      },
      {
        prop: "seckillPrice",
        label: $t("pages.mall.seckill.seckillPrice"),
        minWidth: 110,
        align: "left",
        // 分转元
        formatter: (row: any) => `¥${((row.seckillPrice ?? 0) / 100).toFixed(2)}`,
      },
      {
        prop: "soldCount",
        label: $t("pages.mall.seckill.soldProgress"),
        minWidth: 120,
        align: "left",
        formatter: (row: any) => `${row.soldCount ?? 0} / ${row.totalStock ?? 0}`,
      },
      {
        prop: "perUserLimit",
        label: $t("pages.mall.seckill.perUserLimit"),
        minWidth: 100,
        align: "left",
      },
      {
        prop: "validFrom",
        label: $t("pages.mall.seckill.validFrom"),
        minWidth: 160,
        cellType: "date",
        dateFormat: "YYYY-MM-DD HH:mm:ss",
      },
      {
        prop: "validUntil",
        label: $t("pages.mall.seckill.validUntil"),
        minWidth: 160,
        cellType: "date",
        dateFormat: "YYYY-MM-DD HH:mm:ss",
      },
      {
        prop: "status",
        label: $t("pages.mall.seckill.status"),
        minWidth: 90,
        align: "left",
        formatter: (row: any) =>
          $t(`pages.mall.seckill.statusLabel.${STATUS_KEY[row.status] ?? "active"}`),
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
