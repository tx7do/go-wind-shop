<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <ProPage ref="pageRef" :config="pageConfig" @add="handleAdd" @edit="handleEdit"></ProPage>
    <MTDrawer ref="drawerRef" @success="handleSuccess" />
  </div>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";
import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";
import MTDrawer from "./member-tier-drawer.vue";
import { fetchListMemberTiers, useDeleteMemberTier } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t } from "@/core/i18n";

const { mutateAsync: deleteTier } = useDeleteMemberTier();
const pageRef = ref();
const drawerRef = ref();

const pageConfig = computed<ProPageConfig>(() => ({
  skeleton: true, search: { grid: true, fields: [] },
  table: {
    listAction: async (query: any) => {
      const { page, pageSize, ...queryParams } = query;
      const result = await fetchListMemberTiers(new PaginationQuery({ paging: { page: page || 1, pageSize: pageSize || 10 }, formValues: queryParams }));
      return { items: result.items || [], total: result.total || 0 };
    },
    deleteAction: async (ids: string) => { await deleteTier({ id: ids as any }); },
    toolbar: [], toolbarRight: ["add"], defaultToolbar: ["refresh"],
    tableAttrs: { border: true, stripe: false },
    columns: [
      { type: "index", label: $t("common.table.seq"), width: 60 },
      { prop: "id", label: "ID", minWidth: 60 },
      { prop: "name", label: $t("pages.mall.memberTier.name"), minWidth: 120 },
      { prop: "level", label: $t("pages.mall.memberTier.level"), minWidth: 80 },
      { prop: "minSpend", label: $t("pages.mall.memberTier.minSpend"), minWidth: 120, formatter: (r: any) => `¥${((r.minSpend ?? 0) / 100).toFixed(0)}` },
      { prop: "pointsMultiplier", label: $t("pages.mall.memberTier.pointsMultiplier"), minWidth: 100, formatter: (r: any) => `${r.pointsMultiplier ?? 1}x` },
      { prop: "action", label: $t("common.table.action"), fixed: "right", width: 160, cellType: "tool", buttons: [
        { name: "edit", label: $t("common.button.edit"), icon: "lucide:pen-line" },
        { name: "delete", label: $t("common.button.delete"), icon: "lucide:trash-2", attrs: { type: "danger" } },
      ]},
    ],
  },
}));

function handleAdd() { drawerRef.value?.open(); }
function handleEdit(row: any) { drawerRef.value?.open(row); }
function handleSuccess() { pageRef.value?.refresh(); }
</script>
