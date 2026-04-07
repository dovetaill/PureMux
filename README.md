# PureMux

PureMux `main` 是一个干净、可直接改造的 Go API starter：保留 `server` / `worker` / `scheduler` / `migrate` 四个入口、共享基础设施层，以及一个官方示例模块 `post`。

如果你只想尽快起一个后端项目，请先按下面的本地开发流程跑起来；更完整的多业务、多 surface 示例已经移到 `showcase/multisurface` 分支。

## 本地启动

### 1. 启动依赖

```bash
make up
```

这会启动 starter 自带的本地依赖：

- MySQL
- Redis

对应配置已经放在 `configs/config.local.yaml`，默认指向 `docker-compose.yml` 里的服务。

### 2. 同步 starter schema

```bash
make migrate
```

### 3. 启动 API

```bash
make dev
```

启动后默认访问：

- `http://127.0.0.1:8080/healthz`
- `http://127.0.0.1:8080/readyz`
- `http://127.0.0.1:8080/openapi.json`
- `http://127.0.0.1:8080/docs`

### 4. 关闭依赖

```bash
make down
```

## 一键验证

开发过程中常用这些命令：

```bash
make test
make verify
make smoke
```

- `make test`: 运行 `go test ./...`
- `make verify`: 执行 starter 的标准校验脚本
- `make smoke`: 启动本地依赖、执行 migrate、拉起 server，并检查核心端点

## 默认暴露什么

starter 默认只暴露这些入口：

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
