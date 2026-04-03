# Example Module Guide

`main` 分支现在把 `internal/modules/post` 作为最小参考实现。它展示的是 PureMux 推荐的基础模块分层，而不是一个完整业务系统。

如果你想看此前更丰富的业务组合，请切换到 `showcase/multisurface`。那个分支保留了更完整的示例模块集合与更重的 onboarding 叙事。

## `post module` 展示什么

`internal/modules/post` 负责演示这条依赖方向：

- `handler -> service -> repository`

你可以在这个模块里直接看到：

- Huma 路由注册方式
- 请求体和路径参数绑定
- 统一 envelope 响应
- 分页列表返回结构
- service 层输入校验与 slug 生成
- repository 层的 GORM 访问接口形状
- 独立模块测试如何组织

## 需要优先阅读的文件

建议按下面顺序阅读 starter 示例：

1. `internal/modules/post/post_test.go`
2. `internal/modules/post/handler.go`
3. `internal/modules/post/service.go`
4. `internal/modules/post/repository.go`
5. `internal/modules/post/model.go`

这个顺序可以先看行为，再回到实现细节。

## 模块之外的装配点

`post module` 不是孤立存在的，它还接到这些公共层：

- `internal/api/register/router.go`
- `internal/app/bootstrap/schema.go`
- `internal/api/response/response.go`
- `internal/identity/*`
- `internal/middleware/*`

如果你要替换 starter 示例，通常至少要一起改这几处。

## 如何把 `post` 换成你自己的模块

推荐流程：

1. 复制 `internal/modules/post`
2. 改掉 `Post` 模型和输入输出结构
3. 在 `service.go` 写你的业务规则
4. 在 `repository.go` 接到真实表与查询
5. 在 `handler.go` 改成你的 API 路径
6. 回到 `router.go` 和 `schema.go` 完成注册
7. 补上你自己的模块测试

## 什么时候看 showcase 分支

如果你需要参考下面这些更重的场景，请切到 `showcase/multisurface`：

- 多角色访问控制
- 更复杂的业务模块协作
- 登录、资料、自助能力拆面
- 点赞、收藏等互动流转
- 更接近产品化后台的 API 组织方式

starter 分支只保留最小骨架；showcase 分支保留更完整的参考应用。
