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

const PALETTE = ["#4080ff", "#36d399", "#f7ba1e", "#958ce2"] as const;

// 暗黑模式下的文本色
const TEXT_PRIMARY = computed(() => isDark.value ? "#ffffff" : "#303133");
const TEXT_REGULAR = computed(() => isDark.value ? "#CFD3DC" : "#606266");
const TEXT_SECONDARY = computed(() => isDark.value ? "#8c8c8c" : "#909399");
const TEXT_PLACEHOLDER = computed(() => isDark.value ? "#6b6b6b" : "#C0C4CC");
const TOOLTIP_BG = computed(() => isDark.value ? "rgba(40,40,40,0.96)" : "#fff");
const TOOLTIP_BORDER = computed(() => isDark.value ? "#4c4d4f" : "#eee");
const TOOLTIP_TEXT = computed(() => isDark.value ? "#ffffff" : "#333");
const TOOLTIP_SHADOW = computed(() => isDark.value ? "box-shadow: 0 4px 12px rgba(0,0,0,0.3);" : "box-shadow: 0 4px 12px rgba(0,0,0,0.08);");

const salesData = computed(() =>
  [
    { name: $t("pages.dashboard.outsource"), value: 500 },
    { name: $t("pages.dashboard.custom"), value: 310 },
    { name: $t("pages.dashboard.techSupport"), value: 274 },
    { name: $t("pages.dashboard.remote"), value: 400 },
  ].sort((a, b) => b.value - a.value)
);

const totalValue = computed(() =>
  salesData.value.reduce((sum, item) => sum + item.value, 0)
);

const chartOptions = computed<EChartsOption>(() => ({
  tooltip: {
    trigger: "item",
    backgroundColor: TOOLTIP_BG.value,
    borderColor: TOOLTIP_BORDER.value,
    borderRadius: 8,
    padding: [10, 14],
    textStyle: { color: TOOLTIP_TEXT.value, fontSize: 13 },
    formatter: "{b}<br/>{c} ({d}%)",
    extraCssText: TOOLTIP_SHADOW.value,
  },
  graphic: [
    // 中心标签
    {
      type: "group",
      left: "24.5%",
      top: "center",
      children: [
        {
          type: "text",
          style: {
            text: totalValue.value.toLocaleString(),
            fill: TEXT_PRIMARY.value,
            fontSize: 20,
            fontWeight: 700,
            textAlign: "center",
            fontFamily: "DIN Alternate, Roboto, Helvetica, Arial, sans-serif",
          },
          left: "center",
        },
        {
          type: "text",
          style: {
            text: $t("pages.dashboard.total"),
            fill: TEXT_SECONDARY.value,
            fontSize: 11,
            textAlign: "center",
          },
          left: "center",
          top: 26,
        },
      ],
    },
    // 右侧数据列表
    ...salesData.value.map((item, index) => {
      const percent = ((item.value / totalValue.value) * 100).toFixed(1);
      const topOffset = 13 + index * 56;
      return {
        type: "group",
        left: "56%",
        width: "40%",
        top: topOffset,
        children: [
          {
            type: "rect",
            shape: { width: 10, height: 10, r: 2 },
            style: { fill: PALETTE[index] },
          },
          {
            type: "text",
            left: 16,
            style: {
              text: item.name,
              fill: TEXT_REGULAR.value,
              fontSize: 13,
            },
          },
          {
            type: "text",
            left: 16,
            top: 20,
            style: {
              text: `${item.value.toLocaleString()}`,
              fill: TEXT_PRIMARY.value,
              fontSize: 18,
              fontWeight: 700,
              fontFamily: "DIN Alternate, Roboto, Helvetica, Arial, sans-serif",
            },
          },
          {
            type: "text",
            right: 0,
            top: 24,
            style: {
              text: `${percent}%`,
              fill: TEXT_REGULAR.value,
              fontSize: 13,
              fontWeight: 500,
              textAlign: "right",
            },
          },
        ],
      } as any;
    }),
  ],
  series: [
    {
      type: "pie",
      center: ["25%", "50%"],
      radius: ["52%", "80%"],
      avoidLabelOverlap: false,
      padAngle: 2,
      itemStyle: {
        borderRadius: 4,
        borderWidth: 0,
      },
      label: { show: false },
      labelLine: { show: false },
      emphasis: {
        scaleSize: 4,
        label: { show: false },
      },
      data: salesData.value,
      animationType: "scale",
      animationEasing: "cubicOut",
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
