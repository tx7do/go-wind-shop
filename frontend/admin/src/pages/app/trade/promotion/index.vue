<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <ProPage ref="pageRef" :config="pageConfig" @add="handleAdd" @edit="handleEdit">
    </ProPage>

    <PromotionDrawer ref="drawerRef" @success="handleSuccess" />
  </div>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";

import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";
import PromotionDrawer from "./promotion-drawer.vue";

import { fetchListPromotions, useDeletePromotion } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t } from "@/core/i18n";

const pageRef = ref();
const drawerRef = ref();

const { mutateAsync: deletePromotion } = useDeletePromotion();

const promoTypeTagMap: Record<string, string> = {
  SPECIAL_PRICE: "warning",
  FULL_REDUCTION: "success",
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
      const result = await fetchListPromotions(
        new PaginationQuery({
          paging: { page: page || 1, pageSize: pageSize || 10 },
          formValues: queryParams,
        })
      );
      return { items: result.items || [], total: result.total || 0 };
    },
    deleteAction: async (ids: string) => {
      await deletePromotion({ id: ids as any });
    },
    toolbar: [],
    toolbarRight: ["add"],
    defaultToolbar: ["refresh", "exports", "filter"],
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
        label: $t("pages.mall.promotion.name"),
        minWidth: 160,
        align: "left",
      },
      {
        prop: "promoType",
        label: $t("pages.mall.promotion.promoType"),
        minWidth: 110,
        align: "left",
        cellType: "tag",
        tagTypeMap: promoTypeTagMap as Record<string, "success" | "primary" | "warning" | "danger" | "info">,
        labelMap: {
          SPECIAL_PRICE: $t("pages.mall.promotion.promoTypeLabel.SPECIAL_PRICE"),
          FULL_REDUCTION: $t("pages.mall.promotion.promoTypeLabel.FULL_REDUCTION"),
        },
      },
      {
        prop: "productId",
        label: $t("pages.mall.promotion.productId"),
        minWidth: 100,
        align: "left",
        formatter: (row: any) => row.productId || $t("pages.mall.promotion.storeWide"),
      },
      {
        prop: "specialPrice",
        label: $t("pages.mall.promotion.specialPrice"),
        minWidth: 110,
        align: "right",
        formatter: (row: any) =>
          row.promoType === "SPECIAL_PRICE" ? `¥${((row.specialPrice || 0) / 100).toFixed(2)}` : "—",
      },
      {
        prop: "thresholdAmount",
        label: $t("pages.mall.promotion.thresholdAmount"),
        minWidth: 110,
        align: "right",
        formatter: (row: any) =>
          row.promoType === "FULL_REDUCTION" ? `¥${((row.thresholdAmount || 0) / 100).toFixed(2)}` : "—",
      },
      {
        prop: "reductionAmount",
        label: $t("pages.mall.promotion.reductionAmount"),
        minWidth: 110,
        align: "right",
        formatter: (row: any) =>
          row.promoType === "FULL_REDUCTION" ? `-¥${((row.reductionAmount || 0) / 100).toFixed(2)}` : "—",
      },
      {
        prop: "priority",
        label: $t("pages.mall.promotion.priority"),
        minWidth: 80,
        align: "right",
      },
      {
        prop: "validFrom",
        label: $t("pages.mall.promotion.validFrom"),
        minWidth: 160,
        cellType: "date",
        dateFormat: "YYYY-MM-DD HH:mm",
      },
      {
        prop: "validUntil",
        label: $t("pages.mall.promotion.validUntil"),
        minWidth: 160,
        cellType: "date",
        dateFormat: "YYYY-MM-DD HH:mm",
      },
      {
        prop: "status",
        label: $t("pages.mall.promotion.status"),
        minWidth: 90,
        align: "left",
        cellType: "tag",
        tagTypeMap: {
          ACTIVE: "success",
          INACTIVE: "info",
        } as Record<string, "success" | "primary" | "warning" | "danger" | "info">,
        labelMap: {
          ACTIVE: $t("pages.mall.promotion.statusLabel.ACTIVE"),
          INACTIVE: $t("pages.mall.promotion.statusLabel.INACTIVE"),
        },
      },
      {
        prop: "action",
        label: $t("common.table.action"),
        fixed: "right",
        width: 140,
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
