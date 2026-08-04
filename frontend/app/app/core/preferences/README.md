# core/preferences 偏好设置模块

> 面向店铺前台业务逻辑开发者的使用指南

## 模块定位

提供 **主题、语言、页面切换动画** 等应用级别的偏好配置管理。

核心能力：响应式状态 + localStorage 持久化 + CSS 变量联动。

业务层通过 `updatePreferences()` 修改值，框架自动处理存储、CSS 更新、语言切换等副作用。

> 本模块的字段集与后台 admin 端不同：店铺前台不包含侧边栏、标签页、顶栏、面包屑、导航、快捷键等后台布局类偏好，也没有偏好设置面板。可配置项以下文「Preferences 完整字段速查」为准。

---

## 目录结构

```
core/preferences/
├── index.ts                      # 统一导出（preferences, updatePreferences, resetPreferences 等）
├── preferences.ts                # PreferenceManager 核心类（状态管理、持久化、副作用）
├── use-preferences.ts            # usePreferences() composable（计算属性快捷访问）
├── update-css-variables.ts       # CSS 变量联动（主题色、暗色模式、圆角等）
├── config/
│   ├── default.ts                # 所有偏好项的默认值
│   └── constants.ts              # 内置主题预设色板
└── types/
    ├── preferences-root.ts       # Preferences 总接口 + DeepPartial + InitialOptions
    ├── app.ts                    # 各子项类型（AppPreferences, ThemePreferences 等）
    ├── layout.ts                 # 枚举类型（SupportedLanguagesType, PageTransitionType）
    └── theme.ts                  # ThemeModeType / BuiltinThemeType
```

---

## 快速开始

### 读取偏好值

```ts
import { preferences } from "@/core/preferences";

// 直接读取（响应式 readonly 对象）
const isDark = preferences.theme.mode;          // "dark"
const locale = preferences.app.locale;          // "zh-CN"
```

### 修改偏好值

```ts
import { updatePreferences } from "@/core/preferences";

// 支持深层部分更新（无需传完整对象）
updatePreferences({
  theme: { mode: "dark" },
});
```

### 在组件中使用 composable

```vue
<script setup lang="ts">
import { usePreferences } from "@/core/preferences";

const { isDark, locale, setTheme, toggleTheme } = usePreferences();
</script>

<template>
  <button @click="toggleTheme">切换主题</button>
</template>
```

---

## 核心 API

### `updatePreferences(updates: DeepPartial<Preferences>)`

更新偏好设置。支持深层部分更新，只传需要修改的字段。

```ts
// 修改主题色
updatePreferences({ theme: { colorPrimary: "#ff6600" } });

// 同时修改多个子项
updatePreferences({
  app: { locale: "en-US" },
  theme: { mode: "dark" },
});
```

**自动触发的副作用：**
| 修改项 | 副作用 |
|---|---|
| `theme.*` | 更新 CSS 变量、切换 `dark` 类名 |
| `app.locale` | 加载对应语言包、同步 @nuxtjs/i18n |

### `initPreferences(options: InitialOptions)`

初始化偏好设置。在应用启动流程中调用一次，**不要在业务代码中调用**。

```ts
await initPreferences({
  namespace: "gowind",           // localStorage 键前缀
  overrides: {                   // 覆盖默认值
    app: { name: "My App" },
  },
});
```

### `resetPreferences()`

重置为初始值（默认值 + overrides 合并结果），同时清除 localStorage。

### `clearPreferencesCache()`

清除 localStorage 中的偏好缓存，不重置内存中的值。

---

## `usePreferences()` composable

提供一系列计算属性，方便在组件中快捷访问常用偏好值。

### 主题操作

| 返回值 | 签名 | 说明 |
|---|---|---|
| `setTheme` | `(mode: ThemeModeType) => void` | 设置主题模式 (`light`/`dark`/`auto`) |
| `toggleTheme` | `() => void` | 在 light 和 dark 之间切换 |
| `theme` | `ComputedRef<"dark" \| "light">` | 当前实际主题（已解析 auto） |
| `isDark` | `ComputedRef<boolean>` | 是否暗色模式 |
| `locale` | `ComputedRef<string>` | 当前语言 |

> 店铺前台没有侧边栏、标签页、顶栏、面包屑、导航等布局概念，因此 `usePreferences()` 不暴露 `layout` / `isMobile` / `isSideNav` / `isHeaderNav` 等布局判断属性。

---

## Preferences 完整字段速查

下面列出店铺前台 `Preferences` 全部子项与其字段。字段集以后端 `types/` 中的接口定义为准，本表与 `app/core/preferences/types/` 完全对应。

### `app` — 应用配置

| 字段 | 类型 | 说明 |
|---|---|---|
| `name` | `string` | 应用名 |
| `title` | `string` | 应用标题（用于浏览器标签） |
| `version` | `string` | 应用版本 |
| `locale` | `SupportedLanguagesType` | 当前语言（`zh-CN` / `en-US`） |
| `isMobile` | `boolean` | 是否移动端（自动检测，勿手动设置） |
| `defaultAvatar` | `string` | 应用默认头像 |
| `dynamicTitle` | `boolean` | 开启动态标题 |
| `defaultPageSize` | `number` | 默认每页条数 |
| `compact` | `boolean` | 是否开启紧凑模式 |

### `theme` — 主题配置

| 字段 | 类型 | 说明 |
|---|---|---|
| `mode` | `ThemeModeType` | 主题模式（`light` / `dark` / `auto`） |
| `builtinType` | `BuiltinThemeType` | 内置主题名 |
| `colorPrimary` | `string` | 主题色 |
| `colorSuccess` | `string` | 成功色 |
| `colorWarning` | `string` | 警告色 |
| `colorDestructive` | `string` | 危险色 |
| `radius` | `string` | 圆角 |

### `widget` — 功能部件

| 字段 | 类型 | 说明 |
|---|---|---|
| `themeToggle` | `boolean` | 是否显示主题切换部件 |
| `languageToggle` | `boolean` | 是否显示语言切换部件 |
| `globalSearch` | `boolean` | 是否显示全局搜索部件 |
| `backToTop` | `boolean` | 是否显示回到顶部部件 |

### `logo` / `copyright`

| 子项 | 字段 | 说明 |
|---|---|---|
| `logo` | `enable`, `source` | Logo 可见性与图片地址 |
| `copyright` | `enable`, `companyName`, `companySiteLink`, `date`, `icp`, `icpLink` | 版权与备案信息 |

### `transition` — 页面切换动画

| 字段 | 类型 | 说明 |
|---|---|---|
| `enable` | `boolean` | 页面切换动画是否启用 |
| `loading` | `boolean` | 是否开启页面加载 loading |
| `name` | `PageTransitionType \| string` | 页面切换动画（`fade` / `fade-down` / `fade-slide` / `fade-up`） |
| `progress` | `boolean` | 是否开启页面加载进度动画 |

---

## 内置主题色板

`BuiltinThemeType` 实际可选值（与 `core/preferences/types/theme.ts` 一致）：

| 主题名 | 含义 |
|---|---|
| `default` | 默认（蓝色） |
| `green` | 绿色 |
| `orange` | 橙色 |
| `pink` | 粉色 |
| `red` | 红色 |
| `sky-blue` | 天蓝色 |
| `violet` | 紫罗兰色 |
| `yellow` | 黄色 |
| `custom` | 自定义（由 `colorPrimary` 决定） |

> 注：店铺前台的内置主题集与后台 admin 不同，请勿按 admin 的主题列表填写。

---

## 数据流

```
用户操作 → updatePreferences({ theme: { mode: "dark" } })
    │
    ├─→ 合并到 reactive state
    │
    ├─→ handleUpdates() 副作用
    │     ├─ theme.* → updateCSSVariables() → 更新 CSS 变量 + html.dark
    │     └─ app.locale → loadLocaleMessages() → 加载语言包
    │
    └─→ savePreferences() (防抖 150ms)
          └─→ StorageManager.setItem() → localStorage
```

---

## 常见场景

### 场景 1：切换主题模式

```ts
import { updatePreferences } from "@/core/preferences";

updatePreferences({ theme: { mode: "dark" } });
```

### 场景 2：在组件中判断当前主题

```vue
<script setup lang="ts">
import { usePreferences } from "@/core/preferences";

const { isDark } = usePreferences();
</script>

<template>
  <div>当前为{{ isDark ? "暗色" : "亮色" }}模式</div>
</template>
```

### 场景 3：切换语言

```ts
import { updatePreferences } from "@/core/preferences";

updatePreferences({ app: { locale: "en-US" } });
```

### 场景 4：切换主题色

```ts
import { updatePreferences } from "@/core/preferences";

// 使用内置主题
updatePreferences({ theme: { builtinType: "violet", colorPrimary: "hsl(245 82% 67%)" } });

// 使用自定义颜色
updatePreferences({ theme: { builtinType: "custom", colorPrimary: "#ff6600" } });
```

### 场景 5：重置所有偏好

```ts
import { resetPreferences } from "@/core/preferences";

resetPreferences(); // 恢复为默认值 + overrides，清除 localStorage
```

---

## 注意事项

1. **不要直接修改 `preferences` 对象** — 它是 `readonly` 的，始终通过 `updatePreferences()` 修改
2. **`isMobile` 由框架自动管理** — 无需手动设置
3. **`initPreferences()` 只调用一次** — 已在启动流程中完成，业务代码不要重复调用
4. **颜色值支持多种格式** — `colorPrimary` 等字段接受 HSL、RGB、Hex 格式，内部会统一转换
5. **持久化有防抖** — 连续多次 `updatePreferences()` 不会频繁写入 localStorage（150ms 防抖）
6. **字段集以后端类型定义为准** — 本表字段与 `app/core/preferences/types/` 完全对应；后台 admin 端的布局类字段（sidebar/tabbar/header/breadcrumb/navigation/shortcutKeys）在本店铺前台不存在
