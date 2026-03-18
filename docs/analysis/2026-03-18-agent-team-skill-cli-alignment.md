# agent-team Skill 与 CLI 对齐排查

日期：2026-03-18

## 背景

当前在派发任务时经常出现下面的错误：

```text
Error: worker 'qa-tester' not found: read worker config
/Users/jsonlee/Projects/agent-team/.worktrees/qa-tester/worker.yaml: open
/Users/jsonlee/Projects/agent-team/.worktrees/qa-tester/worker.yaml: no such file or directory
```

本次排查目标不是修改实现，而是确认：

1. `skills/agent-team/SKILL.md` 与当前 CLI 实现是否一致
2. 当前报错的根因是什么
3. 哪些地方属于文档/协议漂移，哪些地方属于真实能力缺口

## 排查方法

按 `systematic-debugging` 流程执行：

1. 先复现报错
2. 读取相关源码和帮助输出
3. 核对当前仓库中的真实 worker 状态
4. 对比 `skills/agent-team`、`skills/workflow`、README 与 CLI 实现
5. 检查近期相关提交，判断是实现问题还是文档漂移

## 复现与现场证据

### 1. 报错可稳定复现

复现命令：

```bash
agent-team worker assign qa-tester "debug-repro"
```

复现结果：

```text
Error: worker 'qa-tester' not found: read worker config /Users/jsonlee/Projects/agent-team/.worktrees/qa-tester/worker.yaml: open /Users/jsonlee/Projects/agent-team/.worktrees/qa-tester/worker.yaml: no such file or directory
Usage:
  agent-team worker assign <worker-id> "<description>" [provider] [flags]
```

结论：这是确定性问题，不是偶发错误。

### 2. 当前仓库里的真实 worker

当前 `.worktrees/` 下实际存在：

```text
.worktrees/go-backend-001
.worktrees/qa-tester-001
.worktrees/rules-writer-001
```

当前 `worker.yaml` 文件实际存在于：

```text
.worktrees/go-backend-001/worker.yaml
.worktrees/qa-tester-001/worker.yaml
.worktrees/rules-writer-001/worker.yaml
```

`qa-tester-001/worker.yaml` 内容表明：

```yaml
worker_id: qa-tester-001
role: qa-tester
```

结论：仓库里存在的是 `qa-tester-001`，不存在 `qa-tester`。

### 3. CLI 当前行为

`agent-team worker status` 展示的也是带序号 worker：

```text
go-backend-001
qa-tester-001
rules-writer-001
```

结论：当前 CLI 的真实实体是 `worker-id`，不是裸 `role name`。

## 根因分析

### 直接根因

`worker assign`、`worker open` 都是直接按参数拼接路径：

```text
.worktrees/<worker-id>/worker.yaml
```

对应实现位于：

- `cmd/worker_assign.go`
- `cmd/worker_open.go`
- `internal/config.go`
- `internal/role.go`

`worker assign qa-tester ...` 时，CLI 会直接读取：

```text
.worktrees/qa-tester/worker.yaml
```

而不是：

1. 先把 `qa-tester` 识别为 role name
2. 再解析成 `qa-tester-001`
3. 或者自动选择唯一匹配 worker

所以当前错误不是底层读取异常，而是“调用方传入 role name，但 CLI 只接受精确 worker-id”导致的必然失败。

### 更深一层原因

问题不只是一条错误命令，而是多处文档/流程把“角色名”和“worker-id”混用了：

1. `agent-team` skill 正文里虽然定义 worker 标识是 `<role-name>-<3-digit-number>`，但其他相关文档并没有始终坚持这一约束。
2. `workflow` skill 写着“把 role alias 解析成 concrete worker”，但当前 `agent-team workflow` CLI 只提供模板与 run-state 管理，不提供实际 dispatch 执行器。
3. controller 侧如果直接把 role 名或 alias 传给 `worker assign`，就会触发这类错误。

## Skill / 文档 / CLI 对比

### A. 与当前报错直接相关的差异

| 项目 | skill / 文档写法 | 实际 CLI / 源码 | 结论 |
| --- | --- | --- | --- |
| worker 标识 | `skills/agent-team/SKILL.md` 定义 worker 为 `<role-name>-<3-digit-number>` | `internal.NextWorkerID()` 生成 `role-001`；`worker status` 也展示带序号 ID | 这一点本身是一致的 |
| `worker assign` 入参 | 文档使用 `<worker-id>` 占位 | 实际只接受精确 `worker-id`，不会把 `qa-tester` 自动解析到 `qa-tester-001` | 当前报错根因 |
| workflow dispatch 语义 | `skills/workflow/SKILL.md` 要求“resolve role alias to a concrete worker at runtime” | 当前 CLI 没有自动执行这一步的 `workflow run/resume` 派发器，只有 `create/validate/state` | controller 层容易漏做映射 |

### B. `agent-team` skill 与实际 CLI 的明确不一致

| 项目 | `skills/agent-team` 当前写法 | 实际 CLI / 源码 | 影响 |
| --- | --- | --- | --- |
| `worker assign` flags | 写成 `[provider] [--proposal] [--design] [--verify-cmd] [--new-window]` | 实际只有 `[provider]`, `--proposal`, `--design`, `--model`, `--new-window`；没有 `--verify-cmd` | 用户会误以为可在 assign 阶段设置 verify 命令 |
| Task Commands | `task show <id>` / `task done <id>` / `task verify <id>` / `task archive <id>` | 实际分别是 `task show <worker-id> <change-name>`、`task done <worker-id> <change-name> <task-id>`、`task verify <worker-id> <change-name>`、`task archive <worker-id> <change-name>` | 任务协议文档已过期 |
| `task create` | skill 的 Task Commands 段落未列出 | 实际 CLI 存在 `task create <worker-id> "<description>"` | 能力漏写 |
| Worker completion protocol | skill 写 `reply-main "Task completed: ...; verify: passed|failed|skipped"` | `cmd/worker_assign.go` 发给 worker 的运行时提醒，以及仓库规则 `.agents/rules/communication.md` / `.agents/rules/task-protocol.md`，要求走 `commit -> task archive -> reply-main "Task completed: ...; change archived: <change-name>"` | 关键协议不一致 |
| Role creation policy | skill 规定“必须通过 /role-creator skill，不要手写或内联创建 role 文件” | 实际 CLI 明确提供 `agent-team role create <role-name>`，且功能完整 | 这是流程治理约束，不是 CLI 能力限制 |

### C. `skills/agent-team/references/*` 与实际 CLI 的明确不一致

| 文件 | 当前写法 | 实际 CLI / 源码 | 影响 |
| --- | --- | --- | --- |
| `skills/agent-team/references/details.md` | 写 `worker assign` 支持 `--verify-cmd` | 实际不支持 | 误导 controller 侧调用 |
| `skills/agent-team/references/worker-workflow.md` | `task done <change-name> <task-id>` | 实际需要 `task done <worker-id> <change-name> <task-id>` | worker 执行命令会失败 |
| `skills/agent-team/references/worker-workflow.md` | `task verify <change-name>` | 实际需要 `task verify <worker-id> <change-name>` | worker 执行命令会失败 |
| `skills/agent-team/references/worker-workflow.md` | 完成后只要求 `reply-main "...; verify: ..."` | 当前仓库规则要求先 archive，再 reply-main | 与现行 task protocol 冲突 |
| `skills/agent-team/references/details.md` | 完成后使用 verify 状态消息 | 当前规则与 `worker_assign` 运行时提醒改为 archive-based completion | 文档未跟上协议演进 |

### D. `skills/workflow` 与实际 CLI 的能力边界差异

| 项目 | `skills/workflow` 当前表述 | 实际 CLI | 结论 |
| --- | --- | --- | --- |
| “Run Workflow” | 描述为可以运行、恢复、推进 workflow | 当前 `agent-team workflow` 只提供 `create`、`validate`、`state` | 这是“controller 手工编排层”，不是自动执行引擎 |
| role alias -> worker mapping | skill 要求运行时解析 | `internal/workflow.go` 只在 run-state 中保存 `role_worker_map` 结构，没有自动派发实现 | 需要 controller 自己做映射，或后续补 executor |

## 近期变更判断

近期相关提交显示，这更像“实现继续演进，文档没有完全跟上”：

- `4d3ed5a feat(task): implement built-in task system`
- `ba9c1d8 refactor(worker): integrate task system into worker commands`
- `68832d0 feat(worker-assign): append completion protocol reminder to task notification`
- `2332a08 feat(worker): unify create/open commands with explicit --provider flag`
- `9646c51 feat: add workflow orchestration layer with template and run-state management`

观察结果：

1. 代码里的 task 协议、worker lifecycle、completion reminder 已经更新。
2. `skills/agent-team` 及其 references 里仍残留旧签名和旧完成协议。
3. `workflow` skill 写的是 controller 期望行为，但当前 CLI 并没有内建自动派发器去强制完成 role alias -> worker-id 解析。

因此，这次报错更接近“协议层和文档层没有收口”，而不是单一函数的实现 bug。

## 结论

### 结论 1

当前 `qa-tester` 报错的根因很明确：

- 调用方传入了 `role name`
- CLI 实际只接受精确 `worker-id`
- 当前仓库真实 worker 是 `qa-tester-001`
- CLI 没有 role name / alias 到 worker-id 的自动解析逻辑

### 结论 2

`skills/agent-team/SKILL.md` 与实际 CLI 不是完全对齐，至少存在以下实质性偏差：

1. `worker assign` 误写 `--verify-cmd`
2. Task 子命令签名整体过期
3. Worker completion 协议仍写旧版 `verify: ...`，未对齐 archive-based protocol
4. `task create` 漏写

### 结论 3

这类错误“容易反复出现”的原因，不是因为 `worker assign` 本身偶发失败，而是因为上层流程很容易把：

- role name
- workflow actor alias
- worker-id

这三个概念混为一谈，而当前 CLI 并没有提供兜底解析。

## 建议优先级

本次不改代码，只记录建议：

### P0：先统一文档与协议

优先同步这些文件：

- `skills/agent-team/SKILL.md`
- `skills/agent-team/references/details.md`
- `skills/agent-team/references/worker-workflow.md`
- `skills/workflow/SKILL.md`

至少要统一：

1. worker-id 必须是 `role-001` 形式
2. task 子命令的真实参数签名
3. completion protocol 以 archive 为准
4. `worker assign` 不支持 `--verify-cmd`

### P1：补 controller 侧规则

如果继续用 workflow/controller 派发任务，需要明确规定：

1. role alias 只能映射为 role
2. dispatch 前必须显式解析成 worker-id
3. 若一个 role 有多个 worker，必须选择具体 worker-id，不能模糊派发

### P2：后续可考虑的实现增强

如果后续决定做产品级兜底，可以考虑但本次未实施：

1. `worker assign <role-name>` 在唯一匹配 worker 存在时自动解析
2. 找不到 worker 时输出候选建议，例如 “did you mean qa-tester-001?”
3. 在 workflow 执行层补一个真正的 dispatcher / runner，统一处理 role alias -> worker-id 解析

## 本次核查结论摘要

一句话总结：

当前问题的根因不是 `qa-tester` worker 损坏，而是上层把 role 名传给了只接受精确 `worker-id` 的 CLI；同时 `skills/agent-team`、其 references、以及 workflow skill 对 task 协议和 dispatch 责任的描述没有完全跟上当前实现。
