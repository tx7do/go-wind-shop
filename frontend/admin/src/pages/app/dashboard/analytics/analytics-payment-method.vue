<template>
  <EchartsUI ref="chartRef" height="100%" />
</template>

<script lang="ts" setup>
import type { EChartsOption } from "echarts";

import { EchartsUI, EchartsUIType, useEcharts } from "@/plugins/echarts";
import { usePreferences } from "@/core/preferences";

import { computed, ref, watch } from "vue";

interface MethodSlice {
  name: string;
  value: number;
  color: string;
}

const props = defineProps<{
  data: MethodSlice[];
}>();

const chartRef = ref<EchartsUIType>();
const { renderEcharts } = useEcharts(chartRef);
const { isDark } = usePreferences();

// 饼图分片统一使用一个中性蓝，避免对“支付方式”做强语义着色。
const SLICE_COLOR = "#4080ff";
const TEXT_PRIMARY = computed(() => (isDark.value ? "#ffffff" : "#303133"));
const TEXT_SECONDARY = computed(() => (isDark.value ? "#8c8c8c" : "#909399"));
const TOOLTIP_BG = computed(() =>
  isDark.value ? "rgba(40,40,40,0.96)" : "rgba(255,255,255,0.96)"
);
const TOOLTIP_BORDER = computed(() => (isDark.value ? "#4c4d4f" : "#eee"));

const chartOptions = computed<EChartsOption>(() => {
  const seriesData = props.data.map((slice) => ({
    name: slice.name,
    value: slice.value,
    itemStyle: { color: SLICE_COLOR },
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
