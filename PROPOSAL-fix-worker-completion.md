# 改造方案：修复 Worker 任务完成后不执行收尾流程的问题

## 问题描述

Worker 完成任务后没有执行 commit → archive → reply-main 收尾流程，导致：
1. 代码改动留在工作区未提交
2. Change 状态停留在 `draft`，未归档
3. 主控制器没有收到完成通知

## 根因分析

`hooks/hooks.json` 中定义了 `TaskCompleted`、`Stop` 等 hook，但这些 hook **从未被安装到 Worker 的 Claude Code 环境中**。

具体链路断裂点：

1. `worker_create.go` 和 `worker_open.go` 在创建/打开 worker 时，**没有将 hooks 写入 worktree 的 `.claude/settings.json`**
2. Claude Code 读取 hook 配置的位置是 `<project>/.claude/settings.json` 或 `~/.claude/settings.json`
3. Worker 的 worktree 中 `.claude/` 目录只有 skills，没有 settings.json
4. 因此 Claude Code 启动后完全不知道这些 hook 的存在，`Stop` 事件从未触发

此外，`TaskCompleted` 不是 Claude Code 原生支持的 hook 事件（支持的事件：`PreToolUse`、`PostToolUse`、`Stop`、`SubagentStop` 等）。即使安装了 hooks.json，`TaskCompleted` 也不会被触发。

## 改造方案

### 核心思路

1. 用 `//go:embed` 将 `hooks/hooks.json` 嵌入二进制，彻底消除运行时文件查找问题
2. 在 worker create/open 时，将 hooks 写入 worktree 的 `.claude/settings.json`
3. 强化 `Stop` hook：从"仅警告"升级为"自动收尾"（commit → archive → notify）
4. 从 `hooks.json` 中移除无效的 `TaskCompleted` 事件，避免误导

---

### 改动 1：新增 `internal/hooks_install.go` — 嵌入 + 安装

创建新文件，职责单一：将嵌入的 hooks.json 按 provider 过滤后写入 worktree。

```go
// internal/hooks_install.go
package internal

import (
	"embed"
	"encoding/json"
	"os"
	"path/filepath"
)

//go:embed hooks.json
var embeddedHooksJSON []byte

// claudeSettings represents .claude/settings.json structure.
// 使用 json.RawMessage 保留未知字段，避免合并时丢失非 hooks 配置。
type claudeSettings struct {
	Hooks  map[string]json.RawMessage `json:"hooks,omitempty"`
	Extras map[string]json.RawMessage `json:"-"`
}

// Claude Code 支持的 hook 事件白名单（2026-03）
var claudeHookEvents = map[string]bool{
	"PreToolUse":   true,
	"PostToolUse":  true,
	"Stop":         true,
	"SubagentStop": true,
}

// InstallHooksToSettings 将嵌入的 hooks.json 中 Claude 兼容的事件
// 写入 worktree 的 .claude/settings.json。
// 已有的 settings.json 会被合并（hooks 按事件名覆盖，其他字段保留）。
func InstallHooksToSettings(wtPath string) error {
	claudeDir := filepath.Join(wtPath, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return err
	}
	settingsPath := filepath.Join(claudeDir, "settings.json")

	// 解析嵌入的 hooks.json
	var raw struct {
		Hooks map[string]json.RawMessage `json:"hooks"`
	}
	if err := json.Unmarshal(embeddedHooksJSON, &raw); err != nil {
		return err
	}

	// 过滤出 Claude 兼容事件
	filtered := make(map[string]json.RawMessage)
	for event, groups := range raw.Hooks {
		if claudeHookEvents[event] {
			filtered[event] = groups
		}
	}

	// 读取已有 settings.json（如有），合并而非覆盖
	existing := make(map[string]json.RawMessage)
	if data, err := os.ReadFile(settingsPath); err == nil {
		_ = json.Unmarshal(data, &existing) // 解析失败则从空开始
	}

	// 合并 hooks：agent-team 的事件覆盖同名事件，保留其他事件
	existingHooks := make(map[string]json.RawMessage)
	if raw, ok := existing["hooks"]; ok {
		_ = json.Unmarshal(raw, &existingHooks)
	}
	for event, groups := range filtered {
		existingHooks[event] = groups
	}
	hooksBytes, err := json.Marshal(existingHooks)
	if err != nil {
		return err
	}
	existing["hooks"] = json.RawMessage(hooksBytes)

	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, data, 0644)
}
```

**与原方案的差异：**
- `//go:embed` 从"可选"提升为唯一方案，删除了不可靠的 `findHooksJSON()` 文件查找逻辑
- 使用 `json.RawMessage` 做合并，保留 settings.json 中非 hooks 的字段（如 `permissions`、`allowedTools` 等）
- 合并粒度从"整个 hooks 对象覆盖"改为"按事件名合并"，不会误删用户自定义的 hook 事件

**注意：** `//go:embed hooks.json` 要求 `hooks.json` 文件位于 `internal/` 目录下（或使用相对路径 `../hooks/hooks.json`）。由于 Go embed 不支持 `..` 路径，需要在构建前将 `hooks/hooks.json` 复制到 `internal/hooks.json`，或改用以下方式：

```go
// 方案 A：在 main.go 或 cmd/ 层嵌入，通过变量传递
// main.go
//go:embed hooks/hooks.json
var EmbeddedHooksJSON []byte

func main() {
    internal.SetEmbeddedHooksJSON(EmbeddedHooksJSON)
    // ...
}
```

```go
// 方案 B（推荐）：Makefile 构建时复制
// Makefile 中 build target 添加：
//   cp hooks/hooks.json internal/hooks.json
// .gitignore 中添加 internal/hooks.json
```

---

### 改动 2：修改 `cmd/worker_create.go` — 创建时安装 hooks

在步骤 9（InjectRolePrompt，第 133-135 行）之后、步骤 10（SpawnPane，第 138 行）之前插入：

```go
	// 9.5. Install hooks into .claude/settings.json
	if err := internal.InstallHooksToSettings(wtPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: install hooks: %v\n", err)
	}
```

**与原方案的差异：**
- 移除了 `if provider == "claude"` 条件判断。hooks 安装到 `.claude/settings.json` 对非 Claude provider 无副作用（它们不读这个文件），但如果未来其他 provider 也支持该格式，就能自动生效。保持简单。

---

### 改动 3：修改 `cmd/worker_open.go` — 重新打开时也安装 hooks

在 InjectRolePrompt（第 70-72 行）之后、SpawnPane（第 75 行）之前插入：

```go
	// Install/update hooks into .claude/settings.json
	if err := internal.InstallHooksToSettings(wtPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: install hooks: %v\n", err)
	}
```

---

### 改动 4：重写 `cmd/hook_stop.go` — 从"警告"升级为"自动收尾"

这是最关键的改动。复用 `hook_task_completed.go` 中已有的 archive + notify 逻辑，并在前面加上 auto-commit。

```go
// cmd/hook_stop.go
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/JsonLee12138/agent-team/internal"
	"github.com/spf13/cobra"
)

func newHookStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Handle stop/session-end event (auto-commit, archive, notify)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runHookStop(cmd)
		},
	}
}

func runHookStop(cmd *cobra.Command) error {
	input, err := internal.ParseHookInput(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] stop: parse input: %v\n", err)
		return nil // hook 不应中断会话退出
	}

	wt, err := internal.ResolveWorktree(input.CWD)
	if err != nil || wt == nil {
		return nil // 不在 agent-team worktree 中，静默退出
	}

	// 1. Auto-commit 未提交的改动
	autoCommit(wt.WtPath, wt.WorkerID)

	// 2. Archive 所有活跃的 changes
	active, err := internal.ListActiveChanges(wt.WtPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] stop: list changes: %v\n", err)
		return nil
	}
	if len(active) == 0 {
		return nil
	}

	archived := 0
	for _, change := range active {
		if err := internal.ApplyChangeTransition(change, internal.ChangeStatusArchived); err != nil {
			fmt.Fprintf(os.Stderr, "[agent-team] stop: archive '%s' skipped (%s → archived not allowed): %v\n",
				change.Name, change.Status, err)
			continue
		}
		if err := internal.SaveChange(wt.WtPath, change); err != nil {
			fmt.Fprintf(os.Stderr, "[agent-team] stop: save '%s' failed: %v\n", change.Name, err)
			continue
		}
		archived++
		fmt.Fprintf(os.Stderr, "[agent-team] stop: archived change '%s'\n", change.Name)
	}

	// 3. 通知主控制器
	if archived > 0 {
		notifyMain(wt, fmt.Sprintf(
			"Session ended: %d change(s) auto-archived by worker '%s'",
			archived, wt.WorkerID))
	}

	return nil
}

// autoCommit 暂存并提交 worktree 中未提交的改动。
// 仅暂存已跟踪文件的修改（git add -u），不会添加新创建的未跟踪文件。
func autoCommit(wtPath, workerID string) {
	statusOut, err := gitExec(wtPath, "status", "--porcelain")
	if err != nil || len(strings.TrimSpace(statusOut)) == 0 {
		return
	}

	fmt.Fprintf(os.Stderr, "[agent-team] stop: uncommitted changes detected, auto-committing...\n")

	// 暂存已跟踪文件的修改
	if _, err := gitExec(wtPath, "add", "-u"); err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] stop: git add -u failed: %v\n", err)
		return
	}

	// 检查是否有暂存内容（git diff --cached --quiet 返回 0 表示无差异）
	if _, err := gitExec(wtPath, "diff", "--cached", "--quiet"); err == nil {
		return // 无暂存内容
	}

	commitMsg := fmt.Sprintf("auto-commit: worker '%s' session ended with uncommitted changes", workerID)
	if _, err := gitExec(wtPath, "commit", "-m", commitMsg); err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] stop: git commit failed: %v\n", err)
		return
	}

	fmt.Fprintf(os.Stderr, "[agent-team] stop: auto-committed changes\n")
}

// notifyMain 通过 reply-main 命令通知主控制器。
// 直接调用 App.RunReplyMain 需要构造 App 实例，这里用 exec 保持解耦。
func notifyMain(wt *internal.WorktreeInfo, msg string) {
	notifyCmd := exec.Command("agent-team", "reply-main", msg)
	notifyCmd.Dir = wt.WtPath
	notifyCmd.Stderr = os.Stderr
	if err := notifyCmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[agent-team] stop: notify main failed: %v\n", err)
	}
}

// gitExec 在指定目录执行 git 命令，返回 stdout。
func gitExec(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return string(out), err
}
```

**与原方案的差异：**
- 提取了 `gitExec` 和 `notifyMain` 辅助函数，消除重复代码
- auto-commit 放在 archive 之前（原方案也是这个顺序，确认正确）
- archive 失败时打印具体的状态转换信息（`draft → archived not allowed`），便于调试
- 统计实际归档数量，只在有归档时才通知主控制器
- 添加了中文注释说明 `git add -u` 不处理未跟踪文件的设计意图

---

### 改动 5：清理 `hooks/hooks.json` — 移除无效的 `TaskCompleted`

从 hooks.json 中删除 `TaskCompleted` 条目（第 55-66 行）。理由：
- 不是任何已知 provider 的原生事件
- 保留会误导维护者以为它能工作
- `hook_task_completed.go` 的逻辑已被 `hook_stop.go` 完整覆盖

同时保留 `cmd/hook_task_completed.go` 文件但标记为 deprecated，以防未来有 provider 支持该事件。

---

### 改动 6（可选）：`hooks/hooks.json` 中 `Stop` 的 timeout 调整

当前 `Stop` hook 的 timeout 是 5000ms（5 秒）。新的 `hook_stop.go` 需要执行 git add + git commit + archive + reply-main，5 秒可能不够。建议调整为 **30000ms**（30 秒）：

```json
"Stop": [
  {
    "hooks": [
      {
        "name": "stop-guard",
        "type": "command",
        "command": "agent-team hook stop",
        "timeout": 30000
      }
    ]
  }
]
```

同理，`SessionEnd`（Gemini 的等价事件）也应调整为 30000ms。

---

## 改动文件清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `internal/hooks_install.go` | 新增 | embed + 过滤 + 合并写入 settings.json |
| `cmd/worker_create.go` | 修改 | 第 135 行后插入 `InstallHooksToSettings` 调用 |
| `cmd/worker_open.go` | 修改 | 第 72 行后插入 `InstallHooksToSettings` 调用 |
| `cmd/hook_stop.go` | 重写 | 警告 → auto-commit + archive + notify |
| `hooks/hooks.json` | 修改 | 移除 `TaskCompleted`；`Stop`/`SessionEnd` timeout → 30s |
| `Makefile` | 修改 | build target 添加 `cp hooks/hooks.json internal/hooks.json` |
| `.gitignore` | 修改 | 添加 `internal/hooks.json` |

## embed 路径问题的解决方案

Go 的 `//go:embed` 不支持 `..` 相对路径，而 `hooks/hooks.json` 在项目根目录，`internal/` 包无法直接嵌入它。

**推荐方案：Makefile 构建时复制**

```makefile
build:
	cp hooks/hooks.json internal/hooks.json
	$(GORUN) build -ldflags "..." -o $(BINARY) .
```

`.gitignore` 中添加 `internal/hooks.json`，确保不会被误提交。

这与项目已有的 `internal/templates/*.tmpl` embed 模式一致（`role_create.go:18`），是最简单可靠的方案。

## 验证步骤

1. **构建验证**：`make build` 成功，`internal/hooks.json` 被正确复制和嵌入
2. **创建 worker 后检查**：
   ```bash
   agent-team worker create <role>
   cat .worktrees/<worker>/.claude/settings.json
   # 应包含 Stop、PreToolUse、PostToolUse 事件，不包含 TaskCompleted
   ```
3. **正常关闭会话**（Ctrl+C 或 /exit）：
   - 未提交的已跟踪文件改动被自动 commit
   - Active changes 被自动 archive
   - 主控制器收到通知消息
4. **无改动时关闭**：不产生空 commit，不发送通知
5. **重新打开 worker**：`agent-team worker open <id>` 后 settings.json 被更新
6. **已有 settings.json 时**：合并而非覆盖，用户自定义配置不丢失
