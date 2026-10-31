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

// 折扣类型/模板状态/可领取 → tag 文案 + 语义色映射
// （FIXED_AMOUNT=primary 满减、PERCENTAGE=warning 折扣；ACTIVE=success 生效、INACTIVE=info 停用；
//   claimable 是 bool，labelMap 走字符串键）
const couponTypeLabelMap: Record<string, string> = {
  DISCOUNT_TYPE_UNSPECIFIED: $t("pages.mall.couponTemplate.typeLabel.DISCOUNT_TYPE_UNSPECIFIED"),
  FIXED_AMOUNT: $t("pages.mall.couponTemplate.typeLabel.FIXED_AMOUNT"),
  PERCENTAGE: $t("pages.mall.couponTemplate.typeLabel.PERCENTAGE"),
};
// @ts-ignore - intentional incomplete record mirroring existing convention
const couponTypeTagTypeMap: Record<
  string,
  "success" | "primary" | "warning" | "danger" | "info"
> = {
  FIXED_AMOUNT: "primary",
  PERCENTAGE: "warning",
};
const couponStatusLabelMap: Record<string, string> = {
  STATUS_UNSPECIFIED: $t("pages.mall.couponTemplate.statusLabel.STATUS_UNSPECIFIED"),
  ACTIVE: $t("pages.mall.couponTemplate.statusLabel.ACTIVE"),
  INACTIVE: $t("pages.mall.couponTemplate.statusLabel.INACTIVE"),
};
// @ts-ignore - intentional incomplete record mirroring existing convention
const couponStatusTagTypeMap: Record<
  string,
  "success" | "primary" | "warning" | "danger" | "info"
> = {
  ACTIVE: "success",
  INACTIVE: "info",
};
const claimableLabelMap: Record<string, string> = {
  true: $t("pages.mall.couponTemplate.claimableLabel.true"),
  false: $t("pages.mall.couponTemplate.claimableLabel.false"),
};
// @ts-ignore - intentional incomplete record mirroring existing convention
const claimableTagTypeMap: Record<string, "success" | "primary" | "warning" | "danger" | "info"> = {
  true: "success",
  false: "info",
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
        cellType: "tag",
        tagTypeMap: couponTypeTagTypeMap,
        labelMap: couponTypeLabelMap,
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
        cellType: "tag",
        tagTypeMap: couponStatusTagTypeMap,
        labelMap: couponStatusLabelMap,
      },
      {
        prop: "claimable",
        label: $t("pages.mall.couponTemplate.claimable"),
        minWidth: 100,
        align: "left",
        cellType: "tag",
        tagTypeMap: claimableTagTypeMap,
        labelMap: claimableLabelMap,
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
