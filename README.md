# PureMux

PureMux 现在已经落地了一套可运行的多 surface API 架构：在保留 `server` / `worker` / `scheduler` / `migrate` 脚手架能力的同时，把业务 API 明确拆成 `public`、`member auth`、`member self`、`admin` 四个 surface。

当前架构切片已经具备：

- 后台管理员登录与当前身份读取
- 管理员用户管理
- 管理员分类管理
- 管理员文稿管理与发布 / 取消发布
- 前台公开分类 / 已发布文稿读取
- 前台会员注册 / 登录 / 自己的资料
- 前台会员点赞 / 收藏 / 我的收藏
- 统一 JWT 鉴权、principal 注入、Huma OpenAPI 注册

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
