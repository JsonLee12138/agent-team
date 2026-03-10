# Skill 缓存方案：项目级安装 + Worker Symlink

## Context

当前 `worker create` 时，所有 skill（包括远程 scoped skill）都安装到 worktree 内部。worktree 是临时的，`worker delete` 删除后 skill 全部丢失。同一角色反复创建 worker 会重复下载远程 skill，浪费时间和带宽。

方案核心：将 skill 安装到项目根目录（`npx skills add` 的默认行为），worker 中只创建 symlink 指向项目级目录。结合 `npx skills check/update` 实现版本管理。

## 改动概览

### 1. 改造 `InstallSkillsForWorkerFromPath` — `internal/skills.go`

当前签名：
```go
func InstallSkillsForWorkerFromPath(wtPath, root, roleName, rolePath, provider string) error
```

新签名（加 `fresh` 参数）：
```go
func InstallSkillsForWorkerFromPath(wtPath, root, roleName, rolePath, provider string, fresh bool) error
```

新逻辑：

**角色自身 skill** → symlink `<wtPath>/.<provider>/skills/<roleName>` → `rolePath`（不再 copyDir）

**依赖 skill 处理**：
- 先检查项目级缓存 `<root>/.<provider>/skills/<shortName>/` 是否存在
- 缓存命中且 `!fresh` → 直接 symlink 到 worktree
- 缓存未命中或 `fresh`：
  - scoped skill → `runNpxSkillsAdd(root, skillName, provider)` 安装到项目根目录，然后 symlink
  - 普通 skill → `findSkillPath` 本地搜索，找到则直接 symlink 到源路径；找不到则 `runNpxSkillsAdd(root, ...)` 安装到项目根目录再 symlink
- symlink 失败时 fallback 到 copyDir（兼容不支持 symlink 的环境）

关键变化：`runNpxSkillsAdd` 的 `cwd` 从 `wtPath` 改为 `root`（项目根目录），这样 `npx skills add` 会安装到 `<root>/.<provider>/skills/`。

### 2. 新增 symlink 工具函数 — `internal/skills.go`

```go
// symlinkSkill 在 worktree 中创建指向目标的 symlink，失败时 fallback 到 copyDir
func symlinkSkill(wtPath, provider, skillName, targetPath string) error

// isSymlink 检查路径是否是 symlink
func isSymlink(path string) bool

// projectSkillPath 返回项目级 skill 缓存路径
func projectSkillPath(root, provider, skillName string) string
// → filepath.Join(root, skillTargetSuffix(provider), skillName)
// 例如 <root>/.claude/skills/vite
```

### 3. 同步更新 `InstallSkillsForWorker` — `internal/skills.go`

```go
func InstallSkillsForWorker(wtPath, root, roleName, provider string) error {
    return InstallSkillsForWorkerFromPath(wtPath, root, roleName, RoleDir(root, roleName), provider, false)
}
```

### 4. 更新 `skillInstaller` 变量签名 — `cmd/worker_create.go`

```go
var defaultTaskSetup = func(wtPath string) error { ... }
var taskSetup = defaultTaskSetup

// 签名加 fresh 参数
var skillInstaller = internal.InstallSkillsForWorkerFromPath
```

### 5. 添加 `--fresh` flag — `cmd/worker_create.go`

`newWorkerCreateCmd` 新增 `--fresh` flag。

`RunWorkerCreate` 签名变为：
```go
func (a *App) RunWorkerCreate(roleName, provider, model string, newWindow, fresh bool) error
```

第 167 行调用改为：
```go
if err := skillInstaller(wtPath, root, roleName, rolePath, provider, fresh); err != nil {
```

### 6. 更新 `worker open` — `cmd/worker_open.go`

第 65 行调用加 `false`：
```go
if err := skillInstaller(wtPath, root, cfg.Role, rolePath, provider, false); err != nil {
```

### 7. 新增 `skill` 子命令组

**`cmd/skill.go`** — 命令组入口：
```go
func newSkillCmd() *cobra.Command {
    cmd := &cobra.Command{Use: "skill", Short: "Manage project skill cache"}
    cmd.AddCommand(newSkillCheckCmd())
    cmd.AddCommand(newSkillUpdateCmd())
    cmd.AddCommand(newSkillCleanCmd())
    return cmd
}
```

**`cmd/skill_check.go`** — 在项目根目录执行 `npx skills check`

**`cmd/skill_update.go`** — 在项目根目录执行 `npx skills update`

**`cmd/skill_clean.go`** — 删除项目根目录下所有 `.<provider>/skills/` 中的远程 skill 缓存

### 8. 注册命令 — `cmd/root.go`

`RegisterCommands` 中添加：
```go
rootCmd.AddCommand(newSkillCmd())
```

### 9. 更新 `.gitignore`

添加远程 skill 缓存目录的忽略规则：
```
# Remote skill cache (installed by npx skills add)
.claude/skills/
.codex/skills/
.opencode/skills/
.gemini/skills/
```

### 10. 更新测试

**`internal/skills_test.go`**：
- `TestInstallSkillsForWorkerLocalOnly` 和 `TestInstallSkillsForWorkerCodexProvider`：调用签名加 `false` 参数
- 新增验证：worktree 中的 skill 是 symlink（`os.Lstat` + `ModeSymlink` 检查），且 symlink 目标正确
- 新增 `TestSymlinkSkill`：验证 symlink 创建、覆盖、fallback 到 copy
- 新增 `TestProjectSkillPath`：验证路径计算
- 新增 `TestFreshFlag`：验证 `fresh=true` 时重新安装

**`CopySkillsToWorktreeFromPath`**：保留不变（仍用 copyDir），标记 `// Deprecated`，现有测试 `TestCopySkillsToWorktree` 和 `TestCopySkillsToWorktreeScopedName` 不需要修改。

## 文件清单

| 文件 | 操作 |
|------|------|
| `internal/skills.go` | 修改：`InstallSkillsForWorkerFromPath` 改为 symlink 模式，新增 `fresh` 参数；新增 `symlinkSkill`、`isSymlink`、`projectSkillPath` 函数 |
| `cmd/worker_create.go` | 修改：`skillInstaller` 签名更新，`RunWorkerCreate` 加 `fresh` 参数，新增 `--fresh` flag |
| `cmd/worker_open.go` | 修改：`skillInstaller` 调用加 `false` |
| `cmd/skill.go` | 新增：skill 子命令组 |
| `cmd/skill_check.go` | 新增：`npx skills check` 封装 |
| `cmd/skill_update.go` | 新增：`npx skills update` 封装 |
| `cmd/skill_clean.go` | 新增：清理缓存 |
| `cmd/root.go` | 修改：注册 `newSkillCmd()` |
| `.gitignore` | 修改：添加 `.<provider>/skills/` 忽略规则 |
| `internal/skills_test.go` | 修改：更新调用签名，新增 symlink 验证测试 |

## 验证方式

1. `go test ./internal/ -run TestInstallSkillsForWorker` — 验证 symlink 模式正常工作
2. `go test ./internal/ -run TestCopySkillsToWorktree` — 验证旧 API 向后兼容
3. `go test ./cmd/ -run TestWorkerDelete` — 验证删除不受影响
4. 手动测试：`agent-team worker create vite-react-dev` → 检查 worktree 中 skill 是 symlink → `agent-team worker delete` → 检查项目级缓存仍在 → 再次 `worker create` → 验证无需重新下载
5. `agent-team skill check` / `agent-team skill update` — 验证版本管理命令
