<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <ProPage ref="pageRef" :config="pageConfig" @add="handleAdd" @edit="handleEdit" @groups="handleGroups">
    </ProPage>

    <GrouponDrawer ref="drawerRef" @success="handleSuccess" />
    <GrouponGroupsDrawer ref="groupsDrawerRef" />
  </div>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";

import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";
import GrouponDrawer from "./groupon-drawer.vue";
import GrouponGroupsDrawer from "./groupon-groups-drawer.vue";

import { fetchListGrouponActivities, useDeleteGrouponActivity } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t } from "@/core/i18n";

const { mutateAsync: deleteGroupon } = useDeleteGrouponActivity();

const pageRef = ref();
const drawerRef = ref();
const groupsDrawerRef = ref();

const grouponStatusLabelMap: Record<string, string> = {
  ACTIVE: $t("pages.mall.groupon.statusLabel.active"),
  INACTIVE: $t("pages.mall.groupon.statusLabel.inactive"),
};
// @ts-ignore - intentional incomplete record mirroring existing convention
const grouponStatusTagTypeMap: Record<
  string,
  "success" | "primary" | "warning" | "danger" | "info"
> = {
  ACTIVE: "success",
  INACTIVE: "info",
};

const pageConfig = computed<ProPageConfig>(() => ({
  skeleton: true,
  search: { grid: true, fields: [] },
  table: {
    listAction: async (query: any) => {
      const { page, pageSize, ...queryParams } = query;
      const result = await fetchListGrouponActivities(
        new PaginationQuery({
          paging: { page: page || 1, pageSize: pageSize || 10 },
          formValues: queryParams,
        })
      );
      return { items: result.items || [], total: result.total || 0 };
    },
    deleteAction: async (ids: string) => {
      await deleteGroupon({ id: ids as any });
    },
    toolbar: [],
    toolbarRight: ["add"],
    defaultToolbar: ["refresh", "filter"],
    tableAttrs: { border: true, stripe: false },
    columns: [
      { type: "index", label: $t("common.table.seq"), width: 60 },
      { prop: "id", label: "ID", minWidth: 70, align: "left" },
      { prop: "name", label: $t("pages.mall.groupon.name"), minWidth: 150, align: "left" },
      { prop: "skuId", label: $t("pages.mall.groupon.skuId"), minWidth: 90, align: "left" },
      {
        prop: "groupPrice",
        label: $t("pages.mall.groupon.groupPrice"),
        minWidth: 110,
        align: "left",
        formatter: (row: any) => `¥${((row.groupPrice ?? 0) / 100).toFixed(2)}`,
      },
      { prop: "groupSize", label: $t("pages.mall.groupon.groupSize"), minWidth: 90, align: "left" },
      { prop: "groupValidHours", label: $t("pages.mall.groupon.groupValidHours"), minWidth: 110, align: "left" },
      {
        prop: "status",
        label: $t("pages.mall.groupon.status"),
        minWidth: 90,
        align: "left",
        cellType: "tag",
        tagTypeMap: grouponStatusTagTypeMap,
        labelMap: grouponStatusLabelMap,
      },
      {
        prop: "action",
        label: $t("common.table.action"),
        fixed: "right",
        width: 220,
        cellType: "tool",
        buttons: [
          { name: "edit", label: $t("common.button.edit"), icon: "lucide:pen-line" },
          {
            name: "groups",
            label: $t("pages.mall.groupon.groups"),
            icon: "lucide:users",
          },
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

function handleGroups(row: any) {
  groupsDrawerRef.value?.open(row);
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
