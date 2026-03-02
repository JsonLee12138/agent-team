# Brainstorming: OpenSpec 移除与迁移方案

**日期**: 2026-03-03
**角色**: General Strategist
**前置文档**: [TDD 驱动的任务分配系统设计](./2026-03-02-task-system-tdd-brainstorming.md)
**状态**: ⏳ 部分完成（Phase 1-2 完成，Phase 3-4 待进行）

---

## 问题陈述

OpenSpec 在项目中的渗透远超 Go 代码层面，涵盖三层：Go 代码、Skill/Prompt 指令、Hook Scripts。移除 OpenSpec 需要同步修改所有三层，否则 AI worker 仍会尝试执行已不存在的命令。

---

## 影响面分析

### 第一层: Go 代码（核心逻辑）

| 文件 | 操作 | 改动内容 |
|------|------|---------|
| `internal/openspec.go` | **删除** | 整个文件 |
| `internal/openspec_test.go` | **删除** | 整个文件 |
| `cmd/worker_create.go:15-24` | **替换** | `openSpecSetup` → `.tasks/` 初始化 |
| `cmd/worker_create.go:130-133` | **替换** | OpenSpec init → task init |
| `cmd/worker_assign.go:21` | **修改** | Short description |
| `cmd/worker_assign.go:77-86` | **替换** | `CreateChange()` → task 系统创建 |
| `cmd/worker_assign.go:103-105` | **替换** | 通知消息从 `/opsx:continue` 改为新指令 |
| `cmd/worker_status.go:64` | **替换** | `openspec/changes` → `.tasks/changes` |
| `cmd/reply_main_test.go` | **修改** | 更新测试中的 openspec 引用 |

### 第二层: Skill/Prompt 指令（AI Worker 行为）

#### `skills/agent-team/SKILL.md`

**Worker 创建流程 (L117-119)**:
```diff
- 4. Generates `.gitignore` (excludes .gitignore, .claude/, .codex/, openspec/)
- 5. Creates worktree `.gitignore` (excludes .gitignore, .claude/, .codex/, openspec/)
- 6. Initializes OpenSpec
+ 4. Generates `.gitignore` (excludes .gitignore, .claude/, .codex/, .tasks/)
+ 5. Creates worktree `.gitignore` (excludes .gitignore, .claude/, .codex/, .tasks/)
+ 6. Initializes task system (.tasks/ directory)
```

**Worker Assign 流程 (L139)**:
```diff
- 1. Creates an OpenSpec change at `openspec/changes/<timestamp>-<slug>/`
+ 1. Creates a task change at `.tasks/changes/<timestamp>-<slug>/`
```

**任务完成协议 (L203-237)** — 完整替换:
```diff
- agent-team reply-main "Task completed: <summary>; archive: success via </openspec archive|/prompts:openspec-archive>"
- Archive fallback rule: try `/openspec archive` first; if command is unavailable, fallback to `/prompts:openspec-archive`.
+ When all tasks are done:
+ 1. Run `agent-team task verify <change-name>` to trigger acceptance tests
+ 2. If verify passes, notify main:
+    agent-team reply-main "Task completed: <summary>; verify: passed"
+ 3. If verify fails, fix issues and re-run verify
+ 4. If verify is skipped (no verify config), notify main:
+    agent-team reply-main "Task completed: <summary>; verify: skipped"
```

#### `skills/agent-team/references/details.md`

**Worktree 目录结构 (L23, L34-43)**:
```diff
-   .gitignore                       <- excludes .gitignore, .claude/, .codex/, openspec/
+   .gitignore                       <- excludes .gitignore, .claude/, .codex/, .tasks/
-   openspec/
-     specs/                         <- project specifications
-     changes/                       <- active changes (managed by OpenSpec)
-       <task-slug>/
-         .openspec.yaml             <- change metadata
-         design.md                  <- brainstorming output (from controller)
-         proposal.md                <- work requirements (from controller)
-         specs/                     <- delta specs (created by worker)
-         tasks.md                   <- task breakdown (created by worker)
-     config.yaml                    <- OpenSpec configuration
+   .tasks/
+     config.yaml                    <- task system configuration (verify defaults, lifecycle)
+     changes/                       <- active changes
+       <timestamp>-<slug>/
+         change.yaml                <- metadata + task list + verify config
+         proposal.md                <- work requirements (from controller)
+         design.md                  <- brainstorming output (optional)
+         tests.md                   <- acceptance test definition (TDD, optional)
+     archive/                       <- completed changes
```

**Worker 工作流指令 (L60-68)**:
```diff
- Changes are managed by OpenSpec. The controller creates a change with design and proposal
- via `agent-team worker assign`. The worker then proceeds through the OpenSpec workflow:
-
- 1. `/opsx:continue` — create remaining artifacts (specs, design, tasks)
- 2. `/opsx:apply` — implement the tasks
- 3. `/opsx:verify` — validate implementation
- 4. Attempt `/openspec archive` for the completed change
- 5. If `/openspec archive` is unavailable, fallback to `/prompts:openspec-archive`
- 6. Send `agent-team reply-main` AFTER the archive attempt with archive status
+ Changes are managed by the built-in task system. The controller creates a change via
+ `agent-team worker assign`. The worker then proceeds through the task workflow:
+
+ 1. Read `proposal.md` and `design.md` (if present) to understand the task
+ 2. Write acceptance tests based on requirements (TDD recommended)
+ 3. Implement the tasks, marking each done via `agent-team task done <change> <id>`
+ 4. When all tasks are done, run `agent-team task verify <change>`
+ 5. If verify passes, notify main: `agent-team reply-main "Task completed: <summary>"`
+ 6. If verify fails, fix issues and re-verify
```

**完成通知模板 (L74-81)**:
```diff
- agent-team reply-main "Task completed: <summary>; archive: success via </openspec archive|/prompts:openspec-archive>"
- agent-team reply-main "Task completed: <summary>; archive failed via </openspec archive|/prompts:openspec-archive>: <error>"
+ agent-team reply-main "Task completed: <summary>; verify: passed"
+ agent-team reply-main "Task completed: <summary>; verify: failed — <reason>"
+ agent-team reply-main "Task completed: <summary>; verify: skipped"
```

#### `skills/agent-team/references/brainstorming.md`

```diff
- 1. **Explore project context** — check the role's `system.md`, existing `openspec/specs/`, project files, docs, and recent commits
+ 1. **Explore project context** — check the role's `system.md`, existing `.tasks/changes/`, project files, docs, and recent commits
```

### 第三层: Hook Scripts

#### `scripts/brainstorming-gate.sh`

```diff
- # 2. openspec/ must exist
- [ ! -d "$CWD/openspec" ] && exit 0
-
- # 3. openspec/changes/ must exist
- CHANGES_DIR="$CWD/openspec/changes"
+ # 2. .tasks/ must exist
+ [ ! -d "$CWD/.tasks" ] && exit 0
+
+ # 3. .tasks/changes/ must exist
+ CHANGES_DIR="$CWD/.tasks/changes"
```

#### `scripts/stop-guard.sh`

```diff
- # Warns if there are unarchived openspec changes in the worktree.
+ # Warns if there are incomplete task changes in the worktree.
- CHANGES_DIR="$CWD/openspec/changes"
+ CHANGES_DIR="$CWD/.tasks/changes"
- echo "[stop-guard] Run 'openspec archive --change $name' before stopping." >&2
+ echo "[stop-guard] Run 'agent-team task archive $name' before stopping." >&2
```

#### `scripts/task-archive.sh`

完整重写 — 去掉 `openspec` CLI 依赖:
```diff
- # Archives the active openspec change and notifies main controller.
+ # Archives the active task change and notifies main controller.
- if ! command -v openspec &>/dev/null; then
-     ...
- fi
- CHANGES_DIR="$CWD/openspec/changes"
+ CHANGES_DIR="$CWD/.tasks/changes"
- if openspec archive --change "$ACTIVE" --dir "$CWD" ...
+ if agent-team task archive "$ACTIVE" ...
```

### 第四层: 配置与文档

| 文件 | 改动 |
|------|------|
| `.gitignore` | `openspec/changes` → `.tasks/` |
| `internal/role.go:375` | `.gitignore` 模板: `openspec/` → `.tasks/` |
| `internal/role.go:492-506` | `GenerateSystemPrompt()` 中的完成协议全部替换 |
| `README.md:62` | 去掉 Node.js 依赖说明 |
| `README.zh.md` | 同步更新 |

---

## 迁移策略

### 推荐: 一次性切换（Big Bang）

**理由**: OpenSpec 和新 task 系统的目录结构完全不同（`openspec/` vs `.tasks/`），并存会造成 worker 困惑。

**步骤**:

1. **Phase 1 — Go 核心** (TDD) ✅ **已完成**
   - ✅ 新建 `internal/task.go` + `internal/task_test.go`
   - ✅ 新建 `internal/task_store.go` + `internal/task_store_test.go`
   - ✅ 新建 `internal/task_lifecycle.go` + `internal/task_lifecycle_test.go`
   - ✅ 新建 `internal/task_verify.go` + `internal/task_verify_test.go`
   - ✅ 新建 `cmd/task.go`, `cmd/task_create.go`, `cmd/task_list.go`, `cmd/task_show.go`, `cmd/task_verify.go`, `cmd/task_archive.go`, `cmd/task_done.go`
   - ✅ 所有 75+ 单元测试通过
   - ✅ 编译成功

2. **Phase 2 — 集成替换** ✅ **已完成**
   - ✅ 修改 `cmd/worker_create.go` — 替换 `openSpecSetup` → `taskSetup`，调用 `InitTasksDir()`
   - ✅ 修改 `cmd/worker_assign.go` — 使用 `CreateTaskChange()` 替换 `CreateChange()`，更新通知消息
   - ✅ 修改 `cmd/worker_status.go` — 使用 `CountActiveChanges()` 读取任务数
   - ✅ 删除 `internal/openspec.go` + `internal/openspec_test.go`
   - ✅ 修改 `cmd/root.go` — 注册 `newTaskCmd()`
   - ✅ 修改 `internal/role.go` — 更新 `WriteWorktreeGitignore()` 和 Task Completion Protocol
   - ✅ 修改 `internal/role_test.go` — 同步更新相关断言
   - ✅ 修改 `cmd/reply_main_test.go` — 更新 `openSpecSetup` 引用

3. **Phase 3 — Skill/Prompt 更新** ⏳ **待进行**
   - [ ] 更新 `skills/agent-team/SKILL.md` — 工作流和完成协议
   - [ ] 更新 `skills/agent-team/references/details.md` — 目录结构和工作流
   - [ ] 更新 `skills/agent-team/references/brainstorming.md` — 上下文探索

4. **Phase 4 — Scripts/Config** ⏳ **待进行**
   - [ ] 更新 `scripts/brainstorming-gate.sh`
   - [ ] 更新 `scripts/stop-guard.sh`
   - [ ] 重写 `scripts/task-archive.sh`
   - [ ] 更新 `.gitignore`（如有必要）
   - [ ] 更新 README.md / README.zh.md

---

## 验证清单

**完成 Phase 1-2 后的验证结果**:

- [x] `grep -r "EnsureOpenSpec\|OpenSpecInit\|CreateChange\|GetOpenSpecStatus" --include="*.go" cmd/ internal/` 无结果（已删除 openspec.go）
- [x] `go test ./... -v` 全部通过 (75+ 单元测试)
- [x] `go build` 编译成功
- [x] `agent-team task create` 命令可用
- [x] `agent-team task list/show/verify/archive/done` 命令可用
- [x] `agent-team worker create` 创建的 worktree 包含 `.tasks/config.yaml`
- [x] `agent-team worker assign` 创建 change 在 `.tasks/changes/<timestamp>-<slug>/` 下
- [x] `cmd/worker_status.go` 使用 `CountActiveChanges()` 读取任务数

**Phase 3-4 验证待进行**:

- [ ] `grep -r "openspec" --include="*.md" skills/` 无 openspec 引用
- [ ] Worker skill 中不再引用任何 `/opsx:*` 或 `/openspec` 命令
- [ ] Hook scripts 正确检查 `.tasks/` 目录
- [ ] README 文档已更新
- [ ] `grep -r "openspec" --include="*.sh" scripts/` 无结果

---

## 开放问题

**已解决**:
1. ✅ Go 核心数据模型和命令层已实现，覆盖完整的 TDD 生命周期
2. ✅ 所有相关的 worker/role 命令已更新为使用 task 系统

**进行中**:
1. 需要更新 Skills 中的工作流指令（`.tasks/` 目录、`agent-team task` 命令）
2. 需要更新 Hook Scripts（brainstorming-gate.sh, stop-guard.sh, task-archive.sh）

**待定**:
1. 现有正在运行的 worker（如果有）是否需要重新创建？还是提供一次性的目录迁移命令？
2. `agent-team migrate` 命令是否需要扩展，支持 `openspec/ → .tasks/` 的迁移？

---

## 执行摘要 (2026-03-03)

### Phase 1-2 已完成

**新增的核心功能** (8 个新文件):
- `internal/task.go` — 数据模型（Change, Task, VerifyConfig, TaskConfig）
- `internal/task_store.go` — 文件系统存储（InitTasksDir, CreateTaskChange, LoadChange, ListChanges）
- `internal/task_lifecycle.go` — 状态机（6 种状态，12 种合法转换，自动转换逻辑）
- `internal/task_verify.go` — 验证引擎（支持超时、跳过、默认配置）
- `cmd/task.go` — 父命令
- `cmd/task_create.go`, `task_list.go`, `task_show.go`, `task_verify.go`, `task_archive.go`, `task_done.go` — 6 个子命令

**测试覆盖**: 75+ 单元测试，全部通过

**关键改动**:
- `cmd/worker_create.go`: `openSpecSetup` → `taskSetup`
- `cmd/worker_assign.go`: `CreateChange` → `CreateTaskChange`
- `cmd/worker_status.go`: 目录遍历 → `CountActiveChanges()`
- `internal/role.go`: `.gitignore` 模板和 Task Completion Protocol
- 删除: `internal/openspec.go`, `internal/openspec_test.go`

**下一步** (Phase 3-4):
1. 更新 Skills 中的工作流和完成协议
2. 更新 Hook Scripts 使用 `.tasks/` 目录
3. 更新 README 和相关文档
