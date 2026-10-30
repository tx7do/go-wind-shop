<script setup lang="ts">
definePageMeta({
  layout: 'account',
  middleware: 'auth',
})
import { computed } from 'vue';
import { useListInvoices } from '@/api/composables';
import { useAccessStore } from '@/stores/modules/core/access.state';
import { useUserStore } from '@/stores/modules/core/user.state';

const { t } = useI18n();

useHead({ title: t('invoices.title') });

const accessStore = useAccessStore();
const userStore = useUserStore();

const isLogin = computed(() => {
  const token = accessStore.accessToken;
  return !!token?.value && !accessStore.loginExpired;
});
const currentUserId = computed(() => userStore.user?.id ?? 0);

type InvoiceType = 'VAT_GENERAL' | 'VAT_SPECIAL' | 'ELECTRONIC' | 'INVOICE_TYPE_UNSPECIFIED';
type InvoiceEntity = {
  id?: number;
  invoiceNumber?: string;
  invoiceType?: InvoiceType;
  amount?: number;
  currency?: string;
  createdAt?: string;
};

// 发票表注入了 UserPrivacy：core 按 viewer.user_id 自动注入 WHERE，
// 跨用户 invoice_id 查不到。BFF fail-closed：List 强制注入 userId=当前登录
// 用户，JSON 解析失败即拒。故此处的 query 不需要带 createdBy/userId 过滤参数——
// BFF 已注入。
// enabled 守卫：未登录 / currentUserId 尚未 hydrate（=0）时不发请求，
// 避免预 hydrate 窗口发出无过滤请求。
const PAGE_SIZE = 10;
const page = ref(1);
const invoicesQuery = useListInvoices(
  computed(() => ({
    page: page.value,
    pageSize: PAGE_SIZE,
    noPaging: false,
    sorting: [{ field: 'id', direction: 'DESC' }],
  })),
  { enabled: computed(() => isLogin.value && currentUserId.value > 0) },
);
const invoices = computed<InvoiceEntity[]>(() => {
  const items = (invoicesQuery.data?.value as any)?.items ?? [];
  return (items as InvoiceEntity[]) ?? [];
});
const invoicesLoading = computed(() => invoicesQuery.isPending.value);
const invoicesError = computed(() => invoicesQuery.isError.value);

const totalCount = computed(() => (invoicesQuery.data?.value as any)?.total ?? 0);
const totalPages = computed(() => Math.max(1, Math.ceil(totalCount.value / PAGE_SIZE)));

function goToPage(p: number) {
  const clamped = Math.min(Math.max(1, p), totalPages.value);
  if (clamped !== page.value) {
    page.value = clamped;
  }
}

const INVOICE_TYPE_LABEL_KEY: Record<InvoiceType, string> = {
  VAT_GENERAL: 'invoices.invoiceType.vat_general',
  VAT_SPECIAL: 'invoices.invoiceType.vat_special',
  ELECTRONIC: 'invoices.invoiceType.electronic',
  INVOICE_TYPE_UNSPECIFIED: 'invoices.invoiceType.unspecified',
};
type BadgeTone = 'neutral' | 'primary' | 'success' | 'warning' | 'danger';
const INVOICE_TYPE_TONE: Record<InvoiceType, BadgeTone> = {
  VAT_GENERAL: 'neutral',
  VAT_SPECIAL: 'primary',
  ELECTRONIC: 'primary',
  INVOICE_TYPE_UNSPECIFIED: 'neutral',
};

function invoiceTypeLabel(it: InvoiceType | undefined): string {
  return t(INVOICE_TYPE_LABEL_KEY[it ?? 'INVOICE_TYPE_UNSPECIFIED']);
}
function invoiceTypeTone(it: InvoiceType | undefined): BadgeTone {
  return INVOICE_TYPE_TONE[it ?? 'INVOICE_TYPE_UNSPECIFIED'];
}

function formatCreatedAt(ts: string | undefined): string {
  if (!ts) return '—';
  try {
    return new Date(ts).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' });
  } catch {
    return '—';
  }
}

// 发票金额单位「分」，uint64 经 JSON 下成字符串，先 Number 再交给价格标签。
function toFen(v: number | undefined): number {
  const fen = Number(v ?? 0);
  return Number.isFinite(fen) ? fen : 0;
}
</script>

<template>
  <LayoutPageHero
    :title="t('invoices.title')"
    :description="t('invoices.subtitle')"
    icon="lucide:file-text"
    size="sm"
  />

  <LayoutSectionContainer>
    <UiAuthGate v-if="!isLogin" />

    <UiListSkeleton v-else-if="invoicesLoading" :rows="5" variant="row" />

    <!-- 错误态：网络/服务端错误，与"空"区分开，并提供重试 -->
    <UiAppEmpty v-else-if="invoicesError" variant="error">
      <template #action>
        <UiButton variant="outline" size="sm" @click="invoicesQuery.refetch()">
          {{ t('ui.button.retry') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <UiAppEmpty
      v-else-if="invoices.length === 0"
      variant="noData"
      :description="t('invoices.empty')"
    />

    <template v-else>
      <div class="overflow-hidden rounded-lg border border-border bg-card shadow-card">
        <div
          class="hidden grid-cols-[1fr_140px_170px] items-center gap-4 border-b border-border px-5 py-2.5 text-xs font-medium text-muted-foreground md:grid"
        >
          <span>{{ t('invoices.table.id') }}</span>
          <span class="text-right">{{ t('invoices.table.amount') }}</span>
          <span class="text-right">{{ t('invoices.table.createdAt') }}</span>
        </div>

        <div
          v-for="invoice in invoices"
          :key="invoice.id"
          class="grid grid-cols-[1fr_auto] items-center gap-x-3 gap-y-1.5 border-b border-border px-4 py-3.5 last:border-b-0 md:grid-cols-[1fr_140px_170px] md:gap-x-4 md:px-5"
        >
          <div class="min-w-0">
            <div class="flex flex-wrap items-center gap-x-2 gap-y-1">
              <span class="text-sm font-semibold tabular-nums text-foreground">
                #{{ invoice.id ?? '—' }}
              </span>
              <UiStatusBadge
                size="sm"
                :tone="invoiceTypeTone(invoice.invoiceType)"
                :label="invoiceTypeLabel(invoice.invoiceType)"
              />
            </div>
            <p class="mt-1 truncate text-[11px] text-muted-foreground md:text-xs">
              {{ invoice.invoiceNumber || '—' }}
            </p>
          </div>

          <div class="flex justify-end">
            <UiPriceTag v-if="toFen(invoice.amount) > 0" :amount="toFen(invoice.amount)" size="md" />
            <span v-else class="text-sm tabular-nums text-muted-foreground">¥0.00</span>
          </div>

          <p class="col-span-2 text-[11px] tabular-nums text-muted-foreground md:col-span-1 md:text-right md:text-xs">
            {{ formatCreatedAt(invoice.createdAt) }}
          </p>
        </div>
      </div>

      <UiPager
        :page="page"
        :total-pages="totalPages"
        class="mt-5"
        @update:page="goToPage"
      />
    </template>
  </LayoutSectionContainer>
</template>
