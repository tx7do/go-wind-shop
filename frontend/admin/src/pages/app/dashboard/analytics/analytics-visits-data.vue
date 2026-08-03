<template>
  <EchartsUI ref="chartRef" height="100%" />
</template>

<script lang="ts" setup>
import type { EChartsOption } from "echarts";

import { EchartsUI, EchartsUIType, useEcharts } from "@/plugins/echarts";
import { $t } from "@/core/i18n";
import { usePreferences } from "@/core/preferences";

const chartRef = ref<EchartsUIType>();
const { renderEcharts } = useEcharts(chartRef);
const { isDark } = usePreferences();

const categories = computed(() => [
  $t("pages.dashboard.web"),
  $t("pages.dashboard.mobile"),
  $t("pages.dashboard.tablet"),
  $t("pages.dashboard.desktop"),
  $t("pages.dashboard.thirdParty"),
  $t("pages.dashboard.other"),
]);

const visitsData = [4200, 3800, 2600, 2100, 1600, 800];
const trendData = [3600, 4100, 2200, 2400, 1200, 1100];

const chartOptions = computed<EChartsOption>(() => ({
  color: ["#4080ff", "#36d399"],
  grid: {
    bottom: 8,
    left: 90,
    right: 32,
    top: 32,
  },
  legend: {
    data: [
      $t("pages.dashboard.visitsLegend"),
      $t("pages.dashboard.trendLegend"),
    ],
    right: 0,
    top: 0,
    itemWidth: 12,
    itemHeight: 8,
    itemGap: 20,
    icon: "roundRect",
    textStyle: {
      color: isDark.value ? "#CFD3DC" : "#606266",
      fontSize: 12,
    },
  },
  series: [
    {
      data: visitsData,
      itemStyle: {
        borderRadius: [0, 4, 4, 0],
      },
      name: $t("pages.dashboard.visitsLegend"),
      type: "bar",
      barMaxWidth: 12,
    },
    {
      data: trendData,
      itemStyle: {
        borderRadius: [0, 4, 4, 0],
      },
      name: $t("pages.dashboard.trendLegend"),
      type: "bar",
      barMaxWidth: 12,
    },
  ],
  tooltip: {
    backgroundColor: isDark.value ? "rgba(40,40,40,0.96)" : "rgba(255,255,255,0.96)",
    borderColor: isDark.value ? "#4c4d4f" : "#eee",
    borderRadius: 8,
    padding: [12, 16],
    textStyle: {
      color: isDark.value ? "#ffffff" : "#303133",
      fontSize: 13,
    },
    trigger: "axis",
    axisPointer: {
      type: "shadow",
    },
  },
  xAxis: {
    axisLine: { show: false },
    axisLabel: {
      color: isDark.value ? "#8c8c8c" : "#909399",
      fontSize: 11,
    },
    axisTick: { show: false },
    splitLine: {
      lineStyle: {
        color: isDark.value ? "rgba(255,255,255,0.08)" : "rgba(0,0,0,0.06)",
        type: "solid",
      },
      show: true,
    },
    type: "value",
  },
  yAxis: {
    axisLine: { show: false },
    axisLabel: {
      color: isDark.value ? "#CFD3DC" : "#606266",
      fontSize: 12,
    },
    axisTick: { show: false },
    data: categories.value,
    splitLine: { show: false },
    type: "category",
  },
}));

watch(
  () => chartOptions.value,
  (options) => {
    renderEcharts(options);
  },
  { immediate: true, deep: true }
);
</script>
