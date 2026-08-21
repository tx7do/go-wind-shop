<template>
  <div class="app-container h-full flex flex-1 flex-col">
    <ProPage ref="pageRef" :config="pageConfig" @add="handleAdd" @edit="handleEdit">
    </ProPage>

    <InvoiceDrawer ref="drawerRef" @success="handleSuccess" />
  </div>
</template>

<script lang="ts" setup>
import { ref, computed } from "vue";

import ProPage from "@/components/Pro/ProPage/index.vue";
import type { ProPageConfig } from "@/components/Pro/ProPage/types";
import InvoiceDrawer from "./invoice-drawer.vue";

import { fetchListInvoices, useDeleteInvoice } from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";
import { $t } from "@/core/i18n";

const { mutateAsync: deleteInvoice } = useDeleteInvoice();

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
      const result = await fetchListInvoices(
        new PaginationQuery({
          paging: { page: page || 1, pageSize: pageSize || 10 },
          formValues: queryParams,
        })
      );
      return { items: result.items || [], total: result.total || 0 };
    },
    deleteAction: async (ids: string) => {
      await deleteInvoice({ id: ids as any });
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
        prop: "orderId",
        label: $t("pages.mall.invoice.orderId"),
        minWidth: 100,
        align: "left",
      },
      {
        prop: "userId",
        label: $t("pages.mall.invoice.userId"),
        minWidth: 100,
        align: "left",
      },
      {
        prop: "invoiceNumber",
        label: $t("pages.mall.invoice.invoiceNumber"),
        minWidth: 140,
        align: "left",
      },
      {
        prop: "invoiceType",
        label: $t("pages.mall.invoice.invoiceType"),
        minWidth: 120,
        align: "left",
      },
      {
        prop: "status",
        label: $t("pages.mall.invoice.status"),
        minWidth: 100,
        align: "left",
      },
      {
        prop: "buyerName",
        label: $t("pages.mall.invoice.buyerName"),
        minWidth: 140,
        align: "left",
      },
      {
        prop: "amount",
        label: $t("pages.mall.invoice.amount"),
        minWidth: 120,
        align: "left",
      },
      {
        prop: "currency",
        label: $t("pages.mall.invoice.currency"),
        minWidth: 100,
        align: "left",
      },
      {
        prop: "createdAt",
        label: $t("pages.mall.invoice.createdAt"),
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
