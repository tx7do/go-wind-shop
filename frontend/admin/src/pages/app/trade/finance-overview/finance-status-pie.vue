<template>
  <EchartsUI ref="chartRef" height="100%" />
</template>

<script lang="ts" setup>
import type { EChartsOption } from "echarts";

import { EchartsUI, EchartsUIType, useEcharts } from "@/plugins/echarts";
import { usePreferences } from "@/core/preferences";

import { computed, ref, watch } from "vue";

interface StatusSlice {
  name: string;
  value: number;
  semantic: "success" | "primary" | "warning" | "danger" | "info";
}

const props = defineProps<{
  data: StatusSlice[];
}>();

const chartRef = ref<EchartsUIType>();
const { renderEcharts } = useEcharts(chartRef);
const { isDark } = usePreferences();

// 主题色——与 Element Plus 语义色解耦的固定调色板，用于饼图分片着色。
// 注入的 color 来自 props（由父组件按状态语义映射为 success/danger 等）。
const TEXT_PRIMARY = computed(() => (isDark.value ? "#ffffff" : "#303133"));
const TEXT_SECONDARY = computed(() => (isDark.value ? "#8c8c8c" : "#909399"));
const TOOLTIP_BG = computed(() =>
  isDark.value ? "rgba(40,40,40,0.96)" : "rgba(255,255,255,0.96)"
);
const TOOLTIP_BORDER = computed(() => (isDark.value ? "#4c4d4f" : "#eee"));

// 将语义标签解析为 Element Plus 当前主题（亮/暗）下的实际颜色值。
// ECharts 渲染到 canvas，无法解析 CSS var()，故需用 getComputedStyle
// 取出 var 在当前主题下解析后的字面量 hex。该 computed 依赖 isDark
// （上方 TEXT_* 已读取），主题切换时整体重算，颜色随之刷新。
function semanticToColor(sem: StatusSlice["semantic"]): string {
  const varName = `--el-color-${sem}`;
  const resolved = getComputedStyle(document.documentElement)
    .getPropertyValue(varName)
    .trim();
  // 解析失败时回退到中性灰字面量（canvas 可用），避免回退到黑色。
  return resolved || "#909399";
}

const chartOptions = computed<EChartsOption>(() => {
  const seriesData = props.data.map((slice) => ({
    name: slice.name,
    value: slice.value,
    itemStyle: { color: semanticToColor(slice.semantic) },
  }));

  return {
    tooltip: {
      trigger: "item",
      backgroundColor: TOOLTIP_BG.value,
      borderColor: TOOLTIP_BORDER.value,
      borderRadius: 8,
      padding: [10, 14],
      textStyle: { color: TEXT_PRIMARY.value, fontSize: 13 },
      formatter: "{b}<br/>{c} ({d}%)",
    },
    legend: {
      type: "scroll",
      bottom: 0,
      left: "center",
      icon: "circle",
      itemWidth: 8,
      itemHeight: 8,
      textStyle: { color: TEXT_SECONDARY.value, fontSize: 11 },
    },
    series: [
      {
        type: "pie",
        center: ["50%", "42%"],
        radius: ["40%", "62%"],
        avoidLabelOverlap: true,
        padAngle: 2,
        itemStyle: { borderRadius: 4, borderWidth: 0 },
        label: { show: false },
        labelLine: { show: false },
        emphasis: { scaleSize: 4, label: { show: false } },
        data: seriesData,
        animationType: "scale",
        animationEasing: "cubicOut",
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
