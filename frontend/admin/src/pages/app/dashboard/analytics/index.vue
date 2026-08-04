<template>
  <div class="analytics-page">
    <!--
      商城分析仪表盘。
      各小部件独立 useQuery 拉取后端数据（staleTime=60s），
      loading/error/empty 各自处理，互不阻塞。
    -->

    <!-- 概览卡片 -->
    <div class="metric-grid mb-5">
      <el-card v-for="(card, index) in metricCards" :key="index" shadow="hover" class="metric-card">
        <div class="metric-header">
          <div class="metric-header__text">
            <div class="metric-title">{{ card.title }}</div>
            <div class="metric-value">
              <el-skeleton v-if="card.isLoading" :rows="0" animated style="width: 60px" />
              <span v-else>{{ card.error ? "—" : card.value }}</span>
            </div>
          </div>
          <div class="metric-header__icon">
            <SvgIcon :icon="card.icon" :size="32" />
          </div>
        </div>
        <div class="metric-footer">
          <span class="metric-footer-label">{{ card.desc }}</span>
          <span class="metric-footer-total">
            {{ $t("pages.dashboard.total") }}
            <strong v-if="!card.isLoading && !card.error">{{ card.totalLabel }}</strong>
          </span>
        </div>
      </el-card>
    </div>

    <!-- 图表 -->
    <div class="chart-grid mb-5">
      <el-card shadow="hover">
        <template #header>
          <div class="card-title-block">
            <span class="card-title">{{ $t("pages.dashboard.chartOrderStatusTitle") }}</span>
            <span class="card-desc">{{ $t("pages.dashboard.chartOrderStatusDesc") }}</span>
          </div>
        </template>
        <div class="chart-state">
          <el-skeleton v-if="orderStatusQuery.isLoading.value" :rows="8" animated />
          <el-alert
            v-else-if="orderStatusQuery.error.value"
            :title="$t('pages.dashboard.loadError')"
            type="error"
            :closable="false"
            show-icon
          />
          <div v-else-if="orderStatusSlices.length === 0" class="chart-empty">
            {{ $t("pages.dashboard.noData") }}
          </div>
          <div v-else class="chart-container chart-container-pie">
            <AnalyticsOrderStatusChart :data="orderStatusSlices" />
          </div>
        </div>
      </el-card>
      <el-card shadow="hover">
        <template #header>
          <div class="card-title-block">
            <span class="card-title">{{ $t("pages.dashboard.chartPaymentMethodTitle") }}</span>
            <span class="card-desc">{{ $t("pages.dashboard.chartPaymentMethodDesc") }}</span>
          </div>
        </template>
        <div class="chart-state">
          <el-skeleton v-if="paymentMethodQuery.isLoading.value" :rows="8" animated />
          <el-alert
            v-else-if="paymentMethodQuery.error.value"
            :title="$t('pages.dashboard.loadError')"
            type="error"
            :closable="false"
            show-icon
          />
          <div v-else-if="paymentMethodSlices.length === 0" class="chart-empty">
            {{ $t("pages.dashboard.noData") }}
          </div>
          <div v-else class="chart-container chart-container-pie">
            <AnalyticsPaymentMethodChart :data="paymentMethodSlices" />
          </div>
        </div>
      </el-card>
      <el-card shadow="hover">
        <template #header>
          <div class="card-title-block">
            <span class="card-title">{{ $t("pages.dashboard.chartRefundStatusTitle") }}</span>
            <span class="card-desc">{{ $t("pages.dashboard.chartRefundStatusDesc") }}</span>
          </div>
        </template>
        <div class="chart-state">
          <el-skeleton v-if="refundStatusQuery.isLoading.value" :rows="8" animated />
          <el-alert
            v-else-if="refundStatusQuery.error.value"
            :title="$t('pages.dashboard.loadError')"
            type="error"
            :closable="false"
            show-icon
          />
          <div v-else-if="refundStatusSlices.length === 0" class="chart-empty">
            {{ $t("pages.dashboard.noData") }}
          </div>
          <div v-else class="chart-container chart-container-pie">
            <AnalyticsRefundStatusChart :data="refundStatusSlices" />
          </div>
        </div>
      </el-card>
    </div>

    <!-- 最近订单 -->
    <el-card shadow="hover">
      <template #header>
        <div class="card-title-block">
          <span class="card-title">{{ $t("pages.dashboard.recentOrdersTitle") }}</span>
          <span class="card-desc">{{ $t("pages.dashboard.recentOrdersDesc") }}</span>
        </div>
      </template>
      <div class="chart-state">
        <el-skeleton v-if="recentOrdersQuery.isLoading.value" :rows="6" animated />
        <el-alert
          v-else-if="recentOrdersQuery.error.value"
          :title="$t('pages.dashboard.loadError')"
          type="error"
          :closable="false"
          show-icon
        />
        <el-table
          v-else
          :data="recentOrders"
          border
          stripe
          size="small"
          :empty-text="$t('pages.dashboard.noData')"
        >
          <el-table-column prop="id" :label="$t('pages.mall.order.id')" width="100" />
          <el-table-column :label="$t('pages.mall.order.status')" width="120">
            <template #default="scope">
              <el-tag
                v-if="scope.row.status && orderStatusLabelMap[scope.row.status]"
                size="small"
                effect="dark"
                round
                :type="orderStatusTagTypeMap[scope.row.status] ?? 'info'"
              >
                {{ orderStatusLabelMap[scope.row.status] }}
              </el-tag>
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column
            prop="totalAmount"
            :label="$t('pages.mall.order.totalAmount')"
            width="120"
            align="right"
          >
            <template #default="scope">
              <span>{{ scope.row.totalAmount ?? "-" }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="currency" :label="$t('pages.mall.order.currency')" width="90" />
          <el-table-column :label="$t('pages.mall.order.createdAt')">
            <template #default="scope">
              <span>{{ formatDateTime(scope.row.createdAt ?? "") }}</span>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </el-card>
  </div>
</template>

<script lang="ts" setup>
import { computed } from "vue";
import { useQuery } from "@tanstack/vue-query";

import SvgIcon from "@/components/SvgIcon/index.vue";
import { $t } from "@/core/i18n";
import { formatDateTime } from "@/utils";

import AnalyticsOrderStatusChart from "./analytics-order-status.vue";
import AnalyticsRefundStatusChart from "./analytics-order-status.vue";
import AnalyticsPaymentMethodChart from "./analytics-payment-method.vue";

import {
  fetchCatalogTotals,
  fetchOrderStatusCounts,
  fetchPaymentMethodCounts,
  fetchOrderTotal,
  fetchPaymentTransactionTotal,
  fetchRefundTotal,
  fetchRefundStatusCounts,
  fetchRecentOrders,
  type DistributionEntry,
} from "@/api/composables/analytics";

// ============================================================
// 订单状态 -> 标签 + 颜色映射（复用 pages.mall.order.statusLabel
// 与既有 trade/order 页面的语义着色，保证一致）。
// ============================================================
const orderStatusLabelMap: Record<string, string> = {
  STATUS_UNSPECIFIED: $t("pages.mall.order.statusLabel.STATUS_UNSPECIFIED"),
  PENDING_PAYMENT: $t("pages.mall.order.statusLabel.PENDING_PAYMENT"),
  PAID: $t("pages.mall.order.statusLabel.PAID"),
  CANCELLED: $t("pages.mall.order.statusLabel.CANCELLED"),
  FULFILLED: $t("pages.mall.order.statusLabel.FULFILLED"),
  CLOSED: $t("pages.mall.order.statusLabel.CLOSED"),
};
// @ts-ignore - intentional incomplete record mirroring existing convention
const orderStatusTagTypeMap: Record<string, "success" | "primary" | "warning" | "danger" | "info"> =
  {
    PENDING_PAYMENT: "warning",
    PAID: "primary",
    CANCELLED: "info",
    FULFILLED: "success",
    CLOSED: "info",
    STATUS_UNSPECIFIED: "info",
  };

// ============================================================
// 退款状态 -> 标签 + 颜色映射（复用 pages.mall.refund.statusLabel
// 与 trade/refund 页面的语义着色，保证一致）。
// 退款状态枚举仅 PENDING/SUCCEEDED/FAILED（无 REFUNDED，
// STATUS_UNSPECIFIED 不在分布统计枚举内）。
// ============================================================
const refundStatusLabelMap: Record<string, string> = {
  PENDING: $t("pages.mall.refund.statusLabel.PENDING"),
  SUCCEEDED: $t("pages.mall.refund.statusLabel.SUCCEEDED"),
  FAILED: $t("pages.mall.refund.statusLabel.FAILED"),
};
// @ts-ignore - intentional incomplete record mirroring existing convention
const refundStatusTagTypeMap: Record<
  string,
  "success" | "primary" | "warning" | "danger" | "info"
> = {
  PENDING: "warning",
  SUCCEEDED: "success",
  FAILED: "danger",
};

// ============================================================
// 概览卡片数据
// ============================================================
const ordersTotalQuery = useQuery({
  queryKey: ["analytics_orderTotal"],
  queryFn: () => fetchOrderTotal(),
  staleTime: 60_000,
});
const paymentsTotalQuery = useQuery({
  queryKey: ["analytics_paymentTxTotal"],
  queryFn: () => fetchPaymentTransactionTotal(),
  staleTime: 60_000,
});
const refundsTotalQuery = useQuery({
  queryKey: ["analytics_refundTotal"],
  queryFn: () => fetchRefundTotal(),
  staleTime: 60_000,
});
const catalogTotalsQuery = useQuery({
  queryKey: ["analytics_catalogTotals"],
  queryFn: () => fetchCatalogTotals(),
  staleTime: 60_000,
});

// ============================================================
// 分布图表数据
// ============================================================
const orderStatusQuery = useQuery({
  queryKey: ["analytics_orderStatusCounts"],
  queryFn: () => fetchOrderStatusCounts(),
  staleTime: 60_000,
});
const paymentMethodQuery = useQuery({
  queryKey: ["analytics_paymentMethodCounts"],
  queryFn: () => fetchPaymentMethodCounts(),
  staleTime: 60_000,
});
const refundStatusQuery = useQuery({
  queryKey: ["analytics_refundStatusCounts"],
  queryFn: () => fetchRefundStatusCounts(),
  staleTime: 60_000,
});

// 将 DistributionEntry[] 映射为图表所需 {name,value,semantic} 形态。
// 颜色不在父组件解析——图表组件内部按 semantic 经 getComputedStyle
// 取当前主题色（ECharts canvas 无法解析 CSS var()）。
// 非法/未知 key 会被过滤。
const orderStatusSlices = computed(() => {
  const raw = orderStatusQuery.data.value ?? [];
  return raw
    .filter((e: DistributionEntry) => orderStatusLabelMap[e.key] !== undefined)
    .map((e: DistributionEntry) => {
      const tagType = orderStatusTagTypeMap[e.key] ?? "info";
      return {
        name: orderStatusLabelMap[e.key] ?? e.key,
        value: e.count,
        semantic: tagType,
      };
    });
});

const refundStatusSlices = computed(() => {
  const raw = refundStatusQuery.data.value ?? [];
  return raw
    .filter((e: DistributionEntry) => refundStatusLabelMap[e.key] !== undefined)
    .map((e: DistributionEntry) => {
      const tagType = refundStatusTagTypeMap[e.key] ?? "info";
      return {
        name: refundStatusLabelMap[e.key] ?? e.key,
        value: e.count,
        semantic: tagType,
      };
    });
});

const paymentMethodSlices = computed(() => {
  const raw = paymentMethodQuery.data.value ?? [];
  return raw
    .filter((e: DistributionEntry) => paymentMethodLabelMap[e.key] !== undefined)
    .map((e: DistributionEntry) => ({
      name: paymentMethodLabelMap[e.key] ?? e.key,
      value: e.count,
      color: "#4080ff",
    }));
});

// 支付方式 -> 显示文案（仅枚举已知渠道，未知归入“其他”）。
const paymentMethodLabelMap: Record<string, string> = {
  ALIPAY: $t("pages.mall.paymentTransaction.methodLabel.ALIPAY"),
  WECHAT: $t("pages.mall.paymentTransaction.methodLabel.WECHAT"),
};

// ============================================================
// 最近订单表格数据
// ============================================================
const recentOrdersQuery = useQuery({
  queryKey: ["analytics_recentOrders"],
  queryFn: () => fetchRecentOrders(20),
  staleTime: 60_000,
});
const recentOrders = computed(() => recentOrdersQuery.data.value?.items ?? []);

// ============================================================
// 概览卡片配置（绑定各 query 的 loading/error/data）。
// ============================================================
const metricCards = computed(() => {
  const ordersValue = ordersTotalQuery.data.value ?? 0;
  const paymentsValue = paymentsTotalQuery.data.value ?? 0;
  const refundsValue = refundsTotalQuery.data.value ?? 0;
  const catalog = catalogTotalsQuery.data.value;

  return [
    {
      title: $t("pages.dashboard.metricCardOrdersTitle"),
      desc: $t("pages.dashboard.metricCardOrdersDesc"),
      icon: "svg:color_card",
      value: ordersValue,
      totalLabel: ordersValue,
      isLoading: ordersTotalQuery.isLoading.value,
      error: ordersTotalQuery.error.value,
    },
    {
      title: $t("pages.dashboard.metricCardPaymentsTitle"),
      desc: $t("pages.dashboard.metricCardPaymentsDesc"),
      icon: "svg:color_bell",
      value: paymentsValue,
      totalLabel: paymentsValue,
      isLoading: paymentsTotalQuery.isLoading.value,
      error: paymentsTotalQuery.error.value,
    },
    {
      title: $t("pages.dashboard.metricCardRefundsTitle"),
      desc: $t("pages.dashboard.metricCardRefundsDesc"),
      icon: "svg:color_bell",
      value: refundsValue,
      totalLabel: refundsValue,
      isLoading: refundsTotalQuery.isLoading.value,
      error: refundsTotalQuery.error.value,
    },
    {
      title: $t("pages.dashboard.metricCardProductsTitle"),
      desc: $t("pages.dashboard.metricCardProductsDesc"),
      icon: "svg:color_download",
      value: catalog?.products ?? 0,
      totalLabel: catalog?.products ?? 0,
      isLoading: catalogTotalsQuery.isLoading.value,
      error: catalogTotalsQuery.error.value,
    },
    {
      title: $t("pages.dashboard.metricCardCatalogTitle"),
      desc: $t("pages.dashboard.metricCardCatalogDesc"),
      icon: "svg:color_cake",
      value: (catalog?.brands ?? 0) + (catalog?.categories ?? 0),
      totalLabel: `${catalog?.brands ?? 0} / ${catalog?.categories ?? 0}`,
      isLoading: catalogTotalsQuery.isLoading.value,
      error: catalogTotalsQuery.error.value,
    },
  ];
});
</script>

<style lang="scss" scoped>
.analytics-page {
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

    .metric-header__icon {
      background: rgba(64, 128, 255, 0.15);
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

    &__icon {
      display: flex;
      flex-shrink: 0;
      align-items: center;
      justify-content: center;
      width: 48px;
      height: 48px;
      background: var(--el-color-primary-light-9);
      border-radius: 12px;
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

    .metric-footer-total {
      color: var(--el-text-color-regular);

      strong {
        font-weight: 600;
        color: var(--el-text-color-primary);
      }
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

.mb-5 {
  margin-bottom: 20px;
}
</style>
