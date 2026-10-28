<template>
  <!-- 商城专属登录插画「风行市集」（docs/brand/ 品牌架构：主标不另造，旋涡徽标作为袋面标识出现）：
     购物袋主体 + 袋面 GoWind 旋涡徽标缓旋，优惠券/购物车/金币/折扣/包裹环绕四周，
     保留品牌双层对旋虚线环流、轨道光点与四向漂移风痕；色板同源 docs/design-language.md 主色 hsl(212 100% 45%)。
     整体不做上下浮动（父级历史 float 已移除）；动效仅保留环流旋转、风痕漂移、旋涡缓旋与星光闪烁，
     且全部尊重 prefers-reduced-motion。 -->
  <svg
    class="gw-mall"
    viewBox="0 0 560 560"
    xmlns="http://www.w3.org/2000/svg"
    role="img"
    aria-label="GoWind 风行商城"
  >
    <defs>
      <linearGradient
        id="gw-mall-grad"
        gradientUnits="userSpaceOnUse"
        x1="0"
        y1="38"
        x2="0"
        y2="474"
      >
        <stop offset="0" stop-color="#70b3ff" />
        <stop offset="0.5" stop-color="#0d74f2" />
        <stop offset="1" stop-color="#005ac2" />
      </linearGradient>
      <linearGradient
        id="gw-mall-bag"
        gradientUnits="userSpaceOnUse"
        x1="0"
        y1="282"
        x2="0"
        y2="436"
      >
        <stop offset="0" stop-color="#fbfdff" />
        <stop offset="1" stop-color="#e9f2fe" />
      </linearGradient>
      <radialGradient id="gw-mall-glow" cx="0.5" cy="0.5" r="0.5">
        <stop offset="0" stop-color="#0d74f2" stop-opacity="0.13" />
        <stop offset="1" stop-color="#0d74f2" stop-opacity="0" />
      </radialGradient>
    </defs>

    <!-- 背景光晕 + 双层对旋虚线环流 + 轨道光点（与品牌插画同构） -->
    <circle cx="280" cy="268" r="252" fill="url(#gw-mall-glow)" />
    <circle
      class="gw-ring2"
      cx="280"
      cy="268"
      r="246"
      fill="none"
      stroke="#0d74f2"
      stroke-opacity="0.07"
      stroke-width="1.5"
      stroke-dasharray="2 16"
      stroke-linecap="round"
    />
    <circle
      class="gw-ring1"
      cx="280"
      cy="268"
      r="212"
      fill="none"
      stroke="#0d74f2"
      stroke-opacity="0.14"
      stroke-width="1.5"
      stroke-dasharray="3 13"
      stroke-linecap="round"
    />
    <g class="gw-orbit">
      <circle cx="280" cy="113" r="4.5" fill="#3b82f6" opacity="0.55" />
      <circle cx="280" cy="423" r="3" fill="#70b3ff" opacity="0.45" />
    </g>
    <g class="gw-orbit2">
      <circle cx="435" cy="268" r="3.5" fill="#0d74f2" opacity="0.4" />
    </g>

    <!-- 四向漂移风痕 -->
    <g
      class="gw-streak"
      fill="none"
      stroke="#0d74f2"
      stroke-opacity="0.12"
      stroke-width="10"
      stroke-linecap="round"
    >
      <path d="M52 190 H140 A20 20 0 0 0 120 156" />
    </g>
    <g
      class="gw-streak gw-s2"
      fill="none"
      stroke="#0d74f2"
      stroke-opacity="0.1"
      stroke-width="10"
      stroke-linecap="round"
    >
      <path d="M400 126 H494 A22 22 0 0 0 474 88" />
    </g>
    <g
      class="gw-streak gw-s3"
      fill="none"
      stroke="#0d74f2"
      stroke-opacity="0.12"
      stroke-width="10"
      stroke-linecap="round"
    >
      <path d="M368 440 H480 A24 24 0 0 0 460 402" />
    </g>
    <g
      class="gw-streak gw-s4"
      fill="none"
      stroke="#0d74f2"
      stroke-opacity="0.09"
      stroke-width="10"
      stroke-linecap="round"
    >
      <path d="M84 465 H161 A18 18 0 0 0 143 435" />
    </g>

    <!-- 环绕的商品元素 -->
    <g transform="translate(112,134)">
      <g transform="rotate(-12 40 28)">
          <path
            d="M17 20 C 10 6 22 -4 38 -8"
            fill="none"
            stroke="#0d74f2"
            stroke-opacity="0.45"
            stroke-width="5.5"
            stroke-linecap="round"
          />
          <rect
            x="0"
            y="0"
            width="80"
            height="56"
            rx="14"
            fill="#f4f9ff"
            stroke="#0d74f2"
            stroke-width="8"
          />
          <circle
            cx="17"
            cy="28"
            r="6"
            fill="#fbfdff"
            stroke="#0d74f2"
            stroke-width="6"
          />
      </g>
    </g>
    <g transform="translate(398,192)">
      <g
        fill="none"
        stroke="#0d74f2"
        stroke-width="9"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <path d="M0 10 H16 L26 52 H72 L84 26 H20" />
        <circle cx="32" cy="68" r="7.5" fill="#0d74f2" stroke="none" opacity="0.85" />
        <circle cx="66" cy="68" r="7.5" fill="#0d74f2" stroke="none" opacity="0.85" />
      </g>
    </g>
    <g transform="translate(120,364)">
      <g>
        <circle
          cx="28"
          cy="28"
          r="26"
          fill="#f4f9ff"
          stroke="#0d74f2"
          stroke-width="9"
        />
        <g
          fill="none"
          stroke="#0d74f2"
          stroke-width="7"
          stroke-linecap="round"
          stroke-linejoin="round"
        >
          <path d="M18 13 L28 25 L38 13" />
          <path d="M28 25 V45" />
          <path d="M19 31 H37" />
          <path d="M19 39 H37" />
        </g>
      </g>
    </g>
    <g transform="translate(446,116)">
      <g>
        <circle
          cx="26"
          cy="26"
          r="24"
          fill="#f4f9ff"
          stroke="#0d74f2"
          stroke-width="8"
        />
        <g fill="none" stroke="#0d74f2" stroke-width="5.5" stroke-linecap="round">
          <circle cx="17.5" cy="17.5" r="5.5" />
          <circle cx="34.5" cy="34.5" r="5.5" />
          <path d="M14 38 L38 14" />
        </g>
      </g>
    </g>
    <g transform="translate(406,372)">
      <g>
        <rect
          x="0"
          y="0"
          width="64"
          height="54"
          rx="10"
          fill="#f4f9ff"
          stroke="#0d74f2"
          stroke-width="9"
        />
        <path
          d="M0 20 H64 M32 20 V54"
          fill="none"
          stroke="#0d74f2"
          stroke-opacity="0.85"
          stroke-width="7"
          stroke-linecap="round"
        />
      </g>
    </g>

    <!-- 点缀星光 -->
    <g class="gw-twinkle" transform="translate(204,104)">
      <path
        d="M0 -10 L2.5 -2.5 L10 0 L2.5 2.5 L0 10 L-2.5 2.5 L-10 0 L-2.5 -2.5 Z"
        fill="#3b82f6"
        opacity="0.5"
      />
    </g>
    <g class="gw-twinkle gw-t2" transform="translate(478,330) scale(0.85)">
      <path
        d="M0 -10 L2.5 -2.5 L10 0 L2.5 2.5 L0 10 L-2.5 2.5 L-10 0 L-2.5 -2.5 Z"
        fill="#3b82f6"
        opacity="0.45"
      />
    </g>
    <g class="gw-twinkle gw-t3" transform="translate(162,300) scale(0.7)">
      <path
        d="M0 -10 L2.5 -2.5 L10 0 L2.5 2.5 L0 10 L-2.5 2.5 L-10 0 L-2.5 -2.5 Z"
        fill="#3b82f6"
        opacity="0.35"
      />
    </g>

    <!-- 主体：购物袋 + 袋面旋涡徽标 -->
    <g>
      <ellipse cx="280" cy="456" rx="84" ry="12" fill="#0d74f2" opacity="0.07" />
      <!-- 宽而低的提手 + 上窄下宽微梯形袋身，避免方圆矩形 + 高把手读成挂锁 -->
      <path
        d="M244 282 V266 A36 36 0 0 1 316 266 V282"
        fill="none"
        stroke="url(#gw-mall-grad)"
        stroke-width="12"
        stroke-linecap="round"
      />
      <path
        d="M212 282 H348 L356 426 A10 10 0 0 1 346 436 H214 A10 10 0 0 1 204 426 Z"
        fill="url(#gw-mall-bag)"
        stroke="url(#gw-mall-grad)"
        stroke-width="13"
        stroke-linejoin="round"
      />
      <!-- 旋涡徽标缓旋：24s/圈、SMIL 以字形局部坐标 (256,256) 为轴（同品牌插画） -->
      <g transform="translate(224,300) scale(0.22)">
        <g>
          <animateTransform
            attributeName="transform"
            type="rotate"
            from="0 256 256"
            to="360 256 256"
            dur="24s"
            repeatCount="indefinite"
          />
          <path
            d="M241 431 A176 176 0 1 1 408 344"
            fill="none"
            stroke="url(#gw-mall-grad)"
            stroke-width="84"
            stroke-linecap="round"
          />
          <path
            d="M264 199 A58 58 0 1 1 204 282"
            fill="none"
            stroke="url(#gw-mall-grad)"
            stroke-width="54"
            stroke-linecap="round"
          />
        </g>
      </g>
    </g>
  </svg>
</template>

<style scoped>
.gw-ring1,
.gw-ring2,
.gw-orbit,
.gw-orbit2 {
  transform-box: view-box;
  transform-origin: 280px 268px;
}
.gw-ring1 {
  animation: gw-spin 48s linear infinite;
}
.gw-ring2 {
  animation: gw-spin 70s linear infinite reverse;
}
.gw-orbit {
  animation: gw-spin 36s linear infinite;
}
.gw-orbit2 {
  animation: gw-spin 52s linear infinite reverse;
}
.gw-streak {
  animation: gw-drift 9s ease-in-out infinite alternate;
}
.gw-s2 {
  animation-duration: 12s;
  animation-delay: -4s;
}
.gw-s3 {
  animation-duration: 10s;
  animation-delay: -2s;
}
.gw-s4 {
  animation-duration: 13s;
  animation-delay: -6s;
}
.gw-twinkle {
  animation: gw-blink 3.6s ease-in-out infinite;
}
.gw-t2 {
  animation-delay: -1.4s;
}
.gw-t3 {
  animation-delay: -2.3s;
}
@keyframes gw-spin {
  to {
    transform: rotate(360deg);
  }
}
@keyframes gw-drift {
  from {
    transform: translateX(-16px);
  }
  to {
    transform: translateX(16px);
  }
}
@keyframes gw-blink {
  0%,
  100% {
    opacity: 0.15;
  }
  50% {
    opacity: 0.6;
  }
}
@media (prefers-reduced-motion: reduce) {
  .gw-ring1,
  .gw-ring2,
  .gw-orbit,
  .gw-orbit2,
  .gw-streak,
  .gw-twinkle {
    animation: none;
  }
}
</style>
