<script setup lang="ts">
import { cn } from '@/lib/utils'

/**
 * 商品图占位组件。
 *
 * 商品实体 proto 层面未定义图片字段，列表/详情页的图位用本组件
 * 生成一张确定性的抽象几何占位图（同心环 + 放射线条 + 噪点）。
 * 纯 SVG、无外部依赖、SSR 安全、无 hydration mismatch。
 *
 * 后端补齐图片字段后应替换为真实图片。
 */
const props = withDefaults(defineProps<{
  /** 用于生成确定性配色的种子，通常是商品 id 或名称。 */
  seed?: string | number
  /** 透传的 class。 */
  class?: string
}>(), {
  seed: 0,
})

// 两组色相，seed 决定具体取值，保证同商品始终同图、不同商品不同图
const HUE_PAIRS: ReadonlyArray<readonly [number, number]> = [
  [142, 172],
  [190, 210],
  [260, 290],
  [340, 20],
  [45, 90],
  [0, 30],
]

function hash(seed: string | number): number {
  const s = String(seed)
  let h = 2166136261
  for (let i = 0; i < s.length; i++) {
    h ^= s.charCodeAt(i)
    h = Math.imul(h, 16777619)
  }
  return Math.abs(h)
}

const hues = computed(() => HUE_PAIRS[hash(props.seed) % HUE_PAIRS.length])
const angle = computed(() => (hash(props.seed) % 360))
const ringCount = computed(() => 5 + (hash(props.seed) % 4))
const uid = `ph-${useId()}`

// 生成放射线条的端点坐标（确定性）
const rays = computed(() => {
  const cnt = 12 + (hash(props.seed >> 3) % 8)
  const arr: { x1: number; y1: number; x2: number; y2: number }[] = []
  const cx = 200, cy = 200
  for (let i = 0; i < cnt; i++) {
    const a = (i / cnt) * Math.PI * 2 + (angle.value * Math.PI) / 180
    const r1 = 40 + ((hash(props.seed + i) % 30))
    const r2 = 180 + ((hash(props.seed + i * 7) % 20))
    arr.push({
      x1: cx + Math.cos(a) * r1,
      y1: cy + Math.sin(a) * r1,
      x2: cx + Math.cos(a) * r2,
      y2: cy + Math.sin(a) * r2,
    })
  }
  return arr
})

const rings = computed(() => {
  const cnt = ringCount.value
  const arr: { r: number; opacity: number }[] = []
  for (let i = 0; i < cnt; i++) {
    arr.push({
      r: 30 + i * 22,
      opacity: 0.05 + (i / cnt) * 0.06,
    })
  }
  return arr
})
</script>

<template>
  <div
    :class="cn('flex items-center justify-center overflow-hidden', props.class)"
    aria-label="商品示意图占位"
    role="img"
  >
    <svg
      width="100%"
      height="100%"
      viewBox="0 0 400 400"
      preserveAspectRatio="xMidYMid slice"
      xmlns="http://www.w3.org/2000/svg"
    >
      <defs>
        <radialGradient :id="`${uid}-bg`" cx="50%" cy="50%" r="70%">
          <stop offset="0%" :stop-color="`hsl(${hues[0]} 70% 25%)`" stop-opacity="0.5" />
          <stop offset="100%" :stop-color="`hsl(${hues[1]} 60% 12%)`" stop-opacity="0.3" />
        </radialGradient>
        <linearGradient :id="`${uid}-ray`" x1="0%" y1="0%" x2="100%" y2="100%">
          <stop offset="0%" :stop-color="`hsl(${hues[0]} 80% 50%)`" stop-opacity="0.15" />
          <stop offset="100%" :stop-color="`hsl(${hues[1]} 80% 50%)`" stop-opacity="0.03" />
        </linearGradient>
        <pattern :id="`${uid}-dot`" width="16" height="16" patternUnits="userSpaceOnUse">
          <circle cx="8" cy="8" r="1" :fill="`hsl(${hues[1]} 60% 80%)`" fill-opacity="0.08" />
        </pattern>
      </defs>

      <!-- 底色 + 噪点 -->
      <rect width="400" height="400" :fill="`url(#${uid}-bg)`" />
      <rect width="400" height="400" :fill="`url(#${uid}-dot)`" />

      <!-- 同心环 -->
      <g :stroke="`hsl(${hues[0]} 80% 60%)`" fill="none">
        <circle
          v-for="(ring, i) in rings"
          :key="`r-${i}`"
          cx="200"
          cy="200"
          :r="ring.r"
          :stroke-opacity="ring.opacity"
          stroke-width="1"
        />
      </g>

      <!-- 放射线条 -->
      <g :stroke="`url(#${uid}-ray)`" stroke-width="1" fill="none">
        <line
          v-for="(ray, i) in rays"
          :key="`l-${i}`"
          :x1="ray.x1"
          :y1="ray.y1"
          :x2="ray.x2"
          :y2="ray.y2"
        />
      </g>

      <!-- 中心装饰圆 -->
      <circle cx="200" cy="200" r="18" :fill="`hsl(${hues[0]} 80% 50%)`" fill-opacity="0.12" />
    </svg>
  </div>
</template>
