# Global Role Resolver Brainstorming

> Role: General Strategist
> Date: 2026-03-02

## Problem Statement & Goals

当前 agent-team 的角色查找仅限项目级 `.agents/teams/`，全局角色目录 `~/.agents/roles/` 虽已有路径定义（`RoleRepoScope`），但未在创建角色和创建员工的核心流程中集成。

**目标：**
1. 创建角色时检查全局是否已有相同/类似角色，避免重复创建
2. 创建员工时支持项目级 → 全局的优先级查找，使全局角色可被直接引用
3. 全局角色采用原地引用方式，不复制到项目本地

## Constraints & Assumptions

- 全局角色目录 `~/.agents/roles/` 可能不存在（需安全处理）
- 全局角色数量通常在几十个以内，无需缓存索引
- `worker.yaml` 新增字段需向后兼容（`omitempty`）
- 已有的 `ResolveAgentsDir()` 模式应被继承

## Candidate Approaches

### Approach A: Unified Resolver (Recommended)

新增 `ResolveRole(root, roleName)` 统一解析函数 + `SearchGlobalRoles(roleName)` 模糊搜索函数。

| Pros | Cons |
|------|------|
| 单一入口，集中管理 | 需修改现有 worker create 角色验证逻辑 |
| worker create / role create 复用 | |
| 符合项目已有 ResolveAgentsDir 风格 | |
| 易于扩展更多搜索层 | |

### Approach B: Separated Implementation

各命令独立实现全局查找逻辑。

| Pros | Cons |
|------|------|
| 改动范围小 | 查找逻辑重复 |
| 各命令独立演进 | 维护成本高 |

### Approach C: Global Index Cache

引入 `~/.agents/.role-index.json` 索引文件。

| Pros | Cons |
|------|------|
| 搜索极快 | 索引维护成本 |
| 支持复杂搜索 | 过度工程化 |

## Recommended Design: Approach A

### 1. Core Data Structures

```go
// internal/role.go

type RoleMatch struct {
    RoleName    string // 角色目录名
    Path        string // 角色完整路径
    Scope       string // "project" | "global"
    Description string // 从 role.yaml 提取的描述
    MatchType   string // "exact" | "keyword"
}
```

### 2. Core Functions

```go
// 统一角色解析（用于 worker create）
func ResolveRole(root, roleName string) (*RoleMatch, error)
// 查找顺序: .agents/teams/<role> → ~/.agents/roles/<role>

// 全局角色搜索（用于 role create 的重复检查）
func SearchGlobalRoles(roleName string) ([]RoleMatch, error)
// 精确名称匹配 + 名称拆词搜 role.yaml 中的 description/systemGoal/inScope

// 辅助：从 role.yaml 提取角色配置
func readRoleYAML(rolePath string) (*RoleConfig, error)

// 辅助：名称拆词
func splitRoleKeywords(roleName string) []string
// "frontend-dev" → ["frontend", "dev"]

// 列出全局角色
func ListGlobalRoles() ([]RoleMatch, error)
```

### 3. Role Create Flow Enhancement

```
agent-team role create <name> ...
    ↓
SearchGlobalRoles(name)
    ├─ 精确名称匹配 ~/.agents/roles/<name>/
    └─ 拆词搜索: split(name) → keywords
       遍历 ~/.agents/roles/*/references/role.yaml 搜索关键词
    ↓
无匹配 → 正常创建
有匹配 → 展示匹配角色详情（名称、描述、路径、匹配类型）
       → fmt.Print("是否仍要创建本地角色？[y/N]")
          ├─ y → 继续创建
          └─ N → 退出，提示可用 worker create 使用全局角色
```

**特殊处理：**
- `--target-dir` 指定时不做全局检查
- `--force` 可跳过全局重复检查（CI/自动化场景）

### 4. Worker Create Flow Enhancement

```
agent-team worker create <role> [provider]
    ↓
ResolveRole(root, roleName)
    ├─ .agents/teams/<role>/ → scope="project" → 正常创建
    ├─ ~/.agents/roles/<role>/ → scope="global" → 用户确认:
    │     [1] 使用全局角色创建员工
    │     [2] 先创建本地角色，再创建员工
    │     [3] 取消
    └─ 两处都未找到 → 提示创建角色
```

**全局角色原地引用时：**
- `InjectRolePrompt` — 从全局路径读 `system.md`
- `InstallSkillsForWorker` — 从全局路径读 `role.yaml` 获取技能列表
- `CopySkillsToWorktree` — 将全局角色的 SKILL.md 等复制到 worktree `.claude/skills/`

### 5. worker.yaml Changes

```yaml
worker_id: frontend-dev-001
role: frontend-dev
role_scope: global          # NEW: "project" | "global"
role_path: ~/.agents/roles/frontend-dev  # NEW: 全局角色时记录绝对路径
provider: claude
created_at: "2026-03-02T..."
```

`role_scope` 和 `role_path` 使用 `omitempty`，向后兼容。

## File Change List

| File | Change | Description |
|------|--------|-------------|
| `internal/role.go` | Modify | 新增 `RoleMatch`, `ResolveRole()`, `SearchGlobalRoles()`, `splitRoleKeywords()`, `readRoleYAML()`, `ListGlobalRoles()` |
| `internal/role_test.go` | Modify | 新增单元测试 |
| `cmd/role_create.go` | Modify | CreateOrUpdateRole 前插入全局检查，支持 `--force` |
| `cmd/worker_create.go` | Modify | 替换角色验证为 `ResolveRole()`，增加全局角色确认交互 |
| `internal/skills.go` | Modify | `InstallSkillsForWorker` / `CopySkillsToWorktree` 支持绝对 rolePath 参数 |

## Risks & Mitigations

| Risk | Mitigation |
|------|-----------|
| `~/.agents/roles/` 不存在 | `ListGlobalRoles` 做 `os.Stat` 检查，不存在返回空列表 |
| 全局角色 role.yaml 格式不统一 | `readRoleYAML` 宽松解析，缺少字段不报错 |
| worker.yaml 新增字段破坏旧版本 | `omitempty` 确保向后兼容 |
| `--force` 滥用 | 仅跳过全局重复检查，不跳过其他验证 |

## Testing Strategy

### Unit Tests (`internal/role_test.go`)

- `TestResolveRole_ProjectFirst` — 项目级优先于全局
- `TestResolveRole_FallbackGlobal` — 项目无角色时找到全局
- `TestResolveRole_NotFound` — 两处都没有返回 error
- `TestSearchGlobalRoles_ExactMatch` — 精确名称匹配
- `TestSearchGlobalRoles_KeywordMatch` — 拆词搜 role.yaml
- `TestSplitRoleKeywords` — 各种名称格式的拆词
- `TestListGlobalRoles` — 扫描全局目录

### Integration Testing

- 使用 `t.TempDir()` 模拟全局目录
- 注入 HOME 环境变量或提取全局路径为函数参数以支持测试

## Open Questions

None — all questions resolved during brainstorming.
