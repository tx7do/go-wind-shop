<div align="center">

# GoWind Shop｜风行商城

**开箱即用的企业级出海电商全栈脚手架**

> 让出海电商开发如风般自由 — GoWind Shop

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![Vue](https://img.shields.io/badge/Vue-3.x-4FC08D?logo=vuedotjs)](https://vuejs.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://www.docker.com/)

**中文** | [English](./README.en-US.md) | [日本語](./README.ja-JP.md)

</div>

---

## 项目简介

GoWind Shop（风行商城）是一套面向出海场景的企业级多语言电商平台，采用 API 优先、前后端分离架构，提供多语言原生内容分发、多币种商品展示与单一结算币种能力。后端基于 [golang](https://go.dev/) + [go-kratos](https://go-kratos.dev/) 微服务框架，采用 Protobuf-first（契约驱动）工作流；前端后台基于 [Vue.js 3](https://vuejs.org/) + [TypeScript](https://www.typescriptlang.org/)，店铺前台基于 [Nuxt](https://nuxt.com/) + Vue 3。

## 项目亮点

- **多语言原生支持**：基于 i18n 的多语言内容分发，面向出海场景
- **契约驱动（Protobuf-first）**：接口、数据模型、错误码均以 proto 为单一事实源，自动生成 Go / TypeScript / OpenAPI
- **三服务 BFF 架构**：`admin` / `app` 瘦网关 + `core` 核心实现，业务逻辑与传输层解耦
- **全栈代码生成**：Ent Schema → ORM，Wire 依赖注入，一键 CRUD 脚手架
- **生产就绪**：JWT 鉴权、Casbin / OPA 授权、SSE 推送、Asynq 异步任务、OpenTelemetry 链路追踪、Swagger 文档、Docker 一键部署

---

## 系统架构

GoWind Shop 采用三服务 BFF 架构：`admin` 与 `app` 为瘦网关，仅做参数校验与 gRPC 转发；`core` 为核心实现，所有数据访问集中于此。

```
HTTP/REST 请求
  └─ admin-service  (后台网关 BFF，REST:6600 / SSE:6601 / gRPC)
       └─ gRPC → core-service  (核心业务 + 数据落点)

app-service (前台网关 BFF，REST:6700) ──gRPC──→ core-service
core-service (gRPC，业务实现与数据访问)
```

| 服务 | 角色 | 是否访问 DB |
|------|------|:---:|
| `app/admin/service` | 后台网关 BFF（接收 REST，转发 gRPC 到 core） | ❌ |
| `app/app/service` | 前台网关 BFF | ❌ |
| `app/core/service` | 核心业务实现（持久化、真实逻辑） | ✅ |

---

## 技术栈

<table>
<tr><th>层级</th><th>技术</th></tr>
<tr><td><strong>后端框架</strong></td><td><code>Golang</code> · <code>go-kratos v2</code> · <code>Wire</code> · <code>Protobuf / Buf</code></td></tr>
<tr><td><strong>ORM</strong></td><td><code>entgo.io/ent</code> · <code>PostgreSQL</code> · <code>MySQL</code></td></tr>
<tr><td><strong>中间件</strong></td><td><code>Redis</code> · <code>MinIO</code>（S3 兼容对象存储） · <code>ElasticSearch / OpenSearch</code> · <code>etcd</code>（服务发现）</td></tr>
<tr><td><strong>认证授权</strong></td><td><code>kratos-authn</code>（JWT HS256） · <code>kratos-authz</code>（Casbin / OPA）</td></tr>
<tr><td><strong>实时通信</strong></td><td><code>SSE</code>（服务端推送） · <code>Asynq</code>（异步任务）</td></tr>
<tr><td><strong>分布式</strong></td><td><code>OpenTelemetry + Jaeger</code>（链路追踪） · <code>DTM</code>（分布式事务）</td></tr>
<tr><td><strong>后台前端</strong></td><td><code>Vue 3</code> · <code>TypeScript</code> · <code>Vite</code> · <code>Element Plus</code></td></tr>
<tr><td><strong>店铺前端</strong></td><td><code>Nuxt</code> · <code>Vue 3</code> · <code>TypeScript</code> · <code>Tailwind CSS</code> · <code>i18n</code></td></tr>
<tr><td><strong>部署运维</strong></td><td><code>Docker</code> · <code>Docker Compose</code> · <code>PM2</code> · <code>Swagger UI</code></td></tr>
</table>

---

## 快速开始

### 环境要求

| 工具 | 版本 |
|------|------|
| Go | 1.25+ |
| Node.js | >= 20.10.0 |
| pnpm | >= 10.0.0 |
| Docker | 20.0+ |

### 环境脚本选型

- Linux / macOS 开发环境：`scripts/env/install_unix_dev.sh`
- Linux / macOS 生产环境：`scripts/env/install_unix_prod.sh`
- Windows 开发环境：`scripts/env/install_windows_dev.ps1`

### Docker 两种部署模式

- **full_deploy 完整模式**：同步启动中间件+后端应用，适用于一键演示、生产部署
- **libs_only 依赖模式（推荐开发）**：仅启动中间件，应用本地 IDE 运行调试

### 后端启动

**Linux / macOS：**

```shell
# 赋予脚本执行权限
chmod +x scripts/**/*.sh

# 开发环境（推荐）
./scripts/env/install_unix_dev.sh
./scripts/docker/libs_only.sh
cd backend/app/<admin|app|core>/service
make run

# 生产环境
./scripts/env/install_unix_prod.sh
./scripts/docker/full_deploy.sh

# PM2 进程托管（生产进阶）
./scripts/deploy/pm2_service.sh
```

**Windows（PowerShell 管理员）：**

```powershell
# 放行脚本策略（首次仅需执行一次）
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser

# 初始化环境
.\scripts\env\install_windows_dev.ps1

# 本地开发
.\scripts\docker\libs_only.ps1
cd backend/app/<admin|app|core>/service
make run

# 一键完整部署
.\scripts\docker\full_deploy.ps1
```

### 前端启动

前端项目统一存放于 `frontend` 目录，依赖安装命令一致：

| 前端 | 目录 | 启动命令 |
|------|------|---------|
| 后台管理 | `frontend/admin` | `pnpm dev` |
| 店铺前台 | `frontend/app` | `pnpm dev` |

```shell
# 安装依赖
pnpm install

# 后台管理（Vue3 + Element Plus）
cd frontend/admin && pnpm dev

# 店铺前台（Nuxt + Vue 3）
cd frontend/app && pnpm dev
```

> 具体监听端口见各前端目录下的 `.env.development`。

---

## 功能列表

### 商品与交易

| 功能 | 说明 |
|------|-----|
| 商品管理 | 商品与 SKU 的发布、上下架与维护 |
| 商品属性 | 商品属性与属性值管理 |
| 类目管理 | 商品类目管理，支持树形结构 |
| 品牌管理 | 品牌信息管理 |
| SKU 与定价 | SKU、属性组合与价格管理 |
| 购物车 | 购物车与购物车项管理 |
| 订单管理 | 订单与订单明细管理 |
| 支付与退款 | 支付交易与退款记录管理（注：当前支付网关为模拟实现，真实支付网关待接入） |
| 物流发货 | 订单履约与物流发货管理，含发货状态机与订单联动 |
| 优惠券系统 | 优惠券模板与用户优惠券管理，含定价折扣引擎与下单核销/退款返还 |
| 文件传输 | 商品图片与资源上传传输 |

### 平台与系统

| 功能 | 说明 |
|------|-----|
| 用户档案 | 用户档案信息管理 |
| 租户管理 | 多租户管理与初始化 |
| 角色管理 | 角色与角色分组管理 |
| 权限管理 | 权限分组与权限点管理 |
| 菜单管理 | 系统菜单与按钮权限配置 |
| 组织管理 | 组织、部门、职位管理 |
| 接口管理 | 接口注册与同步 |
| 字典与多语言 | 数据字典与多语言内容管理 |
| 任务调度 | 异步任务与调度日志管理 |
| 文件管理 | 文件上传与对象存储管理 |
| 内部消息 | 站内消息与消息分类管理 |
| 审计日志 | 登录、操作、权限与数据访问审计日志 |
| 找回密码 | 邮箱验证码找回密码流程 |
| 登录策略 | 登录策略配置管理 |

---

## 项目结构

```
go-wind-shop/
├── backend/                        # 后端项目
│   ├── api/                        # Protobuf API 定义与生成代码
│   │   ├── protos/                 # .proto 源文件（按领域分层）
│   │   └── gen/go/                 # buf 生成的 Go 代码
│   ├── app/                        # 应用实现层（admin / app / core 三个微服务）
│   │   └── <admin|app|core>/service/
│   │       ├── cmd/server/         # 入口 + wire 注入
│   │       ├── configs/            # 运行时配置（YAML）
│   │       └── internal/           # 业务核心（server / service / data）
│   ├── pkg/                        # 跨服务公共库
│   ├── scripts/                    # 部署与环境脚本（env / docker / deploy）
│   ├── sql/                        # 数据库种子 / 演示数据脚本
│   └── Makefile / app.mk           # 构建编排
├── frontend/                       # 前端项目
│   ├── admin/                      # Vue 3 + Element Plus 后台
│   └── app/                        # Nuxt + Vue 3 店铺前台
└── README.md
```

---

## 联系我们

- 微信个人号：`yang_lin_bo`（备注：`go-wind-shop`）
