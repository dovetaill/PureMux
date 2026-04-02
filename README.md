# PureMux
🚀 PureMux: A minimalist, out-of-the-box enterprise Go 1.22+ API scaffold powered by native http.ServeMux and Huma v2. Zero router dependencies! Features auto-generated OpenAPI 3.1 docs, GORM, Redis, slog, and unified JSON responses. Built for modern, type-safe RESTful APIs.

## Stage 1 Status

- 当前仓库已经初始化 Go module：`github.com/dovetaill/PureMux`
- 默认运行时配置文件路径为 `configs/config.yaml`
- 仓库当前提交的是样例配置 `configs/config.example.yaml`
- 阶段一仅完成 `pkg/config`、`pkg/logger`、`pkg/database` 三个基础包
- 本阶段不包含 `cmd/server/main.go`，HTTP 服务启动将在后续阶段接入

## Progress

- 2026-04-02：Stage 1 Core Infra 已完成并合并到 `main`
- 当前已落地：
  - `pkg/config`：YAML + env override 的强类型配置加载
  - `pkg/logger`：JSON `slog`、`stdout/file/both` 输出、跨日主动轮转
  - `pkg/database`：MySQL / Redis bootstrap、fail-fast 检查、资源关闭
- 当前验证记录：`verification.md`
- 下一阶段 TDD 交接输入：`docs/plans/2026-03-18-stage1-core-infra-tdd-spec.md`
- 设计与实施基线文档：
  - `docs/plans/2026-03-18-stage1-core-infra-design.md`
  - `docs/plans/2026-03-18-stage1-core-infra-implementation-plan.md`
