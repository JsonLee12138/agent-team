# 脑暴文档：老模块清理（仅保留治理面）

- 日期：2026-03-20
- 角色：general strategist
- 范围：清理 `req/task/workflow` 老业务逻辑；执行面整体下线；仅保留治理面。

## 1. 问题陈述与目标

当前仓库在治理重构后，存在“新治理链路 + 旧执行链路”并存：

- 新链路：`cmd -> orchestrator -> governance -> modules`
- 旧链路：`req/*`、`task/*`、`workflow create|validate|state` 以及对应 `internal` 执行逻辑

本轮目标：

1. 下线老的 `req/task` 命令族。
2. 删除旧 `workflow` 执行面（template/run-state）命令与逻辑。
3. 仅保留治理面命令：`workflow plan ...`。
4. 为后续“重新设计执行面”腾出干净基线。

## 2. 约束与假设

### 约束

- 本轮不做执行面重建，只做清理与收敛。
- 命令命名不改为 `governance`，继续使用 `workflow`。
- 治理命令形态保留 `workflow plan ...`（不扁平化）。
- 下线命令不做兼容转发（硬切）。

### 假设

- 使用方可以接受 `req/task/workflow state` 的直接不可用。
- 现有治理路径（generate/approve/activate）是当前唯一可用主流程。

## 3. 候选方案与权衡

### 方案A：平滑迁移

先把旧命令改接 orchestrator，再逐步删除旧实现。

- 优点：风险低、回滚简单
- 缺点：并存周期长，清理不彻底

### 方案B：硬切（本次选择）

直接下线 `req/task` 与 `workflow` 执行面，仅保留治理面。

- 优点：清理彻底，边界清晰，后续执行面重设不受历史包袱影响
- 缺点：短期命令断裂，需要一次性完成文档与测试同步

### 方案C：仅列清单不改代码

先冻结设计，下一轮再动手。

- 优点：变更风险最低
- 缺点：本轮无法达到“清理干净”目标

### 推荐

采用方案B（硬切）。

## 4. 推荐设计

### 4.1 目标架构

`cmd (workflow plan)` -> `orchestrator (governance usecases)` -> `governance kernel` -> `modules (最小持久化)`

### 4.2 组件保留/清理清单

#### 必保留

1. `internal/governance/*`
2. `internal/orchestrator/*`（治理用例）
3. `internal/modules/requirement/service.go`（Index 适配）
4. `internal/modules/workflow/service.go`（仅 workflow plan 存取）
5. `cmd/workflow.go` 中 `workflow plan generate|approve|activate`（建议补 `close`）

#### 删除

1. `cmd/req*.go` 全部
2. `cmd/task*.go` 全部
3. `cmd/workflow.go` 中 `create|validate|state` 子命令
4. `internal/modules/task/service.go`
5. `internal/workflow.go`（旧执行引擎）
6. `internal/task*.go`（task 执行逻辑）

#### 收缩改造

- `internal/modules/workflow/service.go` 删除 template/run-state 相关方法，仅保留 plan 存取。
- `cmd/root.go` 移除 `newReqCmd()`、`newTaskCmd()` 注册。
- `cmd/worker_assign.go` 文案移除 `task verify/archive` 提示，改为治理面提示。

## 5. 数据流（仅治理面）

### Generate

`workflow plan generate` -> Gate 校验 -> Advisor 生成 plan -> 保存 -> `proposed`

### Approve

`workflow plan approve` -> Gate 校验 -> owner signoff -> 保存 -> `approved`

### Activate

`workflow plan activate` -> Gate 校验 -> 激活 -> 保存 -> `active`

### Close（建议补齐）

`workflow plan close` -> 关闭 -> 保存 -> `closed`

## 6. 错误处理策略

1. Gate blocker 原样失败返回（不兜底吞错）。
2. owner 非法审批、非法状态迁移直接失败。
3. I/O 错误保留调用上下文前缀。
4. 已下线命令保持 `unknown command`（不做兼容桥接）。
5. CLI 成功输出最小字段（`plan_id/status`），失败输出可识别的 blocker code/message。

## 7. 验证与测试策略

### 保留并通过

- `internal/governance/*_test.go`
- `internal/orchestrator/usecases_integration_test.go`

### 新增/调整

- `cmd/workflow` 治理命令路径测试：`plan generate/approve/activate(/close)`
- 下线断言测试：
  - `agent-team req ...` 不可用
  - `agent-team task ...` 不可用
  - `agent-team workflow state ...` 不可用

### 删除

- 与旧 req/task/workflow 执行面绑定的测试文件。

### 回归门禁

- `make test`
- `make lint`

## 8. 风险与缓解

### 风险1：命令断裂导致使用方困惑

- 缓解：同步更新 `skills/workflow/*` 与命令帮助文案，明确“执行面已下线，后续重设”。

### 风险2：误删治理链路依赖

- 缓解：先做引用扫描与编译检查，再按清单删除；删除后立即全量测试。

### 风险3：worker 指引文案失真

- 缓解：同步修正 `worker assign` 提示，不再提 `task verify/archive`。

## 9. Open Questions

1. 是否本轮一并补齐 `workflow plan close` 命令（推荐：是）？
2. 执行面重新设计时，是否继续挂在 `workflow` 下，还是单独新命令域？
3. 旧执行数据（历史 run-state/change）是否需要迁移脚本，还是视为弃用资产？

---

结论：本轮采用“硬切治理面”策略，保留 `workflow plan`，彻底下线 `req/task` 与旧 `workflow` 执行链路，为后续执行面重构提供干净基线。