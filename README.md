# PureMux

PureMux 现在已经不是一个只有 runtime skeleton 的空壳，它已经落地了一套可运行的后台多用户文稿发布系统，并继续保留原有的 worker / scheduler / migrate 脚手架能力。

当前业务首版已经具备：

- 用户名密码登录
- JWT 鉴权与当前用户上下文注入
- 默认管理员 seed
- 管理员用户管理
- 管理员文稿分类管理
- 多用户文稿 CRUD
- 基于角色与 ownership 的权限控制
- 文稿发布 / 取消发布
- OpenAPI 文档与 Huma 路由注册

## 当前能力概览

### 业务能力

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
- `POST /api/v1/articles`
- `GET /api/v1/articles`
- `GET /api/v1/articles/{id}`
- `PATCH /api/v1/articles/{id}`
- `DELETE /api/v1/articles/{id}`
- `POST /api/v1/articles/{id}/publish`
- `POST /api/v1/articles/{id}/unpublish`

### 基础设施能力

- 原生 `http.ServeMux`
- Huma v2 OpenAPI 输出
- GORM 主数据库支持（`mysql` / `postgres`）
- Redis bootstrap
- Asynq worker skeleton
- cron scheduler skeleton
- migrate command skeleton
- `slog` 结构化日志

## 角色与权限规则

系统当前只有两种角色：

- `admin`
  - 可以管理所有用户
  - 可以管理所有分类
  - 可以查看、修改、删除、发布任意文稿
- `user`
  - 不能访问 `/api/v1/admin/*`
  - 只能查看、修改、删除、发布自己创建的文稿

ownership 规则现在只作用在文稿资源：

- 管理员可操作全部文稿
- 普通用户只能操作自己的文稿
- JWT 合法但用户状态为 `disabled` 时，也会被拒绝访问受保护接口

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

下面用一个“管理员创建用户与分类，普通用户发布文稿”的完整流程说明如何使用当前系统。

### Step 1. 管理员登录

```bash
curl -X POST http://127.0.0.1:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{
    "username": "admin",
    "password": "admin123456"
  }'
```

返回示例：

```json
{
  "code": 0,
  "message": "login success",
  "data": {
    "token": "<jwt>",
    "expires_at": "2026-04-03T10:00:00+08:00",
    "user": {
      "id": 1,
      "username": "admin",
      "role": "admin",
      "status": "active"
    }
  }
}
```

后续请求统一带：

```text
Authorization: Bearer <jwt>
```

### Step 2. 查看当前登录用户

```bash
curl http://127.0.0.1:8080/api/v1/auth/me \
  -H 'Authorization: Bearer <admin-jwt>'
```

### Step 3. 管理员创建普通用户

```bash
curl -X POST http://127.0.0.1:8080/api/v1/admin/users \
  -H 'Authorization: Bearer <admin-jwt>' \
  -H 'Content-Type: application/json' \
  -d '{
    "username": "editor01",
    "password": "editor123456",
    "role": "user",
    "status": "active"
  }'
```

查看用户列表：

```bash
curl 'http://127.0.0.1:8080/api/v1/admin/users?page=1&page_size=20' \
  -H 'Authorization: Bearer <admin-jwt>'
```

### Step 4. 管理员创建文稿分类

```bash
curl -X POST http://127.0.0.1:8080/api/v1/admin/categories \
  -H 'Authorization: Bearer <admin-jwt>' \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "公告",
    "slug": "notice",
    "description": "站内公告分类"
  }'
```

查看分类列表：

```bash
curl 'http://127.0.0.1:8080/api/v1/admin/categories?page=1&page_size=20' \
  -H 'Authorization: Bearer <admin-jwt>'
```

### Step 5. 普通用户登录

```bash
curl -X POST http://127.0.0.1:8080/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{
    "username": "editor01",
    "password": "editor123456"
  }'
```

### Step 6. 普通用户创建自己的文稿

```bash
curl -X POST http://127.0.0.1:8080/api/v1/articles \
  -H 'Authorization: Bearer <user-jwt>' \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "首篇站点公告",
    "summary": "这是一个简短摘要",
    "content": "这里是正文内容",
    "category_id": 1
  }'
```

查看自己的文稿列表：

```bash
curl 'http://127.0.0.1:8080/api/v1/articles?page=1&page_size=20' \
  -H 'Authorization: Bearer <user-jwt>'
```

普通用户在这个列表里只能看到自己创建的文稿。

### Step 7. 普通用户发布自己的文稿

```bash
curl -X POST http://127.0.0.1:8080/api/v1/articles/1/publish \
  -H 'Authorization: Bearer <user-jwt>'
```

取消发布：

```bash
curl -X POST http://127.0.0.1:8080/api/v1/articles/1/unpublish \
  -H 'Authorization: Bearer <user-jwt>'
```

### Step 8. 管理员管理任意文稿

管理员可以直接查看或修改任意用户的文稿：

```bash
curl http://127.0.0.1:8080/api/v1/articles/1 \
  -H 'Authorization: Bearer <admin-jwt>'
```

```bash
curl -X PATCH http://127.0.0.1:8080/api/v1/articles/1 \
  -H 'Authorization: Bearer <admin-jwt>' \
  -H 'Content-Type: application/json' \
  -d '{
    "title": "管理员修订后的标题"
  }'
```

## 当前模块映射

如果你之前是按 `handler / service / repository` 抽象来理解 PureMux，现在可以直接对照真实业务模块：

- 鉴权模块：`internal/modules/auth`
  - 登录
  - JWT 签发 / 解析
  - 当前用户上下文
- 用户管理模块：`internal/modules/user`
  - 管理员用户 CRUD
  - 用户列表分页
- 分类管理模块：`internal/modules/category`
  - 管理员分类 CRUD
  - 分类列表分页
- 文稿模块：`internal/modules/article`
  - 文稿 CRUD
  - ownership 控制
  - 发布 / 取消发布
- 中间件：`internal/middleware`
  - `Authenticate`
  - `RequireAuthenticated`
  - `RequireAdmin`
- 统一依赖装配：`internal/api/register/router.go`
- 统一响应：`internal/api/response/response.go`

## 业务 Demo：代码应该怎么扩展

当前仓库已经把“示例分层”真正落到了真实模块里。你要继续加新业务时，建议直接参照现有模块，不要重新发明一套目录规范。

### 推荐扩展顺序

1. 在 `internal/modules/<module>` 下定义 `model.go`
2. 在 `repository.go` 写数据库访问
3. 在 `service.go` 写业务规则与权限判断
4. 在 `handler.go` 暴露 Huma 路由
5. 在 `internal/api/register/router.go` 装配依赖并注册路由
6. 如有异步逻辑，再接入 `internal/queue`
7. 如有周期任务，再由 `internal/scheduler` 负责 enqueue

### 可以直接参考的真实实现

- 登录与 Bearer JWT：`internal/modules/auth`
- 管理员 CRUD：`internal/modules/user`、`internal/modules/category`
- ownership 业务判断：`internal/modules/article/service.go`
- 统一分页 envelope：`internal/api/response/response.go`
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
