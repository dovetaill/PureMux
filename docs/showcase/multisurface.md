# Showcase: `showcase/multisurface`

`main` 分支已经收敛为 starter，因此旧的完整示例应用被保留在 `showcase/multisurface` 分支。

## 为什么要拆分

这个仓库此前默认展示的是一个更完整的多面 API 示例。它很适合说明 PureMux 能如何组织角色、模块和路由，但对于刚开始搭项目的人来说，默认分支的信息量偏大。

现在的拆分方式是：

- `main`: 最小 starter，保留共享基础设施和一个 `post module`
- `showcase/multisurface`: 保留更完整的角色化、多模块示例应用

## 这个 showcase 分支保留什么

`showcase/multisurface` 会继续保留：

- 更完整的模块集合
- 面向角色拆分的路由面
- 更重的业务示例叙事
- 旧版 README 中更详细的能力清单
- 默认示例账号和相关启动说明

如果你的团队想先参考一个更完整的 PureMux 业务组合，再决定如何裁剪，先看这个分支会更合适。

## 什么时候该用哪个分支

优先用 `main`，如果你想：

- 从最小骨架开始搭自己的系统
- 自己定义领域模型
- 逐步接入身份、权限和业务模块

优先看 `showcase/multisurface`，如果你想：

- 直接参考一个更完整的组织方式
- 学习多模块如何一起装配
- 对照更丰富的 API 设计和模块边界

## 切换方式

```bash
git switch showcase/multisurface
```

切换后，你会看到保留下来的完整示例文档和更丰富的业务模块布局。
