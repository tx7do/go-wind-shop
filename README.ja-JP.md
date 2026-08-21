<div align="center">

# GoWind Shop｜風行ショップ

**すぐ使える企業向け越境ECフルスタックスキャフォールド**

> 越境EC開発を風のように自由に — GoWind Shop

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev/)
[![Vue](https://img.shields.io/badge/Vue-3.x-4FC08D?logo=vuedotjs)](https://vuejs.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)](https://www.docker.com/)

[English](./README.en-US.md) | [中文](./README.md) | **日本語**

</div>

---

## プロジェクト概要

GoWind Shop（風行ショップ）は、越境シナリオ向けのエンタープライズ級多言語ECプラットフォームです。APIファースト・フロントエンド/バックエンド分離アーキテクチャを採用し、ネイティブな多言語コンテンツ配信、多通貨の商品表示、単一決済通貨の機能を提供します。バックエンドは [golang](https://go.dev/) + [go-kratos](https://go-kratos.dev/) マイクロサービスフレームワークを基盤とし、Protobufファースト（契約駆動）のワークフローを採用。管理画面フロントエンドは [Vue.js 3](https://vuejs.org/) + [TypeScript](https://www.typescriptlang.org/)、店舗フロントエンドは [Nuxt](https://nuxt.com/) + Vue 3 で構築しています。

## 特徴

- **多言語ネイティブサポート**：i18n に基づく多言語コンテンツ配信、越境シナリオ向け設計。
- **契約駆動（Protobufファースト）**：インターフェース・データモデル・エラーコードは proto を唯一の真実の情報源とし、Go / TypeScript / OpenAPI を自動生成。
- **3サービスBFFアーキテクチャ**：`admin` / `app` の薄いゲートウェイと `core` のコア実装により、ビジネスロジックとトランスポート層を分離。
- **フルスタックコード生成**：Ent スキーマ → ORM、Wire による依存性注入、ワンクリック CRUD スキャフォールド。
- **本番対応**：JWT 認証、Casbin / OPA 認可、SSE プッシュ、Asynq 非同期タスク、OpenTelemetry トレーシング、Swagger ドキュメント、Docker ワンクリックデプロイ。

---

## システムアーキテクチャ

GoWind Shop は 3 サービスの BFF アーキテクチャを採用します。`admin` と `app` は薄いゲートウェイであり、パラメータ検証と gRPC 転送のみを行います。`core` はコア実装であり、すべてのデータアクセスはここに集約されます。

```
HTTP/REST リクエスト
  └─ admin-service  (管理画面ゲートウェイ BFF、REST:6600 / SSE:6601 / gRPC)
       └─ gRPC → core-service  (コア業務 + データ永続化)

app-service (店舗ゲートウェイ BFF、REST:6700) ──gRPC──→ core-service
core-service (gRPC、業務実装とデータアクセス)
```

| サービス | 役割 | DB アクセス |
|------|------|:---:|
| `app/admin/service` | 管理画面ゲートウェイ BFF（REST を受信し core へ gRPC 転送） | ❌ |
| `app/app/service` | 店舗ゲートウェイ BFF | ❌ |
| `app/core/service` | コア業務実装（永続化、実ロジック） | ✅ |

---

## 技術スタック

<table>
<tr><th>階層</th><th>技術</th></tr>
<tr><td><strong>バックエンドフレームワーク</strong></td><td><code>Golang</code> · <code>go-kratos v2</code> · <code>Wire</code> · <code>Protobuf / Buf</code></td></tr>
<tr><td><strong>ORM</strong></td><td><code>entgo.io/ent</code> · <code>PostgreSQL</code> · <code>MySQL</code></td></tr>
<tr><td><strong>ミドルウェア</strong></td><td><code>Redis</code> · <code>MinIO</code>（S3 互換オブジェクトストレージ） · <code>ElasticSearch / OpenSearch</code> · <code>etcd</code>（サービスディスカバリ）</td></tr>
<tr><td><strong>認証・認可</strong></td><td><code>kratos-authn</code>（JWT HS256） · <code>kratos-authz</code>（Casbin / OPA）</td></tr>
<tr><td><strong>リアルタイム通信</strong></td><td><code>SSE</code>（サーバープッシュ） · <code>Asynq</code>（非同期タスク）</td></tr>
<tr><td><strong>分散</strong></td><td><code>OpenTelemetry + Jaeger</code>（トレーシング） · <code>DTM</code>（分散トランザクション）</td></tr>
<tr><td><strong>管理画面フロントエンド</strong></td><td><code>Vue 3</code> · <code>TypeScript</code> · <code>Vite</code> · <code>Element Plus</code></td></tr>
<tr><td><strong>店舗フロントエンド</strong></td><td><code>Nuxt</code> · <code>Vue 3</code> · <code>TypeScript</code> · <code>Tailwind CSS</code> · <code>i18n</code></td></tr>
<tr><td><strong>デプロイ・運用</strong></td><td><code>Docker</code> · <code>Docker Compose</code> · <code>PM2</code> · <code>Swagger UI</code></td></tr>
</table>

---

## クイックスタート

### 環境要件

| ツール | バージョン |
|------|------|
| Go | 1.25+ |
| Node.js | >= 20.10.0 |
| pnpm | >= 10.0.0 |
| Docker | 20.0+ |

### 環境スクリプト選択

- Linux / macOS 開発環境：`scripts/env/install_unix_dev.sh`
- Linux / macOS 本番環境：`scripts/env/install_unix_prod.sh`
- Windows 開発環境：`scripts/env/install_windows_dev.ps1`

### Docker 2つのデプロイモード

- **full_deploy 完全モード**：ミドルウェア+バックエンドアプリを同時起動、ワンクリックデモ・本番デプロイに適用。
- **libs_only 依存モード（推奨）**：ミドルウェアのみ起動、アプリはローカルIDEで実行・デバッグ、日常開発に適用。

### バックエンド起動

**Linux / macOS：**

```shell
# スクリプトに実行権限を付与
chmod +x scripts/**/*.sh

# 開発環境（推奨）
./scripts/env/install_unix_dev.sh
./scripts/docker/libs_only.sh
cd backend/app/<admin|app|core>/service
make run

# 本番環境
./scripts/env/install_unix_prod.sh
./scripts/docker/full_deploy.sh

# PM2 プロセス管理（本番上級）
./scripts/deploy/pm2_service.sh
```

**Windows（PowerShell 管理者）：**

```powershell
# スクリプト実行ポリシーの許可（初回のみ1回実行）
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser

# 環境初期化
.\scripts\env\install_windows_dev.ps1

# ローカル開発
.\scripts\docker\libs_only.ps1
cd backend/app/<admin|app|core>/service
make run

# ワンクリック完全デプロイ
.\scripts\docker\full_deploy.ps1
```

### フロントエンド起動

フロントエンドプロジェクトは `frontend` ディレクトリに統一して配置されています。依存関係のインストールコマンドは共通です：

| フロントエンド | ディレクトリ | 起動コマンド |
|----------|-----------|---------|
| 管理画面 | `frontend/admin` | `pnpm dev` |
| 店舗フロント | `frontend/app` | `pnpm dev` |

```shell
# 依存関係のインストール
pnpm install

# 管理画面（Vue3 + Element Plus）
cd frontend/admin && pnpm dev

# 店舗フロント（Nuxt + Vue 3）
cd frontend/app && pnpm dev
```

> リッスンポートは各フロントエンドの `.env.development` で設定されます。

---

## 機能リスト

### 商品・取引

| 機能 | 説明 |
|------|-----|
| 商品管理 | 商品と SKU の公開・公開停止・公開解除、および保守。 |
| 商品属性 | 商品属性および属性値の管理。 |
| カテゴリ管理 | 商品カテゴリの管理、ツリー構造をサポート。 |
| ブランド管理 | ブランド情報の管理。 |
| SKU と価格設定 | SKU、属性の組み合わせ、価格の管理。 |
| ショッピングカート | カートおよびカート項目の管理。 |
| 注文管理 | 注文および注文明細の管理。 |
| 決済と返金 | 決済トランザクションと返金記録の管理（注：現在の決済ゲートウェイはシミュレーション実装であり、実際のゲートウェイ統合は今後対応予定）。 |
| 物流・出荷 | 注文履行と出荷管理。出荷状態機械と注文状態連携を含む。 |
| クーポンシステム | クーポンテンプレートとユーザークーポンの管理。決済時の割引エンジン、チェックアウトでの適用・返金時の復元を含む。 |
| ファイル転送 | 商品画像やリソースのアップロード・転送。 |

### プラットフォーム・システム

| 機能 | 説明 |
|------|-----|
| ユーザープロファイル | ユーザープロファイル情報の管理。 |
| テナント管理 | マルチテナントの管理と初期化。 |
| ロール管理 | ロールおよびロールグループの管理。 |
| 権限管理 | 権限グループと権限ポイントの管理。 |
| メニュー管理 | システムメニューとボタン権限の設定。 |
| 組織管理 | 組織・部署・役職の管理。 |
| インターフェース管理 | インターフェースの登録と同期。 |
| ディクショナリと多言語 | データディクショナリと多言語コンテンツの管理。 |
| タスクスケジューリング | 非同期タスクとスケジューリングログの管理。 |
| ファイル管理 | ファイルアップロードとオブジェクトストレージの管理。 |
| 内部メッセージ | サイト内メッセージとメッセージ分類の管理。 |
| 監査ログ | ログイン・操作・権限・データアクセスの監査ログ。 |
| パスワード再設定 | メール認証コードによるパスワード再設定フロー。 |
| ログインポリシー | ログインポリシーの設定と管理。 |

---

## プロジェクト構造

```
go-wind-shop/
├── backend/                        # バックエンドプロジェクト
│   ├── api/                        # Protobuf API 定義と生成コード
│   │   ├── protos/                 # .proto ソースファイル（ドメイン別階層）
│   │   └── gen/go/                 # buf 生成 Go コード
│   ├── app/                        # アプリケーション層（admin / app / core の3マイクロサービス）
│   │   └── <admin|app|core>/service/
│   │       ├── cmd/server/         # エントリポイント + wire インジェクション
│   │       ├── configs/            # 実行時設定（YAML）
│   │       └── internal/           # 業務コア（server / service / data）
│   ├── pkg/                        # サービス間共通ライブラリ
│   ├── scripts/                    # デプロイ・環境スクリプト（env / docker / deploy）
│   ├── sql/                        # データベースシード / デモデータスクリプト
│   └── Makefile / app.mk           # ビルドオーケストレーション
├── frontend/                       # フロントエンドプロジェクト
│   ├── admin/                      # Vue 3 + Element Plus 管理画面
│   └── app/                        # Nuxt + Vue 3 店舗フロント
└── README.md
```

---

## お問い合わせ

- WeChat 個人アカウント：`yang_lin_bo`（備考：`go-wind-shop`）
