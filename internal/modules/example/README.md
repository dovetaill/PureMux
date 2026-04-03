# Example Module Layout

这个目录现在主要承担“对照说明”的作用：告诉你 PureMux 里抽象的模块化分层，已经如何映射到当前真实业务模块。

## 抽象分层如何映射到真实代码

当前仓库里已经有四个真实业务模块：

- `internal/modules/auth`
  - 登录
  - JWT 签发 / 解析
  - 当前用户上下文
- `internal/modules/user`
  - 管理员用户 CRUD
  - 用户分页列表
- `internal/modules/category`
  - 管理员分类 CRUD
  - 分类分页列表
- `internal/modules/article`
  - 文稿 CRUD
  - ownership 校验
  - 发布 / 取消发布

每个模块都延续同一条依赖方向：

- `handler -> service -> repository`

## 当前真实模块中的职责边界

### `handler`

负责：

- Huma 路由注册
- HTTP 输入绑定
- 调用 `service`
- 返回统一 envelope 响应

可以直接参考：

- `internal/modules/auth/handler.go`
- `internal/modules/user/handler.go`
- `internal/modules/category/handler.go`
- `internal/modules/article/handler.go`

### `service`

负责：

- 输入规则校验
- 业务状态流转
- ownership / 角色权限判断
- 调用 repository 完成数据读写

可以直接参考：

- `internal/modules/auth/service.go`
- `internal/modules/user/service.go`
- `internal/modules/category/service.go`
- `internal/modules/article/service.go`

尤其是文稿模块里的 ownership 判断：

- `admin` 可管理所有文稿
- `user` 只能管理自己的文稿

### `repository`

负责：

- GORM 数据访问
- 列表查询
- 单条查询
- 持久化更新

可以直接参考：

- `internal/modules/auth/repository.go`
- `internal/modules/user/repository.go`
- `internal/modules/category/repository.go`
- `internal/modules/article/repository.go`

## 模块之外的公共拼装点

业务模块不是各自单独运行的，它们还依赖这些公共层：

- `internal/middleware/auth.go`
  - 解析 `Authorization: Bearer <token>`
  - 将当前用户写入 context
- `internal/middleware/authorize.go`
  - 提供受保护路由与管理员路由守卫
- `internal/api/register/router.go`
  - 装配 repository / service / handler
  - 注册 auth / user / category / article 路由
- `internal/api/response/response.go`
  - 统一成功 / 失败 / 分页响应结构

## 如果你要新增一个真实业务模块

建议直接复制当前模式，而不是再写一个“教学版假模块”。

例如你要加一个 `comment` 模块，可以按下面顺序：

1. 建 `internal/modules/comment/model.go`
2. 建 `internal/modules/comment/repository.go`
3. 建 `internal/modules/comment/service.go`
4. 建 `internal/modules/comment/handler.go`
5. 在 `internal/api/register/router.go` 装配并注册
6. 如需鉴权，复用 `internal/middleware`
7. 如需统一响应，复用 `internal/api/response`

## 推荐你优先模仿哪个模块

按常见业务复杂度，建议参考顺序：

1. 最简单 CRUD：参考 `internal/modules/category`
2. 带管理员权限：参考 `internal/modules/user`
3. 带 ownership 与状态流转：参考 `internal/modules/article`
4. 带 JWT 与当前用户上下文：参考 `internal/modules/auth`

## 本目录现在的意义

这个 `example` 目录保留的意义不再是“未来再补一个假示例”，而是：

- 作为新模块开发前的说明入口
- 帮助你快速找到现有真实模块对照实现
- 约束后续新增模块继续沿用同一套分层与装配方式
