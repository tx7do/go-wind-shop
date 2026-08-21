<script setup lang="ts">
definePageMeta({
  layout: 'account',
  middleware: 'auth',
})
import { computed } from 'vue';
import { XIcon } from '@/plugins/xicon';
import { useListInvoices } from '@/api/composables';
import { useAccessStore } from '@/stores/modules/core/access.state';
import { useUserStore } from '@/stores/modules/core/user.state';

const { t } = useI18n();
const localePath = useLocalePath();

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

function invoiceTypeLabel(it: InvoiceType | undefined): string {
  return t(INVOICE_TYPE_LABEL_KEY[it ?? 'INVOICE_TYPE_UNSPECIFIED']);
}

function formatCreatedAt(ts: string | undefined): string {
  if (!ts) return '—';
  try {
    return new Date(ts).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' });
  } catch {
    return '—';
  }
}

function displayAmount(v: number | undefined, currency?: string): string {
  const prefix = currency === 'CNY' ? t('mall.product.currencyCny') : '';
  return prefix + (v ?? 0);
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
    <!-- 未登录 -->
    <div
      v-if="!isLogin"
      class="flex flex-col items-center gap-6 rounded-2xl border border-border bg-card p-12 text-center"
    >
      <XIcon icon="carbon:locked" :size="48" class="text-muted-foreground" />
      <p class="text-lg text-muted-foreground">{{ t('authentication.login.please_login') }}</p>
      <UiButton @click="navigateTo(localePath('/login'))">
        {{ t('navbar.user.login') }}
      </UiButton>
    </div>

    <!-- 加载中 -->
    <div v-else-if="invoicesLoading" class="overflow-hidden rounded-2xl border border-border bg-card">
      <div class="border-b border-border bg-muted/40 px-6 py-3">
        <UiSkeleton class="h-4 w-48" />
      </div>
      <div v-for="i in 4" :key="i" class="border-b border-border px-6 py-4 last:border-b-0">
        <div class="flex items-center gap-4">
          <UiSkeleton class="h-4 w-10" />
          <UiSkeleton class="h-5 w-20" />
          <UiSkeleton class="h-4 w-14 ml-auto" />
          <UiSkeleton class="h-4 w-24" />
        </div>
      </div>
    </div>

    <!-- 错误态：网络/服务端错误，与"空"区分开，并提供重试 -->
    <UiAppEmpty
      v-else-if="invoicesError"
      variant="error"
    >
      <template #action>
        <UiButton variant="outline" size="sm" @click="invoicesQuery.refetch()">
          {{ t('ui.button.retry') }}
        </UiButton>
      </template>
    </UiAppEmpty>

    <!-- 已加载：空 / 列表 -->
    <div v-else class="flex flex-col gap-4">
      <!-- 空态 -->
      <UiAppEmpty
        v-if="invoices.length === 0"
        variant="noData"
        :description="t('invoices.empty')"
      />

      <!-- 发票列表 -->
      <div v-else class="overflow-x-auto rounded-2xl border border-border bg-card">
        <div class="min-w-[640px]">
          <div class="border-b border-border bg-muted/40 px-6 py-3 text-xs font-medium uppercase tracking-wide text-muted-foreground">
            <div class="grid grid-cols-[80px_1fr_1fr_120px_140px] items-center gap-4">
              <span>{{ t('invoices.table.id') }}</span>
              <span>{{ t('invoices.table.invoiceNumber') }}</span>
              <span>{{ t('invoices.table.invoiceType') }}</span>
              <span class="text-right">{{ t('invoices.table.amount') }}</span>
              <span class="text-right">{{ t('invoices.table.createdAt') }}</span>
            </div>
          </div>

          <div
            v-for="invoice in invoices"
            :key="invoice.id"
            class="border-b border-border px-6 py-4 last:border-b-0"
          >
            <div class="grid grid-cols-[80px_1fr_1fr_120px_140px] items-center gap-4">
              <span class="text-sm tabular-nums text-foreground">#{{ invoice.id ?? '—' }}</span>

              <span class="text-sm text-foreground">
                {{ invoice.invoiceNumber || '—' }}
              </span>

              <span class="text-xs text-muted-foreground">
                {{ invoiceTypeLabel(invoice.invoiceType) }}
              </span>

              <span class="text-right text-sm tabular-nums text-foreground">
                {{ displayAmount(invoice.amount, invoice.currency) }}
              </span>
              <span class="text-right text-xs tabular-nums text-muted-foreground">
                {{ formatCreatedAt(invoice.createdAt) }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- 分页：仅当超过一页时显示 -->
      <div
        v-if="totalPages > 1"
        class="flex items-center justify-center gap-4 py-2"
      >
        <UiButton
          variant="outline"
          size="sm"
          :disabled="page <= 1"
          @click="goToPage(page - 1)"
        >
          {{ t('ui.pagination.previous') }}
        </UiButton>
        <span class="text-xs tabular-nums text-muted-foreground">
          {{ t('ui.pagination.page', { current: page, total: totalPages }) }}
        </span>
        <UiButton
          variant="outline"
          size="sm"
          :disabled="page >= totalPages"
          @click="goToPage(page + 1)"
        >
          {{ t('ui.pagination.next') }}
        </UiButton>
      </div>
    </div>
  </LayoutSectionContainer>
</template>
