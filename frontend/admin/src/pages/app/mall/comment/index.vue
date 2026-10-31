<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <ProPage ref="pageRef" :config="pageConfig" @edit="handleEdit">
    </ProPage>

    <CommentDrawer ref="drawerRef" @success="handleSuccess" />
  </div>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";

import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";
import CommentDrawer from "./comment-drawer.vue";

import { fetchListComments } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t } from "@/core/i18n";

const pageRef = ref();
const drawerRef = ref();

// 内容类型/审核状态 -> 显示文案映射（裸枚举直接展示无翻译）
const commentTypeLabelMap: Record<string, string> = {
  CONTENT_TYPE_PRODUCT: $t("pages.mall.comment.typeLabel.CONTENT_TYPE_PRODUCT"),
  CONTENT_TYPE_SHOP: $t("pages.mall.comment.typeLabel.CONTENT_TYPE_SHOP"),
  CONTENT_TYPE_POST: $t("pages.mall.comment.typeLabel.CONTENT_TYPE_POST"),
};

// @ts-ignore - intentional incomplete record mirroring existing convention
const commentTypeTagTypeMap: Record<
  string,
  "success" | "primary" | "warning" | "danger" | "info"
> = {
  CONTENT_TYPE_PRODUCT: "primary",
  CONTENT_TYPE_SHOP: "warning",
  CONTENT_TYPE_POST: "info",
};

const commentStatusLabelMap: Record<string, string> = {
  STATUS_PENDING: $t("pages.mall.comment.statusLabel.STATUS_PENDING"),
  STATUS_APPROVED: $t("pages.mall.comment.statusLabel.STATUS_APPROVED"),
  STATUS_REJECTED: $t("pages.mall.comment.statusLabel.STATUS_REJECTED"),
  STATUS_SPAM: $t("pages.mall.comment.statusLabel.STATUS_SPAM"),
};
// @ts-ignore - intentional incomplete record mirroring existing convention
const commentStatusTagTypeMap: Record<
  string,
  "success" | "primary" | "warning" | "danger" | "info"
> = {
  STATUS_PENDING: "warning",
  STATUS_APPROVED: "success",
  STATUS_REJECTED: "danger",
  STATUS_SPAM: "info",
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
      const result = await fetchListComments(
        new PaginationQuery({
          paging: { page: page || 1, pageSize: pageSize || 10 },
          formValues: queryParams,
        })
      );
      return { items: result.items || [], total: result.total || 0 };
    },
    toolbar: [],
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
        prop: "contentType",
        label: $t("pages.mall.comment.contentType"),
        minWidth: 100,
        align: "left",
        cellType: "tag",
        tagTypeMap: commentTypeTagTypeMap,
        labelMap: commentTypeLabelMap,
      },
      {
        prop: "objectId",
        label: $t("pages.mall.comment.objectId"),
        minWidth: 100,
        align: "left",
      },
      {
        prop: "content",
        label: $t("pages.mall.comment.content"),
        minWidth: 200,
        align: "left",
      },
      {
        prop: "status",
        label: $t("pages.mall.comment.status"),
        minWidth: 100,
        align: "left",
        cellType: "tag",
        tagTypeMap: commentStatusTagTypeMap,
        labelMap: commentStatusLabelMap,
      },
      {
        prop: "createdBy",
        label: $t("pages.mall.comment.createdBy"),
        minWidth: 100,
        align: "left",
      },
      {
        prop: "createdAt",
        label: $t("pages.mall.comment.createdAt"),
        minWidth: 160,
        cellType: "date",
        dateFormat: "YYYY-MM-DD HH:mm:ss",
      },
      {
        prop: "action",
        label: $t("common.table.action"),
        fixed: "right",
        width: 100,
        cellType: "tool",
        buttons: [
          { name: "edit", label: $t("common.button.edit"), icon: "lucide:pen-line" },
        ],
      },
    ],
  },
}));

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
