# Example Module Layout

这个目录只是阶段二的说明性 stub，用来约定后续业务模块的基本分层方式。

## 推荐拆分

- `handler`
  - 负责 HTTP / task 输入输出绑定、参数校验、响应映射
- `service`
  - 负责业务编排、事务边界、跨资源协调
- `repository`
  - 负责数据库或外部资源访问，不承载业务流程判断

## 依赖方向

- `handler -> service -> repository`
- `handler` 不直接访问数据库
- `repository` 不感知 HTTP / Huma / transport 细节

## 阶段二约束

- 先保持模块边界清晰，再逐步填充真实业务实现
- 后续新增模块时，优先复用 `internal/app/bootstrap` 提供的共享资源与运行时约定
