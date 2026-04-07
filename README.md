# PureMux

PureMux `main` 是一个干净、可直接改造的 Go API starter：保留 `server` / `worker` / `scheduler` / `migrate` 四个入口、共享基础设施层，以及一个官方示例模块 `post`。

更完整的多业务、多 surface 示例已经移到 `showcase/multisurface` 分支；如果你只是要起一个新后端项目，请先看这里，再按自己的领域替换示例模块。

## Quickstart

### 1. 复制配置

```bash
cp configs/config.example.yaml configs/config.yaml
```

至少确认这些配置：

- `database.driver`
- `database.mysql.*` 或 `database.postgres.*`
- `redis.*`
- `auth.jwt.*`

### 2. 执行 schema sync

```bash
go run ./cmd/migrate -config configs/config.yaml
```

### 3. 启动 API

```bash
go run ./cmd/server -config configs/config.yaml
```

### 4. 可选：启动 worker / scheduler

```bash
go run ./cmd/worker -config configs/config.yaml
go run ./cmd/scheduler -config configs/config.yaml
```

## 默认暴露什么

启动后，starter 默认只暴露这些入口：

- `GET /healthz`
- `GET /readyz`
- `GET /openapi.json`
- `GET /docs`
- `GET /api/v1/posts`
- `GET /api/v1/posts/{id}`
- `POST /api/v1/posts`
- `PATCH /api/v1/posts/{id}`
- `DELETE /api/v1/posts/{id}`

## 官方 demo 模块

`main` 分支只保留一个官方 demo 模块：`internal/modules/post`。

它用最小成本展示了 PureMux 推荐的模块分层：

- `model.go`
- `repository.go`
- `service.go`
- `handler.go`
- `post_test.go`

相关 wiring 入口位于：

- `internal/api/register/router.go`
- `internal/app/bootstrap/schema.go`

如果你要把 starter 改造成自己的项目，先从复制 `post` 模块开始最省时间。更具体的替换步骤见 `internal/modules/example/README.md`。

## 运行时基础能力

`main` 分支保留的是 starter 级共享能力，而不是业务示例集合：

- Huma v2 + OpenAPI
- GORM 主数据库接入（MySQL / PostgreSQL）
- Redis bootstrap
- `slog` 结构化日志
- 统一 JSON envelope
- `/healthz` 与 `/readyz`
- 通用 middleware / identity / bootstrap 约定

## 完整 showcase 在哪里

如果你想参考更丰富的业务叙事、多角色 API surface 和更多模块组合，请切换到 `showcase/multisurface`：

```bash
git switch showcase/multisurface
```

简要说明见 `docs/showcase/multisurface.md`。
