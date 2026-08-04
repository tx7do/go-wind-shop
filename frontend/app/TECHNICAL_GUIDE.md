# 基于 Nuxt 4 的电商店铺前台：架构深度解析与二次开发指南

> 本文面向希望基于 GoWind Shop 店铺前台进行二次开发的前端工程师，系统性地讲解项目的技术选型、架构设计与模块划分，并提供扩展开发的实操指引。

---

## 一、项目概览

本项目是 GoWind Shop（风行商城）的店铺前台，使用 Nuxt 4（Vue 3）构建，支持 SSR/SSG 双模式部署，提供商品浏览、类目导航、购物车、下单结算、订单与支付等完整的电商购物链路，并内置多语言（中英文）和暗色模式支持。

### 核心特性一览

| 特性 | 技术方案 |
|------|----------|
| 框架 | Nuxt 4（Vue 3.5+） |
| 样式 | Tailwind CSS v4 + CSS 变量主题系统 |
| UI 组件库 | shadcn-vue（基于 Reka UI） |
| 状态管理 | Pinia + 持久化插件 |
| 数据请求 | Axios + TanStack Vue Query |
| 国际化 | @nuxtjs/i18n（prefix 路由策略） |
| API 协议 | Protobuf 生成 TypeScript HTTP 客户端 |
| 部署 | SSG 静态生成 + SPA fallback |

---

## 二、技术栈详解

### 2.1 Nuxt 4 + Vue 3

项目基于 Nuxt 4，采用文件系统路由（`app/pages/`），组件自动扫描注册，并通过 `compatibilityDate` 锁定行为一致性。SSR 模式开启，Nitro 引擎负责服务端渲染与静态生成。

```ts
// nuxt.config.ts 核心配置
export default defineNuxtConfig({
  ssr: true,
  nitro: {
    prerender: {
      routes: ['/'],
      crawlLinks: true,
    },
  },
})
```

### 2.2 Tailwind CSS v4 + 语义化主题

项目使用 **Tailwind CSS v4**（通过 `@tailwindcss/vite` 插件集成），并采用 CSS 变量构建语义化配色系统：

```css
/* main.css 主题变量定义 */
:root {
  --primary: 142.1 76.2% 36.3%;
  --background: 210 40% 98%;
  --foreground: 224 71.4% 4.1%;
  /* ...更多语义 token */
}

.dark {
  --primary: 142.1 86.2% 50.3%;
  --background = 224 45% 6%;
  /* ... */
}
```

暗色模式通过 `.dark` class 精确切换，并设置了 `@media (prefers-color-scheme: dark)` 作为纯 CSS 后备方案，防止 JS 未执行时出现 FOUC（闪烁）问题。

**关键设计决策**：禁用了 Tailwind 的 `dark:` 变体写法，统一使用语义化 CSS 变量（如 `hsl(var(--foreground))`）控制样式，确保主题切换的一致性。

### 2.3 shadcn-vue 组件体系

UI 组件基于 [shadcn-vue](https://www.shadcn-vue.com/) 构建，底层依赖 Reka UI（无头组件库）。项目在 `app/components/ui/` 下维护了基础组件集合（button / input / select / dropdown-menu / pagination / skeleton / sonner / carousel / avatar / switch / image 等）。这些组件可以通过 `npx shadcn-vue@latest add <component>` 按需添加，遵循 shadcn 的 "拥有代码" 哲学——组件代码直接存在于你的项目中，可完全自定义。

> 店铺前台的 `app/components/` 仅包含三个子目录：`auth`（登录/注册等认证组件）、`layout`（页面布局组件）、`ui`（shadcn-vue 基础组件）。不存在文章、评论、内容渲染等 CMS 专属组件。

---

## 三、架构设计

### 3.1 整体分层架构

```
┌───────────────────────────────────────────────────┐
│                   Pages / Views                    │  页面层
├───────────────────────────────────────────────────┤
│              Components (UI / Business)            │  组件层
├────────────┬────────────┬─────────────────────────┤
│  Stores    │ Preferences│  Composables (Hooks)     │  状态/逻辑层
├────────────┴────────────┴─────────────────────────┤
│                 API Layer (两层架构 + ApiClient)     │  数据层
├───────────────────────────────────────────────────┤
│            Core (Transport / Storage)              │  基础设施层
└───────────────────────────────────────────────────┘
```

### 3.2 API 两层架构 + 统一 ApiClient

API 模块是本项目最核心的架构设计之一，采用 **Generated → Composables** 两层架构，并通过 `apiClient` 单例统一管理所有 Service Client：

```
api/
├── client.ts            ← ApiClient 单例（统一入口，懒加载各 Service Client）
├── generated/           ← 第一层：Protobuf 自动生成的 TypeScript 客户端
│   └── index.ts             # 类型定义 + Service Client 工厂 + ApiClient 类
└── composables/         ← 第二层：业务逻辑封装 + Vue Query 集成
    ├── product.ts            # 纯函数(listProduct...) + Hook(useListProduct...) + fetch(fetchListProduct...)
    └── ...
```

**第一层：Generated（协议层）**

由 `protoc-gen-typescript-http` 自动生成，定义了与后端 API 的类型契约。除了每个服务的 `createXxxServiceClient` 工厂方法外，还提供了统一的 `ApiClient` 类，内置懒加载缓存：

```ts
// generated/index.ts（自动生成，不要修改）
export class ApiClient {
  private _productService?: ProductService;
  get productService(): ProductService {
    return this._productService ??= createProductServiceClient(this._transport);
  }
  // ... 其他服务同理
}
export function createApiClient(transport: ClientTransport): ApiClient;
```

`ApiClient` 实际暴露的全部 Service Client getter 见 [`app/api/README.md` 的「支持的服务」](./app/api/README.md#支持的服务)。覆盖认证、用户资料、商品/属性/属性值、SKU/SKU 价格/SKU 属性组合、类目、品牌、购物车/购物车项、订单/订单明细、支付交易/退款、文件传输等店铺业务域。

**ApiClient 单例**

`client.ts` 将已有的 `requestApi`（基于 Axios）适配为 `ClientTransport` 接口，创建全局唯一的 `apiClient` 实例：

```ts
// client.ts
import { createApiClient, type ClientTransport } from '@/api/generated/app/service/v1';
import { requestApi } from '@/core/transport/rest';

const transport: ClientTransport = {
  unary(path, method, body) {
    return requestApi({ path, method, body });
  },
};
export const apiClient = createApiClient(transport);
```

**第二层：Composables（业务逻辑 + Vue Query 集成）**

每个 Composables 文件同时包含三层内容：业务纯函数、Vue Query Hooks、fetch 方法。业务逻辑直接调用 `apiClient`：

```ts
// composables/product.ts
import { apiClient } from '@/api/client';

// 业务纯函数（组装参数，调用 apiClient）
export async function listProduct(paging?, formValues?, fieldMask?, orderBy?, options?) {
  return await apiClient.productService.List({
    fieldMask,
    orderBy: makeOrderBy(orderBy),
    query: makeQueryString(merged, options?.isTenantUser),
    page: paging?.page,
    pageSize: paging?.pageSize,
    noPaging,
  });
}

// Vue Query Hook（用于组件）
export function useListProduct(options?) {
  return useMutation({
    mutationFn: (params) => {
      const locale = getCurrentLocale();  // 自动注入当前语言
      return listProduct(params.paging, params.formValues, ...);
    },
    ...options,
  });
}

// fetch 方法（供 Store / 非组件上下文使用）
export async function fetchListProduct(params: ListProductParams) {
  return queryClient.fetchQuery({
    queryKey: ['listProduct', params, locale],
    queryFn: () => listProduct(...),
  });
}
```

这种架构设计的优势：
- **简洁性**：消除了独立的 Service 层，减少文件数量和样板代码
- **统一管理**：`apiClient` 单例统一管理所有 Service Client 的创建和缓存
- **灵活调用**：组件中使用 `useXxx` Hook，Store 中使用 `fetchXxx` 方法，无需缓存时直接调用纯函数
- **语言感知**：Composable 层自动注入当前 locale，组件无需关心

> 完整的命名规范、类型规范、错误处理规范与最佳实践见 [`app/api/README.md`](./app/api/README.md)。

### 3.3 请求客户端（RequestClient）

`core/transport/rest/` 实现了一个功能完备的 HTTP 客户端：

```
transport/
├── rest/
│   ├── request-client.ts      # Axios 单例封装
│   ├── request-api.ts         # Protobuf 适配器
│   ├── preset-interceptors.ts # 预置拦截器（401、错误消息）
│   ├── pagination.ts          # 分页参数处理
│   ├── utils.ts               # 查询字符串、排序等工具
│   └── modules/
│       ├── interceptor.ts     # 拦截器管理器
│       ├── uploader.ts        # 文件上传
│       └── downloader.ts      # 文件下载
└── sse/
    └── sse_client.ts          # Server-Sent Events 客户端
```

**内置拦截器链（按执行顺序）**：

1. **Token 注入**：自动添加 `Authorization: Bearer xxx`
2. **Request ID**：注入 `X-Request-ID` 和 `XMLHttpRequest` 标识
3. **Locale 注入**：自动添加 `Accept-Language` 请求头
4. **Auth 拦截器**：401 时自动刷新 Token 或跳转登录页
5. **响应解构**：提取 `response.data`，简化调用方处理
6. **错误消息**：统一提取错误文本并通过回调通知

**初始化时机**：在 Nuxt 插件 `01.init-client.ts` 中完成初始化，注入 Token 获取、语言获取、Token 刷新、重新认证等回调：

```ts
// plugins/01.init-client.ts
export default defineNuxtPlugin(async (nuxtApp) => {
  if (import.meta.server) return;

  RequestClient.init(config.apiURL, {
    getToken: () => accessStore.accessToken?.value ?? null,
    getLocale: () => nuxtApp.$i18n?.locale?.value || 'zh-CN',
    refreshToken: async () => { /* ... */ },
    onReAuthenticate: async () => { /* ... */ },
  });
});
```

### 3.4 状态管理

项目使用 Pinia 进行状态管理，按职责分为 `core` 和 `app` 两个模块：

```
stores/modules/
├── core/
│   ├── access.state.ts   # 访问控制（Token、登录状态）
│   ├── user.state.ts     # 用户信息
│   ├── navbar.state.ts   # 导航栏状态
│   └── loading.state.ts  # 全局加载状态
└── app/
    └── auth.state.ts     # 认证逻辑（登录、登出、Token 刷新）
```

**认证流程**（`auth.state.ts`）：

```
登录 → API 登录 → 存储 AccessToken → 获取用户信息 → 跳转首页
                                                     ↓
Token 过期 ← 401 拦截 → 自动刷新 Token → 失败则跳转登录页
```

密码在传输前通过 AES-CBC 加密（`CryptoJS`），密钥通过运行时配置注入。

### 3.5 偏好设置系统（Preferences）

`core/preferences/` 实现了一套用户偏好管理：

```
preferences/
├── preferences.ts           # PreferenceManager 单例
├── use-preferences.ts       # Vue Composable
├── update-css-variables.ts  # CSS 变量同步
├── config/
│   └── default.ts           # 默认偏好值
└── types/                   # 类型定义
```

**核心设计**：
- **PreferenceManager** 是响应式单例，通过 `reactive` 管理状态，`readonly` 对外暴露
- 偏好变更自动同步到 CSS 变量（控制主题色、圆角等）和 localStorage
- 语言切换自动同步到 `@nuxtjs/i18n`
- 支持主题模式：`light` / `dark` / `auto`（跟随系统）
- 使用 `useDebounceFn` 防抖写入，避免频繁操作存储

```ts
// 组件中使用
const { isDark, theme, locale, toggleTheme, setLocale } = usePreferences();
```

> **店铺前台的偏好字段集与后台 admin 不同**：本端只包含 `app`（应用配置）/`theme`（主题）/`widget`（功能部件）/`logo`/`copyright`/`transition`（页面切换动画）六类，不包含侧边栏、标签页、顶栏、面包屑、导航、快捷键等后台布局类偏好，也没有偏好设置面板。完整字段见 [`app/core/preferences/README.md`](./app/core/preferences/README.md)。

### 3.6 存储管理器（StorageManager）

`core/storage/` 提供了增强版的浏览器存储抽象：

- **TTL 过期**：每个存储项支持独立的过期时间
- **驱逐策略**：支持 LRU / LFU / Hybrid 三种驱逐算法
- **批量操作**：`getItems` / `setItems` / `removeItems`，自动让出主线程避免阻塞
- **跨标签页同步**：基于 BroadcastChannel API
- **监控埋点**：命中率、过期次数、错误数等指标
- **配额管理**：自动清理过期项，处理 `QuotaExceededError`

### 3.7 国际化方案

使用 `@nuxtjs/i18n` 模块，采用 **prefix 路由策略**（所有路由带语言前缀）：

```ts
i18n: {
  locales: [
    { code: 'zh-CN', file: 'zh-CN/index.ts' },
    { code: 'en-US', file: 'en-US/index.ts' },
  ],
  defaultLocale: 'zh-CN',
  strategy: 'prefix',       // /zh-CN/、/en-US/
}
```

翻译文件按模块拆分在 `locales/` 目录下。

**语言感知 API 调用**：通过 `getCurrentLocale()` 工具函数统一获取当前语言，确保 API 请求的语言参数与 UI 语言同步：

```ts
// utils/locale.ts
export function getCurrentLocale(): SupportedLanguagesType {
  const nuxtApp = useNuxtApp();
  const locale = nuxtApp.$i18n?.locale?.value;
  // zh → zh-CN, en → en-US
  return map[locale] || 'zh-CN';
}
```

---

## 四、核心功能模块

### 4.1 商品浏览

店铺前台的购物链路由以下页面承载：

```
pages/
├── index.vue              # 首页
├── login.vue              # 登录页
├── register.vue           # 注册页
├── settings.vue           # 设置页
├── user.vue               # 用户中心
├── product/
│   └── [id].vue           # 商品详情
├── category/
│   ├── index.vue          # 类目列表
│   └── [id].vue           # 类目详情
├── cart.vue               # 购物车
├── checkout.vue           # 结算
└── orders/
    ├── index.vue          # 订单列表
    └── [id].vue           # 订单详情
```

商品与类目数据通过 `productService` / `categoryService` / `brandService` / `skuService` / `skuPriceService` / `skuAttributeCombinationService` / `productAttributeService` / `productAttributeValueService` 等 Composable 获取，列表/详情接口在 app BFF 的公开读白名单内，无需登录即可浏览。

### 4.2 交易链路

购物车 → 结算 → 订单 → 支付构成完整交易闭环：

- **购物车**：`cartService` / `cartItemService` 管理购物车与购物车项
- **下单结算**：`checkout.vue` 调用 `orderService` 创建订单
- **订单管理**：`orderService` / `orderItemService` 查询与管理订单
- **支付**：`paymentTransactionService` / `paymentRefundService` 处理支付交易与退款

> 交易域（cart / order / payment）接口默认强制 JWT，不在 app BFF 的公开读白名单内，需登录后访问。

首页使用了 `IntersectionObserver` + `MutationObserver` 实现滚动渐显动画，性能开销极低。

---

## 五、环境配置与构建

### 5.1 多环境配置

通过 `.env` 文件管理不同环境的配置：

```bash
# .env.development
NUXT_PUBLIC_API_BASE_URL=http://localhost:6700
NUXT_PUBLIC_ENABLE_MOCK=false
NUXT_PUBLIC_AES_KEY=

# .env.production
NUXT_PUBLIC_API_BASE_URL=https://api.example.com
NUXT_PUBLIC_ENABLE_MOCK=false
NUXT_PUBLIC_AES_KEY=your-secret-key
```

> `NUXT_PUBLIC_API_BASE_URL` 在开发环境指向 app-service（前台 BFF，REST:6700）。

### 5.2 构建命令

```bash
pnpm dev          # 开发服务器（加载 .env.development）
pnpm build        # 服务端构建（加载 .env.production）
pnpm generate     # 静态站点生成（SSG）
pnpm preview      # 预览构建结果
```

### 5.3 SSG 部署策略

项目支持静态站点生成（SSG）部署：

1. **预渲染**：Nitro 引擎预渲染 `/` 首页，并自动爬取内部链接
2. **动态路由**：`/product/:id`、`/category/:id` 等动态路由由客户端渲染（SPA fallback）
3. **根路径重定向**：构建后自动生成 `index.html`，将根路径 `/` 重定向到 `/zh-CN/`

```ts
// nuxt.config.ts — 构建后钩子
nitro: {
  hooks: {
    'prerender:done'() {
      writeFileSync(resolve(outputDir, 'index.html'), `
        <meta http-equiv="refresh" content="0;url=/zh-CN/">
        <script>location.replace("/zh-CN/"+location.search+location.hash)</script>
      `);
    },
  },
}
```

配合 Nginx 的 `try_files $uri $uri/ /index.html` 即可完成 SPA fallback。

---

## 六、二次开发指南

### 6.1 新增业务模块

以添加一个「商品评价」模块为例（假设后端已提供 `productReview` 服务）：

**第一步：确认 API 类型定义**

在 `api/generated/` 中确保后端已生成对应的 TypeScript 客户端（如 `createProductReviewServiceClient`，并出现在 `ApiClient` 的 getter 中）。

**第二步：编写 Composables 层**

在 `api/composables/product-review.ts` 中同时编写业务纯函数、Vue Query Hooks 和 fetch 方法，直接使用 `apiClient`。命名遵循 `listXxx` / `useListXxx` / `fetchListXxx` 的三态约定（详见 [`app/api/README.md`](./app/api/README.md)）。

**第三步：注册导出**

在 `api/composables/index.ts` 中添加：
```ts
export * from './product-review';
```

**第四步：创建页面**

在 `pages/` 下创建路由页面，使用 Composable 获取数据。

### 6.2 新增 UI 组件

项目使用 shadcn-vue，可通过 CLI 添加标准组件：

```bash
npx shadcn-vue@latest add dialog
```

自定义业务组件放在 `components/` 对应的目录中，Nuxt 会自动扫描注册。**注意**：组件只扫描 `.vue` 文件，`index.ts` 文件会被忽略，避免命名冲突。

### 6.3 新增多语言支持

1. 在 `locales/` 下创建新的语言目录（如 `ja-JP/`）
2. 复制现有语言文件并翻译
3. 在 `nuxt.config.ts` 的 `i18n.locales` 中注册新语言
4. 在 `core/preferences/types/layout.ts` 的 `SupportedLanguagesType` 中添加新语言类型

### 6.4 自定义主题

主题通过 CSS 变量控制，修改 `assets/css/main.css` 中的变量即可：

```css
:root {
  --primary: 210 100% 50%;       /* 改为蓝色主题 */
  --accent: 280 80% 60%;         /* 改为紫色点缀 */
}
```

如需切换内置主题，可通过偏好系统设置 `theme.builtinType`（可选值见 [`app/core/preferences/README.md`](./app/core/preferences/README.md) 的内置主题色板表）。

### 6.5 扩展请求拦截器

在需要自定义请求/响应处理时，可通过 `RequestClient` 的拦截器 API：

```ts
const client = RequestClient.getInstance();

// 添加请求拦截器
client.addRequestInterceptor({
  fulfilled: (config) => {
    config.headers['X-Custom-Header'] = 'value';
    return config as never;
  },
});

// 添加响应拦截器
client.addResponseInterceptor({
  fulfilled: (response) => {
    // 自定义响应处理
    return response;
  },
});
```

### 6.6 关键注意事项

| 场景 | 注意事项 |
|------|----------|
| Composable 函数命名 | 需精确匹配单复数（如 `useListProducts` vs `useListProduct`） |
| 组件中使用 i18n | 需使用 `useI18n()` composable，不要直接引用 `$t` |
| 路由导航 | 使用 `navigateTo()` 而非 `useRouter().push()` |
| 语言切换 | 通过 `switchLocalePath()` 并传入配置的 locale 联合类型 |
| Store 中导航 | 使用 `navigateTo()` 替代 `useRouter()`（Pinia 中无法使用路由 composables） |
| 暗色模式样式 | 使用 CSS 变量 `hsl(var(--foreground))` 而非 Tailwind `dark:` 变体 |
| 图片占位 | 使用 `<UiImage>` 公共组件，自动处理加载失败 |
| Dev server 缓存 | 新增组件后需重启 dev server 刷新自动导入缓存 |
| JSON 翻译文件 | 禁止使用裸露花括号 `{}`，需使用 `{'@'}` 转义特殊字符 |

---

## 七、项目结构速查

```
app/
├── api/                    # API 两层架构 + ApiClient
│   ├── client.ts           #   ApiClient 单例
│   ├── generated/          #   Protobuf 生成代码
│   └── composables/        #   业务逻辑 + Vue Query Hooks（auth/user-profile/file-transfer/product/category/brand/cart/cart-item/order/payment-transaction/sku/sku-price/...）
├── assets/css/             # 全局样式 + 主题变量
├── components/             # Vue 组件
│   ├── auth/               #   认证组件（登录/注册）
│   ├── layout/             #   布局组件（Header/Footer/Nav）
│   └── ui/                 #   基础 UI 组件（shadcn-vue）
├── constants/              # 常量定义
├── core/                   # 核心基础设施
│   ├── preferences/        #   偏好设置系统
│   ├── storage/            #   存储管理器
│   └── transport/          #   HTTP / SSE 传输层
├── hooks/                  # 通用 Composables
├── layouts/                # 页面布局
├── pages/                  # 路由页面
├── plugins/                # Nuxt 插件
├── stores/                 # Pinia 状态管理
├── utils/                  # 工具函数
└── typings/                # 全局类型定义

locales/                    # 国际化翻译文件
├── zh-CN/
└── en-US/
```

---

## 八、总结

本项目采用清晰的分层架构，将 **协议层 → 组合函数层** 通过 `apiClient` 单例解耦，使每一层都可以独立演进和替换。核心基础设施（Transport、Storage、Preferences）高度模块化，可方便地扩展或替换实现。

对于二次开发者而言，最常见的扩展模式是：**新增 Composable → 创建页面组件**。每个 Composable 文件集中了业务纯函数、Vue Query Hooks 和 fetch 方法，大部分情况下无需触碰基础设施代码。

项目在暗色模式、国际化、SSR/SSG 兼容性等方面积累了大量工程实践细节，建议开发者在深入开发前完整阅读 `core/` 目录下的实现代码，理解整体设计意图后再进行扩展。
