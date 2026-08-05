<template>
  <div class="finance-page">
    <!--
      财务总览。
      数据来源：payment_transaction / payment_refund 的 List RPC（pageSize=100 近期样本）。
      支付为 STUB（force-SUCCEEDED），故金额非真实结算依据——顶部醒目提示。
      各小部件独立 useQuery 拉取（staleTime=60s），loading/error/empty 各自处理。
    -->

    <!-- 样本提示 -->
    <el-alert
      :title="$t('pages.mall.financeOverview.disclaimer')"
      type="warning"
      :closable="false"
      show-icon
      class="mb-5"
    />

    <!-- 金额概览卡片 -->
    <div class="metric-grid mb-5">
      <el-card v-for="(card, index) in metricCards" :key="index" shadow="hover" class="metric-card">
        <div class="metric-header">
          <div class="metric-header__text">
            <div class="metric-title">{{ card.title }}</div>
            <div class="metric-value">
              <el-skeleton v-if="card.isLoading" :rows="0" animated style="width: 80px" />
              <span v-else>{{ card.error ? "—" : card.value }}</span>
            </div>
          </div>
        </div>
        <div class="metric-footer">
          <span class="metric-footer-label">{{ card.desc }}</span>
        </div>
      </el-card>
    </div>

    <!-- 图表 -->
    <div class="chart-grid mb-5">
      <el-card shadow="hover">
        <template #header>
          <div class="card-title-block">
            <span class="card-title">{{ $t("pages.mall.financeOverview.chartTrendTitle") }}</span>
            <span class="card-desc">{{ $t("pages.mall.financeOverview.chartTrendDesc") }}</span>
          </div>
        </template>
        <div class="chart-state">
          <el-skeleton v-if="txQuery.isLoading.value" :rows="8" animated />
          <el-alert
            v-else-if="txQuery.error.value"
            :title="$t('pages.dashboard.loadError')"
            type="error"
            :closable="false"
            show-icon
          />
          <div v-else-if="txTrend.length === 0" class="chart-empty">
            {{ $t("pages.dashboard.noData") }}
          </div>
          <div v-else class="chart-container chart-container-trend">
            <FinanceTrendChart :data="txTrend" />
          </div>
        </div>
      </el-card>
      <el-card shadow="hover">
        <template #header>
          <div class="card-title-block">
            <span class="card-title">{{
              $t("pages.mall.financeOverview.chartTxStatusTitle")
            }}</span>
            <span class="card-desc">{{
              $t("pages.mall.financeOverview.chartTxStatusDesc")
            }}</span>
          </div>
        </template>
        <div class="chart-state">
          <el-skeleton v-if="txQuery.isLoading.value" :rows="8" animated />
          <el-alert
            v-else-if="txQuery.error.value"
            :title="$t('pages.dashboard.loadError')"
            type="error"
            :closable="false"
            show-icon
          />
          <div v-else-if="txStatusSlices.length === 0" class="chart-empty">
            {{ $t("pages.dashboard.noData") }}
          </div>
          <div v-else class="chart-container chart-container-pie">
            <FinanceStatusPieChart :data="txStatusSlices" />
          </div>
        </div>
      </el-card>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed } from "vue";
import { useQuery } from "@tanstack/vue-query";

import { $t } from "@/core/i18n";

import FinanceTrendChart from "./finance-trend.vue";
import FinanceStatusPieChart from "./finance-status-pie.vue";

import {
  fetchListPaymentTransactions,
  fetchListPaymentRefunds,
} from "@/api/composables";
import { PaginationQuery } from "@/core/transport/rest";

// ============================================================
// 数据拉取（pageSize=100 近期样本，非全量）。
// 支付为 STUB，金额非真实结算依据——见顶部提示。
// ============================================================
const txQuery = useQuery({
  queryKey: ["finance_txSample"],
  queryFn: () =>
    fetchListPaymentTransactions(
      new PaginationQuery({
        paging: { page: 1, pageSize: 100 },
        orderBy: ["-created_at"],
      })
    ),
  staleTime: 60_000,
});
const refundQuery = useQuery({
  queryKey: ["finance_refundSample"],
  queryFn: () =>
    fetchListPaymentRefunds(
      new PaginationQuery({
        paging: { page: 1, pageSize: 100 },
        orderBy: ["-created_at"],
      })
    ),
  staleTime: 60_000,
});

// ============================================================
// 支付流水状态 -> 标签 + 语义颜色映射（复用
// pages.mall.paymentTransaction.statusLabel，与 trade/payment-transaction
// 页面一致）。
// ============================================================
const txStatusLabelMap: Record<string, string> = {
  STATUS_UNSPECIFIED: $t("pages.mall.paymentTransaction.statusLabel.STATUS_UNSPECIFIED"),
  PENDING: $t("pages.mall.paymentTransaction.statusLabel.PENDING"),
  SUCCEEDED: $t("pages.mall.paymentTransaction.statusLabel.SUCCEEDED"),
  FAILED: $t("pages.mall.paymentTransaction.statusLabel.FAILED"),
  REFUNDED: $t("pages.mall.paymentTransaction.statusLabel.REFUNDED"),
};
// @ts-ignore - intentional incomplete record mirroring existing convention
const txStatusTagTypeMap: Record<
  string,
  "success" | "primary" | "warning" | "danger" | "info"
> = {
  SUCCEEDED: "success",
  PENDING: "warning",
  FAILED: "danger",
  REFUNDED: "danger",
  STATUS_UNSPECIFIED: "info",
};

// ============================================================
// 金额聚合（样本内 SUCCEEDED 记录求和）。
// 跨币种不可直接相减——按币种分组，仅取样本中频次最高的主币种做展示，
// 其他币种忽略并在卡片标注。金额字段为 number（int64 经 protojson 以 number
// 下发），Number 解析失败回退 0。
// ============================================================
const dominantCurrency = computed(() => {
  const items = (txQuery.data.value?.items ?? []) as any[];
  const counts = new Map<string, number>();
  for (const it of items) {
    const c = it.currency;
    if (typeof c === "string" && c) counts.set(c, (counts.get(c) ?? 0) + 1);
  }
  let dom = "";
  let domN = 0;
  for (const [c, n] of counts) {
    if (n > domN) {
      dom = c;
      domN = n;
    }
  }
  return dom;
});
const txVolume = computed(() => {
  const items = (txQuery.data.value?.items ?? []) as any[];
  const dom = dominantCurrency.value;
  let sum = 0;
  for (const it of items) {
    if (it.status === "SUCCEEDED" && it.currency === dom) {
      sum += Number(it.amount ?? 0);
    }
  }
  return sum;
});
const refundVolume = computed(() => {
  const items = (refundQuery.data.value?.items ?? []) as any[];
  const dom = dominantCurrency.value;
  let sum = 0;
  for (const it of items) {
    if (it.status === "SUCCEEDED" && it.currency === dom) {
      sum += Number(it.amount ?? 0);
    }
  }
  return sum;
});

// ============================================================
// 趋势数据：按 createdAt 的日期（YYYY-MM-DD）聚合主币种 SUCCEEDED 流水金额。
// 按日期升序输出，供折线图。
// ============================================================
const txTrend = computed(() => {
  const items = (txQuery.data.value?.items ?? []) as any[];
  const dom = dominantCurrency.value;
  const byDate = new Map<string, number>();
  for (const it of items) {
    if (it.status !== "SUCCEEDED" || it.currency !== dom) continue;
    const ts = it.createdAt ?? "";
    const day = typeof ts === "string" ? ts.slice(0, 10) : "";
    if (!day) continue;
    byDate.set(day, (byDate.get(day) ?? 0) + Number(it.amount ?? 0));
  }
  const points = Array.from(byDate.entries()).map(([date, value]) => ({
    date,
    value,
  }));
  points.sort((a, b) => (a.date < b.date ? -1 : a.date > b.date ? 1 : 0));
  return points;
});

// ============================================================
// 状态分布：按 status 聚合金额。颜色由图表组件内按 semantic 解析。
// 口径与 txVolume 对齐：REFUNDED 属退款联动后的终态，其金额已计入退款侧，
// 不再纳入"支付状态金额分布"，避免与交易总额卡片口径矛盾。
// 非法/未知 key 被过滤。
// ============================================================
const txStatusSlices = computed(() => {
  const items = (txQuery.data.value?.items ?? []) as any[];
  const byStatus = new Map<string, number>();
  for (const it of items) {
    const st = it.status;
    if (typeof st !== "string") continue;
    if (st === "REFUNDED") continue; // 终态不进分布饼，与 txVolume 口径一致
    byStatus.set(st, (byStatus.get(st) ?? 0) + Number(it.amount ?? 0));
  }
  const slices: {
    name: string;
    value: number;
    semantic: "success" | "primary" | "warning" | "danger" | "info";
  }[] = [];
  for (const [key, val] of byStatus) {
    const label = txStatusLabelMap[key];
    if (label === undefined) continue;
    const sem = txStatusTagTypeMap[key] ?? "info";
    slices.push({ name: label, value: val, semantic: sem });
  }
  return slices;
});

// ============================================================
// 概览卡片（绑定各 query 的 loading/error/data）。
// ============================================================
const metricCards = computed(() => {
  return [
    {
      title: $t("pages.mall.financeOverview.metricCardTransactionVolumeTitle"),
      desc: $t("pages.mall.financeOverview.metricCardTransactionVolumeDesc"),
      value: txVolume.value,
      isLoading: txQuery.isLoading.value,
      error: txQuery.error.value,
    },
    {
      title: $t("pages.mall.financeOverview.metricCardRefundVolumeTitle"),
      desc: $t("pages.mall.financeOverview.metricCardRefundVolumeDesc"),
      value: refundVolume.value,
      isLoading: refundQuery.isLoading.value,
      error: refundQuery.error.value,
    },
    {
      title: $t("pages.mall.financeOverview.metricCardNetRevenueTitle"),
      desc: $t("pages.mall.financeOverview.metricCardNetRevenueDesc"),
      value: txVolume.value - refundVolume.value,
      isLoading: txQuery.isLoading.value || refundQuery.isLoading.value,
      error: txQuery.error.value || refundQuery.error.value,
    },
  ];
});
</script>

<style lang="scss" scoped>
.finance-page {
  padding: 20px;
}

.metric-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.chart-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;

  > :deep(.el-card) {
    flex: 1 1 360px;
    min-width: 0;
  }
}

.metric-card {
  flex: 1 1 220px;
  min-width: 0;
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 12px;
  transition: all 0.3s ease;

  &:hover {
    border-color: var(--el-color-primary-light-5);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.03);
  }

  html.dark & {
    &:hover {
      border-color: var(--el-color-primary-light-3);
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
    }
  }

  :deep(.el-card__body) {
    display: flex;
    flex-direction: column;
    gap: 12px;
    justify-content: space-between;
    height: 100%;
    padding: 20px;
  }

  .metric-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;

    &__text {
      flex: 1;
      min-width: 0;
    }
  }

  .metric-title {
    margin-bottom: 8px;
    font-size: 14px;
    font-weight: 500;
    color: var(--el-text-color-regular);
  }

  .metric-value {
    font-size: 28px;
    font-weight: 700;
    line-height: 1;
    color: var(--el-text-color-primary);
    letter-spacing: -0.5px;
  }

  .metric-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-top: 10px;
    font-size: 12px;
    border-top: 1px solid var(--el-border-color-lighter);

    .metric-footer-label {
      color: var(--el-text-color-regular);
    }
  }
}

.card-title-block {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.card-title {
  display: block;
  padding-top: 2px;
  font-size: 15px;
  font-weight: 600;
  color: var(--el-text-color-primary);
}

.card-desc {
  font-size: 12px;
  font-weight: 400;
  color: var(--el-text-color-secondary);
}

.chart-state {
  width: 100%;
}

.chart-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 320px;
  font-size: 13px;
  color: var(--el-text-color-secondary);
}

.chart-container {
  width: 100%;
  height: 100%;
}

.chart-container-pie {
  height: 320px;
}

.chart-container-trend {
  height: 320px;
}

.mb-5 {
  margin-bottom: 20px;
}
</style>
