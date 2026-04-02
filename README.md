# PureMux

PureMux 是一个基于原生 `http.ServeMux`、Huma v2、GORM、Redis 和 `slog` 的模块化单体 API 脚手架。当前仓库已经从阶段一基础设施包扩展到阶段二运行时骨架，具备 `server / worker / scheduler / migrate` 四个入口，并支持按配置切换 `mysql` 或 `postgres` 作为主数据库。

## Stage 2 Status

- Go module：`github.com/dovetaill/PureMux`
- 默认运行时配置路径：`configs/config.yaml`
- 仓库当前提交的是样例配置：`configs/config.example.yaml`
- 当前已具备：
  - `pkg/config`：YAML + env override 强类型配置加载
  - `pkg/logger`：结构化日志与文件轮转
  - `pkg/database`：`mysql` / `postgres` 主库切换 + `redis` bootstrap
  - `cmd/server`：HTTP API skeleton（Huma + middleware）
  - `cmd/worker`：Asynq worker skeleton
  - `cmd/scheduler`：cron 触发 enqueue skeleton
  - `cmd/migrate`：迁移命令 skeleton

## Available Commands

- `go run ./cmd/server -config configs/config.yaml`
  - 启动 HTTP API 入口，暴露 `/healthz`、`/readyz`、`/openapi.json`、`/docs`
- `go run ./cmd/worker -config configs/config.yaml`
  - 启动 Asynq worker，消费已注册的队列任务
- `go run ./cmd/scheduler -config configs/config.yaml`
  - 启动 cron scheduler，按 `scheduler.spec` 定时 enqueue 队列任务
- `go run ./cmd/migrate -config configs/config.yaml`
  - 校验迁移运行所需配置并构建迁移 URL / source skeleton

## Primary Database Support

- 支持的主数据库驱动：
  - `mysql`
  - `postgres`
- 通过 `database.driver` 切换主库驱动
- `redis` 始终作为后台任务与缓存相关运行时依赖，由 `pkg/database.Bootstrap` 一并初始化

## Runtime Roles

- `server`
  - 负责 API 暴露、健康检查、OpenAPI 文档和基础 HTTP middleware
- `worker`
  - 负责消费 Asynq 任务并执行后台作业
- `scheduler`
  - 负责用 `cron` 注册周期任务，并把任务投递到 Asynq 队列
- `migrate`
  - 负责根据当前数据库驱动构造迁移运行配置，约定迁移目录为 `migrations/`

## Basic Startup Flow

1. 读取 `configs/config.yaml`（并叠加环境变量）
2. 初始化 `slog` logger
3. bootstrap 主数据库与 Redis 资源
4. 根据入口命令构建对应 runtime：
   - `server`：注册 Huma router 与 middleware
   - `worker`：启动 Asynq server 与 handlers
   - `scheduler`：注册 cron jobs，并只做 enqueue
   - `migrate`：构造 migrate source / database URL
5. 收到退出信号后按统一 lifecycle 逆序释放资源

## Progress

- 2026-03-18：Stage 1 Core Infra 完成
- 2026-04-02：Stage 2 Modular API Scaffold skeleton 完成
- 当前验证记录：`verification.md`
- 阶段二设计与实施基线文档：
  - `docs/plans/2026-04-02-modular-api-scaffold-design.md`
  - `docs/plans/2026-04-02-modular-api-scaffold-implementation-plan.md`
