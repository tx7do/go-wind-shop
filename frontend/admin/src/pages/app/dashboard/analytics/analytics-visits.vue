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

const chartOptions = computed<EChartsOption>(() => ({
  color: ["#4080ff"],
  grid: {
    bottom: 24,
    left: 40,
    right: 16,
    top: 16,
  },
  series: [
    {
      barMaxWidth: 32,
      data: [3000, 2000, 3333, 5000, 3200, 4200, 3200, 2100, 3000, 5100, 6000, 3200, 4800],
      itemStyle: {
        borderRadius: [4, 4, 0, 0],
        color: {
          colorStops: [
            { offset: 0, color: "#4080ff" },
            { offset: 1, color: "rgba(64,128,255,0.6)" },
          ],
          x: 0,
          x2: 0,
          y: 0,
          y2: 1,
        },
      },
      type: "bar",
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
    axisPointer: {
      lineStyle: {
        color: "#4080ff",
        opacity: 0.3,
        width: 1,
      },
    },
    trigger: "axis",
  },
  xAxis: {
    axisLine: {
      show: false,
    },
    axisLabel: {
      color: isDark.value ? "#CFD3DC" : "#606266",
      fontSize: 11,
    },
    axisTick: {
      show: false,
    },
    data: Array.from({ length: 12 }).map(
      (_item, index) => `${index + 1}${$t("pages.dashboard.month")}`
    ),
    splitLine: {
      show: false,
    },
    type: "category",
  },
  yAxis: {
    axisLine: {
      show: false,
    },
    axisLabel: {
      color: isDark.value ? "#8c8c8c" : "#909399",
      fontSize: 11,
      formatter: (val: number) => {
        if (val >= 1000) return `${Math.round(val / 1000)}k`;
        return `${val}`;
      },
    },
    axisTick: {
      show: false,
    },
    max: 8000,
    splitArea: {
      show: false,
    },
    splitLine: {
      lineStyle: {
        color: isDark.value ? "rgba(255,255,255,0.06)" : "rgba(0,0,0,0.05)",
        type: "solid",
      },
      show: true,
    },
    splitNumber: 4,
    type: "value",
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
