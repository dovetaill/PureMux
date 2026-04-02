# PureMux

PureMux 是一个基于原生 `http.ServeMux`、Huma v2、GORM、Redis、Asynq、cron 和 `slog` 的模块化单体 API 脚手架。

它的目标不是直接塞满业务，而是先把一套稳定的运行时骨架搭起来：

- HTTP API 入口：`server`
- 后台任务消费者：`worker`
- 定时任务调度器：`scheduler`
- 数据库迁移入口：`migrate`

当前仓库已经完成阶段二 scaffold：可以跑起来、可以切主库、可以挂 API、可以接异步任务和定时任务，但业务模块仍然需要你按项目需求继续填充。

## 适合什么场景

- 你想做一个 Go 的模块化单体 API
- 你希望 HTTP 层尽量轻，直接用标准库 `http.ServeMux`
- 你需要 OpenAPI 文档和类型化接口定义
- 你有后台任务、定时任务、数据库迁移这些常见运行时需求
- 你想把代码按 `handler / service / repository` 分层，而不是把所有逻辑塞进一个包

## 当前能力边界

当前已经具备：

- 配置加载：YAML + env override
- 结构化日志：`slog`
- 主数据库切换：`mysql` / `postgres`
- Redis bootstrap
- HTTP skeleton：`/healthz`、`/readyz`、`/openapi.json`、`/docs`
- Asynq worker skeleton
- cron scheduler skeleton
- migrate command skeleton

当前还没有做的部分：

- 真实业务模块实现
- 真实 migration SQL
- `cmd/migrate` 对 `golang-migrate` 的完整执行接线
- 基于真实 MySQL / PostgreSQL / Redis 的 smoke test

## 5 分钟快速启动

### 1. 准备依赖

你至少需要：

- Go `1.25+`
- Redis
- MySQL 或 PostgreSQL 二选一

### 2. 准备配置文件

仓库提交的是样例配置，你需要先复制成运行配置：

```bash
cp configs/config.example.yaml configs/config.yaml
```

### 3. 选择主数据库

如果你要用 MySQL：

```yaml
database:
  driver: mysql
```

如果你要用 PostgreSQL：

```yaml
database:
  driver: postgres
```

然后分别补好对应连接信息：

- `database.mysql.*`
- `database.postgres.*`
- `redis.*`

### 4. 启动 API 服务

```bash
go run ./cmd/server -config configs/config.yaml
```

启动后你可以访问：

- `http://127.0.0.1:8080/healthz`
- `http://127.0.0.1:8080/readyz`
- `http://127.0.0.1:8080/openapi.json`
- `http://127.0.0.1:8080/docs`

### 5. 启动后台 worker

```bash
go run ./cmd/worker -config configs/config.yaml
```

这会启动 Asynq worker，消费已经注册的任务类型。

### 6. 启动 scheduler

如果你开启了：

```yaml
scheduler:
  enabled: true
```

就可以运行：

```bash
go run ./cmd/scheduler -config configs/config.yaml
```

`scheduler` 的职责应该是“按 cron 规则 enqueue 任务”，而不是直接写业务逻辑。

### 7. 执行 migrate skeleton

```bash
go run ./cmd/migrate -config configs/config.yaml
```

当前这一步主要做：

- 加载配置
- 按主数据库驱动生成 migrate 所需 URL
- 约定迁移目录为 `migrations/`

它还没有真正接入 migration SQL 执行流程。

## 四个入口分别做什么

### `cmd/server`

负责 HTTP API 入口：

- 初始化共享 runtime
- 注册 Huma router
- 挂载 `/healthz`、`/readyz`、`/openapi.json`、`/docs`
- 统一接 request id / recover / timeout / access log middleware

### `cmd/worker`

负责后台任务消费：

- 初始化共享 runtime
- 使用 Redis 构建 Asynq server
- 注册任务 handler
- 等待信号并优雅退出

### `cmd/scheduler`

负责定时任务：

- 初始化共享 runtime
- 注册 cron jobs
- job 本身只做 enqueue
- 不在 scheduler 里直接执行业务逻辑

### `cmd/migrate`

负责迁移入口骨架：

- 加载配置
- 根据 `database.driver` 构造数据库 URL
- 固定迁移 source 为 `file://migrations`

## 最常改的配置项

你通常只需要先关注这些字段：

```yaml
app:
  host: 0.0.0.0
  port: 8080

database:
  driver: mysql # 或 postgres

redis:
  addr: 127.0.0.1:6379

queue:
  asynq:
    concurrency: 10
    queue_name: default

scheduler:
  enabled: false
  spec: "@every 1m"

docs:
  enabled: true
  openapi_path: /openapi.json
  ui_path: /docs
```

## 推荐开发流程

推荐按下面顺序推进：

1. 先定义业务模块的 `repository` 接口
2. 再写 `service` 组织业务流程
3. 再写 `handler` 暴露 HTTP 接口
4. 如果有异步逻辑，再把它拆成 queue task
5. 如果有周期任务，再用 scheduler 只做 enqueue
6. 最后再补 migration SQL 和真实联调

一句话：

- `handler` 管输入输出
- `service` 管业务编排
- `repository` 管数据访问
- `worker` 管异步消费
- `scheduler` 管定时投递

## 业务 Demo：用户 / 文章 模块怎么写

下面是推荐写法，用来说明 PureMux 里的业务代码应该怎么组织。

注意：这是“推荐模式示例”，不是当前仓库已经完整存在的真实业务实现。

### 目录建议

```text
internal/
  modules/
    user/
      handler.go
      service.go
      repository.go
    article/
      handler.go
      service.go
      repository.go
```

如果你的业务更复杂，也可以拆成：

```text
internal/modules/article/
  handler/
  service/
  repository/
  model/
```

但在模块还不复杂时，优先保持简单。

### `handler / service / repository` 分层职责

#### `handler`

负责：

- 接 HTTP 请求
- 参数绑定和基础校验
- 调用 `service`
- 返回统一响应结构

例如你可以暴露这几个接口：

- `POST /users`
- `POST /articles`
- `GET /articles/{id}`

#### `service`

负责：

- 业务规则
- 事务边界
- 跨资源协调
- 决定什么时候发异步任务

例如“创建文章”通常不是简单写库：

- 检查作者是否存在
- 创建文章记录
- 投递一个 `article:created` 的后台任务

#### `repository`

负责：

- 数据库访问
- 查询与持久化
- 不关心 HTTP、Huma、OpenAPI

不要在 `repository` 里塞：

- HTTP 参数解析
- 业务状态流转
- 调用 worker / scheduler

## 业务 Demo：示意代码

### 1. 用户仓储接口

```go
// internal/modules/user/repository.go
package user

type Repository interface {
    Create(ctx context.Context, input CreateUserInput) (*User, error)
    GetByID(ctx context.Context, id int64) (*User, error)
}
```

### 2. 文章服务编排

```go
// internal/modules/article/service.go
package article

type UserRepository interface {
    GetByID(ctx context.Context, id int64) (*User, error)
}

type ArticleRepository interface {
    Create(ctx context.Context, input CreateArticleInput) (*Article, error)
    GetByID(ctx context.Context, id int64) (*Article, error)
}

type Service struct {
    users    UserRepository
    articles ArticleRepository
    enqueuer queueasynq.Enqueuer
}

func (s *Service) Create(ctx context.Context, input CreateArticleInput) (*Article, error) {
    if _, err := s.users.GetByID(ctx, input.AuthorID); err != nil {
        return nil, err
    }

    article, err := s.articles.Create(ctx, input)
    if err != nil {
        return nil, err
    }

    // 示例：创建文章后投递异步任务
    _, _ = queueasynq.EnqueueTask(
        s.enqueuer,
        "default",
        tasks.Payload{Source: "article-service"},
    )

    return article, nil
}
```

上面这段的重点不是任务名本身，而是职责边界：

- `service` 决定“什么时候要发任务”
- `repository` 不负责发任务
- `handler` 也不直接发任务

### 3. handler 只管输入输出

```go
// internal/modules/article/handler.go
package article

type Handler struct {
    service *Service
}

func (h *Handler) Register(api huma.API) {
    // 在这里注册 POST /articles、GET /articles/{id}
    // 然后调用 h.service.Create / h.service.GetByID
}
```

### 4. 在 router 里统一注册模块

推荐在 `internal/api/register/router.go` 里做模块装配：

```go
// 示意：在 NewRouter 里把 article handler 注册进去
articleRepo := article.NewRepository(rt.Resources.MySQL)
articleSvc := article.NewService(userRepo, articleRepo, asynqClient)
articleHandler := article.NewHandler(articleSvc)
articleHandler.Register(api)
```

也就是说：

- `router` 负责装配依赖
- `handler` 负责注册 API
- `service` 负责业务逻辑

## 异步任务怎么接入业务

当前仓库里的 queue 是 skeleton，但接法应该遵循下面的模式：

### 1. 定义任务类型和 payload

推荐放在：

- `internal/queue/tasks/tasks.go`
- `internal/queue/tasks/payload.go`

未来业务里可以新增类似：

- `article:created`
- `article:publish-due`
- `user:welcome-email`

### 2. 在 service 里 enqueue

在业务成功后投递任务，而不是在 HTTP handler 里直接写 Redis / Asynq 细节。

### 3. 在 worker 里注册 handler

推荐把任务处理器统一注册到：

- `internal/queue/asynq/handlers.go`

worker handler 里只处理任务消费逻辑，不参与 HTTP 层。

## 定时任务怎么接入业务

scheduler 的原则很重要：

- scheduler 只负责“到点了，投递一个任务”
- 真正的业务处理应该交给 worker

例如“每天扫描待发布文章”这个需求，推荐这样拆：

1. scheduler 到点触发
2. scheduler enqueue 一个 `article:publish-due` 任务
3. worker 消费任务并执行真正的发布逻辑

不要把“扫描数据库 + 业务更新 + 通知发送”整套逻辑都写在 scheduler job 里。

## 建议你接下来怎么扩展

如果你准备开始写真实业务，我建议按这个顺序：

1. 先做一个最小 `user` 模块
2. 再做 `article` 模块
3. 再给 `article` 加一个异步任务
4. 最后再加 scheduler 触发文章类任务

这样你能先把：

- API 分层
- 数据访问边界
- queue 接法
- scheduler 接法

一次走通，而不是一上来把整个系统写复杂。

## 相关文档

- 验证记录：`verification.md`
- 阶段二实施计划：`docs/plans/2026-04-02-modular-api-scaffold-implementation-plan.md`
- 示例模块说明：`internal/modules/example/README.md`
