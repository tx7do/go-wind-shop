<template>
  <div class="analytics-page">
    <!-- Overview Cards -->
    <el-row :gutter="16" class="mb-5">
      <el-col v-for="(item, index) in overviewItems" :key="index" :xs="24" :sm="12" :md="6">
        <el-card shadow="hover" class="overview-card">
          <div class="overview-header">
            <div class="overview-header__text">
              <div class="title">{{ item.title }}</div>
              <div class="value-row">
                <span class="value">{{ item.value.toLocaleString() }}</span>
                <span :class="['trend', item.trend >= 0 ? 'trend--up' : 'trend--down']">
                  <SvgIcon :icon="item.trend >= 0 ? 'lucide:trending-up' : 'lucide:trending-down'" :size="14" />
                  {{ Math.abs(item.trend) }}%
                </span>
              </div>
            </div>
            <div class="overview-header__icon">
              <SvgIcon :icon="item.icon" :size="32" />
            </div>
          </div>
          <div class="overview-footer">
            <span class="footer-label">{{ $t("pages.dashboard.vsYesterday") }}</span>
            <span class="footer-total">
              {{ $t("pages.dashboard.total") }}
              <strong>{{ item.totalValue.toLocaleString() }}</strong>
            </span>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Trends Chart -->
    <el-card shadow="hover" class="mb-5">
      <template #header>
        <div class="card-header-tabs">
          <el-radio-group v-model="activeTab" size="small">
            <el-radio-button value="trends">
              {{ $t("pages.dashboard.visitsTrend") }}
            </el-radio-button>
            <el-radio-button value="visits">
              {{ $t("pages.dashboard.monthVisits") }}
            </el-radio-button>
          </el-radio-group>
        </div>
      </template>
      <div class="chart-container chart-container-trend">
        <AnalyticsTrends v-if="activeTab === 'trends'" />
        <AnalyticsVisits v-else />
      </div>
    </el-card>

    <!-- Chart Cards Grid -->
    <el-row :gutter="16">
      <el-col :xs="24" :sm="24" :md="8">
        <el-card shadow="hover">
          <template #header>
            <span class="card-title">{{ $t("pages.dashboard.visitCount") }}</span>
          </template>
          <div class="chart-container chart-container-small">
            <AnalyticsVisitsData />
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="24" :md="8">
        <el-card shadow="hover">
          <template #header>
            <span class="card-title">{{ $t("pages.dashboard.visitSource") }}</span>
          </template>
          <div class="chart-container chart-container-small">
            <AnalyticsVisitsSource />
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="24" :md="8">
        <el-card shadow="hover">
          <template #header>
            <span class="card-title">{{ $t("pages.dashboard.salesDistribution") }}</span>
          </template>
          <div class="chart-container chart-container-small">
            <AnalyticsVisitsSales />
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script lang="ts" setup>
import { ref } from "vue";

import SvgIcon from "@/components/SvgIcon/index.vue";
import { $t } from "@/core/i18n";

import AnalyticsTrends from "./analytics-trends.vue";
import AnalyticsVisits from "./analytics-visits.vue";
import AnalyticsVisitsData from "./analytics-visits-data.vue";
import AnalyticsVisitsSales from "./analytics-visits-sales.vue";
import AnalyticsVisitsSource from "./analytics-visits-source.vue";

// 定义 OverviewItem 接口
interface OverviewItem {
  icon: string;
  title: string;
  totalValue: number;
  trend: number;
  value: number;
}

const overviewItems = ref<OverviewItem[]>([
  {
    icon: "svg:color_card",
    title: $t("pages.dashboard.currentUserCount"),
    totalValue: 120_000,
    trend: 12,
    value: 2000,
  },
  {
    icon: "svg:color_cake",
    title: $t("pages.dashboard.currentAccessCount"),
    totalValue: 500_000,
    trend: -5,
    value: 20_000,
  },
  {
    icon: "svg:color_download",
    title: $t("pages.dashboard.currentDownloadCount"),
    totalValue: 120_000,
    trend: 18,
    value: 8000,
  },
  {
    icon: "svg:color_bell",
    title: $t("pages.dashboard.currentUsageCount"),
    totalValue: 50_000,
    trend: 8,
    value: 5000,
  },
]);

// 当前激活的标签
const activeTab = ref<"trends" | "visits">("trends");
</script>

<style lang="scss" scoped>
.analytics-page {
  padding: 20px;
}

.overview-card {
  border-radius: 12px;
  transition: all 0.3s ease;
  border: 1px solid var(--el-border-color-lighter);

  &:hover {
    border-color: var(--el-color-primary-light-5);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.03);
  }

  // 暗黑模式 hover 阴影
  html.dark & {
    &:hover {
      border-color: var(--el-color-primary-light-3);
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
    }

    .overview-header__icon {
      background: rgba(64, 128, 255, 0.15);
    }
  }

  :deep(.el-card__body) {
    padding: 20px;
    height: 100%;
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    gap: 12px;
  }

  .overview-header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;

    &__text {
      flex: 1;
      min-width: 0;
    }

    &__icon {
      flex-shrink: 0;
      display: flex;
      align-items: center;
      justify-content: center;
      width: 48px;
      height: 48px;
      border-radius: 12px;
      background: var(--el-color-primary-light-9);
    }
  }

  .title {
    font-size: 14px;
    font-weight: 500;
    color: var(--el-text-color-regular);
    margin-bottom: 8px;
  }

  .value-row {
    display: flex;
    align-items: center;
    gap: 10px;
  }

  .value {
    font-size: 28px;
    font-weight: 700;
    color: var(--el-text-color-primary);
    line-height: 1;
    letter-spacing: -0.5px;
  }

  .trend {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    font-size: 12px;
    font-weight: 600;
    line-height: 1;
    padding: 2px 6px;
    border-radius: 4px;

    &--up {
      color: var(--el-color-success);
      background: var(--el-color-success-light-9);
    }

    &--down {
      color: var(--el-color-danger);
      background: var(--el-color-danger-light-9);
    }
  }

  .overview-footer {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding-top: 10px;
    border-top: 1px solid var(--el-border-color-lighter);
    font-size: 12px;

    .footer-label {
      color: var(--el-text-color-regular);
    }

    .footer-total {
      color: var(--el-text-color-regular);

      strong {
        color: var(--el-text-color-primary);
        font-weight: 600;
      }
    }
  }
}

.card-header-tabs {
  display: flex;
  align-items: center;
}

.card-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--el-text-color-primary);
  display: block;
  padding-top: 2px;
}

.chart-container {
  width: 100%;
  height: 100%;
}

.chart-container-trend {
  height: 380px;
}

.chart-container-small {
  height: 300px;
}
</style>
