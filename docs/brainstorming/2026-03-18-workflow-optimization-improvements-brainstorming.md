# Workflow Optimization — 四项核心改进脑暴

> 基于：2026-03-18-workflow-optimization-task-breakdown.md（已完成）
> 日期：2026-03-18
> 状态：已批准

---

## 改进背景

在 Workflow Optimization 47 子任务执行过程中，发现以下四点需要优化：

1. **技能加载逻辑** — debug 时需使用 `systematic-debugging` skill，增量加载应使用 `find-skill`
2. **构建规则静态化** — `build-verification.md` 应在 init 时由 AI 扫描项目后动态生成
3. **上下文管理规则位置** — `context-management.md` 是强制要求，应在 `index.md` 中标记必读
4. **初始化强制检查** — 未初始化时必须强制初始化

---

## 改进方案总览

| # | 改进点 | 最终方案 | 复杂度 |
|---|--------|----------|--------|
| 1 | 技能加载优化 | 当前逻辑已满足（项目缓存 + 软连接） | 无需改动 |
| 2 | 构建规则动态化 | init 时 AI 扫描生成 + worker create 时 hash 比对警告 | 中 |
| 3 | 上下文管理强制化 | index.md 标记必读 | 低 |
| 4 | 未初始化强制检查 | PersistentPreRunE 交互式确认 | 中 |

---

## 改进点 1：技能加载优化

### 当前状态分析

阅读 `internal/skills.go` 后确认当前逻辑已满足需求：

1. **worker create 时** → `InstallSkillsForWorkerFromPath()` 安装 skills
2. **优先检查项目缓存** (`.agents/.cache/skills/`) → 有则软连接
3. **缓存未命中** → `npx skills add` 下载到项目缓存 → 软连接

**关键函数：**
- `findSkillPath()` — 5 层搜索 + 远程下载回退
- `projectSkillPath()` — 项目缓存路径
- `symlinkSkill()` — 创建软连接，失败回退到复制

### 决策

**无需改动** — 当前逻辑已符合"先检查缓存，未命中再增量下载"的设计。

---

## 改进点 2：构建规则动态化

### 问题

当前 `build-verification.md` 是静态模板文件，无法适配不同项目的构建脚本差异。

### 方案

**init 时 AI 扫描生成 + worker create 时 hash 比对警告**

| 阶段 | 行为 | 复杂度 |
|------|------|--------|
| `init` | AI 扫描项目文件（Makefile, go.mod, package.json, build.sh 等），生成 `build-verification.md` | 重（仅一次） |
| `worker create` | **只做 hash 比对** — 对构建相关文件算 hash，与 `.build-hash` 比较。若变更 → 输出警告 | 极轻 |

### 实现细节

1. **扫描文件列表**：
   - `Makefile`, `makefile`
   - `go.mod`, `go.sum`
   - `package.json`, `package-lock.json`
   - `Cargo.toml`
   - `build.sh`, `.github/workflows/*.yml`

2. **存储 hash 记录**：
   - 文件：`.agents/rules/.build-hash`
   - 格式：
     ```yaml
     generated_at: 2026-03-18T10:00:00Z
     files:
       Makefile: abc123...
       go.mod: def456...
     ```

3. **变更警告**：
   ```
   ⚠️  Build scripts changed since last init
      Modified: Makefile, go.mod
      Suggestion: Run 'agent-team rules sync --rebuild' to regenerate build-verification.md
   ```

4. **手动刷新命令**：
   ```bash
   agent-team rules sync --rebuild
   ```

### 优点

- **安全性** — 构建脚本变更会被发现
- **便捷性** — worker create 不会变慢（只做 hash 比对）
- **幂等性** — init 已有幂等机制，用户可随时手动编辑

---

## 改进点 3：上下文管理强制化

### 问题

`context-management.md` 包含 `/compact` 触发条件等强制规则，但当前在 `index.md` 中只是普通列表项，容易被忽略。

### 方案

**在 `index.md` 中标记必读**

修改前：
```markdown
- `context-management.md` — /compact 触发条件、handoff 摘要、Provider 差异
```

修改后：
```markdown
- `context-management.md` — **必读**：`/compact` 触发条件与上下文管理
```

### 配套改动

在 `CLAUDE.md` 根文件中已明确列出 5 条强制规则：

```markdown
- MUST read `.agents/rules/index.md` at task start and load the rule files required by the task.
- MUST call `/compact` whenever any trigger in `.agents/rules/context-management.md` fires.
```

此项无需代码改动，只需更新 `index.md` 文案。

---

## 改进点 4：未初始化强制检查

### 问题

当前 `cmd/root.go` 的 `PersistentPreRunE` 有 bootstrap 逻辑，但某些命令会跳过（如 `_inject-role-prompt`、`init` 等）。若用户未执行 `agent-team init` 就直接创建 worker，会导致 `.agents/rules/` 不存在。

### 方案

**在 `PersistentPreRunE` 中交互式确认**

```go
func PersistentPreRunE(cmd *cobra.Command, args []string) error {
    // 跳过命令列表
    skipCommands := []string{"init", "setup", "migrate", "help", "version"}
    if isSubCommand(cmd, skipCommands) {
        return nil
    }

    // 检查 .agents/rules/ 是否存在
    if !HasRulesDir(root) {
        fmt.Println("⚠️  Project not initialized.")
        fmt.Println("   Run 'agent-team init' now? [Y/n]")

        var answer string
        fmt.Scanln(&answer)
        if strings.ToLower(answer) != "n" {
            // 自动执行 init
            return cmd.Execute() // 或调用 init.InitProject()
        }
        return fmt.Errorf("initialization required")
    }

    // 现有 bootstrap 逻辑...
}
```

### 交互流程

```
$ agent-team worker create go-backend
⚠️  Project not initialized.
   Run 'agent-team init' now? [Y/n] y

[执行 init 流程...]
✓  Project initialized.
[继续执行 worker create]
```

### 跳过选项

在 CI/自动化场景中，用户可能需要非交互式失败：

```bash
# 非交互式模式：直接失败不询问
AGENT_TEAM_NONINTERACTIVE=1 agent-team worker create go-backend
# → error: project not initialized, run 'agent-team init' first
```

---

## 风险与缓解

| 风险 | 缓解措施 |
|------|----------|
| init 时 AI 扫描耗时 | 仅首次执行，后续 worker create 只做 hash 比对 |
| 交互式确认阻塞自动化 | 提供 `AGENT_TEAM_NONINTERACTIVE=1` 环境变量 |
| hash 比对误报 | `.build-hash` 仅跟踪构建相关文件，不跟踪文档 |

---

## 实施建议

### 任务拆分（预估）

| 任务 | 负责人 | 预估子任务数 |
|------|--------|-------------|
| `.build-hash` 存储 + hash 比对逻辑 | go-backend | 3 |
| `rules sync --rebuild` 命令 | go-backend | 2 |
| `index.md` 必读标记更新 | rules-writer | 1 |
| `PersistentPreRunE` 交互式确认 | go-backend | 3 |
| 测试验证 | qa-tester | 4 |

**总计约 13 个子任务**，可复用 Phase 1A + Phase 2B + Phase 3 的角色。

---

*Generated: 2026-03-18 · Approved design*
