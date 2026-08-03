<template>
  <el-icon :size="20" class="theme-switch" @click="handleToggle">
    <component :is="isDark ? Sunny : Moon" />
  </el-icon>
</template>

<script setup lang="ts">
import { Moon, Sunny } from "@element-plus/icons-vue";
import { usePreferences } from "@/core/preferences";

const { isDark, toggleTheme } = usePreferences();

function handleToggle(event: MouseEvent) {
  if (!(document as any).startViewTransition) {
    toggleTheme();
    return;
  }

  const goingDark = !isDark.value;

  // 按钮点击坐标（切亮色时的圆心）
  const btnX = event.clientX;
  const btnY = event.clientY;

  // 左下角坐标（切暗色时的圆心）
  const cornerX = 0;
  const cornerY = window.innerHeight;

  // 计算从左下角到按钮的最大半径（切暗色用）
  const darkMaxRadius = Math.hypot(
    Math.max(btnX, window.innerWidth - cornerX),
    Math.max(cornerY - btnY, cornerY),
  );

  // 计算从按钮覆盖全屏的最大半径（切亮色用）
  const lightMaxRadius = Math.hypot(
    Math.max(btnX, window.innerWidth - btnX),
    Math.max(btnY, window.innerHeight - btnY),
  );

  // 动态注入 z-index 样式
  const style = document.createElement("style");
  style.dataset.themeTransition = "";
  document.head.appendChild(style);

  const transition = (document as any).startViewTransition(async () => {
    toggleTheme();
  });

  transition.ready.then(() => {
    if (goingDark) {
      // 切暗色：黑暗从左下角扩散，直到吞没按钮
      style.textContent =
        "::view-transition-old(root){z-index:-1}::view-transition-new(root){z-index:auto}";

      document.documentElement.animate(
        {
          clipPath: [
            `circle(0px at ${cornerX}px ${cornerY}px)`,
            `circle(${darkMaxRadius}px at ${cornerX}px ${cornerY}px)`,
          ],
        },
        {
          duration: 600,
          easing: "ease-in-out",
          pseudoElement: "::view-transition-new(root)",
        },
      );
    } else {
      // 切亮色：白色光明从按钮释放，扩散到整个页面
      style.textContent =
        "::view-transition-old(root){z-index:-1}::view-transition-new(root){z-index:auto}";

      document.documentElement.animate(
        {
          clipPath: [
            `circle(0px at ${btnX}px ${btnY}px)`,
            `circle(${lightMaxRadius}px at ${btnX}px ${btnY}px)`,
          ],
        },
        {
          duration: 600,
          easing: "ease-in-out",
          pseudoElement: "::view-transition-new(root)",
        },
      );
    }
  });

  transition.finished.then(() => {
    style.remove();
  });
}
</script>

<style>
/* 禁用默认的 view-transition 淡入淡出，让圆形动画独占 */
::view-transition-old(root),
::view-transition-new(root) {
  animation: none !important;
  mix-blend-mode: normal;
}
</style>
