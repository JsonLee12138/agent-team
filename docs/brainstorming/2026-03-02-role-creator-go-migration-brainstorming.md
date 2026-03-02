# Role-Creator Python → Go Migration

- **Role**: General Strategist
- **Date**: 2026-03-02

## Problem Statement

`skills/role-creator/` 中的 `create_role_skill.py` 是唯一的 Python 脚本，与项目主体（Go CLI 插件）技术栈不统一。既然 agent-team 已经是插件架构，脚本应统一到 Go 项目中，消除 Python 运行时依赖。

## Goals

1. 将 `create_role_skill.py` 的全部功能用 Go 重写
2. 以 `agent-team role create <role-name>` 子命令形式暴露
3. 模板通过 `//go:embed` 嵌入二进制，零外部依赖
4. 测试一对一迁移，保持完整覆盖率
5. 清理 Python 相关文件，保留 SKILL.md 作为技能入口

## Constraints & Assumptions

- 用户已安装 `agent-team` CLI，SKILL.md 中调用 `agent-team role create` 而非 `go` 或 `python3`
- 模板语法从 Python `string.Template`（`$var`）迁移到 Go `text/template`（`{{.Var}}`）
- CLI 参数与原 Python 脚本完全一致，唯一变更：`--role-name` 改为 positional 参数
- 交互逻辑（`confirm_overwrite`）放在 `cmd/` 层，`internal/` 保持纯逻辑可测试

## Candidate Approaches

### A. internal 包 + cmd 入口（Recommended）

```
cmd/role_create.go            # 子命令，参数解析，调用 internal
internal/role_create.go       # 核心逻辑
internal/role_create_test.go  # 完整测试
internal/templates/*.tmpl     # //go:embed 模板
```

- 与现有 `cmd/` + `internal/` 分层一致
- 命令层薄，逻辑可测试、可复用

### B. 全部放在 cmd 层

- 简单直接，文件少
- 逻辑和 CLI 耦合，不利于测试

### C. 独立 pkg 包

- 最大隔离性
- 项目无 `pkg/` 目录，过度设计

**Decision**: 方案 A

## Recommended Design

### File Structure

#### New Files

| File | Responsibility |
|------|---------------|
| `cmd/role_create.go` | `agent-team role create` 子命令，参数解析 |
| `internal/role_create.go` | 核心逻辑：validation、模板渲染、备份、写入 |
| `internal/role_create_test.go` | 16 个测试场景一对一迁移 |
| `internal/templates/SKILL.md.tmpl` | 嵌入模板 |
| `internal/templates/role.yaml.tmpl` | 嵌入模板 |
| `internal/templates/system.md.tmpl` | 嵌入模板 |

#### Modified Files

| File | Change |
|------|--------|
| `cmd/root.go` | 注册 `role create` 子命令 |
| `skills/role-creator/SKILL.md` | 调用方式改为 `agent-team role create ...` |

#### Deleted Paths

| Path | Reason |
|------|--------|
| `skills/role-creator/scripts/` | Python 脚本被 Go 替代 |
| `skills/role-creator/tests/` | Python 测试被 Go 测试替代 |
| `skills/role-creator/assets/` | 模板迁移到 `internal/templates/` |

### CLI Interface

```bash
agent-team role create <role-name> [flags]
```

**Required**:
- `role-name` (positional): kebab-case 角色名
- `--description`: 角色描述
- `--system-goal`: 系统目标

**Optional**:
| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--in-scope` | `[]string` | `[]` | 职责范围内，可重复或逗号分隔 |
| `--out-of-scope` | `[]string` | `[]` | 职责范围外 |
| `--skills` | `[]string` | `[]` | 技能列表 |
| `--recommended-skills` | `[]string` | `[]` | 推荐技能 |
| `--add-skills` | `[]string` | `[]` | 追加技能 |
| `--remove-skills` | `[]string` | `[]` | 移除技能 |
| `--manual-skills` | `[]string` | `[]` | 手动指定技能 |
| `--target-dir` | `string` | `skills` | 输出目录 |
| `--overwrite` | `string` | `ask` | 覆盖模式：ask/yes/no |
| `--repo-root` | `string` | `.` | 仓库根目录 |

### Core Data Structures

```go
type RoleConfig struct {
    RoleName    string
    Description string
    SystemGoal  string
    InScope     []string
    OutOfScope  []string
    Skills      []string
}

type GenerationResult struct {
    TargetDir  string
    BackupPath string // empty = no backup
}
```

### Function Mapping (Python → Go)

| Python | Go | Notes |
|--------|-----|-------|
| `is_kebab_case` | `IsKebabCase` | regex `^[a-z0-9]+(?:-[a-z0-9]+)*$` |
| `normalize_role_name` | `NormalizeRoleName` | lower → replace non-alnum with `-` → collapse → trim |
| `validate_role_name` | `ValidateRoleName` | validate + suggest |
| `parse_csv_list` | `ParseCSVList` | comma split |
| `dedupe_keep_order` | `DedupeKeepOrder` | ordered dedup |
| `parse_checkbox_indices` | `ParseCheckboxIndices` | checkbox format |
| `parse_numeric_indices` | `ParseNumericIndices` | numeric format |
| `parse_selection_reply` | `ParseSelectionReply` | checkbox-first precedence |
| `resolve_final_skills` | `ResolveFinalSkills` | merge + dedup |
| `render_files` | `RenderFiles` | `text/template` + `//go:embed` |
| `backup_existing_role` | `BackupExistingRole` | `<parent>/.backup/<name>-<ts>/` |
| `confirm_overwrite` | cmd layer | interactive, injected via callback |
| `create_or_update_role` | `CreateOrUpdateRole` | main entry |
| `collect_scope` | `CollectScope` | merge + fallback |

### Template Migration

- Python `string.Template` (`$var`) → Go `text/template` (`{{.Var}}`)
- Variable names camelCase in Go substitutions
- Three template files embedded via `//go:embed internal/templates/*`

### Test Migration (16 cases)

| # | Go Test Function | Scope |
|---|-----------------|-------|
| 1 | `TestValidateRoleName_InvalidIncludesSuggestion` | name validation error message |
| 2 | `TestRenderFiles_Deterministic` | idempotent rendering |
| 3 | `TestRenderFiles_SystemPromptIncludesSkillPolicy` | skill policy text in system.md |
| 4 | `TestCreateOrUpdateRole_OverwriteCreatesBackup` | backup + managed file overwrite + legacy cleanup |
| 5 | `TestResolveFinalSkills_ManualFallback` | manual skills when recommendation empty |
| 6 | `TestResolveFinalSkills_EmptyAllowed` | empty skills = `[]` |
| 7 | `TestRenderFiles_EmptySkillsYAMLArray` | `skills: []` in YAML |
| 8 | `TestRenderFiles_NonEmptySkillsYAMLList` | `skills:\n  - "x"` in YAML |
| 9 | `TestParseSelectionReply_CheckboxMode` | checkbox parsing |
| 10 | `TestParseSelectionReply_NumericMode` | numeric parsing |
| 11 | `TestParseSelectionReply_CheckboxPrecedence` | mixed input, checkbox wins |
| 12 | `TestCreateOrUpdateRole_HappyPath` | new role creation end-to-end |
| 13 | `TestCreateOrUpdateRole_OverwriteModeNoError` | overwrite=no → error |
| 14 | `TestNormalizeRoleName_EdgeCases` | normalization edge cases |
| 15 | `TestCollectScope_Fallback` | scope merge + fallback |
| 16 | `TestCreateOrUpdateRole_TargetDirVariants` | .agents/teams + custom path |

**Testing approach**:
- Pure functions: table-driven tests
- File system operations: `t.TempDir()`
- Interactive confirm: injected `io.Reader` / callback

### SKILL.md Changes

1. **Generate Command** section: `python3 skills/role-creator/scripts/create_role_skill.py` → `agent-team role create`
2. **Validation** section: `python3 -m unittest ...` → `go test ./internal/... -run TestRoleCreate -v`
3. **Workflow step 7**: "Execute generator script" → "Execute `agent-team role create`"

## Risks & Mitigations

| Risk | Impact | Mitigation |
|------|--------|-----------|
| 模板语法迁移引入渲染差异 | 生成文件内容不一致 | 测试 #2 (deterministic) 对比验证 |
| Go `text/template` 对 YAML 缩进处理不同 | role.yaml 格式错误 | 测试 #7, #8 验证 YAML 格式 |
| 遗漏 Python 脚本的边界情况 | 回归 bug | 16 个测试一对一迁移 |
| 用户未安装最新 CLI 版本 | SKILL.md 调用失败 | SKILL.md 中注明最低版本要求 |

## Validation Strategy

1. Go 测试全部通过（16 cases）
2. 手动执行 `agent-team role create test-role ...` 验证生成文件
3. 对比 Python 和 Go 版本对相同输入的输出一致性
4. 确认 `skills/role-creator/scripts/`、`tests/`、`assets/` 已删除
5. 确认 `skills/role-creator/SKILL.md` 已更新且无 Python 引用

## Open Questions

None — all decisions confirmed.
