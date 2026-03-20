# Worker 优化方案 Brainstorming

**日期**: 2026-03-19
**角色**: General Strategist
**状态**: 已批准

---

## 问题陈述

当前 agent-team 存在以下问题：

1. **Init 缺少关键规则**：初始化项目时没有生成代码合并规则和 agent-team 命令调用规则，导致 AI 在执行 workflow 时不知道该用哪些 CLI 命令，可能手动操作绕过工具链。
2. **Worker 创建过早**：`worker create` 会立即创建 worktree，但实际分配任务可能在很久之后，导致 worktree 中的代码不是最新的。
3. **合并安全性不足**：Worker 提交时可能包含 agent-team 生成的文件（`CLAUDE.md`, `GEMINI.md`, `AGENTS.md`, `.gitignore`），合并回 main 时会覆盖 controller 的版本。

---

## 目标

- Worker worktree 在任务分配时才创建，确保代码始终基于最新的 main
- 建立完整的代码合并规则，防止生成文件被错误提交/合并
- 建立 agent-team CLI 命令调用规则，确保 workflow 执行时走正确的工具链
- 无活跃任务的 worker 在接受新任务前自动 rebase

---

## 约束与假设

- Worker 可以 commit 到自己的 `team/<worker-id>` 分支
- Worker 不能提交 `CLAUDE.md`, `GEMINI.md`, `AGENTS.md`, `.gitignore`, `worker.yaml`
- Controller 负责合并 worker 分支到 main
- `.gitignore` 增强作为第一道防线，规则作为第二道防线
- `worker merge` 命令不做代码增强（纯规则指导 controller）

---

## 设计方案

### 1. Worker 延迟创建 + Open 适配

#### `worker create` 变更

- 只创建配置文件：`.agents/workers/<worker-id>/worker.yaml`
- **不创建 worktree**、不写 `.gitignore`、不注入 role prompt、不安装 skills
- `worker.yaml` 新增字段：`worktree_created: false`

#### `worker open` 变更

检测 `worktree_created: false` 时，先执行完整的 worktree 创建流程：

1. `git worktree add .worktrees/<worker-id> -b team/<worker-id>`
2. 写 `.gitignore`（含增强内容）
3. 注入 role prompt（`CLAUDE.md`, `GEMINI.md`, `AGENTS.md`）
4. 安装 skills
5. 初始化 `.tasks/` 目录
6. 更新 `worktree_created: true`

然后继续正常 open 流程（spawn pane, launch AI）。

#### `worker assign` 变更

- 已有逻辑：pane 不活跃时自动调用 `worker open`
- 新增：在 open 之前，检测 worker 无活跃任务时，自动执行 `git rebase main`
- 如果 worktree 还没创建（`worktree_created: false`），`open` 中会自动创建（基于 main 最新代码，无需额外 rebase）

### 2. .gitignore 增强

在 `internal/worktree.go` 的 `WriteWorktreeGitignore` 中新增：

```
CLAUDE.md
GEMINI.md
AGENTS.md
```

完整 `.gitignore` 内容：

```
.gitignore
.claude/
.codex/
.gemini/
.opencode/
.tasks/
worker.yaml
CLAUDE.md
GEMINI.md
AGENTS.md
```

### 3. Init 新增规则文件

#### `.agents/rules/merge-workflow.md`

**Worker 提交规则**：
- Worker 可以 commit 到 `team/<worker-id>` 分支
- 禁止提交 `CLAUDE.md`, `GEMINI.md`, `AGENTS.md`, `.gitignore`, `worker.yaml`
- 提交前必须 `git status` 检查，如果这些文件出现在 staged 中必须 `git reset HEAD <file>`

**Controller 合并规则**：
- 使用 `agent-team worker merge <worker-id>` 合并
- 合并后检查并恢复被覆盖的文件：`git checkout HEAD~ -- CLAUDE.md GEMINI.md AGENTS.md .gitignore`
- 冲突时优先保留 main（ours）版本

**Rebase 规则**：
- 指派新任务前，无活跃任务的 worker 必须 rebase main
- 并行 worker 无强关联任务时也应 rebase
- Rebase 命令：在 worktree 中执行 `git rebase main`

#### `.agents/rules/agent-team-commands.md`

**Workflow 节点与 CLI 命令映射**：

| Workflow Node Type | 必须使用的 agent-team 命令 |
|---|---|
| `assign_role_task` | `agent-team worker assign <worker-id> "<task>"` |
| `wait_for_completion` | 等待 worker 调用 `agent-team reply-main` |
| `verify_or_test` | `agent-team task verify <worker-id> <change-name>` |
| `merge` | `agent-team worker merge <worker-id>` |
| `controller_task` | 直接在 controller 中执行 |
| `decision` | `agent-team workflow state confirm --node <id> --outcome <label>` |

**禁止手动替代**：
- 禁止直接用 `git merge` 替代 `agent-team worker merge`
- 禁止直接在 worker worktree 中手动执行任务（绕过 assign）
- 禁止跳过 `workflow state` 更新

**任务完成链**：
- Worker 完成后：commit → `agent-team task archive` → `agent-team reply-main`
- Controller 收到回复后：review → `agent-team worker merge` → 检查合并结果 → 下一任务

**Rebase 触发时机**：
- Controller 合并完一个 worker 后，在给下一个 worker 分配任务前检查是否需要 rebase
- 无活跃任务的 worker：必须 rebase
- 有并行但无强关联的 worker：建议 rebase

#### `index.md` 更新

添加引用：
- `.agents/rules/merge-workflow.md` — 代码合并、提交排除、rebase 时机
- `.agents/rules/agent-team-commands.md` — workflow 执行时 CLI 命令映射

---

## 影响的文件清单

| 文件 | 变更类型 | 说明 |
|---|---|---|
| `internal/init.go` | 修改 | `InitRulesDir` 新增 2 个规则模板；更新 `index.md` 模板 |
| `cmd/worker_create.go` | 修改 | 延迟 worktree 创建，只生成 worker.yaml |
| `internal/role.go` | 修改 | `WorkerConfig` 新增 `worktree_created` 字段 |
| `cmd/worker_open.go` | 修改 | 检测并自动创建 worktree |
| `cmd/worker_assign.go` | 修改 | 无活跃任务时自动 rebase |
| `internal/worktree.go` | 修改 | `.gitignore` 增加 3 个条目 |
| `cmd/worker_delete.go` | 可能修改 | 处理 worktree 不存在的情况 |

---

## 风险与缓解

| 风险 | 缓解措施 |
|---|---|
| Rebase 冲突导致 assign 失败 | 规则指导 controller 手动解决冲突后重试 |
| 旧 worker.yaml 无 `worktree_created` 字段 | 代码中做兼容处理：字段不存在时视为 `true`（已创建） |
| Worker 绕过 `.gitignore` 手动 `git add -f` | 规则中明确禁止，作为最后防线 |

---

## 验证策略

- 单元测试：`worker create` 不创建 worktree 目录
- 单元测试：`worker open` 在 `worktree_created: false` 时创建 worktree
- 单元测试：`worker assign` 无活跃任务时触发 rebase
- 集成测试：完整流程 create → assign → commit → merge → 下一次 assign 带 rebase
- 手动验证：`.gitignore` 排除生成文件生效

---

## 开放问题

1. `worker.yaml` 配置文件的存放位置：当前是 `<worktree>/worker.yaml`，但延迟创建 worktree 后需要提前存放。建议使用 `.agents/workers/<worker-id>/worker.yaml`。
2. 是否需要 `agent-team worker rebase <worker-id>` 独立命令，还是只在 assign 中自动执行。
