<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <ProPage ref="pageRef" :config="pageConfig" @add="handleAdd" @edit="handleEdit">
    </ProPage>

    <CouponTemplateDrawer ref="drawerRef" @success="handleSuccess" />
  </div>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";

import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";
import CouponTemplateDrawer from "./coupon-template-drawer.vue";

import { fetchListCouponTemplates, useDeleteCouponTemplate } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t } from "@/core/i18n";

const { mutateAsync: deleteCouponTemplate } = useDeleteCouponTemplate();

const pageRef = ref();
const drawerRef = ref();

const pageConfig = computed<ProPageConfig>(() => ({
  skeleton: true,
  search: {
    grid: true,
    fields: [],
  },

  table: {
    listAction: async (query: any) => {
      const { page, pageSize, ...queryParams } = query;
      const result = await fetchListCouponTemplates(
        new PaginationQuery({
          paging: { page: page || 1, pageSize: pageSize || 10 },
          formValues: queryParams,
        })
      );
      return { items: result.items || [], total: result.total || 0 };
    },
    deleteAction: async (ids: string) => {
      await deleteCouponTemplate({ id: ids as any });
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
        minWidth: 80,
        align: "left",
      },
      {
        prop: "discountType",
        label: $t("pages.mall.couponTemplate.discountType"),
        minWidth: 120,
        align: "left",
      },
      {
        prop: "discountValue",
        label: $t("pages.mall.couponTemplate.discountValue"),
        minWidth: 120,
        align: "left",
      },
      {
        prop: "discountPercentage",
        label: $t("pages.mall.couponTemplate.discountPercentage"),
        minWidth: 120,
        align: "left",
      },
      {
        prop: "validFrom",
        label: $t("pages.mall.couponTemplate.validFrom"),
        minWidth: 160,
        cellType: "date",
        dateFormat: "YYYY-MM-DD HH:mm:ss",
      },
      {
        prop: "validUntil",
        label: $t("pages.mall.couponTemplate.validUntil"),
        minWidth: 160,
        cellType: "date",
        dateFormat: "YYYY-MM-DD HH:mm:ss",
      },
      {
        prop: "maxRedemptions",
        label: $t("pages.mall.couponTemplate.maxRedemptions"),
        minWidth: 120,
        align: "left",
      },
      {
        prop: "maxRedemptionsPerUser",
        label: $t("pages.mall.couponTemplate.maxRedemptionsPerUser"),
        minWidth: 120,
        align: "left",
      },
      {
        prop: "redeemedCount",
        label: $t("pages.mall.couponTemplate.redeemedCount"),
        minWidth: 120,
        align: "left",
      },
      {
        prop: "status",
        label: $t("pages.mall.couponTemplate.status"),
        minWidth: 100,
        align: "left",
      },
      {
        prop: "claimable",
        label: $t("pages.mall.couponTemplate.claimable"),
        minWidth: 100,
        align: "left",
        cellType: "tag",
      },
      {
        prop: "createdAt",
        label: $t("pages.mall.couponTemplate.createdAt"),
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
