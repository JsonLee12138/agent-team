# 脑暴文档：单任务 Worker 任务记录与归档模型

- 日期：2026-03-21
- 角色：general strategist
- 范围：将当前基于消息分配的 worker 执行模型，收敛为“任务先行、单 worker 单任务、归档记录 merged SHA”的文本工件驱动模型。

## 1. 问题陈述与目标

当前 worker 任务分配主要依赖会话消息与临时上下文：

- `worker assign` 通过 pane 消息把任务发给 worker
- worker 启动后的主要任务线索来自会话消息，而不是稳定任务工件
- 会话上下文膨胀、compact、pane 丢失或 provider 切换后，任务恢复能力不足

本次目标：

1. 让任务记录从“消息态”改为“文件态”。
2. 明确一个 worker 只绑定一个任务。
3. 让 worker 通过 worktree 本地 `worker.yaml` 稳定读取任务入口。
4. 让任务状态、归档、merge 结果形成可审计闭环。
5. 保留注入式 role/rules/skills 上下文，但不再把任务正文塞入会话消息。

## 2. 约束与假设

### 约束

- 任务目录统一放在 `.agent-team/task/`。
- 归档目录统一放在 `.agent-team/archive/task/`。
- worker 不预创建，必须先有任务，再创建绑定该任务的 worker。
- 去掉中央 `.agents/workers/<worker-id>/worker.yaml` 模式。
- worker 本地配置固定放在 `<worktree>/worker.yaml`。
- 状态更新通过 CLI 命令完成，不允许直接手改 `task.yaml`。
- `task create` 时确定 role。
- `task assign` 可对 `draft` 和 `assigned` 执行，但 `assigned` 只能同角色重绑 worker。
- worker 可因打回产生多个提交，但主线最终通过 squash merge 保留一个交付提交。
- archive 只记录 `merged_sha`。

### 假设

- main 作为控制器，承担任务创建、分配、验收、归档、merge、cleanup 的主流程责任。
- worker 完成任务后自行执行 `task done`，main 只在验收通过后归档。
- 有设计方案时，直接写入 `context.md` 的 `Design` 段，而不是额外引入更多任务文件。

## 3. 候选方案与权衡

### 方案 A：继续使用消息分配

保留当前模式：worker 依赖 pane 消息接收任务，会话消息作为主要任务入口。

- 优点：实现最轻，变更最少
- 缺点：恢复能力差，任务可审计性弱，容易受上下文压缩影响

### 方案 B：任务工件驱动 + 单任务 worker（本次选择）

main 先在 `.agent-team/task/<task-id>/` 创建任务工件，再创建并绑定 worker，worker 通过本地 `worker.yaml` 读取任务入口。

- 优点：恢复稳定、可审计、边界清晰、适合单 worker 单任务治理
- 缺点：需要重建 task 命令面与最小任务生命周期

### 方案 C：恢复旧 task 全生命周期系统

回到旧 task 模型，恢复更完整的 task 状态、校验与展示逻辑。

- 优点：功能更全
- 缺点：过重，会重新引入此前已清理的历史包袱

### 推荐

采用方案 B：**任务先行、文本工件驱动、单 worker 单任务、归档记录 `merged_sha`**。

## 4. 推荐设计

### 4.1 架构

核心模型：

- 一个 worker 只绑定一个任务
- 一个任务只绑定一个 role
- worker 可因打回产生多个提交
- main 验收通过后 squash merge
- archive 只记录 `merged_sha`

目录结构：

- 活跃任务：`.agent-team/task/<task-id>/`
  - `task.yaml`
  - `context.md`
- 归档任务：`.agent-team/archive/task/<task-id>/`
- worker 本地配置：`<worktree>/worker.yaml`

生命周期：

- `draft -> assigned -> done -> archived`
- `task create`：main 先建任务
- `task assign`：自动创建 worker + 绑定任务 + 自动 open 会话
- `task done`：worker 完成后触发
- `task archive`：main merge 后归档

绑定规则：

- `task create` 时确定 `role`
- `task assign` 可用于：
  - `draft`：首次分配
  - `assigned`：同角色重绑 worker
- 禁止跨角色重绑

上下文读取规则：

- 不依赖启动时自动注入任务正文
- worker 启动后必须先读：
  - `worker.yaml`
  - `task.yaml`
  - `context.md`
- `task assign` 只发送轻量提醒消息，提示按 `worker.yaml` 的 `task_path` 读取任务

第一版最小命令面：

- `task create`
- `task list`
- `task show`
- `task assign`
- `task done`
- `task archive`

### 4.2 组件

#### Task Package

位置：
- `.agent-team/task/<task-id>/task.yaml`
- `.agent-team/task/<task-id>/context.md`

职责：
- 作为任务唯一事实来源
- `task.yaml` 存结构化状态
- `context.md` 存任务背景、范围、验收、约束、设计内容

建议 `task.yaml` 最小字段：
- `task_id`
- `title`
- `role`
- `status`
- `worker_id`
- `task_path`
- `created_at`
- `assigned_at`
- `done_at`
- `archived_at`
- `merged_sha`

说明：
- `task_path` 直接写当前任务目录路径
- 有设计时，写入 `context.md` 的 `Design` 段

#### Worker Local Config

位置：
- `<worktree>/worker.yaml`

职责：
- 表示这个 worktree 当前绑定的是哪个任务
- 作为 worker 启动后的第一入口

建议最小字段：
- `worker_id`
- `role`
- `task_id`
- `task_path`
- `status`
- `created_at`

说明：
- 不再依赖中央 `.agents/workers/<worker-id>/worker.yaml`
- worker 不预创建，必须由任务驱动生成

#### CLI Layer

命令面：
- `task create`
- `task list`
- `task show`
- `task assign`
- `task done`
- `task archive`

职责边界：
- `create`：创建任务目录与初始文件
- `list`：列出活跃任务与基础状态
- `show`：查看单任务详情
- `assign`：按任务创建/重绑同角色 worker，并自动 open
- `done`：worker 标记完成
- `archive`：main 在 merge 后归档到 `.agent-team/archive/task/`

#### Git / Delivery Boundary

职责：
- worker 分支允许多提交
- main 历史只保留 squash 后的最终交付
- 归档只记录 `merged_sha`

这样：
- 过程问题可以在 worker 分支追溯
- 主线交付记录保持干净

#### Session Reminder Layer

职责：
- `task assign` 发送轻量消息给 worker
- 内容只提醒：
  - 先读 `worker.yaml`
  - 按 `task_path` 读取 `task.yaml` 和 `context.md`
- 不再把任务正文塞进会话消息

## 5. 数据流

### A. 创建任务

1. main 执行 `task create`
2. 在 `.agent-team/task/<task-id>/` 生成：
   - `task.yaml`
   - `context.md`
3. 初始状态为 `draft`
4. role 在此时确定

### B. 分配任务

1. main 执行 `task assign <task-id>`
2. 系统读取任务：
   - 若状态为 `draft`，创建新 worker
   - 若状态为 `assigned`，允许同角色重绑 worker
3. 创建 worktree
4. 在 worktree 根目录写入 `worker.yaml`
5. open worker 会话
6. 发一条轻量提醒消息：
   - 先读 `worker.yaml`
   - 再按 `task_path` 读取任务文件
7. 任务状态进入 `assigned`

### C. worker 执行任务

1. worker 进入 worktree
2. 先读：
   - `<worktree>/worker.yaml`
   - `.agent-team/task/<task-id>/task.yaml`
   - `.agent-team/task/<task-id>/context.md`
3. 根据任务文件执行开发
4. 若 main 打回，可继续在同一 worker 分支追加提交

### D. worker 完成任务

1. worker 完成代码并提交
2. 执行 `task done <task-id>`
3. 系统将状态从 `assigned` 改为 `done`
4. worker 再执行：
   - `agent-team reply-main "Task completed: ..."`

### E. main 验收与打回

1. main 检查 worker 结果
2. 若不通过：
   - main 告知 worker 修改
   - 任务状态回到 `assigned`
   - worker 继续提交修复
3. 若通过：
   - main 执行 squash merge

### F. 归档任务

1. merge 成功后，main 执行 `task archive <task-id>`
2. 系统将任务目录从：
   - `.agent-team/task/<task-id>/`
   移动到：
   - `.agent-team/archive/task/<task-id>/`
3. 写入归档信息：
   - `status=archived`
   - `archived_at`
   - `merged_sha`
4. cleanup worker

### G. 恢复场景

如果 worker 会话丢失或异常：
1. 对 `assigned` 状态任务重新执行 `task assign`
2. 只能重绑到同角色 worker
3. 新 worker 启动后继续按 `worker.yaml -> task.yaml -> context.md` 恢复

## 6. 错误处理策略

1. `task assign` 对非 `draft/assigned` 任务直接失败。
2. `assigned` 状态重绑时，如果新 worker 角色与任务 role 不一致，直接失败。
3. `task done` 仅允许 `assigned -> done`。
4. 打回时仅允许 `done -> assigned`。
5. `task archive` 仅允许在 merge 成功后执行，并要求能提供有效 `merged_sha`。
6. 任务目录移动失败时，不应写入部分归档状态，避免“状态已归档但目录未迁移”的半完成态。
7. worker 读取任务时，如果 `worker.yaml` 指向的 `task_path` 不存在，应立即报错并停止继续开发。

## 7. 验证与测试策略

### 最小验证目标

1. `task create` 能正确生成任务目录、`task.yaml`、`context.md`
2. `task assign` 能：
   - 创建/重绑同角色 worker
   - 写入 worktree 根目录 `worker.yaml`
   - 自动 open worker 会话
   - 将任务状态更新为 `assigned`
3. worker 能通过 `worker.yaml` 恢复到任务上下文
4. `task done` 能将状态从 `assigned` 更新为 `done`
5. 打回后能回到 `assigned`
6. `task archive` 能在 merge 后把目录迁移到 `.agent-team/archive/task/` 并写入 `merged_sha`

### 建议测试层次

- 单元测试：
  - task id 生成
  - 状态迁移
  - task 文件读写
  - archive 迁移
  - role 一致性校验

- 集成测试：
  - `task create -> assign -> done -> archive`
  - `assigned` 同角色重绑恢复
  - `done -> assigned` 打回重做

- 回归检查：
  - worker 启动后仍正确注入 role/rules/skills
  - task 路径读取不依赖消息正文

## 8. 风险与缓解

### 风险 1：task 与 worker 绑定信息漂移

- 缓解：以 `task.yaml` 和 worktree 本地 `worker.yaml` 为唯一绑定面，禁止中央 worker 配置双写。

### 风险 2：打回后 worker 分支多提交，主线交付口径混乱

- 缓解：明确 worker 分支允许多提交，但 main 只通过 squash merge 交付，archive 仅记录 `merged_sha`。

### 风险 3：恢复路径不清晰

- 缓解：固定恢复入口为 `worker.yaml -> task.yaml -> context.md`，assign 消息只做提醒，不承载正文。

### 风险 4：任务目录结构膨胀

- 缓解：保持双文件模型（`task.yaml + context.md`），有设计时写入 `context.md` 的 `Design` 段，不额外扩展更多任务文件。

## 9. Open Questions

1. `task list` 第一版是否只列活跃任务，还是支持切换查看 archived？
2. `task show` 是否需要同时显示关联 worker 当前 pane/session 信息？
3. `task archive` 是否应同时触发 worker cleanup，还是保持为显式独立步骤？
4. `merged_sha` 的来源是命令参数传入，还是通过 git merge 结果自动解析？

---

结论：本次采用“**任务先行 + 单任务 worker + worktree 本地任务指针 + merge 后归档**”模型，替代当前依赖会话消息的任务分配方式。该方案更符合文本工件优先、上下文可恢复、Git 可审计的治理目标。