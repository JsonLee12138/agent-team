# Workflow Optimization — 任务拆分与并发编排

> 基于：2026-03-18-workflow-optimization-brainstorming.md（已批准）
> 日期：2026-03-18
> 状态：待审批

---

## 角色定义

| 角色 ID | 职责 | 技术栈 | 说明 |
|---------|------|--------|------|
| `go-backend` | Go CLI 核心代码开发 | Go 1.24, Cobra, YAML | 负责 internal/ 和 cmd/ 的所有 Go 代码改动 |
| `rules-writer` | 规则文件与 Provider 文件编写 | Markdown | 编写 .agents/rules/*.md、CLAUDE.md、AGENTS.md、GEMINI.md |
| `qa-tester` | 测试编写与集成验证 | Go test, shell | 编写单元测试、集成测试，验收 4 个模块 |

---

## 并发编排总览

```
阶段 1（并发 2 workers）   规则内容 + Go 数据模型
阶段 2（并发 3 workers）   InjectRolePrompt 改造 + Req CLI 命令 + init 拆分
阶段 3（并发 2 workers）   测试 + 集成验证
```

**总计需要 worker 数量：最多 3 个并发**（go-backend ×2 + rules-writer ×1，QA 在阶段 3 复用）

---

## 阶段 1：基础层（并发 2 workers）

> 目标：产出规则文件内容 + Requirement 数据模型，为后续阶段提供基础

### 1A. 规则文件体系创建 — `rules-writer`

| 任务 ID | 任务 | 输出文件 | 细节 |
|---------|------|---------|------|
| 1A-1 | 创建 rules 目录结构 | `.agents/rules/` | 创建目录，添加 `.gitkeep` |
| 1A-2 | 编写 `index.md` | `.agents/rules/index.md` | Level 0 规则索引，≤500 字符；列出所有规则文件及其触发条件 |
| 1A-3 | 编写 `debugging.md` | `.agents/rules/debugging.md` | 调试规则：systematic-debugging 流程、日志检查、错误复现步骤 |
| 1A-4 | 编写 `build-verification.md` | `.agents/rules/build-verification.md` | 构建验证规则：`go build` 前检查、`go vet`/`go test` 要求、提交前校验清单 |
| 1A-5 | 编写 `communication.md` | `.agents/rules/communication.md` | 沟通规则：reply-main 格式、问题上报协议、进度汇报频率 |
| 1A-6 | 编写 `context-management.md` | `.agents/rules/context-management.md` | 上下文管理规则（来自模块 2）：/compact 触发条件（5 条规则）、各 provider 差异处理 |
| 1A-7 | 编写 `task-protocol.md` | `.agents/rules/task-protocol.md` | 任务协议规则：任务完成流程、commit→archive→reply-main 链、Change 状态机说明 |
| 1A-8 | 编写 `worktree.md` | `.agents/rules/worktree.md` | Worktree 规则：禁止 checkout/switch、分支约束、.gitignore 路径说明 |
| 1A-9 | 创建/更新根目录 `CLAUDE.md` | `CLAUDE.md` | Claude 专属指令 + `<!-- agent-team:rules-start -->` 引用 .agents/rules/ |
| 1A-10 | 更新根目录 `AGENTS.md` | `AGENTS.md` | Codex 专属指令 + rules 引用区域 |
| 1A-11 | 更新根目录 `GEMINI.md` | `GEMINI.md` | 保留现有内容 + 追加 rules 引用区域（注意：Gemini 无原生 /compact，需在引用中说明替代方案） |

**验收标准：**
- `index.md` 字符数 ≤ 500
- 每个规则文件结构清晰，使用 MUST/ALWAYS 强指令
- 根目录 provider 文件包含 `<!-- agent-team:rules-start -->...<!-- agent-team:rules-end -->` tag 区域
- context-management.md 包含完整的 5 条 compact 规则

---

### 1B. Requirement 数据模型 — `go-backend` (Worker 1)

| 任务 ID | 任务 | 输出文件 | 细节 |
|---------|------|---------|------|
| 1B-1 | 定义 Requirement + SubTask 数据结构 | `internal/requirement.go` | 新文件：`Requirement` struct（name, description, status, created_at, sub_tasks）；`SubTask` struct（id, title, assigned_to, status, change_name）；`RequirementStatus`（open, in_progress, done）；`SubTaskStatus`（pending, assigned, done, skipped） |
| 1B-2 | 定义 RequirementIndex 数据结构 | `internal/requirement.go` | `RequirementIndex` struct（requirements 列表，每项含 name/status/sub_task_count/done_count）；`RequirementIndexEntry` struct |
| 1B-3 | 实现 Requirement YAML 存储 | `internal/requirement_store.go` | 新文件：路径函数（`RequirementsDir()`→`.tasks/requirements/`、`RequirementDir()`、`RequirementYAMLPath()`、`RequirementIndexPath()`）；`SaveRequirement()`、`LoadRequirement()`、`ListRequirements()` |
| 1B-4 | 实现 RequirementIndex CRUD | `internal/requirement_store.go` | `LoadRequirementIndex()`、`SaveRequirementIndex()`、`RebuildRequirementIndex()`（遍历所有 requirement 重建索引）、`UpdateIndexEntry()`（单条更新） |
| 1B-5 | 实现 SubTask→Requirement 状态回滚 | `internal/requirement_lifecycle.go` | 新文件：`MarkSubTaskDone()`（标记子任务完成，检查是否所有子任务都 done→自动标记 Requirement done）；`ValidateSubTaskTransition()`；`AutoPromoteRequirement()` |
| 1B-6 | 实现 SubTask 分配逻辑 | `internal/requirement_lifecycle.go` | `AssignSubTask()`（设置 assigned_to + change_name，将 SubTask 状态从 pending→assigned，Requirement 状态从 open→in_progress） |

**验收标准：**
- 所有 struct 有完整的 YAML tag
- 存储路径与 brainstorming 文档一致（`.tasks/requirements/<req-name>/requirement.yaml`）
- index.yaml 与 requirement.yaml 数据同步
- 状态机：SubTask done → 检查全部 done → Requirement 自动 done

---

## 阶段 2：功能实现层（并发 3 workers）

> 依赖：阶段 1 全部完成
> 目标：CLI 命令 + InjectRolePrompt 改造 + init 拆分

### 2A. InjectRolePrompt 精简改造 — `go-backend` (Worker 1)

| 任务 ID | 任务 | 输出文件 | 细节 |
|---------|------|---------|------|
| 2A-1 | 重构 `buildRoleSectionFromPath()` | `internal/role.go` | 改造现有函数（L522-555）：不再读取完整 system.md 作为 prompt body；改为生成极简角色身份段（角色名 + 一句话描述）|
| 2A-2 | 新增 `buildRulesIndexSection()` | `internal/role.go` | 读取 `.agents/rules/index.md` 内容，拼接到注入内容中；若 index.md 不存在则跳过（向后兼容） |
| 2A-3 | 新增 `buildSkillIndexSection()` | `internal/role.go` | 改造现有技能注入：只输出技能名称 + 触发条件（从 SKILL.md 提取），不内联完整 SKILL.md 内容 |
| 2A-4 | 重构 `roleSectionTmpl` 模板 | `internal/role.go` | 精简模板（L435-520）：移除内联的 Git Rules、Task Completion Protocol 等大段内容；改为引用 `.agents/rules/worktree.md` 和 `.agents/rules/task-protocol.md` |
| 2A-5 | 更新 tag 名称约定 | `internal/role.go` | 确保 `InjectSection()` 使用 `agent-team:rules-start` / `agent-team:rules-end` tag（检查是否需要同步改 `InjectSection` 函数） |
| 2A-6 | 向后兼容处理 | `internal/role.go` | 若 `.agents/rules/` 不存在，回退到现有的完整注入模式（保证未 init 的项目不受影响） |

**验收标准：**
- 注入内容总长度（无 rules 目录时）不超过现有长度
- 注入内容总长度（有 rules 目录时）≤ 主提示词 500 字符 + 角色身份 + 技能索引
- CLAUDE.md / AGENTS.md / GEMINI.md 三个文件注入行为一致
- 未 init 的项目执行 `worker create` 仍然正常工作

---

### 2B. Requirement CLI 命令 — `go-backend` (Worker 2)

| 任务 ID | 任务 | 输出文件 | 细节 |
|---------|------|---------|------|
| 2B-1 | 创建 `req` 父命令 | `cmd/req.go` | 新文件：`newReqCmd()` 返回 `*cobra.Command`；Short 描述："Manage requirements and sub-tasks" |
| 2B-2 | 实现 `req create` | `cmd/req_create.go` | `agent-team req create <name> --description "..."` ：调用 `internal.CreateRequirement()`；初始化 `.tasks/requirements/<name>/requirement.yaml`；更新 index.yaml |
| 2B-3 | 实现 `req split` | `cmd/req_split.go` | `agent-team req split <name> --task "title1" --task "title2"` ：支持多个 `--task` flag 添加子任务；自动分配递增 SubTask ID；更新 index.yaml 的 sub_task_count |
| 2B-4 | 实现 `req assign` | `cmd/req_assign.go` | `agent-team req assign <name> <task-id> <worker-id>` ：调用 `internal.AssignSubTask()`；验证 worker 存在；创建对应 Change（调用现有 `CreateTaskChange()`）；将 change_name 写回 SubTask |
| 2B-5 | 实现 `req status` | `cmd/req_status.go` | `agent-team req status [name]`：无参数时显示所有 Requirement 概览（从 index.yaml）；有参数时显示单个 Requirement 的所有 SubTask 状态；表格格式输出（ID / Title / Status / AssignedTo / Change） |
| 2B-6 | 实现 `req done` | `cmd/req_done.go` | `agent-team req done <name>`：强制标记 Requirement 为 done；检查是否有未完成的 SubTask，若有则警告但允许强制完成（`--force` flag） |
| 2B-7 | 注册 req 到 root | `cmd/root.go` | 在 `RegisterCommands()` 中添加 `rootCmd.AddCommand(newReqCmd())` |

**验收标准：**
- 5 个子命令全部可用：create, split, assign, status, done
- `req status` 无参数输出格式清晰，包含进度百分比
- `req assign` 自动创建 Change 并关联
- index.yaml 在每次操作后自动更新

---

### 2C. init 职责拆分 — `go-backend` (Worker 1 或 rules-writer)

| 任务 ID | 任务 | 输出文件 | 细节 |
|---------|------|---------|------|
| 2C-1 | 创建 `setup` 命令（迁移原 init 逻辑） | `cmd/setup.go` | 新文件：将 `cmd/init.go` 中的全局安装逻辑（provider 检测、plugin 角色扫描安装、全局目录创建）迁移到 `setup` 命令 |
| 2C-2 | 重写 `init` 命令（新语义） | `cmd/init.go` | 改造：项目初始化专用；创建 `.agents/rules/` 目录；生成默认规则文件（调用模板）；创建/更新根目录 provider 文件（CLAUDE.md / AGENTS.md / GEMINI.md） |
| 2C-3 | 实现规则文件模板生成 | `internal/init.go` | 新增 `InitRulesDir(root)` 函数：创建 `.agents/rules/` 目录；写入默认 index.md + 6 个规则文件（内容可从 embedded 模板或字符串常量）；幂等执行（不覆盖已有文件） |
| 2C-4 | 实现 provider 文件生成/更新 | `internal/init.go` | 新增 `InitProviderFiles(root)` 函数：CLAUDE.md / AGENTS.md / GEMINI.md；若文件不存在则生成完整内容；若已存在则只更新 `<!-- agent-team:rules-start -->...<!-- agent-team:rules-end -->` tag 区域（调用 `InjectSection`） |
| 2C-5 | init 别名兼容 | `cmd/setup.go` | `setup` 命令注册 `init` 作为 Aliases（`Aliases: []string{"init"}` — 不行，init 已被重新定义）；改为：在 `setup` 的 help 中说明它替代了旧的 init；`init` 命令中检测用户可能需要 setup，给出提示 |
| 2C-6 | PersistentPreRunE 更新 | `cmd/root.go` | 将 `setup` 加入跳过 bootstrap 的命令列表（与现有 `init` 同等对待） |
| 2C-7 | 新增 `rules sync` 子命令 | `cmd/init.go` 或 `cmd/rules.go` | `agent-team rules sync`：手动同步 .agents/rules/ 内容到根目录 provider 文件的 tag 区域；可选，用于规则文件更新后手动同步 |

**验收标准：**
- `agent-team setup` 执行原 init 的全局安装功能
- `agent-team init` 执行项目级初始化（创建 rules + provider 文件）
- `init` 幂等：多次执行不覆盖用户自定义内容
- provider 文件只更新 tag 区域，不影响用户手写区域

---

## 阶段 3：测试与验证（并发 2 workers）

> 依赖：阶段 2 全部完成
> 目标：全面测试 + 集成验证

### 3A. 单元测试 — `qa-tester` (Worker 1)

| 任务 ID | 任务 | 输出文件 | 细节 |
|---------|------|---------|------|
| 3A-1 | Requirement 数据模型测试 | `internal/requirement_test.go` | 测试 YAML 序列化/反序列化、状态枚举验证 |
| 3A-2 | Requirement Store 测试 | `internal/requirement_store_test.go` | 测试 CRUD 操作、index.yaml 同步、路径函数、边界情况（空目录、重复创建） |
| 3A-3 | Requirement Lifecycle 测试 | `internal/requirement_lifecycle_test.go` | 测试状态转换：SubTask done→Requirement auto done；部分完成场景；AssignSubTask 联动 |
| 3A-4 | InjectRolePrompt 改造测试 | `internal/role_test.go` | 更新现有测试：验证精简注入模式；测试向后兼容（无 rules 目录）；测试 rules 索引注入内容长度 |
| 3A-5 | init/setup 拆分测试 | `internal/init_test.go` | 测试 `InitRulesDir()` 幂等性；测试 `InitProviderFiles()` tag 区域更新；测试不覆盖用户内容 |

### 3B. CLI 命令测试 + 集成验证 — `qa-tester` (Worker 2)

| 任务 ID | 任务 | 输出文件 | 细节 |
|---------|------|---------|------|
| 3B-1 | req 命令测试 | `cmd/req_*_test.go` | 各子命令的 CLI 层测试：参数解析、错误处理、输出格式 |
| 3B-2 | 端到端流程验证 | `internal/integration_test.go` | 完整流程：`req create` → `req split` → `req assign` → worker `task done` → 验证状态回滚到 Requirement |
| 3B-3 | init + setup 分离验证 | 手动验证 | `agent-team setup` + `agent-team init` 分别执行；验证幂等性 |
| 3B-4 | Worker 注入验证 | 手动验证 | 创建 worker → 检查注入的 CLAUDE.md 内容是否精简；验证 rules/index.md 出现在注入内容中 |
| 3B-5 | 构建验证 | — | `go build` + `go vet` + `go test ./...` 全部通过 |

---

## 依赖关系图

```
阶段 1（基础层）
  ┌─────────────────┐    ┌──────────────────────┐
  │ 1A: rules-writer │    │ 1B: go-backend (W1)  │
  │ 规则文件 + Provider│    │ Requirement 数据模型  │
  └────────┬────────┘    └──────────┬───────────┘
           │                        │
           ▼                        ▼
阶段 2（功能层）── 全部依赖阶段 1 完成 ──
  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
  │ 2A: go-backend│  │ 2B: go-backend│  │ 2C: go-backend│
  │ (W1)         │  │ (W2)         │  │ (W1/writer)  │
  │ Inject 改造   │  │ Req CLI 命令  │  │ init 拆分     │
  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘
         │                 │                  │
         ▼                 ▼                  ▼
阶段 3（测试层）── 全部依赖阶段 2 完成 ──
  ┌──────────────────┐  ┌──────────────────────┐
  │ 3A: qa-tester(W1)│  │ 3B: qa-tester (W2)   │
  │ 单元测试          │  │ CLI 测试 + 集成验证    │
  └──────────────────┘  └──────────────────────┘
```

**阶段内并发细节：**

| 阶段 | 并发 workers | Worker 分配 | 预估改动量 |
|------|-------------|------------|-----------|
| 阶段 1 | 2 | rules-writer ×1, go-backend ×1 | 11 个 md 文件 + 3 个 go 文件 |
| 阶段 2 | 3 | go-backend ×2, go-backend/writer ×1 | 约 10 个 go 文件改动/新增 |
| 阶段 3 | 2 | qa-tester ×2 | 约 6 个 test 文件 |

---

## 任务总计

| 类别 | 任务数 |
|------|-------|
| 阶段 1A（规则内容） | 11 个 |
| 阶段 1B（数据模型） | 6 个 |
| 阶段 2A（Inject 改造） | 6 个 |
| 阶段 2B（Req CLI） | 7 个 |
| 阶段 2C（init 拆分） | 7 个 |
| 阶段 3A（单元测试） | 5 个 |
| 阶段 3B（集成测试） | 5 个 |
| **总计** | **47 个子任务** |

---

## 文件改动清单（按角色分配）

### `rules-writer` 负责（14 个文件）

| 文件 | 操作 | 阶段 |
|------|------|------|
| `.agents/rules/index.md` | 新建 | 1A |
| `.agents/rules/debugging.md` | 新建 | 1A |
| `.agents/rules/build-verification.md` | 新建 | 1A |
| `.agents/rules/communication.md` | 新建 | 1A |
| `.agents/rules/context-management.md` | 新建 | 1A |
| `.agents/rules/task-protocol.md` | 新建 | 1A |
| `.agents/rules/worktree.md` | 新建 | 1A |
| `CLAUDE.md` | 新建 | 1A |
| `AGENTS.md` | 新建 | 1A |
| `GEMINI.md` | 修改 | 1A |

### `go-backend` 负责（约 15 个文件）

| 文件 | 操作 | 阶段 |
|------|------|------|
| `internal/requirement.go` | 新建 | 1B |
| `internal/requirement_store.go` | 新建 | 1B |
| `internal/requirement_lifecycle.go` | 新建 | 1B |
| `internal/role.go` | 修改 | 2A |
| `cmd/req.go` | 新建 | 2B |
| `cmd/req_create.go` | 新建 | 2B |
| `cmd/req_split.go` | 新建 | 2B |
| `cmd/req_assign.go` | 新建 | 2B |
| `cmd/req_status.go` | 新建 | 2B |
| `cmd/req_done.go` | 新建 | 2B |
| `cmd/root.go` | 修改 | 2B |
| `cmd/setup.go` | 新建 | 2C |
| `cmd/init.go` | 修改 | 2C |
| `internal/init.go` | 修改 | 2C |

### `qa-tester` 负责（约 7 个文件）

| 文件 | 操作 | 阶段 |
|------|------|------|
| `internal/requirement_test.go` | 新建 | 3A |
| `internal/requirement_store_test.go` | 新建 | 3A |
| `internal/requirement_lifecycle_test.go` | 新建 | 3A |
| `internal/role_test.go` | 修改 | 3A |
| `internal/init_test.go` | 新建 | 3A |
| `cmd/req_create_test.go` 等 | 新建 | 3B |
| `internal/integration_test.go` | 修改 | 3B |

---

## 开放问题决策建议

| # | 问题 | 建议决策 |
|---|------|---------|
| 1 | `req split` 交互式 vs 自动？ | **先做 `--task` flag 手动拆分**，后续可加 AI 辅助 |
| 2 | 是否需要 `rules sync` 命令？ | **做**，作为 `init` 的补充；worker create 时不自动同步 |
| 3 | Gemini 无 /compact 怎么处理？ | context-management.md 中写明 "Gemini CLI 暂不支持 compact，改用手动总结上文" |

---

## 执行建议

1. **先创建角色**：`go-backend` 和 `rules-writer`（qa-tester 可后续创建）
2. **阶段 1 立即启动**：两个 worker 无依赖可并行
3. **阶段 2 中 2A 和 2C 有文件冲突**（都改 `internal/role.go` 和 `cmd/init.go`），建议 2A + 2C 由同一个 worker 串行完成，2B 由另一个 worker 并行
4. **阶段 3 在合并阶段 2 所有分支后执行**，避免冲突

---

*Generated: 2026-03-18 · Based on approved brainstorming doc*
