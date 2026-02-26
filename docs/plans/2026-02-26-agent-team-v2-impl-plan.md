# agent-team v2 实施计划

基于 [设计文档](2026-02-26-agent-team-v2-design.md)。

## 阶段 1：internal 层重构

### 1.1 新增 WorkerConfig 结构体
**文件**: `internal/config.go`
- 新增 `WorkerConfig` 结构体（worker_id, role, default_provider, default_model, pane_id, controller_pane_id, created_at）
- 新增 `LoadWorkerConfig(path)` / `Save(path)` 方法
- 保留 `RoleConfig` 暂不删除（后续阶段移除）

### 1.2 重构路径函数
**文件**: `internal/role.go`
- 新增 `RoleDir(root, roleName) string` → `agents/teams/<role-name>/`
- 新增 `WorkerDir(root, workerID) string` → `agents/workers/<worker-id>/`
- 新增 `WorkerConfigPath(root, workerID) string` → `agents/workers/<worker-id>/config.yaml`
- 新增 `ListAvailableRoles(root) []string` → 扫描 `agents/teams/*/SKILL.md`
- 新增 `ListWorkers(root) []WorkerInfo` → 扫描 `agents/workers/*/config.yaml`
- 新增 `NextWorkerID(root, roleName) string` → 计算下一个序号（如 `frontend-dev-001`）
- 更新 `FindWtBase()` 保持不变

### 1.3 新增 Skills 复制逻辑
**文件**: `internal/skills.go`（新文件）
- `CopySkillsToWorktree(wtPath, root, roleName) error`
  - 读取 `agents/teams/<role>/references/role.yaml` 获取 skills 列表
  - 复制角色 skill 目录到 `.claude/skills/<role>/` 和 `.codex/skills/<role>/`
  - 遍历依赖 skills，从全局 `skills/` 或 `~/.claude/skills/` 查找并复制
  - `.claude/skills/` 和 `.codex/skills/` 内容保持一致

### 1.4 更新 CLAUDE.md 生成逻辑
**文件**: `internal/role.go`
- 更新 `buildRoleSection()` → 从 `agents/teams/<role>/system.md` 读取（替代 `prompt.md`）
- 更新 `InjectRolePrompt()` → 接受 workerID 参数

### 1.5 更新 OpenSpec Change 创建
**文件**: `internal/openspec.go`
- 更新 `CreateChange()` → 支持 design.md 和 proposal.md 双文件写入
- 函数签名改为 `CreateChange(wtPath, changeName, proposal, design string)`

### 1.6 新增 .gitignore 生成
**文件**: `internal/role.go` 或 `internal/skills.go`
- `WriteWorktreeGitignore(wtPath) error` → 生成 `.gitignore`（排除 .gitignore, .claude/, .codex/, openspec/）

## 阶段 2：CLI 命令重构

### 2.1 Worker 父命令
**文件**: `cmd/worker.go`（新文件）
- 创建 `worker` 父命令，注册所有子命令

### 2.2 worker create
**文件**: `cmd/worker_create.go`（新文件）
- 参数: `<role-name>`
- 流程:
  1. 检查 `agents/teams/<role-name>/` 存在
  2. 如果不存在，检查全局 skills 中是否有，提示复制
  3. `NextWorkerID()` 计算序号
  4. `git worktree add .worktrees/<worker-id> -b team/<worker-id>`
  5. `WriteWorktreeGitignore()`
  6. 创建 `agents/workers/<worker-id>/config.yaml`
  7. OpenSpec init

### 2.3 worker open
**文件**: `cmd/worker_open.go`（新文件）
- 参数: `<worker-id> [provider] [--model] [--new-window]`
- 流程:
  1. 加载 worker config 获取 role
  2. `CopySkillsToWorktree()` 复制 skills
  3. `InjectRolePrompt()` 生成 CLAUDE.md
  4. 打开终端会话

### 2.4 worker assign
**文件**: `cmd/worker_assign.go`（新文件）
- 参数: `<worker-id> "desc" [--proposal file] [--design file]`
- 流程:
  1. `CreateChange()` 创建 design.md + proposal.md
  2. 自动打开会话（如果未运行）
  3. 通知会话

### 2.5 worker status
**文件**: `cmd/worker_status.go`（新文件）
- 显示所有 worker 及其角色、会话状态、活跃任务

### 2.6 worker merge
**文件**: `cmd/worker_merge.go`（新文件）
- 合并 `team/<worker-id>` 到当前分支

### 2.7 worker delete
**文件**: `cmd/worker_delete.go`（新文件）
- 关闭会话、删除 worktree 和分支、删除 `agents/workers/<worker-id>/`

### 2.8 role list
**文件**: `cmd/role.go`（新文件）
- 列出 `agents/teams/` 下所有可用角色

### 2.9 reply / reply-main 更新
**文件**: `cmd/reply.go`, `cmd/reply_main.go`
- `reply` 改为接受 `<worker-id>` 而非 `<name>`
- `reply-main` 更新为从 `agents/workers/` 查找 config

### 2.10 更新 root.go
**文件**: `cmd/root.go`
- `RegisterCommands()` 替换为新的子命令注册
- 移除 `open-all` 命令（可后续按需添加）

## 阶段 3：删除旧代码

### 3.1 删除旧命令文件
- 删除: `cmd/create.go`, `cmd/open.go`, `cmd/assign.go`, `cmd/status.go`, `cmd/merge.go`, `cmd/delete.go`
- 删除对应测试: `cmd/create_test.go`, `cmd/open_test.go`, `cmd/assign_test.go`, `cmd/delete_test.go`, `cmd/commands_test.go`

### 3.2 清理 internal
- 移除旧的 `RoleConfig` 结构体（如果不再使用）
- 移除 `PromptMDContent()` 函数
- 移除旧的路径函数（`TeamsDir`, `ConfigPath` 等）

## 阶段 4：role-creator 改造

### 4.1 更新 Python 脚本
**文件**: `skills/role-creator/scripts/create_role_skill.py`
- 新增 `--target-dir` 参数（默认 `skills`，可选 `agents/teams` 或自定义路径）
- 修改 `create_or_update_role()` 使用 target-dir 而非硬编码 `skills/`

### 4.2 更新 SKILL.md
**文件**: `skills/role-creator/SKILL.md`
- 更新 Generate Command 示例，包含 `--target-dir` 选项
- 新增目标目录选择步骤

### 4.3 更新测试
**文件**: `skills/role-creator/tests/test_create_role_skill.py`
- 添加 `--target-dir` 参数相关测试

## 阶段 5：agent-team SKILL.md 改造

### 5.1 重写 SKILL.md
**文件**: `skills/agent-team/SKILL.md`
- 更新为 v2 命令体系
- 添加角色/员工概念说明
- 添加角色创建流程（指引使用 role-creator skill）
- 添加任务完成规范（/openspec archive + reply-main）
- 更新通信协议

### 5.2 更新 references
**文件**: `skills/agent-team/references/details.md`
- 更新目录布局为新结构
- 更新命令参考

**文件**: `skills/agent-team/references/brainstorming.md`
- 更新步骤 6-8：design.md/proposal.md 存储说明
- 更新 assign 命令格式

## 阶段 6：测试

### 6.1 Go 单元测试
- 新增 `internal/skills_test.go`
- 新增 `cmd/worker_create_test.go`, `cmd/worker_open_test.go` 等
- 运行 `make test` 确保全部通过

### 6.2 集成验证
- 手动测试完整流程：角色创建 → worker create → worker open → worker assign → worker merge → worker delete
- 验证 .gitignore 正确排除文件
- 验证 skills 复制到 .claude/ 和 .codex/ 一致性

## 执行顺序

1. 阶段 1（internal 层） → 2（CLI 命令） → 3（删除旧代码）作为一个大 PR
2. 阶段 4（role-creator）可并行或作为单独 PR
3. 阶段 5（SKILL.md）随 CLI 变更一起更新
4. 阶段 6（测试）贯穿始终，每个阶段完成后运行
