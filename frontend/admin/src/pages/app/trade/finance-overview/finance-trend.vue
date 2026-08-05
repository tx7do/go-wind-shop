<template>
  <EchartsUI ref="chartRef" height="100%" />
</template>

<script lang="ts" setup>
import type { EChartsOption } from "echarts";

import { EchartsUI, EchartsUIType, useEcharts } from "@/plugins/echarts";
import { usePreferences } from "@/core/preferences";

import { computed, ref, watch } from "vue";

interface TrendPoint {
  date: string;
  value: number;
}

const props = defineProps<{
  data: TrendPoint[];
}>();

const chartRef = ref<EchartsUIType>();
const { renderEcharts } = useEcharts(chartRef);
const { isDark } = usePreferences();

// 主题色——坐标轴/文本/网格线随亮暗主题切换。ECharts 渲染到 canvas，
// 无法解析 CSS var()，故用字面量 hex 表达。颜色取自 Element Plus 边框色
// 的语义等价（非直接读取 var，因 canvas 上下文不参与 CSS 解析）。
const TEXT_PRIMARY = computed(() => (isDark.value ? "#ffffff" : "#303133"));
const TEXT_SECONDARY = computed(() => (isDark.value ? "#8c8c8c" : "#909399"));
const TOOLTIP_BG = computed(() =>
  isDark.value ? "rgba(40,40,40,0.96)" : "rgba(255,255,255,0.96)"
);
const TOOLTIP_BORDER = computed(() => (isDark.value ? "#4c4d4f" : "#eee"));
const AXIS_LINE = computed(() => (isDark.value ? "#4c4d4f" : "#dcdfe6"));
const SPLIT_LINE = computed(() =>
  isDark.value ? "rgba(255,255,255,0.08)" : "rgba(0,0,0,0.08)"
);
// 趋势线统一中性蓝，避免对“金额趋势”做强语义着色。
const LINE_COLOR = "#4080ff";

const chartOptions = computed<EChartsOption>(() => {
  const dates = props.data.map((p) => p.date);
  const values = props.data.map((p) => p.value);

  return {
    tooltip: {
      trigger: "axis",
      backgroundColor: TOOLTIP_BG.value,
      borderColor: TOOLTIP_BORDER.value,
      borderRadius: 8,
      padding: [10, 14],
      textStyle: { color: TEXT_PRIMARY.value, fontSize: 13 },
    },
    grid: { left: 40, right: 20, top: 20, bottom: 30, containLabel: true },
    xAxis: {
      type: "category",
      data: dates,
      boundaryGap: false,
      axisLine: { lineStyle: { color: AXIS_LINE.value } },
      axisLabel: { color: TEXT_SECONDARY.value, fontSize: 11 },
      axisTick: { lineStyle: { color: AXIS_LINE.value } },
    },
    yAxis: {
      type: "value",
      axisLine: { lineStyle: { color: AXIS_LINE.value } },
      axisLabel: { color: TEXT_SECONDARY.value, fontSize: 11 },
      splitLine: { lineStyle: { color: SPLIT_LINE.value } },
    },
    series: [
      {
        type: "line",
        data: values,
        smooth: true,
        symbol: "circle",
        symbolSize: 6,
        lineStyle: { color: LINE_COLOR, width: 2 },
        itemStyle: { color: LINE_COLOR },
        areaStyle: {
          color: {
            type: "linear",
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: "rgba(64,128,255,0.25)" },
              { offset: 1, color: "rgba(64,128,255,0.02)" },
            ],
          },
        },
      },
    ],
  };
});

watch(
  () => chartOptions.value,
  (options) => {
    renderEcharts(options);
  },
  { immediate: true, deep: true }
);
</script>
