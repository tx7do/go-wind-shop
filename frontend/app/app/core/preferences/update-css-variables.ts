import type { Preferences } from "./types";

/**
 * 更新 CSS 变量 — 适配 shadcn-vue
 *
 * 暗色模式策略（单层）：
 * 1. JS 是暗色 token 的唯一来源：始终落到具体的 .dark / .light class
 *    - mode='auto'  → 按 matchMedia 解析成具体 class
 *    - mode='dark'  → 添加 .dark，移除 .light
 *    - mode='light' → 添加 .light，移除 .dark
 * 2. 只写 class，不写任何颜色/几何变量（token 的唯一来源是 assets/css/main.css）
 *
 * 注意：main.css 里与 .dark 等价的 @media 副本已删除。若这里退回"两个 class 都不加"，
 * 暗色偏好用户会拿到亮色主题——改这里时同步改 app/app.vue 里 useHead({ script }) 的
 * 首屏防闪脚本（app.html 是死文件，Nuxt 4 不加载它）。
 */
function updateCSSVariables(preferences: Preferences) {
  if (typeof document === "undefined") return;

  const root = document.documentElement;
  if (!root) return;

  const theme = preferences?.theme ?? {};
  const { mode } = theme;

  // 1. 暗色模式 — 通过 .dark / .light class 精确控制
  if (Reflect.has(theme, "mode")) {
    if (mode === "dark") {
      root.classList.add("dark");
      root.classList.remove("light");
    } else if (mode === "light") {
      root.classList.add("light");
      root.classList.remove("dark");
    } else {
      // auto：解析成具体 class，不留给 @media 兜底
      const dark = isDarkTheme(mode);
      root.classList.toggle("dark", dark);
      root.classList.toggle("light", !dark);
    }
  }

  // 2. 圆角不再由这里落到 <html> 的 style 上。
  //    theme.radius 在整个应用里没有任何设置入口（偏好面板不含圆角控件），
  //    但内联 style 的优先级高于 main.css，于是 preferences 一旦持久化成功，
  //    一份历史 "0.6rem" 就会永久压住设计系统的 --radius，改版也改不动。
  //    几何值与颜色同源：只在 assets/css/main.css 的 :root/.dark 里定义。
}

/**
 * 判断当前是否为暗色主题
 */
function isDarkTheme(theme: string | undefined): boolean {
  if (!theme) return false;
  if (theme === "dark") return true;
  if (theme === "auto") {
    if (typeof window === "undefined") return false;
    return window.matchMedia("(prefers-color-scheme: dark)").matches;
  }
  return false;
}

export { isDarkTheme, updateCSSVariables };
