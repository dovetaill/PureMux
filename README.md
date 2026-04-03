# PureMux

PureMux 现在已经落地了一套可运行的多 surface API 架构：在保留 `server` / `worker` / `scheduler` / `migrate` 脚手架能力的同时，把业务 API 明确拆成 `public`、`member auth`、`member self`、`admin` 四个 surface。

> 说明：当前这套完整多 surface 示例会保留在 `showcase/multisurface` 分支；`main` 会逐步收敛为 starter 主线。

当前架构切片已经具备：

- 后台管理员登录与当前身份读取
- 管理员用户管理
- 管理员分类管理
- 管理员文稿管理与发布 / 取消发布
- 前台公开分类 / 已发布文稿读取
- 前台会员注册 / 登录 / 自己的资料
- 前台会员点赞 / 收藏 / 我的收藏
- 统一 JWT 鉴权、principal 注入、Huma OpenAPI 注册

## 目录导航

- [当前能力概览](#当前能力概览)
  - [Admin API](#admin-api)
  - [Public API](#public-api)
  - [Member Auth API](#member-auth-api)
  - [Member Self API](#member-self-api)
  - [基础设施能力](#基础设施能力)
- [Principal 与权限规则](#principal-与权限规则)
- [模块地图](#模块地图)
- [统一响应结构](#统一响应结构)
- [快速启动](#快速启动)
  - [1. 准备依赖](#1-准备依赖)
  - [2. 准备配置文件](#2-准备配置文件)
  - [3. 选择主数据库](#3-选择主数据库)
  - [4. 启动 API 服务](#4-启动-api-服务)
  - [5. 启动 worker 与 scheduler](#5-启动-worker-与-scheduler)
  - [6. migrate 入口](#6-migrate-入口)
- [使用教程](#使用教程)
- [当前模块映射](#当前模块映射)
- [业务 Demo：代码应该怎么扩展](#业务-demo代码应该怎么扩展)
- [运行时入口分别做什么](#运行时入口分别做什么)
- [首版发布后暂缓项](#首版发布后暂缓项)
- [相关文档](#相关文档)
- [仓库目录与文件说明](#仓库目录与文件说明)
  - [顶层目录速查](#顶层目录速查)
  - [顶层文件说明](#顶层文件说明)
  - [详细文件地图（按目录）](#详细文件地图按目录)
  - [阅读源码时可以按这个顺序进入](#阅读源码时可以按这个顺序进入)

## 当前能力概览

### Admin API

- `POST /api/v1/auth/login`
- `GET /api/v1/auth/me`
- `POST /api/v1/admin/users`
- `GET /api/v1/admin/users`
- `GET /api/v1/admin/users/{id}`
- `PATCH /api/v1/admin/users/{id}`
- `DELETE /api/v1/admin/users/{id}`
- `POST /api/v1/admin/categories`
- `GET /api/v1/admin/categories`
- `GET /api/v1/admin/categories/{id}`
- `PATCH /api/v1/admin/categories/{id}`
- `DELETE /api/v1/admin/categories/{id}`
- `POST /api/v1/admin/articles`
- `GET /api/v1/admin/articles`
- `GET /api/v1/admin/articles/{id}`
- `PATCH /api/v1/admin/articles/{id}`
- `DELETE /api/v1/admin/articles/{id}`
- `POST /api/v1/admin/articles/{id}/publish`
- `POST /api/v1/admin/articles/{id}/unpublish`

### Public API

- `GET /api/v1/articles`
- `GET /api/v1/articles/{slug}`
- `GET /api/v1/categories`

### Member Auth API

- `POST /api/v1/member/auth/register`
- `POST /api/v1/member/auth/login`

### Member Self API

- `GET /api/v1/me`
- `GET /api/v1/me/favorites`
- `POST /api/v1/articles/{id}/likes`
- `DELETE /api/v1/articles/{id}/likes`
- `POST /api/v1/articles/{id}/favorites`
- `DELETE /api/v1/articles/{id}/favorites`

### 基础设施能力

- 原生 `http.ServeMux`
- Huma v2 OpenAPI 输出
- GORM 主数据库支持（`mysql` / `postgres`）
- Redis bootstrap
- Asynq worker skeleton
- cron scheduler skeleton
- migrate command skeleton
- `slog` 结构化日志

## Principal 与权限规则

系统当前围绕三类 principal 运转：

- `anonymous`
  - 可以访问 `public` surface
- `admin`
  - 可以访问全部 `admin` surface
  - 可以管理后台用户、分类、文稿
- `member`
  - 可以访问 `member auth` 登录后的 `member self` / `engagement` 能力
  - 不可访问 `/api/v1/admin/*`

权限规则统一为：

- `public` 路由允许匿名访问
- `admin` 路由只允许 `admin` principal
- `member self` 与互动路由只允许 `member` principal
- JWT 合法但账号状态为 `disabled` 时，受保护路由会拒绝访问
- 文稿业务规则仍留在 `article/service.go`，但 public surface 不再承载 ownership 决策

## 模块地图

当前业务模块与公共拼装点如下：

- `internal/modules/auth`
  - 后台管理员登录
  - 后台当前身份读取
- `internal/modules/user`
  - 后台管理员用户 CRUD
- `internal/modules/member`
  - 前台会员注册 / 登录 / 自己的资料
- `internal/modules/category`
  - `public_handler.go` 负责公开分类读取
  - `admin_handler.go` 负责后台分类管理
- `internal/modules/article`
  - `public_handler.go` 负责公开文章列表 / slug 详情
  - `admin_handler.go` 负责后台文稿管理与发布流转
- `internal/modules/engagement`
  - 点赞、收藏、取消操作
  - 我的收藏列表
- `internal/identity`
  - token / password / principal 公共能力
- `internal/middleware`
  - `Authenticate`、`RequireAdmin`、`RequireMember`
- `internal/api/register/router.go`
  - 只做依赖装配与 route groups wiring

## 统一响应结构

成功响应统一返回：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

列表接口统一返回分页结构：

```json
{
  "code": 0,
  "message": "user list",
  "data": {
    "page": 1,
    "page_size": 20,
    "total": 1,
    "items": []
  }
}
```

错误响应统一返回：

```json
{
  "code": 401,
  "message": "unauthorized"
}
```

## 快速启动

### 1. 准备依赖

你至少需要：

- Go `1.25+`
- Redis
- MySQL 或 PostgreSQL 二选一

### 2. 准备配置文件

```bash
cp configs/config.example.yaml configs/config.yaml
```

默认配置里已经包含首版业务所需的 JWT 和 seed admin 配置：

```yaml
auth:
  jwt:
    secret: change-me-in-production
    issuer: PureMux
    ttl_minutes: 120
  seed_admin:
    enabled: true
    username: admin
    password: admin123456
```

建议在本地启动前至少修改：

- `database.driver`
- `database.mysql.*` 或 `database.postgres.*`
- `redis.*`
- `auth.jwt.secret`
- `auth.seed_admin.username`
- `auth.seed_admin.password`

### 3. 选择主数据库

如果你使用 MySQL：

```yaml
database:
  driver: mysql
```

如果你使用 PostgreSQL：

```yaml
database:
  driver: postgres
```

### 4. 启动 API 服务

```bash
go run ./cmd/server -config configs/config.yaml
```

启动后可访问：

- `http://127.0.0.1:8080/healthz`
- `http://127.0.0.1:8080/readyz`
- `http://127.0.0.1:8080/openapi.json`
- `http://127.0.0.1:8080/docs`

第一次启动时，如果数据库中还没有管理员账号，系统会按 `auth.seed_admin` 自动创建默认管理员。

### 5. 启动 worker 与 scheduler

worker：

```bash
go run ./cmd/worker -config configs/config.yaml
```

scheduler：

```bash
go run ./cmd/scheduler -config configs/config.yaml
```

如果没有启用定时任务，也可以只启动 `server`。

### 6. migrate 入口

```bash
go run ./cmd/migrate -config configs/config.yaml
```

当前 `cmd/migrate` 仍然是 skeleton，已经能根据配置生成数据库 URL，但还没有接入真实 migration SQL 执行流程。

## 使用教程

下面用一个“admin 管理内容、member 使用前台能力、anonymous 读取公开内容”的流程说明当前系统。

### Step 1. 管理员登录

```bash
curl -X POST http://127.0.0.1:8080/api/v1/auth/login   -H 'Content-Type: application/json'   -d '{
    "username": "admin",
    "password": "admin123456"
  }'
```

### Step 2. 管理员创建分类与文稿

创建分类：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/admin/categories   -H 'Authorization: Bearer <admin-jwt>'   -H 'Content-Type: application/json'   -d '{
    "name": "公告",
    "slug": "notice",
    "description": "站内公告分类"
  }'
```

创建文稿：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/admin/articles   -H 'Authorization: Bearer <admin-jwt>'   -H 'Content-Type: application/json'   -d '{
    "title": "首篇站点公告",
    "summary": "这是一个简短摘要",
    "content": "这里是正文内容",
    "category_id": 1
  }'
```

发布文稿：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/admin/articles/1/publish   -H 'Authorization: Bearer <admin-jwt>'
```

### Step 3. 前台会员注册并登录

注册：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/member/auth/register   -H 'Content-Type: application/json'   -d '{
    "username": "alice",
    "password": "alice123456"
  }'
```

登录：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/member/auth/login   -H 'Content-Type: application/json'   -d '{
    "username": "alice",
    "password": "alice123456"
  }'
```

查看自己的资料：

```bash
curl http://127.0.0.1:8080/api/v1/me   -H 'Authorization: Bearer <member-jwt>'
```

### Step 4. 匿名访问公开内容

读取公开文章列表：

```bash
curl 'http://127.0.0.1:8080/api/v1/articles?page=1&page_size=20'
```

按 slug 读取公开文章详情：

```bash
curl http://127.0.0.1:8080/api/v1/articles/first-post
```

读取公开分类列表：

```bash
curl 'http://127.0.0.1:8080/api/v1/categories?page=1&page_size=20'
```

### Step 5. 会员点赞、收藏与查看自己的收藏

点赞：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/articles/1/likes   -H 'Authorization: Bearer <member-jwt>'
```

收藏：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/articles/1/favorites   -H 'Authorization: Bearer <member-jwt>'
```

查看我的收藏：

```bash
curl 'http://127.0.0.1:8080/api/v1/me/favorites?page=1&page_size=20'   -H 'Authorization: Bearer <member-jwt>'
```

## 当前模块映射

如果你之前是按 `handler / service / repository` 抽象来理解 PureMux，现在可以直接对照真实业务模块：

- 鉴权模块：`internal/modules/auth`
  - 后台管理员登录
  - 后台当前身份读取
- 管理员用户模块：`internal/modules/user`
  - 管理员用户 CRUD
  - 用户列表分页
- 前台会员模块：`internal/modules/member`
  - 会员注册 / 登录
  - 自己的资料
- 分类模块：`internal/modules/category`
  - `public_handler.go`：公开分类列表
  - `admin_handler.go`：后台分类管理
- 文稿模块：`internal/modules/article`
  - `public_handler.go`：公开文章列表 / slug 详情
  - `admin_handler.go`：后台文稿管理与发布
- 互动模块：`internal/modules/engagement`
  - 点赞 / 取消点赞
  - 收藏 / 取消收藏
  - 我的收藏
- 中间件：`internal/middleware`
  - `Authenticate`
  - `RequireAuthenticated`
  - `RequireAdmin`
  - `RequireMember`
- 统一依赖装配：`internal/api/register/router.go`
- 统一响应：`internal/api/response/response.go`

## 业务 Demo：代码应该怎么扩展

当前仓库已经把“示例分层”真正落到了真实模块里。你要继续加新业务时，建议先决定 surface，再决定模块边界，而不是把所有路由继续混回一个 handler。

### 推荐扩展顺序

1. 先判断这是 `public`、`member auth`、`member self` 还是 `admin` 能力
2. 在 `internal/modules/<module>` 下定义 `model.go`
3. 在 `repository.go` 写数据访问，不做权限决策
4. 在 `service.go` 写业务规则、状态流转与授权判断
5. 如有双 contract，优先拆成 `public_handler.go` / `admin_handler.go`
6. 把路由 ownership 留在模块 handler，不把业务路径规则塞回 `router.go`
7. 在 `internal/api/register/router.go` 只做依赖装配与 group 注册
8. 如有异步逻辑，再接入 `internal/queue`
9. 如有周期任务，再由 `internal/scheduler` 负责 enqueue

### 可以直接参考的真实实现

- 后台管理员登录：`internal/modules/auth`
- 会员注册 / 登录 / 自己：`internal/modules/member`
- public/admin 双面拆分：`internal/modules/category`、`internal/modules/article`
- member 互动能力：`internal/modules/engagement`
- principal / token / password 公共层：`internal/identity`
- 最终路由装配：`internal/api/register/router.go`

## 运行时入口分别做什么

### `cmd/server`

负责 HTTP API：

- 初始化 runtime
- 自动迁移业务表
- seed 默认管理员
- 注册业务路由
- 挂载健康检查 / 文档 / 中间件

### `cmd/worker`

负责异步任务消费：

- 初始化 runtime
- 构建 Asynq worker
- 注册任务处理器
- 等待信号并优雅退出

### `cmd/scheduler`

负责定时任务投递：

- 初始化 runtime
- 注册 cron job
- 到点后 enqueue 任务
- 不直接承担核心业务逻辑

### `cmd/migrate`

负责迁移入口：

- 加载配置
- 生成数据库 URL
- 约定迁移目录
- 为后续真实 migration 流程保留入口

## 首版发布后暂缓项

当前首版业务已经可运行，但下面这些能力仍然是明确延后项：

- refresh token
- 审计日志
- 文稿版本历史
- 审核流
- 真实 migration SQL 执行流程
- MySQL / PostgreSQL / Redis 外部依赖 smoke test
- 更细粒度权限模型（如分类级授权、发布审批人）

## 相关文档

- 验证记录：`verification.md`
- 示例模块说明：`internal/modules/example/README.md`
- 本轮业务设计：`docs/plans/2026-04-03-backoffice-publishing-design.md`
- 本轮业务实施计划：`docs/plans/2026-04-03-backoffice-publishing-implementation-plan.md`

## 仓库目录与文件说明

这一节专门回答“目录、文件、函数都做什么”。为了不打断前面的快速上手，下面按目录分组说明，优先覆盖 Git 已跟踪文件；像 `.git/` 这类运行时元数据目录不展开。

### 顶层目录速查

| 路径 | 作用 |
| --- | --- |
| `cmd/` | 4 个可执行入口：API 服务、worker、scheduler、migrate。 |
| `configs/` | 配置样例文件。 |
| `docs/plans/` | 历史设计稿、实施计划和 TDD 说明。 |
| `internal/` | 核心业务代码：API 装配、bootstrap、鉴权、中间件、业务模块、队列、调度器。 |
| `migrations/` | migration 占位目录，当前还没有真实 SQL。 |
| `pkg/` | 可复用基础设施包：配置、数据库、日志。 |
| `.codex/` | 本地 AI/代理协作说明文件。 |
| `.worktree/`、`.worktrees/` | 本地 worktree 辅助目录，不参与线上业务逻辑。 |

### 顶层文件说明

- `README.md`：项目总说明，包含能力概览、启动方式、API 示例，以及这份仓库地图。
- `.gitignore`：忽略日志、构建物、本地缓存、私有配置等不应提交的文件。
- `go.mod`：Go 模块定义与直接依赖入口。
- `go.sum`：Go 依赖校验和锁文件，用来保证依赖下载结果一致。
- `verification.md`：阶段性验证记录，汇总测试、构建、格式化、静态检查的执行结果。
- `.codex/AGENTS.md`：给 AI/Codex 的本地操作规约，说明工具、流程和约束。

### 详细文件地图（按目录）

<details>
<summary><code>cmd/</code>：四个运行入口</summary>

- `cmd/server/main.go`：HTTP API 服务入口。
  - `main`：读取配置、构建 server runtime、注册路由、启动 `http.Server`、处理优雅退出。
  - `envOrDefault`：优先读环境变量，否则回退默认配置路径。
- `cmd/worker/main.go`：异步任务消费入口。
  - `main`：初始化 worker runtime、创建 Asynq server、注册任务处理器、监听退出信号。
  - `envOrDefault`：读取环境变量中的配置路径。
- `cmd/scheduler/main.go`：定时任务投递入口。
  - `main`：初始化 scheduler runtime、创建 Asynq client、注册 cron job、等待退出信号并停止调度器。
  - `envOrDefault`：读取环境变量中的配置路径。
- `cmd/migrate/main.go`：数据库迁移命令入口。
  - `main`：加载配置并执行迁移骨架流程。
  - `envOrDefault`：读取环境变量中的配置路径。

</details>

<details>
<summary><code>configs/</code>：配置样例</summary>

- `configs/config.example.yaml`：本地启动示例配置，覆盖 app/http/database/redis/auth/queue/scheduler/docs/log 等配置段，复制成 `configs/config.yaml` 后即可改成本地值。

</details>

<details>
<summary><code>docs/plans/</code>：设计与实施记录</summary>

- `docs/plans/2026-03-18-stage1-core-infra-design.md`：Stage 1 核心基础设施设计稿。
- `docs/plans/2026-03-18-stage1-core-infra-implementation-plan.md`：Stage 1 基础设施实施计划拆解。
- `docs/plans/2026-03-18-stage1-core-infra-tdd-spec.md`：Stage 1 的 TDD 规格说明。
- `docs/plans/2026-04-02-modular-api-scaffold-design.md`：模块化 API 脚手架设计稿。
- `docs/plans/2026-04-02-modular-api-scaffold-implementation-plan.md`：模块化 API 脚手架实施计划。
- `docs/plans/2026-04-03-backoffice-publishing-design.md`：后台发布系统设计稿。
- `docs/plans/2026-04-03-backoffice-publishing-implementation-plan.md`：后台发布系统实施计划。
- `docs/plans/2026-04-03-multi-surface-api-architecture-design.md`：`public / member auth / member self / admin` 多 surface API 架构设计稿。
- `docs/plans/2026-04-03-multi-surface-api-architecture-implementation-plan.md`：多 surface API 架构的实施计划。

</details>

<details>
<summary><code>internal/api/handlers/</code>：健康检查与就绪检查</summary>

- `internal/api/handlers/health.go`：注册健康检查接口。
  - `RegisterHealth`：挂载 `/healthz`，返回 `alive` 和 `status: up`。
- `internal/api/handlers/ready.go`：注册 readiness 接口。
  - `RegisterReady`：挂载 `/readyz`，按 runtime 中的 MySQL/Redis 资源状态返回依赖可用性。
- `internal/api/handlers/health_test.go`：验证健康检查、就绪检查和统一响应结构。
  - 主要测试函数：`TestHealthzReturnsAlive`、`TestReadyzReturnsDependencyStatus`、`TestResponseHelpersReturnStandardShape`。

</details>

<details>
<summary><code>internal/api/register/</code>：最终 HTTP 路由装配层</summary>

- `internal/api/register/router.go`：整个 API 的总装配入口，只负责 wiring，不写具体业务规则。
  - `NewRouter`：创建 Huma API、划分 `public/member auth/member self/admin` route groups、挂中间件、注册所有业务模块。
  - `normalizeOpenAPIPath`：兼容 `/openapi` 与 `/openapi.json` 配置写法。
  - `newAuthService`、`newUserService`、`newCategoryService`、`newArticleService`、`newMemberService`、`newEngagementService`：按 runtime 资源构造各业务 service。
  - `nilLogger`：从 runtime 安全取出 logger。
  - `requireAdminMiddleware`、`requireMemberMiddleware`：把 admin/member surface 的访问约束做成 Huma middleware。
  - `Authenticate`：`tokenAuthenticator` 的组合接口，用于多鉴权源兜底。
  - `memberJWTConfig`：给 member surface 派生一套独立 JWT 配置。
- `internal/api/register/router_test.go`：验证总路由 wiring 是否正确。
  - 主要测试函数：`TestRouterRegistersAuthAndBusinessRoutes`、`TestPublicArticleRoutesAreAccessibleWithoutAuth`、`TestMemberRoutesRequireMemberAuth`、`TestAdminRoutesRejectNonAdminPrincipal`。
  - 辅助函数：`newRouterTestRuntime`、`assertOperation`、`httpMethodKey` 用于生成测试 runtime 和校验 OpenAPI 注册结果。

</details>

<details>
<summary><code>internal/api/response/</code>：统一响应封装</summary>

- `internal/api/response/response.go`：定义统一响应 envelope。
  - `Envelope`：统一的 `code/message/data` 响应结构。
  - `OK`：返回标准成功响应。
  - `Paged`：返回标准分页成功响应。
  - `Fail`：返回标准错误响应。
- `internal/api/response/response_test.go`：验证分页响应结构。
  - 主要测试函数：`TestPagedReturnsStandardShape`。

</details>

<details>
<summary><code>internal/app/bootstrap/</code>：运行时资源装配</summary>

- `internal/app/bootstrap/runtime.go`：定义所有入口共享的 runtime。
  - `Runtime`：聚合配置、日志、数据库/Redis 资源以及退出回调。
  - `RegisterCloser`：注册退出时要执行的回收动作。
  - `Shutdown`：按逆序统一关闭 runtime 资源。
  - `buildRuntime`：加载配置、初始化 logger 和数据库资源，拼出最小运行时。
- `internal/app/bootstrap/server.go`：server 专用 bootstrap。
  - `BuildServerRuntime`：创建 runtime，并追加 schema 自动迁移与默认管理员 seed。
  - `bootstrapServerBusinessSchema`：执行业务表迁移和默认管理员初始化。
- `internal/app/bootstrap/worker.go`：worker 专用 bootstrap。
  - `BuildWorkerRuntime`：为 worker 入口构建共享 runtime。
- `internal/app/bootstrap/scheduler.go`：scheduler 专用 bootstrap。
  - `BuildSchedulerRuntime`：为定时任务入口构建共享 runtime。
- `internal/app/bootstrap/migrate.go`：migrate 入口骨架。
  - `MigrateConfig`：migration 命令需要的最小配置结构。
  - `BuildMigrateConfig`：按主库驱动生成数据库 URL 与 `file://migrations` 源地址。
  - `RunMigrateCommand`：migrate 入口的配置校验流程，当前还未接入真实 SQL 执行。
  - `normalizeDatabaseDriver`、`resolveMigrateMySQLConfig`、`buildMySQLMigrateURL`、`buildPostgresMigrateURL`：迁移配置辅助函数。
- `internal/app/bootstrap/schema.go`：业务模型自动迁移与默认管理员 seed。
  - `RegisterBusinessModels`：收集所有业务模型，供自动迁移使用。
  - `RegisterSeedAdminSupport`：注册默认管理员仓储和密码哈希实现。
  - `AutoMigrateBusinessTables`：把已注册业务模型交给 GORM 自动建表。
  - `SeedDefaultAdmin`：在没有管理员时创建默认后台管理员。
  - `SeedAdminAccount`、`SeedAdminStore`、`passwordHasher`：seed admin 所需的契约定义。
- `internal/app/bootstrap/bootstrap_test.go`：验证 runtime 资源共享和关闭顺序。
  - 主要测试函数：`TestBuildServerRuntimeReturnsSharedResources`、`TestBuildWorkerRuntimeReturnsSharedResources`、`TestShutdownRunsClosersInReverseOrder`。
- `internal/app/bootstrap/migrate_test.go`：验证 migration 配置生成逻辑。
  - 主要测试函数：`TestBuildMigrateConfigUsesSelectedDriver`、`TestMigrateCommandRejectsUnsupportedDriver`。
- `internal/app/bootstrap/schema_test.go`：验证模型迁移和默认管理员 seed 流程。
  - 主要测试函数：`TestAutoMigrateRegistersAllBusinessModels`、`TestSeedAdminCreatesDefaultAdminWhenMissing`、`TestSeedAdminSkipsWhenDisabled`、`TestBuildServerRuntimeRunsSchemaAndSeedAdmin`。

</details>

<details>
<summary><code>internal/app/lifecycle/</code>：退出生命周期</summary>

- `internal/app/lifecycle/shutdown.go`：统一的资源关闭工具。
  - `Closer`：退出阶段执行的关闭函数类型。
  - `Shutdown`：按逆序执行所有关闭动作，并聚合错误。

</details>

<details>
<summary><code>internal/identity/</code>：principal、密码、token 的公共桥接层</summary>

- `internal/identity/claims.go`：principal 抽象。
  - `PrincipalKind`、`Principal`：统一描述 admin/member 身份。
  - `PrincipalFromCurrentUser`：把 `auth.CurrentUser` 转成 principal，供中间件与路由守卫复用。
- `internal/identity/context.go`：principal 的 context 读写。
  - `ContextWithPrincipal`：把 principal 写入请求上下文。
  - `PrincipalFromContext`：从上下文中取 principal。
- `internal/identity/password.go`：密码函数桥接。
  - `HashPassword`：复用 auth 模块的哈希逻辑。
  - `VerifyPassword`：复用 auth 模块的校验逻辑。
- `internal/identity/token.go`：token manager 桥接。
  - `TokenManager`、`TokenClaims`：复用 auth 模块中的类型别名。
  - `NewTokenManager`：用统一配置创建 JWT 管理器。

</details>

<details>
<summary><code>internal/middleware/</code>：HTTP 中间件</summary>

- `internal/middleware/requestid.go`：中间件基础设施。
  - `Middleware`：中间件函数类型。
  - `Chain`：按顺序组合中间件。
  - `RequestID`：为每个请求注入 `X-Request-ID`。
- `internal/middleware/accesslog.go`：访问日志中间件。
  - `AccessLog`：记录 method、path、耗时等信息。
- `internal/middleware/auth.go`：JWT 解析与上下文注入。
  - `Authenticate`：解析 `Authorization: Bearer <token>`，写入 `CurrentUser` 和 principal。
  - `bearerToken`：从请求头中提取 Bearer token。
  - `writeAuthError`：输出统一 JSON 鉴权错误。
- `internal/middleware/authorize.go`：角色守卫。
  - `RequireAuthenticated`：要求请求里存在当前用户。
  - `RequireAdmin`：要求当前 principal/user 是 admin。
  - `RequireMember`：要求当前 principal 是 member。
- `internal/middleware/recover.go`：panic 恢复。
  - `Recover`：兜底捕获 panic，返回统一 500 响应。
- `internal/middleware/timeout.go`：超时控制。
  - `Timeout`：把 handler 包装成带超时的 `http.TimeoutHandler`。
- `internal/middleware/auth_test.go`：验证鉴权与角色守卫。
  - 主要测试函数：`TestAuthenticateStoresAdminPrincipal`、`TestAuthenticateStoresMemberPrincipal`、`TestRequireAdminRejectsMemberPrincipal`、`TestRequireMemberRejectsAnonymous`。

</details>

<details>
<summary><code>internal/modules/auth/</code>：后台管理员登录与当前身份</summary>

- `internal/modules/auth/model.go`：管理员用户模型与当前用户上下文。
  - `User`、`CurrentUser`：后台用户存储模型与请求期身份模型。
  - `TableName`：声明表名 `users`。
  - `ToCurrentUser`：把数据库模型转成上下文可用的轻量身份对象。
  - `ContextWithCurrentUser`、`CurrentUserFromContext`：在请求上下文写入/读取当前用户。
  - `init`：注册业务模型，并把默认管理员 seed 能力挂到 bootstrap。
- `internal/modules/auth/password.go`：密码处理。
  - `HashPassword`：生成密码哈希。
  - `VerifyPassword`：校验密码哈希。
- `internal/modules/auth/jwt.go`：JWT 签发与解析。
  - `NewTokenManager`：根据配置创建 JWT 管理器。
  - `Sign`：签发 token。
  - `Parse`：解析 token。
- `internal/modules/auth/repository.go`：后台用户仓储。
  - `NewRepository`：创建 GORM 仓储。
  - `FindByUsername`、`FindByID`：按用户名/ID 查用户。
  - `HasAdmin`：判断是否已有管理员。
  - `CreateAdmin`：创建默认管理员账号。
- `internal/modules/auth/service.go`：后台登录业务。
  - `NewService`：构造 auth service。
  - `Login`：验证管理员用户名密码并签发 JWT。
  - `Authenticate`：解析 token 并回查数据库，得到当前管理员身份。
  - `StatusFromError`：把 auth 领域错误映射到 HTTP 状态码。
- `internal/modules/auth/handler.go`：后台登录与 `/auth/me` 路由。
  - `RegisterRoutes`：注册 `/api/v1/auth/login` 和 `/api/v1/auth/me`。
- `internal/modules/auth/auth_test.go`：auth 模块测试。
  - 主要测试函数：`TestLoginReturnsJWTForValidCredentials`、`TestLoginRejectsInvalidPassword`、`TestAuthMiddlewareLoadsCurrentUserFromBearerToken`、`TestAuthMiddlewareRejectsDisabledUser`、`TestRequireAdminRejectsNonAdmin`。
  - 辅助函数：`newAuthService`、`newAuthHandler`、`mustHashPassword`、`mustLogin`、`mustSignToken` 等用于快速拼装测试上下文。

</details>

<details>
<summary><code>internal/modules/user/</code>：后台管理员用户 CRUD</summary>

- `internal/modules/user/model.go`：用户模型别名与角色/状态常量复用。
  - `User`：直接复用 auth 模块的 `User` 模型。
- `internal/modules/user/repository.go`：管理员用户仓储。
  - `NewRepository`：创建仓储。
  - `Create`、`Update`、`Delete`：写操作。
  - `List`、`FindByID`、`FindByUsername`：查询操作。
- `internal/modules/user/service.go`：管理员用户业务层。
  - `NewService`：构造 service。
  - `Create`：创建后台用户并处理用户名冲突。
  - `List`：分页列出用户。
  - `Get`：读取单个用户。
  - `Update`：修改用户名、密码、角色、状态。
  - `Delete`：删除用户。
  - `StatusFromError`：把领域错误转换为 HTTP 状态。
  - `normalizeRole`、`normalizeStatus`、`normalizePage`：做输入规范化。
- `internal/modules/user/handler.go`：管理员用户路由。
  - `RegisterRoutes`：注册 `/api/v1/admin/users` 下的创建、列表、详情、更新、删除接口。
  - `parseID`：解析路径中的用户 ID。
  - `stringValue`：把可选字符串指针转成稳定值。
- `internal/modules/user/user_test.go`：用户模块测试。
  - 主要测试函数：`TestAdminCanCreateUser`、`TestAdminCanListUsers`、`TestNonAdminCannotAccessUserAdminEndpoints`、`TestCreateUserRejectsDuplicateUsername`、`TestAdminCanUpdateUser`、`TestAdminCanDeleteUser`。
  - 辅助函数：`newAdminUserHandler`、`newActorAuthService`、`mustActorToken`、`decodeEnvelope` 等用于模拟后台请求链路。

</details>

<details>
<summary><code>internal/modules/member/</code>：前台会员注册、登录、我的资料</summary>

- `internal/modules/member/model.go`：会员模型与 profile 视图。
  - `Member`：会员表模型。
  - `Profile`：对前台暴露的会员信息结构。
  - `TableName`：声明表名 `members`。
  - `ToCurrentUser`：把会员转成通用 `CurrentUser`，角色固定为 `member`。
  - `ToProfile`：把会员转成前台 profile 结构。
  - `init`：注册会员模型到自动迁移清单。
- `internal/modules/member/repository.go`：会员仓储。
  - `NewRepository`：创建仓储。
  - `Create`：创建会员。
  - `FindByUsername`、`FindByID`：按用户名/ID 读取会员。
- `internal/modules/member/service.go`：会员业务层。
  - `NewService`：构造 member service。
  - `Register`：注册会员并立即签发 token。
  - `Login`：会员登录并返回 token。
  - `Authenticate`：解析 member token 并回查数据库。
  - `GetSelf`：读取当前会员资料。
  - `issueToken`：内部签发 token 帮助函数。
  - `StatusFromError`：错误到 HTTP 状态码映射。
- `internal/modules/member/public_handler.go`：会员公开认证接口。
  - `RegisterPublicRoutes`：注册 `/api/v1/member/auth/register` 和 `/api/v1/member/auth/login`。
- `internal/modules/member/self_handler.go`：会员“我的资料”接口。
  - `RegisterSelfRoutes`：注册 `/api/v1/me`。
- `internal/modules/member/member_test.go`：会员模块测试。
  - 主要测试函数：`TestMemberRegister`、`TestMemberLogin`、`TestMemberCanFetchSelfProfile`。
  - 辅助函数：`newMemberHandler`、`mustHashPassword`、`mustLogin`、`decodeEnvelope` 等用于构造测试上下文。

</details>

<details>
<summary><code>internal/modules/category/</code>：分类模块（public + admin 双 surface）</summary>

- `internal/modules/category/model.go`：分类模型。
  - `Category`：分类表模型。
  - `TableName`：声明表名 `categories`。
  - `init`：把分类模型注册到自动迁移清单。
- `internal/modules/category/repository.go`：分类仓储。
  - `NewRepository`：创建仓储。
  - `Create`、`Update`、`Delete`：写操作。
  - `List`、`FindByID`、`FindBySlug`：查询操作。
- `internal/modules/category/service.go`：分类业务层。
  - `NewService`：构造 service。
  - `Create`、`Update`、`Delete`：后台分类管理。
  - `List`、`ListPublic`、`Get`：后台/公开读取。
  - `StatusFromError`：错误到 HTTP 状态码映射。
  - `normalizeSlug`、`normalizePage`：slug 和分页规范化。
- `internal/modules/category/handler.go`：分类请求结构与公共工具。
  - `parseID`：解析分类 ID。
  - `stringValue`：提取可选字符串。
- `internal/modules/category/public_handler.go`：公开分类接口。
  - `RegisterPublicRoutes`：注册 `/api/v1/categories` 列表接口。
- `internal/modules/category/admin_handler.go`：后台分类接口。
  - `RegisterAdminRoutes`：注册 `/api/v1/admin/categories` 的创建、列表、详情、更新、删除接口。
- `internal/modules/category/category_test.go`：分类模块测试。
  - 主要测试函数：`TestPublicCategoryListIsAccessibleWithoutAuth`、`TestAdminCategoryCrudStillRequiresAdmin`、`TestAdminCanCreateCategory`、`TestAdminCanListCategories`、`TestNonAdminCannotAccessCategoryAdminEndpoints`、`TestCreateCategoryRejectsDuplicateSlug`、`TestAdminCanUpdateCategory`、`TestAdminCanDeleteCategory`。

</details>

<details>
<summary><code>internal/modules/article/</code>：文稿模块（public + admin 双 surface）</summary>

- `internal/modules/article/model.go`：文稿模型与过滤条件。
  - `Article`：文稿表模型，包含标题、slug、摘要、正文、发布状态、作者和分类信息。
  - `ListFilter`：文稿列表查询过滤条件。
  - `TableName`：声明表名 `articles`。
  - `init`：注册文稿模型到自动迁移清单。
- `internal/modules/article/repository.go`：文稿仓储。
  - `NewRepository`：创建仓储。
  - `Create`、`Update`、`Delete`：文稿写操作。
  - `List`、`FindByID`、`FindBySlug`：文稿查询操作。
- `internal/modules/article/service.go`：文稿业务层。
  - `NewService`：构造 service。
  - `Create`：创建草稿文稿。
  - `List`：后台查看文稿列表；普通用户只能看到自己的文稿。
  - `ListPublic`：公开列表，只返回已发布文稿。
  - `Get`、`GetPublicBySlug`：后台和公开详情读取。
  - `Update`、`Delete`：编辑或删除文稿。
  - `Publish`、`Unpublish`：发布状态流转。
  - `StatusFromError`：错误到 HTTP 状态码映射。
  - `loadOwnedArticle`：校验 ownership/admin 权限。
  - `normalizePage`、`normalizeSlug`：分页与 slug 规范化。
- `internal/modules/article/handler.go`：文稿路由共用 DTO 与工具函数。
  - `parseID`：解析文稿 ID。
  - `stringValue`：处理可选字符串字段。
- `internal/modules/article/public_handler.go`：公开文稿接口。
  - `RegisterPublicRoutes`：注册公开列表和按 slug 取详情接口。
- `internal/modules/article/admin_handler.go`：后台文稿接口。
  - `RegisterAdminRoutes`：注册创建、列表、详情、更新、删除、发布、取消发布接口。
- `internal/modules/article/article_test.go`：文稿模块测试。
  - 主要测试函数：`TestAdminCanCreateDraftArticle`、`TestAdminCanListDraftAndPublishedArticles`、`TestPublicArticleListOnlyReturnsPublishedItems`、`TestPublicArticleDetailLoadsBySlug`、`TestAdminArticleRoutesPreserveDraftManagement`、`TestNonAdminCannotManageAdminArticleRoutes`、`TestAdminCanManageAnyArticle`、`TestAdminPublishAndUnpublishTransitionsStatus`。
  - 辅助函数：`newArticleHandler`、`newActorAuthService`、`mustActorToken`、`decodeEnvelope` 等用于模拟前后台请求。

</details>

<details>
<summary><code>internal/modules/engagement/</code>：会员互动模块（点赞、收藏、我的收藏）</summary>

- `internal/modules/engagement/model.go`：互动模型。
  - `Like`：点赞记录模型，对 `(member_id, article_id)` 做联合唯一约束。
  - `Favorite`：收藏记录模型，对 `(member_id, article_id)` 做联合唯一约束。
  - `TableName`：声明 `article_likes` / `article_favorites` 表名。
  - `init`：注册互动模型到自动迁移清单。
- `internal/modules/engagement/repository.go`：互动仓储。
  - `NewRepository`：创建仓储。
  - `CreateLike`、`DeleteLike`：点赞增删。
  - `CreateFavorite`、`DeleteFavorite`：收藏增删。
  - `ListFavorites`：分页列出会员收藏。
  - `isDuplicateError`：识别唯一索引冲突。
- `internal/modules/engagement/service.go`：互动业务层。
  - `NewService`：构造 service。
  - `Like`、`Unlike`：点赞/取消点赞。
  - `Favorite`、`Unfavorite`：收藏/取消收藏。
  - `ListFavorites`：分页返回我的收藏。
  - `StatusFromError`：错误到 HTTP 状态码映射。
  - `normalizePage`：规范分页参数。
- `internal/modules/engagement/handler.go`：互动路由。
  - `RegisterRoutes`：注册点赞、取消点赞、收藏、取消收藏、我的收藏接口。
  - `currentMember`：从上下文提取当前会员 ID。
  - `parseID`：解析文稿 ID。
- `internal/modules/engagement/engagement_test.go`：互动模块测试。
  - 主要测试函数：`TestMemberCanLikeArticle`、`TestMemberCanFavoriteArticle`、`TestDuplicateFavoriteReturnsConflict`、`TestAnonymousCannotFavoriteArticle`、`TestMyFavoritesReturnsMemberScopedList`。
  - 辅助函数：`newEngagementHandler`、`Authenticate`、`engagementKey`、`decodeEnvelope` 等用于构造 member 请求上下文。

</details>

<details>
<summary><code>internal/modules/example/</code>：模块化开发说明</summary>

- `internal/modules/example/README.md`：不是业务代码，而是“如何照着现有模块扩展新业务”的说明文档，解释 `handler -> service -> repository` 的职责划分，以及 public/member/admin 多 surface 的落点。

</details>

<details>
<summary><code>internal/queue/asynq/</code>：异步任务执行层</summary>

- `internal/queue/asynq/client.go`：Asynq 投递端。
  - `NewClient`：用 runtime 里的 Redis 客户端创建 Asynq client。
  - `EnqueueTask`：把标准任务投递到指定队列。
  - `runtimeRedis`：安全读取 runtime 中的 Redis 资源。
- `internal/queue/asynq/server.go`：Asynq worker server。
  - `NewServer`：按配置创建 Asynq worker，设置并发度和队列名。
- `internal/queue/asynq/handlers.go`：任务处理器注册。
  - `RegisterHandlers`：注册当前 worker 支持的任务处理函数；现阶段只处理 `runtime:heartbeat`。
- `internal/queue/asynq/asynq_test.go`：异步任务测试。
  - 主要测试函数：`TestNewTaskBuildsStableTaskNameAndPayload`、`TestEnqueueHelperBuildsAsynqTask`、`TestRegisterHandlersReturnsMuxWithKnownTaskTypes`。
  - `Enqueue`：测试替身中的投递实现。

</details>

<details>
<summary><code>internal/queue/tasks/</code>：标准任务定义</summary>

- `internal/queue/tasks/tasks.go`：任务定义入口。
  - `TypeRuntimeHeartbeat`：当前预置的任务类型常量。
  - `NewTask`：把标准 payload 编码成 Asynq task。
- `internal/queue/tasks/payload.go`：任务载荷定义。
  - `Payload`：当前标准任务载荷，包含 `source` 字段。
  - `DecodePayload`：把 Asynq task 反解成业务可读 payload。

</details>

<details>
<summary><code>internal/scheduler/</code>：cron 调度层</summary>

- `internal/scheduler/scheduler.go`：调度器主入口。
  - `New`：创建最小 cron scheduler。
  - `RegisterJobs`：按配置注册所有定时任务。
- `internal/scheduler/jobs.go`：定时任务定义。
  - `Enqueuer`：调度器依赖的最小投递接口。
  - `NewAsynqEnqueuer`：把 Asynq client 适配成调度器所需的 `Enqueuer`。
  - `EnqueueRuntimeHeartbeat`：投递 heartbeat 任务。
  - `NewRuntimeHeartbeatJob`：返回一个只负责 enqueue 的 cron job 函数。
- `internal/scheduler/scheduler_test.go`：调度器测试。
  - 主要测试函数：`TestRegisterJobsAddsCronEntries`、`TestScheduledJobOnlyEnqueuesTask`。
  - `EnqueueRuntimeHeartbeat`：测试替身里的投递实现。

</details>

<details>
<summary><code>migrations/</code>：数据库迁移目录</summary>

- `migrations/.keep`：占位文件，用来保证空目录能被 Git 跟踪；当前仓库还没有真正的 migration SQL 文件。

</details>

<details>
<summary><code>pkg/config/</code>：配置读取与配置结构</summary>

- `pkg/config/doc.go`：package 级说明文件，用于放置 `config` 包文档注释。
- `pkg/config/config.go`：所有强类型配置结构。
  - `Config`：总配置入口。
  - `AppConfig`、`HTTPConfig`、`MySQLConfig`、`RedisConfig`、`DatabaseConfig`、`DBMySQLConfig`、`PostgresConfig`、`AuthConfig`、`JWTConfig`、`SeedAdminConfig`、`QueueConfig`、`AsynqConfig`、`SchedulerConfig`、`DocsConfig`、`LogConfig`：分别描述应用、网络、数据库、认证、队列、调度、文档、日志配置。
- `pkg/config/load.go`：配置加载实现。
  - `Load`：从 YAML 读配置，并支持环境变量覆盖。
- `pkg/config/config_test.go`：配置包测试。
  - 主要测试函数：`TestLoadReadsYAML`、`TestLoadEnvOverridesYAML`、`TestLoadReturnsErrorForMissingRequiredFields`、`TestLoadReadsDatabaseDriver`、`TestLoadReadsPostgresConfig`、`TestLoadReadsQueueAndSchedulerConfig`、`TestLoadReadsJWTConfig`、`TestLoadReadsSeedAdminConfig`、`TestLoadAppliesHTTPTimeoutDefaults`。
  - `writeConfigFile`：测试里生成临时配置文件的辅助函数。

</details>

<details>
<summary><code>pkg/database/</code>：数据库与 Redis bootstrap</summary>

- `pkg/database/doc.go`：package 级说明文件，用于放置 `database` 包文档注释。
- `pkg/database/resources.go`：资源聚合入口。
  - `Resources`：聚合 GORM 主库和 Redis 客户端。
  - `Bootstrap`：按顺序初始化主数据库与 Redis。
  - `Close`：统一关闭 Redis 与 SQL 连接。
- `pkg/database/driver.go`：数据库驱动分发层。
  - `openPrimaryDatabase`：根据配置决定走 MySQL 还是 PostgreSQL。
  - `normalizedDriver`：规范化驱动名称。
  - `resolveMySQLConfig`：兼容 `database.mysql` 与顶层 `mysql` 配置来源。
- `pkg/database/mysql.go`：MySQL 启动逻辑。
  - `openMySQL`：建立 GORM MySQL 连接并做 ping、连接池设置。
  - `buildMySQLDSN`：构建 MySQL DSN。
- `pkg/database/postgres.go`：PostgreSQL 启动逻辑。
  - `openPostgres`：建立 GORM PostgreSQL 连接并做 ping、连接池设置。
  - `buildPostgresDSN`：构建 PostgreSQL DSN。
- `pkg/database/redis.go`：Redis 启动逻辑。
  - `openRedis`：建立 Redis 客户端并做 ping。
  - `buildRedisOptions`：根据配置构建 Redis 选项。
- `pkg/database/database_test.go`：数据库 bootstrap 测试。
  - 主要测试函数：`TestBuildMySQLDSN`、`TestBuildRedisOptions`、`TestResourcesCloseIsSafe`、`TestBootstrapUsesMySQLDialectorWhenDriverIsMySQL`、`TestBootstrapUsesPostgresDialectorWhenDriverIsPostgres`、`TestBuildPostgresDSN`、`TestBootstrapReturnsErrorForUnsupportedDriver`。

</details>

<details>
<summary><code>pkg/logger/</code>：结构化日志与轮转</summary>

- `pkg/logger/doc.go`：package 级说明文件，用于放置 `logger` 包文档注释。
- `pkg/logger/logger.go`：日志创建入口。
  - `New`：根据配置创建 JSON `slog`，并返回 cleanup 函数。
  - `buildWriter`：决定输出到 stdout、文件或双写。
  - `newFileLogger`：创建 lumberjack 文件日志器。
  - `parseLevel`：解析日志级别。
- `pkg/logger/rotate.go`：按日轮转逻辑。
  - `rotationTicker`、`stdTicker`、`fileRotator`：轮转所需的时钟/轮转抽象。
  - `newStdTicker`：创建标准 ticker。
  - `startDailyRotation`：日期变化后主动执行一次日志轮转。
  - `dayKey`：把时间转成按天比较的 key。
  - `C`、`Stop`：ticker 适配方法。
- `pkg/logger/logger_test.go`：日志包测试。
  - 主要测试函数：`TestNewReturnsJSONLogger`、`TestNewSupportsStdoutOnly`、`TestDailyRotatorCallsRotateAfterDayChange`。
  - `C`、`Stop`、`Rotate`：测试替身中的时钟/轮转实现。

</details>

### 阅读源码时可以按这个顺序进入

1. 先看 `cmd/server/main.go`，理解启动链路从哪里开始。
2. 再看 `internal/app/bootstrap/`，理解配置、日志、数据库、seed admin 是怎么拼起来的。
3. 接着看 `internal/api/register/router.go`，理解四个 surface 如何装配到一起。
4. 然后按业务模块阅读 `internal/modules/auth -> user -> member -> category -> article -> engagement`。
5. 最后再补 `internal/middleware/`、`internal/identity/`、`internal/queue/`、`internal/scheduler/` 和 `pkg/` 基础设施层。
