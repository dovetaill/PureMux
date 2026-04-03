# Example Module Layout

这个目录现在主要承担“对照说明”的作用：告诉你 PureMux 里的模块化分层，已经如何映射到当前真实业务模块与多 surface API 装配方式。

## 抽象分层如何映射到真实代码

当前仓库里已经有 6 个真实业务模块：

- `internal/modules/auth`
  - 后台管理员登录
  - 后台当前身份读取
- `internal/modules/user`
  - 后台管理员用户 CRUD
  - 用户分页列表
- `internal/modules/member`
  - 前台会员注册 / 登录
  - 自己的资料
- `internal/modules/category`
  - `public_handler.go`：公开分类列表
  - `admin_handler.go`：后台分类管理
- `internal/modules/article`
  - `public_handler.go`：公开文章列表 / slug 详情
  - `admin_handler.go`：后台文稿管理与发布流转
- `internal/modules/engagement`
  - 点赞 / 取消点赞
  - 收藏 / 取消收藏
  - 我的收藏

每个模块都继续沿用同一条依赖方向：

- `handler -> service -> repository`

## 当前真实模块中的职责边界

### `handler`

负责：

- Huma 路由注册
- HTTP 输入绑定
- 调用 `service`
- 返回统一 envelope 响应

可以直接参考：

- `internal/modules/member/public_handler.go`
- `internal/modules/member/self_handler.go`
- `internal/modules/category/public_handler.go`
- `internal/modules/category/admin_handler.go`
- `internal/modules/article/public_handler.go`
- `internal/modules/article/admin_handler.go`
- `internal/modules/engagement/handler.go`

### `service`

负责：

- 输入规则校验
- 业务状态流转
- principal / ownership / member 作用域判断
- 调用 repository 完成数据读写

可以直接参考：

- `internal/modules/auth/service.go`
- `internal/modules/member/service.go`
- `internal/modules/category/service.go`
- `internal/modules/article/service.go`
- `internal/modules/engagement/service.go`

### `repository`

负责：

- GORM 数据访问
- 列表查询
- 单条查询
- 持久化更新

可以直接参考：

- `internal/modules/auth/repository.go`
- `internal/modules/member/repository.go`
- `internal/modules/category/repository.go`
- `internal/modules/article/repository.go`
- `internal/modules/engagement/repository.go`

## 模块之外的公共拼装点

业务模块不是各自单独运行的，它们还依赖这些公共层：

- `internal/identity/*`
  - token / password / principal 公共能力
- `internal/middleware/auth.go`
  - 解析 `Authorization: Bearer <token>`
  - 将当前 principal 写入 context
- `internal/middleware/authorize.go`
  - 提供管理员守卫与 member 守卫
- `internal/api/register/router.go`
  - 按 `public / member auth / member self / admin` groups 装配依赖与注册路由
- `internal/api/response/response.go`
  - 统一成功 / 失败 / 分页响应结构

## 如果你要新增一个真实业务模块

建议先决定 surface，再决定模块边界，而不是直接加一个混合 handler。

例如你要加一个 `comment` 模块，可以按下面顺序：

1. 先判断它属于 `public`、`member auth`、`member self` 还是 `admin`
2. 建 `internal/modules/comment/model.go`
3. 建 `internal/modules/comment/repository.go`
4. 建 `internal/modules/comment/service.go`
5. 如果存在双 contract，拆成 `public_handler.go` / `admin_handler.go`
6. 把路由 ownership 留在模块 handler，不回退到 `router.go` 做路径判断
7. 在 `internal/api/register/router.go` 装配并注册
8. 如需鉴权，复用 `internal/identity` 与 `internal/middleware`
9. 如需统一响应，复用 `internal/api/response`

## 推荐你优先模仿哪个模块

按常见业务复杂度，建议参考顺序：

1. 最简单 public/admin 双面：参考 `internal/modules/category`
2. 带 slug 与内容状态流转：参考 `internal/modules/article`
3. 带 member 身份：参考 `internal/modules/member`
4. 带 member 互动作用域：参考 `internal/modules/engagement`
5. 带后台管理员登录：参考 `internal/modules/auth`

## 本目录现在的意义

这个 `example` 目录保留的意义不再是“未来再补一个假示例”，而是：

- 作为新模块开发前的说明入口
- 帮助你快速找到现有真实模块对照实现
- 约束后续新增模块继续沿用同一套分层与多 surface 装配方式
