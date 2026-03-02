# Brainstorming: 多 Provider 适配方案（OpenCode / Codex / Gemini CLI）

**日期**: 2026-03-02
**角色**: General Strategist
**状态**: 已批准

---

## 问题陈述与目标

### 当前状态

`agent-team` 已支持 3 个 provider（claude/codex/opencode），并有完整的 Claude Code 插件系统：
- `.claude-plugin/` 包含 plugin.json、marketplace.json（不含 hooks 目录）
- `hooks/hooks.json` 在 repo root，注册了 4 个 hooks（SessionStart、Stop、PreToolUse、PostToolUse）
- 7 个 hook 脚本（shell）在 `scripts/` 中，但其中 3 个（TaskCompleted、SubagentStart、TeammateIdle）未在 hooks.json 注册
- Worker 配置已迁移到 worktree 内的 `worker.yaml`（Go 代码），但部分 shell 脚本仍引用旧路径 `.agents/workers/$ID/config.yaml`
- 现有 hooks 以 shell 脚本实现，调用 `agent-team` Go binary 或直接操作文件系统

### 目标

1. **为 Gemini CLI 创建原生扩展**，可通过 `gemini extensions install` 安装
2. **为 OpenCode 创建 npm 插件**，可通过 `opencode.json` plugin 数组安装
3. **统一 hook 逻辑到 Go CLI**，避免多套 shell 脚本维护
4. **Codex 简化适配** — 仅 skills 安装 + prompt-driven 工作流
5. **兼容新 .tasks/ TDD 任务系统**（替代 OpenSpec）

---

## 约束与假设

- **Gemini 扩展 `gemini-extension.json` 必须在 repo root** — 不支持子目录安装
- **OpenCode 插件支持 npm 包分发** — TypeScript 实现，Bun runtime
- **Codex 无插件/hook 系统** — 只能通过 skills 安装 + prompt 约定
- **现有 Claude Code 插件继续工作** — `.claude-plugin/` 结构保留
- **Go binary `agent-team` 是所有 hook 的统一入口**
- **新 `.tasks/` 任务系统将替代 OpenSpec** — hook 实现需依赖新 task 代码

---

## 各 Provider 能力调研

### Gemini CLI

- **扩展系统**: `gemini-extension.json` 清单 + `gemini extensions install <github-url>`
- **Skills 目录**: `.gemini/skills/` + `.agents/skills/`（跨工具兼容别名）
- **11 种 Hook 事件**: SessionStart、SessionEnd、BeforeTool、AfterTool、BeforeAgent、AfterAgent、BeforeModel、AfterModel、BeforeToolSelection、PreCompress、Notification
- **Hook 协议**: stdin JSON（含 `session_id`、`cwd`、`hook_event_name`、`timestamp`）→ stdout JSON
- **退出码**: 0=成功，2=紧急阻断，其他=非致命警告
- **模板变量**: `${extensionPath}`、`${workspacePath}`、`${/}`
- **环境变量**: `GEMINI_PROJECT_DIR`、`GEMINI_SESSION_ID`、`GEMINI_CWD`、`CLAUDE_PROJECT_DIR`（兼容别名）
- **SKILL.md 格式**: YAML frontmatter（name + description）+ markdown，按需激活加载
- **Context 文件**: `GEMINI.md`（可配置）

### OpenCode

- **插件系统**: TypeScript/JavaScript，npm 包 或 本地 `.opencode/plugins/*.ts`
- **Plugin 接口**: `@opencode-ai/plugin` 类型，异步函数返回 hooks 对象
- **可用 Hook**: `tool.execute.before`、`tool.execute.after`、`event`（25+ 事件类型）、`stop`、`experimental.chat.system.transform`
- **外部二进制调用**: `ctx.$` 提供 Bun shell，可直接调用 Go binary
- **分发**: npm publish，用户在 `opencode.json` 的 `plugin` 数组中声明
- **已知限制**: MCP 工具调用不触发 `tool.execute.before/after` hooks

### Codex

- **无插件/hook 系统**
- **Skills 目录**: `.agents/skills/`、`~/.agents/skills/`、`/etc/codex/skills/`
- **支持 symlinked skill folders**
- **配置**: `~/.codex/config.toml` 的 `[[skills.config]]`

### Claude Code（现有）

- **插件系统**: `.claude-plugin/` 只含 plugin.json + marketplace.json；hooks.json 在 repo root `hooks/hooks.json`
- **Hook 事件**: SessionStart、Stop、PreToolUse、PostToolUse（已注册）；TaskCompleted、SubagentStart、TeammateIdle（脚本存在但未注册）
- **Hook 协议**: stdin JSON（`cwd` 等）→ stderr 日志 + exit code
- **hooks.json 格式**: 嵌套结构 `{ "hooks": { "<Event>": [{ "matcher?": "...", "hooks": [{ "type": "command", "command": "..." }] }] } }`，与 Gemini 格式高度相似

---

## 候选方案与权衡

### 方案 A: 共享 Shell 脚本 + 薄 Provider 适配层

核心 hook 逻辑保持为共享 shell 脚本，每个 provider 适配器做事件映射+配置清单。

**优点**: 与现有脚本一致，迁移成本低
**缺点**: Shell 跨平台问题、各 provider JSON 格式差异需在 shell 中处理

### 方案 B: Go CLI 统一 Hook 处理器 (推荐 ✓)

在 Go CLI 中新增 `agent-team hook <event>` 子命令群，所有 hook 逻辑用 Go 实现。各 provider 适配器只配置事件到 Go binary 的映射。

**优点**: 统一语言、可测试、跨平台、stdin JSON 差异在 Go 内统一解析、与新 .tasks/ 系统天然集成
**缺点**: 初始开发量较大、需确保 binary 在 PATH 中

### 方案 C: 混合 — Go 核心 + Shell 胶水

关键逻辑用 Go 子命令，轻量检查用 shell。

**优点**: 务实折中
**缺点**: 两种语言维护，界限模糊

### 选择理由

方案 B 是最佳选择：
1. 已有 `_inject-role-prompt` Go 子命令作为先例
2. 新 `.tasks/` 系统是 Go 实现，hook 可直接调用内部函数
3. 各 provider JSON 协议差异在 Go 中处理更可靠
4. Go binary 只需一个 PATH 配置即可跨所有 provider 工作
5. 可测试性远优于 shell 脚本

---

## 推荐设计

### 1. 目录结构

```
agent-team/                              # Repo root = Claude 插件 + Gemini 扩展
├── gemini-extension.json                # Gemini 扩展清单（root 必须）
├── GEMINI.md                            # Gemini context 文件
│
├── .claude-plugin/                      # Claude Code 插件元数据（无 hooks 目录）
│   ├── plugin.json
│   └── marketplace.json
│
├── hooks/                               # Claude + Gemini 共用（格式兼容）
│   └── hooks.json                       # 统一 hooks 配置 → agent-team hook <event>
│
├── adapters/
│   └── opencode/                        # OpenCode npm 插件
│       ├── package.json                 # name: "opencode-agent-team"
│       ├── tsconfig.json
│       └── src/
│           └── plugin.ts               # 调用 agent-team hook <event>
│
├── cmd/                                 # Go CLI hook 子命令
│   ├── hook.go                          # hook 父命令
│   ├── hook_session_start.go
│   ├── hook_pre_tool_use.go
│   ├── hook_post_tool_use.go
│   ├── hook_stop.go
│   ├── hook_task_completed.go
│   ├── hook_subagent_start.go
│   └── hook_teammate_idle.go
│
├── internal/
│   └── hook/                            # Hook 核心实现
│       ├── provider.go                  # Provider 检测 + JSON 协议适配
│       ├── session_start.go             # 角色注入逻辑
│       ├── brainstorming_gate.go        # Pre-tool 检查
│       ├── post_edit_check.go           # Post-tool 质量检查
│       ├── stop_guard.go               # 退出警告
│       ├── task_completed.go            # 任务归档
│       ├── subagent_context.go          # 子代理上下文
│       └── teammate_idle.go             # 空闲通知
│
├── scripts/                             # 现有 shell 脚本（迁移完成后删除）
└── settings.json                        # 插件设置
```

### 2. hooks.json: Claude 与 Gemini 共用一份

**关键发现**: Claude Code 和 Gemini CLI 的 hooks.json 格式高度相似，可以共用同一份文件。

Claude 实际格式:
```json
{ "hooks": { "SessionStart": [{ "hooks": [{ "type": "command", "command": "..." }] }] } }
```

Gemini 格式:
```json
{ "enabled": true, "hooks": { "SessionStart": [{ "matcher": "*", "hooks": [{ "name": "...", "type": "command", "command": "...", "timeout": 10000 }] }] } }
```

Gemini 额外的字段（`enabled`、`name`、`timeout`）对 Claude 无影响，Claude 会忽略未知字段。因此 **一份 Gemini 格式的 hooks.json 可以同时被两个 provider 读取**。

注意: 两个 provider 的事件名不同（如 Claude 的 `PreToolUse` vs Gemini 的 `BeforeTool`），因此同一个 hooks.json 中会包含两套事件定义，各 provider 只读取自己识别的事件。

### 3. Go CLI Hook 子命令系统

```
agent-team hook <event> [--provider auto|claude|gemini|opencode] [--root <path>]
```

**统一行为：**
1. 从 stdin 读 JSON
2. 自动检测 provider（通过环境变量，可 --provider 覆盖）
3. 解析为内部统一格式（HookInput）
4. 执行 hook 逻辑（依赖 internal/ 包）
5. 格式化为 provider-specific JSON（HookOutput）
6. 输出到 stdout，日志到 stderr

**Provider 自动检测：**

```go
func DetectProvider() Provider {
    if os.Getenv("CLAUDE_PLUGIN_ROOT") != "" { return ProviderClaude }
    if os.Getenv("GEMINI_PROJECT_DIR") != "" { return ProviderGemini }
    if os.Getenv("OPENCODE_SESSION") != "" { return ProviderOpenCode }
    return ProviderUnknown
}
```

**统一数据结构：**

```go
type HookInput struct {
    CWD        string
    SessionID  string
    ToolName   string            // pre/post tool use
    ToolInput  map[string]any    // pre tool use
    ToolOutput string            // post tool use
    ParentCWD  string            // subagent
    Provider   Provider
    Raw        json.RawMessage   // 原始 JSON
}

type HookOutput struct {
    Decision      string // "allow" / "deny"
    Reason        string
    SystemMessage string
    Context       string // 注入的额外上下文
}
```

### 4. Hook 事件映射表

| 功能 | Claude Code | Gemini CLI | OpenCode | Codex |
|------|------------|------------|----------|-------|
| 角色注入 | `SessionStart` | `SessionStart` | `session.created` event | worker create 预写入 |
| Brainstorming gate | `PreToolUse (Write\|Edit)` | `BeforeTool (write_file\|edit_file)` | `tool.execute.before` | Prompt 约定 |
| Post-edit check | `PostToolUse (Write\|Edit)` | `AfterTool (write_file\|edit_file)` | `tool.execute.after` | Prompt 约定 |
| 退出警告 | `Stop` | `SessionEnd` | `stop` hook | Prompt 约定 |
| 任务归档 | `TaskCompleted` | `AfterAgent` (判断完成) | `session.idle` event | Prompt 约定 |
| 子代理上下文 | `SubagentStart` | `BeforeAgent` (子代理) | 无直接对应 | N/A |
| 空闲通知 | `TeammateIdle` | `Notification` | `session.idle` event | N/A |

### 5. Claude Code 插件（重构）

重构后 `hooks/hooks.json` 同时服务 Claude 和 Gemini，包含两套事件名：

```json
{
  "enabled": true,
  "description": "agent-team hooks: role injection, quality checks, brainstorming gate, task archiving.",
  "hooks": {
    "SessionStart": [{
      "hooks": [{
        "name": "role-inject",
        "type": "command",
        "command": "agent-team hook session-start",
        "timeout": 10000
      }]
    }],
    "Stop": [{
      "hooks": [{
        "name": "stop-guard",
        "type": "command",
        "command": "agent-team hook stop",
        "timeout": 5000
      }]
    }],
    "PreToolUse": [{
      "matcher": "Write|Edit",
      "hooks": [{
        "name": "brainstorming-gate",
        "type": "command",
        "command": "agent-team hook pre-tool-use",
        "timeout": 5000
      }]
    }],
    "PostToolUse": [{
      "matcher": "Write|Edit",
      "hooks": [{
        "name": "post-edit-check",
        "type": "command",
        "command": "agent-team hook post-tool-use",
        "timeout": 15000
      }]
    }],
    "TaskCompleted": [{
      "hooks": [{
        "name": "task-archive",
        "type": "command",
        "command": "agent-team hook task-completed",
        "timeout": 30000
      }]
    }],
    "SubagentStart": [{
      "hooks": [{
        "name": "subagent-context",
        "type": "command",
        "command": "agent-team hook subagent-start",
        "timeout": 5000
      }]
    }],
    "SessionEnd": [{
      "matcher": "*",
      "hooks": [{
        "name": "stop-guard-gemini",
        "type": "command",
        "command": "agent-team hook stop",
        "timeout": 5000
      }]
    }],
    "BeforeTool": [{
      "matcher": "write_file|edit_file",
      "hooks": [{
        "name": "brainstorming-gate-gemini",
        "type": "command",
        "command": "agent-team hook pre-tool-use",
        "timeout": 5000
      }]
    }],
    "AfterTool": [{
      "matcher": "write_file|edit_file",
      "hooks": [{
        "name": "post-edit-check-gemini",
        "type": "command",
        "command": "agent-team hook post-tool-use",
        "timeout": 15000
      }]
    }]
  }
}
```

说明：
- Claude 识别 `SessionStart`、`Stop`、`PreToolUse`、`PostToolUse`、`TaskCompleted`、`SubagentStart`
- Gemini 识别 `SessionStart`（共享）、`SessionEnd`、`BeforeTool`、`AfterTool`
- 两个 provider 各取所需，忽略不识别的事件名

### 6. Gemini CLI 扩展

`gemini-extension.json`（repo root）:

```json
{
  "name": "agent-team",
  "version": "1.0.0",
  "description": "AI team role and worker manager for multi-agent development",
  "contextFileName": "GEMINI.md",
  "settings": [
    {
      "name": "Agent Team Binary Path",
      "description": "Path to agent-team binary (if not in PATH)",
      "envVar": "AGENT_TEAM_BIN"
    }
  ]
}
```

Gemini hooks 已包含在共用的 `hooks/hooks.json` 中（见 Section 5），Gemini 读取其中的 `SessionStart`（共享）、`SessionEnd`、`BeforeTool`、`AfterTool` 事件。

### 7. OpenCode 插件（npm 包）

`adapters/opencode/package.json`:

```json
{
  "name": "opencode-agent-team",
  "version": "1.0.0",
  "type": "module",
  "main": "./dist/plugin.js",
  "dependencies": {
    "@opencode-ai/plugin": "^0.4.45"
  }
}
```

`adapters/opencode/src/plugin.ts`:

```typescript
import type { Plugin } from "@opencode-ai/plugin"

export const AgentTeamPlugin: Plugin = async ({ $, worktree }) => {
  return {
    event: async ({ event }) => {
      if (event.type === "session.created") {
        await $`agent-team hook session-start --provider opencode --root ${worktree}`
      }
      if (event.type === "session.idle") {
        await $`agent-team hook teammate-idle --provider opencode --root ${worktree}`
      }
    },

    "tool.execute.before": async (meta, data) => {
      if (["edit", "Write"].includes(meta.tool)) {
        const input = JSON.stringify({
          cwd: process.cwd(),
          tool_name: meta.tool,
          tool_input: data.args,
        })
        const result = await $`echo ${input} | agent-team hook pre-tool-use --provider opencode`
        // 解析结果判断是否 abort
      }
    },

    "tool.execute.after": async (meta, data) => {
      if (["edit", "Write"].includes(meta.tool)) {
        await $`agent-team hook post-tool-use --provider opencode --root ${worktree}`
      }
    },

    stop: async (input) => {
      await $`agent-team hook stop --provider opencode --root ${worktree}`
    },
  }
}

export default AgentTeamPlugin
```

### 8. Codex 适配（仅 Skills 安装 + Prompt-Driven）

Codex 不做 hook 适配，保持现有 `CopySkillsToWorktree()` 中 `.codex/skills/` 路径逻辑。

**Prompt-Driven Task 工作流（注入角色 system.md / SKILL.md）：**

```markdown
## 任务工作流（必须遵循）

### 开始工作前
1. 读取 `.tasks/changes/` 目录，找到 status=assigned 的 change
2. 阅读 `proposal.md` 了解任务目标
3. 如果 `design.md` 不存在，先创建技术设计文档再开始编码
4. 如果 `tests.md` 存在，先阅读验收标准

### 工作过程中
- 每完成一个 task，执行: `agent-team task done <change-name> <task-id>`
- 提交代码前确保与 design.md 一致

### 完成任务后
1. 确认所有 tasks 标记为 done
2. 执行: `agent-team task verify <change-name>` 触发验收测试
3. 执行: `agent-team reply-main "任务完成: <change-name>"`
```

**Hook 功能的 Prompt 替代映射：**

| Hook 功能 | Hook 实现（Claude/Gemini/OpenCode） | Codex Prompt 替代 |
|-----------|-------------------------------------|-------------------|
| 角色注入 | SessionStart → 自动注入 | worker create 时预写入 AGENTS.md |
| Brainstorming gate | PreToolUse → 自动拦截 | "编码前必须检查 design.md 存在" |
| Post-edit check | PostToolUse → 自动检查 | "每次编辑后运行 quality checks" |
| Stop guard | Stop → 警告 | "退出前确认所有 changes 已归档" |
| Task archive | TaskCompleted → 自动归档 | "完成后执行 agent-team task archive" |

### 9. Hook 与 .tasks/ 系统的依赖关系

Go hook 子命令直接依赖新 `.tasks/` TDD 任务系统的内部函数：

```
internal/hook/
├── session_start.go      → 依赖 internal/role.go (角色查找)
│                           + internal/config.go (WorkerConfig, 含 RolePath 全局角色支持)
├── brainstorming_gate.go → 依赖 internal/task_store.go (ListActiveChanges, ChangeDirPath)
├── post_edit_check.go    → 依赖 internal/task_store.go + role.go (quality_checks)
├── stop_guard.go         → 依赖 internal/task_store.go (ListActiveChanges)
├── task_completed.go     → 依赖 internal/task_lifecycle.go (ApplyChangeTransition → archived)
│                           + internal/task_verify.go (RunVerify)
├── subagent_context.go   → 依赖 internal/role.go (读 CLAUDE.md/GEMINI.md AGENT_TEAM block)
└── teammate_idle.go      → 依赖 internal/session.go (通知主控)
```

**关键: 角色注入需支持全局角色**

`inject_role_prompt.go` 已实现: 从 worktree 内 `worker.yaml` 读取 `WorkerConfig`，若 `RolePath` 非空则使用全局角色路径（`~/.agents/roles/`），否则从项目 `.agents/teams/` 查找。Go hook 实现需保留此逻辑。

```go
// WorkerConfig 关键字段:
type WorkerConfig struct {
    Role      string // 角色名
    RoleScope string // "project" | "global"
    RolePath  string // 全局角色时的绝对路径
    Provider  string // "claude" | "codex" | "opencode" | "gemini"
    // ...
}
```

---

## 已知代码问题（Phase 1 实施前需修复）

以下是当前代码中 shell 脚本与 Go 代码不一致的问题。在 Phase 1 中将 hook 迁移到 Go 后，这些问题将自动解决：

### 1. shell 脚本引用旧路径 config.yaml

- `session-init.sh:26` 和 `post-edit-check.sh:23` 仍查找 `.agents/workers/$ID/config.yaml`
- Go 代码已迁移到 worktree 内 `worker.yaml`（`internal/config.go:28`）
- **影响**: shell 脚本在新创建的 worker 中可能找不到 config 文件
- **修复**: Go hook 实现将直接调用 `LoadWorkerConfig(WorkerYAMLPath(wtPath))`

### 2. task-archive.sh 调用签名不匹配

- `cmd/task_archive.go` 要求 `cobra.ExactArgs(2)`: `<worker-id> <change-name>`
- `task-archive.sh:47` 只传 1 个参数: `agent-team task archive "$ACTIVE" --dir "$CWD"`
- **影响**: 脚本运行时会报参数数量错误
- **修复**: Go hook 实现将直接调用 `internal` 函数，不走 CLI 子命令

### 3. 三个 hooks 未在 hooks.json 注册

- `TaskCompleted` → `task-archive.sh`
- `SubagentStart` → `subagent-context.sh`
- `TeammateIdle` → `teammate-idle-notify.sh`
- **影响**: 这三个 hook 实际上不会被触发
- **修复**: 新的统一 hooks.json 将注册所有事件

---

## 数据流

### Hook 调用流程（以 SessionStart 为例）

```
┌─────────────┐     stdin JSON      ┌──────────────────┐     角色查找      ┌────────────────┐
│ Claude Code  │ ──────────────────→ │                  │ ──────────────→ │ .agents/teams/  │
│ Gemini CLI   │   {cwd, session_id} │  agent-team hook │                  │ ~/.agents/roles │
│ OpenCode     │                     │  session-start   │ ←────────────── │ worker.yaml     │
└─────────────┘                     │                  │   role + prompt  └────────────────┘
                                    │  1. DetectProvider│
                                    │  2. ParseInput    │     stdout JSON
                                    │  3. 查找 worktree │ ──────────────→ Provider-specific
                                    │  4. 读取 role     │                  response
                                    │  5. 注入 prompt   │
                                    │  6. FormatOutput  │
                                    └──────────────────┘
```

### 各 Provider 的 stdin/stdout 差异

**SessionStart 输入：**

| Provider | stdin JSON |
|----------|-----------|
| Claude | `{"cwd": "/path/to/worktree"}` |
| Gemini | `{"cwd": "/path", "session_id": "abc", "hook_event_name": "SessionStart", "source": "startup"}` |
| OpenCode | 无 stdin（TypeScript 直接传参调用 Go binary） |

**Pre-tool-use 输出（deny 场景）：**

| Provider | stdout JSON |
|----------|------------|
| Claude | `{"error": "..."}` + exit 2 |
| Gemini | `{"decision": "deny", "reason": "..."}` |
| OpenCode | Go binary exit 1，TypeScript 侧 `data.abort = "reason"` |

---

## 错误处理

| 场景 | 处理方式 |
|------|---------|
| `agent-team` binary 不在 PATH | Hook 优雅退出（exit 0），stderr 输出安装提示 |
| 非 worktree 环境执行 hook | 静默退出（exit 0），不影响正常使用 |
| worker.yaml / role 文件缺失 | stderr 警告，hook 返回空结果（不阻断） |
| .tasks/ 目录不存在 | task 相关 hook 跳过检查 |
| Go binary 版本不匹配 | hook output 中 `systemMessage` 提示升级 |
| Hook 超时 | Gemini 有 `timeout` 字段控制；Claude/OpenCode 由 provider 自行处理 |

---

## 测试策略

| 层级 | 测试内容 | 方式 |
|------|---------|------|
| **单元测试** | `internal/hook/` 各 handler 逻辑 | Go `testing`，mock stdin/stdout |
| **协议测试** | 各 provider JSON 解析/序列化 | 固定 JSON fixtures 对比 |
| **集成测试** | Go CLI hook 子命令端到端 | `echo '<json>' \| agent-team hook <event>` |
| **插件测试** | OpenCode TypeScript 插件 | Bun test / vitest |
| **手工验证** | 各 provider 真实安装 + hook 触发 | 在各 CLI 中实际运行 |

---

## 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| `gemini-extension.json` 在 repo root 与 Claude 插件共存 | 低 | 文件无冲突，Claude 不读此文件 |
| Claude/Gemini 共用 hooks.json 时未知字段兼容性 | 低 | 两者都忽略不识别的事件名和字段；需实测验证 |
| OpenCode `tool.execute.before/after` 不触发 MCP 工具 | 中 | 记录限制，MCP 工具不受 gate 保护 |
| Gemini hook 工具名（`write_file`）与 Claude（`Write`）不同 | 中 | Go hook handler 内做工具名映射 |
| Go binary 需在 PATH 中 | 中 | `gemini-extension.json` settings 允许配置 binary path |
| 现有 shell scripts 有已知 bug（见"已知代码问题"） | 高 | Phase 1 迁移到 Go 后自动修复，迁移前不修 shell |
| Codex 无 hook，prompt 约定不被 AI 严格遵守 | 中 | 强化 SKILL.md 指令，增加检查点提醒 |

---

## 实现优先级

| 阶段 | 内容 | 前置依赖 |
|------|------|---------|
| **Phase 0** | .tasks/ TDD 任务系统开发 | 无（已完成 ✅） |
| **Phase 1** | Go hook 子命令框架 + `session-start`（角色注入）+ 统一 hooks.json | Phase 0 |
| **Phase 2** | Gemini 扩展清单 (`gemini-extension.json`) + GEMINI.md | Phase 1 |
| **Phase 3** | OpenCode npm 插件 (opencode-agent-team) | Phase 1 |
| **Phase 4** | 迁移剩余 5 个 hooks 到 Go（依赖 .tasks/） + 删除旧 shell 脚本 | Phase 0 + Phase 1 |
| **Phase 5** | 更新 hooks.json 为统一格式（Claude+Gemini 共用） | Phase 4 |
| **Phase 6** | Codex prompt 工作流模板 + 测试 + 文档 | Phase 0 |

---

## 开放问题

1. **Gemini SKILL.md frontmatter**: 现有角色 SKILL.md 需要添加 YAML frontmatter 才能被 Gemini 发现。是否统一加上，还是在复制时自动注入？
2. **OpenCode npm 包命名**: `opencode-agent-team` 还是 `@agent-team/opencode`？
3. **Go binary 分发**: 是否需要 Homebrew formula 或其他包管理器支持，以简化安装？
4. **Gemini 扩展发布**: 是否需要提交到 [geminicli.com/extensions](https://geminicli.com/extensions/) marketplace？
5. **Context 文件统一**: CLAUDE.md / GEMINI.md / AGENTS.md 是否应该统一模板，由 Go CLI 生成？
6. **hooks.json 兼容性验证**: Claude Code 是否确实忽略 Gemini 特有字段（`enabled`、`name`、`timeout`）？需实测。
7. **Gemini provider 注册**: 需要在 Go 代码 `SupportedProviders` map 中添加 `"gemini"` provider，包括 launch command 和 skills 目标目录（`.gemini/skills/`）。

---

## 参考文档

- OpenCode Plugins: https://opencode.ai/docs/plugins/
- OpenCode Config: https://opencode.ai/docs/config/
- OpenCode Skills: https://opencode.ai/docs/skills/
- Codex Skills: https://developers.openai.com/codex/skills
- Codex Config Reference: https://developers.openai.com/codex/config-reference
- Gemini CLI Get Started: https://geminicli.com/docs/get-started/
- Gemini CLI Extensions: https://geminicli.com/docs/extensions/
- Gemini CLI Writing Extensions: https://geminicli.com/docs/extensions/writing-extensions/
- Gemini CLI Extension Reference: https://geminicli.com/docs/extensions/reference/
- Gemini CLI Hooks: https://geminicli.com/docs/hooks/
- Gemini CLI Writing Hooks: https://geminicli.com/docs/hooks/writing-hooks/
- Gemini CLI Hooks Reference: https://geminicli.com/docs/hooks/reference/
- Gemini CLI Skills: https://geminicli.com/docs/cli/skills/
- Gemini CLI Configuration: https://google-gemini.github.io/gemini-cli/docs/get-started/configuration.html
- TDD 任务系统 Brainstorming: docs/brainstorming/2026-03-02-task-system-tdd-brainstorming.md
- OpenCode/Codex Skills 扩展方案: docs/opencode-codex-skills-extension.md
