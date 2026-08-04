<div align="center">

# GoWind Shop

**Out-of-the-box enterprise-grade cross-border e-commerce full-stack scaffold**

> Make cross-border e-commerce development as free as the wind — GoWind Shop

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![Vue](https://img.shields.io/badge/Vue-3.x-4FC08D?logo=vuedotjs)](https://vuejs.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://www.docker.com/)

**English** | [中文](./README.md) | [日本語](./README.ja-JP.md)

</div>

---

## Introduction

GoWind Shop is an enterprise-grade multi-language e-commerce platform built for cross-border scenarios. It adopts an API-first, front-end/back-end separated architecture, providing native multi-language content distribution, multi-currency product display, and a single settlement currency. The backend is built on [golang](https://go.dev/) + the [go-kratos](https://go-kratos.dev/) microservice framework with a Protobuf-first (contract-driven) workflow; the admin frontend is built with [Vue.js 3](https://vuejs.org/) + [TypeScript](https://www.typescriptlang.org/), and the storefront with [Nuxt](https://nuxt.com/) + Vue 3.

## Highlights

- **Multi-language native support**: i18n-based multi-language content distribution, designed for cross-border scenarios.
- **Contract-driven (Protobuf-first)**: interfaces, data models, and error codes use proto as the single source of truth, auto-generating Go / TypeScript / OpenAPI.
- **Three-service BFF architecture**: `admin` / `app` thin gateways plus a `core` implementation, decoupling business logic from the transport layer.
- **Full-stack code generation**: Ent schema → ORM, Wire dependency injection, one-click CRUD scaffolding.
- **Production-ready**: JWT authentication, Casbin / OPA authorization, SSE push, Asynq async tasks, OpenTelemetry tracing, Swagger docs, one-click Docker deployment.

---

## System Architecture

GoWind Shop uses a three-service BFF architecture: `admin` and `app` are thin gateways that only perform parameter validation and gRPC forwarding; `core` is the core implementation, where all data access is centralized.

```
HTTP/REST request
  └─ admin-service  (admin gateway BFF, REST:6600 / SSE:6601 / gRPC)
       └─ gRPC → core-service  (core business + data persistence)

app-service (storefront gateway BFF, REST:6700) ──gRPC──→ core-service
core-service (gRPC, business implementation and data access)
```

| Service | Role | DB Access |
|------|------|:---:|
| `app/admin/service` | Admin gateway BFF (receives REST, forwards gRPC to core) | ❌ |
| `app/app/service` | Storefront gateway BFF | ❌ |
| `app/core/service` | Core business implementation (persistence, real logic) | ✅ |

---

## Tech Stack

<table>
<tr><th>Layer</th><th>Technology</th></tr>
<tr><td><strong>Backend framework</strong></td><td><code>Golang</code> · <code>go-kratos v2</code> · <code>Wire</code> · <code>Protobuf / Buf</code></td></tr>
<tr><td><strong>ORM</strong></td><td><code>entgo.io/ent</code> · <code>PostgreSQL</code> · <code>MySQL</code></td></tr>
<tr><td><strong>Middleware</strong></td><td><code>Redis</code> · <code>MinIO</code> (S3-compatible object storage) · <code>ElasticSearch / OpenSearch</code> · <code>etcd</code> (service discovery)</td></tr>
<tr><td><strong>Authentication & Authorization</strong></td><td><code>kratos-authn</code> (JWT HS256) · <code>kratos-authz</code> (Casbin / OPA)</td></tr>
<tr><td><strong>Real-time communication</strong></td><td><code>SSE</code> (server push) · <code>Asynq</code> (async tasks)</td></tr>
<tr><td><strong>Distributed</strong></td><td><code>OpenTelemetry + Jaeger</code> (tracing) · <code>DTM</code> (distributed transactions)</td></tr>
<tr><td><strong>Admin frontend</strong></td><td><code>Vue 3</code> · <code>TypeScript</code> · <code>Vite</code> · <code>Element Plus</code></td></tr>
<tr><td><strong>Storefront frontend</strong></td><td><code>Nuxt</code> · <code>Vue 3</code> · <code>TypeScript</code> · <code>Tailwind CSS</code> · <code>i18n</code></td></tr>
<tr><td><strong>Deployment & Operations</strong></td><td><code>Docker</code> · <code>Docker Compose</code> · <code>PM2</code> · <code>Swagger UI</code></td></tr>
</table>

---

## Quick Start

### Environment Requirements

| Tool | Version |
|------|------|
| Go | 1.25+ |
| Node.js | >= 20.10.0 |
| pnpm | >= 10.0.0 |
| Docker | 20.0+ |

### Environment Script Selection

- Linux / macOS Development: `scripts/env/install_unix_dev.sh`
- Linux / macOS Production: `scripts/env/install_unix_prod.sh`
- Windows Development: `scripts/env/install_windows_dev.ps1`

### Docker Deployment Modes

- **full_deploy (full mode)**: Starts middleware + backend application, suitable for one-click demo or production deployment.
- **libs_only (Recommended)**: Starts middleware only; run the application locally in your IDE for daily development.

### Backend Startup

**Linux / macOS:**

```shell
# Grant script execution permissions
chmod +x scripts/**/*.sh

# Development (Recommended)
./scripts/env/install_unix_dev.sh
./scripts/docker/libs_only.sh
cd backend/app/<admin|app|core>/service
make run

# Production
./scripts/env/install_unix_prod.sh
./scripts/docker/full_deploy.sh

# PM2 Process Management (Advanced Production)
./scripts/deploy/pm2_service.sh
```

**Windows (PowerShell Administrator):**

```powershell
# Allow script execution (only needed once)
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser

# Initialize environment
.\scripts\env\install_windows_dev.ps1

# Local development
.\scripts\docker\libs_only.ps1
cd backend/app/<admin|app|core>/service
make run

# One-click full deployment
.\scripts\docker\full_deploy.ps1
```

### Frontend Startup

Frontend projects are located in the `frontend` directory. Dependency installation is unified:

| Frontend | Directory | Command |
|----------|-----------|---------|
| Admin | `frontend/admin` | `pnpm dev` |
| Storefront | `frontend/app` | `pnpm dev` |

```shell
# Install dependencies
pnpm install

# Admin (Vue3 + Element Plus)
cd frontend/admin && pnpm dev

# Storefront (Nuxt + Vue 3)
cd frontend/app && pnpm dev
```

> Listening ports are configured in each frontend's `.env.development`.

---

## Features

### Commerce

| Feature | Description |
|------|-----|
| Product Management | Publishing, listing/delisting, and maintenance of products and SKUs. |
| Product Attributes | Management of product attributes and attribute values. |
| Category Management | Product category management with tree structure. |
| Brand Management | Brand information management. |
| SKU & Pricing | Management of SKUs, attribute combinations, and prices. |
| Shopping Cart | Shopping cart and cart item management. |
| Order Management | Order and order detail management. |
| Payment & Refund | Payment transaction and refund record management. |
| File Transfer | Upload and transfer of product images and resources. |

### Platform & System

| Feature | Description |
|------|-----|
| User Profile | User profile information management. |
| Tenant Management | Multi-tenant management and initialization. |
| Role Management | Role and role group management. |
| Permission Management | Permission group and permission point management. |
| Menu Management | System menu and button permission configuration. |
| Organization Management | Organization, department, and position management. |
| API Management | API registration and synchronization. |
| Dictionary & i18n | Data dictionary and multi-language content management. |
| Task Scheduling | Async task and scheduling log management. |
| File Management | File upload and object storage management. |
| Internal Message | In-app message and message category management. |
| Audit Logs | Login, operation, permission, and data access audit logs. |

---

## Project Structure

```
go-wind-shop/
├── backend/                        # Backend project
│   ├── api/                        # Protobuf API definitions and generated code
│   │   ├── protos/                 # .proto source files (layered by domain)
│   │   └── gen/go/                 # buf-generated Go code
│   ├── app/                        # Application layer (admin / app / core microservices)
│   │   └── <admin|app|core>/service/
│   │       ├── cmd/server/         # Entry point + wire injection
│   │       ├── configs/            # Runtime configuration (YAML)
│   │       └── internal/           # Business core (server / service / data)
│   ├── pkg/                        # Cross-service shared libraries
│   ├── scripts/                    # Deployment and environment scripts (env / docker / deploy)
│   ├── sql/                        # Database seed / demo data scripts
│   └── Makefile / app.mk           # Build orchestration
├── frontend/                       # Frontend project
│   ├── admin/                      # Vue 3 + Element Plus admin
│   └── app/                        # Nuxt + Vue 3 storefront
└── README.md
```

---

## Contact

- WeChat: `yang_lin_bo` (note: `go-wind-shop`)
