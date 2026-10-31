<script setup lang="ts">
import { useId } from 'vue'

// 同一页会出现两个徽标（登录页的移动条 + 桌面列），再叠上共享布局里的 Header。
// 渐变 id 必须逐个唯一：重复 id 时浏览器会挑文档里第一个，而那个可能正好在
// display:none 的兄弟节点里——描边会退化成不绘制（实测桌面登录页整块空白）。
const gradientId = `gwind-${useId().replace(/[^a-zA-Z0-9_-]/g, '')}`
</script>

<template>
  <svg
    xmlns="http://www.w3.org/2000/svg"
    viewBox="0 0 48 48"
    fill="none"
    role="img"
    aria-label="GoWind"
  >
    <defs>
      <linearGradient :id="gradientId" x1="6" y1="40" x2="42" y2="8" gradientUnits="userSpaceOnUse">
        <stop stop-color="var(--logo-from)" />
        <stop offset=".5" stop-color="var(--logo-mid)" />
        <stop offset="1" stop-color="var(--logo-to)" />
      </linearGradient>
    </defs>
    <path d="M25.6 39.2a4 4 0 1 0 2.4-7.2H4" :stroke="`url(#${gradientId})`" stroke-width="4.5" stroke-linecap="round" />
    <path d="M35 16a5 5 0 1 1 4 8H4" :stroke="`url(#${gradientId})`" stroke-width="4.5" stroke-linecap="round" />
    <path d="M19.6 8.8a4 4 0 1 1 2.4 7.2H4" :stroke="`url(#${gradientId})`" stroke-width="4.5" stroke-linecap="round" />
  </svg>
</template>
