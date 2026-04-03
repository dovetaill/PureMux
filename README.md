# PureMux

PureMux `main` 现在定位为一个可直接改造的 Go API starter：保留 `server` / `worker` / `scheduler` / `migrate` 四个运行入口、共享基础设施、统一响应约定，以及一个最小 `post` CRUD 示例模块。

如果你想查看之前完整的多 surface 业务示例，请切换到 `showcase/multisurface` 分支。该分支保留了更重的角色划分、业务模块组合和更完整的示例叙事。补充说明见 `docs/showcase/multisurface.md`。

## Starter 提供什么

`main` 分支默认保留这些能力：

- 原生 `http.ServeMux`
- Huma v2 OpenAPI 文档与 `/docs`
- GORM 主数据库接入（`mysql` / `postgres`）
- Redis bootstrap
- `slog` 结构化日志
- worker / scheduler / migrate 命令入口
- 统一 JSON envelope
- `/healthz` 与 `/readyz`
- 可复用的 `internal/identity` / `internal/middleware`
- 一个最小 `post module`，演示 handler / service / repository 分层

## Starter API 面

启动 server 后，`main` 分支默认暴露这些入口：

- `GET /healthz`
- `GET /readyz`
- `GET /openapi.json`
- `GET /docs`
- `GET /api/v1/posts`
- `GET /api/v1/posts/{id}`
- `POST /api/v1/posts`
- `PATCH /api/v1/posts/{id}`
- `DELETE /api/v1/posts/{id}`

这套 API 足够展示 PureMux 的请求绑定、分页响应、服务层校验、仓储分层，以及路由注册方式，而不会把默认分支绑死在某个业务领域。

## 快速启动

### 1. 准备依赖

你至少需要：

- Go `1.25+`
- Redis
- MySQL 或 PostgreSQL

### 2. 准备配置文件

```bash
cp configs/config.example.yaml configs/config.yaml
```

建议至少确认这些配置：

- `database.driver`
- `database.mysql.*` 或 `database.postgres.*`
- `redis.*`
- `auth.jwt.secret`
- `auth.jwt.issuer`

`main` 分支不再默认 seed 管理员账号；身份能力保留为可复用基础设施，由你按自己的业务模型接入。

### 3. 启动 API 服务

```bash
go run ./cmd/server -config configs/config.yaml
```

启动后可访问：

- `http://127.0.0.1:8080/healthz`
- `http://127.0.0.1:8080/readyz`
- `http://127.0.0.1:8080/openapi.json`
- `http://127.0.0.1:8080/docs`

### 4. 启动 worker 与 scheduler

```bash
go run ./cmd/worker -config configs/config.yaml
go run ./cmd/scheduler -config configs/config.yaml
```

### 5. 使用 migrate 命令

```bash
go run ./cmd/migrate -config configs/config.yaml
```

## Demo 模块在哪里

当前 starter 的参考模块位于：

- `internal/modules/post/model.go`
- `internal/modules/post/repository.go`
- `internal/modules/post/service.go`
- `internal/modules/post/handler.go`
- `internal/modules/post/post_test.go`

它配套的装配点在：

- `internal/api/register/router.go`
- `internal/app/bootstrap/schema.go`

如果你想快速理解 PureMux 的推荐结构，先从 `post module` 开始最省时间。

## 如何替换掉 demo 模块

推荐按下面顺序把 starter 改造成你自己的领域：

1. 复制 `internal/modules/post` 作为你的业务模块骨架
2. 调整 `model.go` 定义自己的表结构与字段
3. 在 `service.go` 写输入校验与业务规则
4. 在 `repository.go` 接入你的数据访问逻辑
5. 在 `handler.go` 暴露自己的路由与响应文案
6. 在 `internal/api/register/router.go` 注册你的模块
7. 在 `internal/app/bootstrap/schema.go` 注册你的模型
8. 如需鉴权，复用 `internal/identity` 与 `internal/middleware`

更详细的模块说明见 `internal/modules/example/README.md`。

## 运行时入口说明

- `cmd/server`: HTTP API 服务
- `cmd/worker`: 异步任务 worker 骨架
- `cmd/scheduler`: 定时任务入口
- `cmd/migrate`: 数据库迁移入口

## 统一响应结构

成功响应：

```json
{
  "code": 0,
  "message": "ok",
  "data": {}
}
```

分页响应：

```json
{
  "code": 0,
  "message": "post list",
  "data": {
    "page": 1,
    "page_size": 20,
    "total": 1,
    "items": []
  }
}
```

错误响应：

```json
{
  "code": 404,
  "message": "post not found"
}
```

## Showcase 分支

如果你需要更完整的参考应用，而不是 starter，请查看 `showcase/multisurface`：

- 保留拆分更细的业务模块图
- 保留角色驱动的更复杂 API surface
- 保留更重的 onboarding 文档和业务示例

切换方式：

```bash
git switch showcase/multisurface
```

补充背景见 `docs/showcase/multisurface.md`。
