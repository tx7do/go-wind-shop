<template>
  <EchartsUI ref="chartRef" height="100%" />
</template>

<script lang="ts" setup>
import type { EChartsOption } from "echarts";

import { EchartsUI, EchartsUIType, useEcharts } from "@/plugins/echarts";
import { usePreferences } from "@/core/preferences";

const chartRef = ref<EchartsUIType>();
const { renderEcharts } = useEcharts(chartRef);
const { isDark } = usePreferences();

const chartOptions = computed<EChartsOption>(() => ({
  color: ["#4080ff", "#36d399"],
  grid: {
    bottom: 24,
    left: 40,
    right: 16,
    top: 16,
  },
  legend: {
    data: ["\u8bbf\u95ee\u8d8b\u52bf", "\u6708\u8bbf\u95ee\u91cf"],
    right: 0,
    top: 0,
    itemWidth: 20,
    itemHeight: 10,
    itemGap: 20,
    icon: "roundRect",
    textStyle: {
      color: isDark.value ? "#CFD3DC" : "#606266",
      fontSize: 12,
    },
  },
  series: [
    {
      areaStyle: {
        color: {
          colorStops: [
            { offset: 0, color: "rgba(64,128,255,0.25)" },
            { offset: 1, color: "rgba(64,128,255,0.02)" },
          ],
          x: 0,
          x2: 0,
          y: 0,
          y2: 1,
        },
      },
      data: [
        111, 2000, 6000, 16_000, 33_333, 55_555, 64_000, 33_333, 18_000, 36_000, 70_000, 42_444,
        23_222, 13_000, 8000, 4000, 1200, 333, 222, 111,
      ],
      lineStyle: {
        width: 2,
      },
      name: "访问趋势",
      smooth: 0.4,
      symbol: "none",
      type: "line",
    },
    {
      areaStyle: {
        color: {
          colorStops: [
            { offset: 0, color: "rgba(54,211,153,0.2)" },
            { offset: 1, color: "rgba(54,211,153,0.01)" },
          ],
          x: 0,
          x2: 0,
          y: 0,
          y2: 1,
        },
      },
      data: [
        33, 66, 88, 333, 3333, 6200, 20_000, 3000, 1200, 13_000, 22_000, 11_000, 2221, 1201, 390,
        198, 60, 30, 22, 11,
      ],
      lineStyle: {
        width: 2,
      },
      name: "月访问量",
      smooth: 0.4,
      symbol: "none",
      type: "line",
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
    boundaryGap: false,
    data: Array.from({ length: 18 }).map((_item, index) => `${index + 6}:00`),
    splitLine: {
      show: false,
    },
    type: "category",
  },
  yAxis: [
    {
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
      max: 80_000,
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
  ],
}));

watch(
  () => chartOptions.value,
  (options) => {
    renderEcharts(options);
  },
  { immediate: true, deep: true }
);
</script>
