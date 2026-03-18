# Agent-Team 流程优化 Brainstorming

> 角色：General Strategist
> 日期：2026-03-18
> 状态：已批准

---

## 问题陈述与目标

基于 Claude Code Insights 报告（651 sessions, 23 次错误方向），当前 agent-team 项目存在以下核心问题：

1. **Token 消耗过高** — worker 长时间运行无上下文压缩，主控分发任务后上下文膨胀
2. **任务粒度过粗** — `assign_role_task` 只接受字符串描述，无结构化拆分机制
3. **规则不可见** — 无项目级 CLAUDE.md/AGENTS.md/GEMINI.md，worker 无行为指引
4. **提示词臃肿** — InjectRolePrompt() 一次性注入所有内容，浪费 token

### 成功标准

- Worker 每个子任务完成后自动 compact，上下文不超过必要范围
- 需求可拆分为可追踪的子任务，主控有全局视图
- 规则系统渐进式加载，主提示词控制在 500 字符内
- 所有 provider（Claude/Codex/Gemini/OpenCode）统一规则体验

---

## 约束与假设

- 不创建新 skills（用户自行寻找合适的 skills）
- 保持 YAML 作为数据存储格式（与现有系统一致）
- Go 代码改动需一次性交付（方案 B）
- compact 由 worker 自行执行（规则驱动），不由 CLI 强制发送
- 测试/验收循环不纳入主控 Requirement 跟踪

---

## 参考项目研究

| 项目 | 类型 | 可借鉴模式 |
|------|------|-----------|
| [garrytan/gstack](https://github.com/garrytan/gstack) | 单 agent 技能切换 | CLAUDE.md 记录可用 skills；`CONDUCTOR_PORT` 多实例隔离 |
| [NousResearch/hermes-agent](https://github.com/NousResearch/hermes-agent) | Python agent 框架 | **50% 阈值上下文压缩**；bounded memory（2200 字符）；**3 级渐进技能加载**（list→view→deep-view）；summary-only 子 agent 返回；depth-limited delegation |
| [msitarzewski/agency-agents](https://github.com/msitarzewski/agency-agents) | 提示词库 + 编排器 | Dev-QA 循环（最多 3 次重试升级）；RICE 评分任务优先级；Assignment Matrix 任务路由 |
| [chenhg5/cc-connect](https://github.com/chenhg5/cc-connect) | Go 消息路由桥 | `/compress` 透传原生 compact；统一 Agent 接口（Go interface）；Relay 跨 agent 通信 |

**核心借鉴**：hermes-agent 的渐进式加载 + bounded memory 模式最直接解决 token 问题。

---

## 推荐设计

### 模块 1：渐进式多文件规则系统

**架构：**

```
.agents/rules/
├── index.md              ← Level 0: 规则索引（注入主提示词，≤500 字符）
├── debugging.md          ← Level 1: 按需读取
├── build-verification.md
├── communication.md
├── context-management.md ← compact 规则
├── task-protocol.md      ← 任务完成协议
└── worktree.md
```

**Level 0 — index.md（直接注入主提示词）：**

```markdown
# Project Rules
以下规则文件必须在相关场景下阅读并遵守：
- 调试 bug → 读 .agents/rules/debugging.md
- 提交代码 → 读 .agents/rules/build-verification.md
- 方案讨论 → 读 .agents/rules/communication.md
- 上下文管理 → 读 .agents/rules/context-management.md
- 任务完成 → 读 .agents/rules/task-protocol.md
- Worktree 操作 → 读 .agents/rules/worktree.md
```

**Level 1 — 各规则文件（worker 按场景自行读取）**

**根目录多平台文件：**

```
项目根/
├── CLAUDE.md      ← Claude Code 专属指令 + 引用 .agents/rules/
├── AGENTS.md      ← Codex 等 agents 专属 + 引用 .agents/rules/
├── GEMINI.md      ← Gemini CLI 专属 + 引用 .agents/rules/
└── .agents/rules/ ← 共享规则（所有平台通用）
```

各平台文件包含：
1. 该平台特有指令（构建命令、格式偏好等）
2. `<!-- agent-team:rules-start -->...<!-- agent-team:rules-end -->` tag 区域引用通用规则

**InjectRolePrompt() 改造：**

- 只写入：角色身份（极简）+ rules/index.md 内容 + 技能索引（名称+触发条件）
- 使用 `<!-- agent-team:rules-start -->...<!-- agent-team:rules-end -->` tag 模式注入
- 不再内联完整规则或 SKILL.md 内容

---

### 模块 2：Context Management（Auto-Compact）

**纯规则驱动，由 worker 自行执行：**

`.agents/rules/context-management.md`：

```markdown
## Context Management Rules

1. 每完成一个子任务后，立即执行 /compact
2. 每次 reply-main 汇报后，立即执行 /compact
3. 读取大量文件（超过 5 个）后，执行 /compact
4. 感知到上下文变长时，主动执行 /compact
5. 每 10-15 轮对话强制执行一次 /compact
```

**不做 CLI 层自动发送** — 各 provider 的 compact 命令不完全统一，且 worker 自行判断更灵活。

---

### 模块 3：两级任务树

**数据模型：**

```yaml
# .tasks/requirements/<req-name>/requirement.yaml
name: "user-auth-feature"
description: "实现用户认证功能"
status: open | in_progress | done
created_at: "2026-03-18T10:00:00Z"
sub_tasks:
  - id: 1
    title: "实现登录 API"
    assigned_to: "backend-dev"        # worker ID
    status: pending | assigned | done | skipped
    change_name: "login-api"          # 关联 worker 的 Change
  - id: 2
    title: "实现登录页面 UI"
    assigned_to: "frontend-dev"
    status: pending
    change_name: ""
```

**索引文件（快速查询）：**

```yaml
# .tasks/requirements/index.yaml
requirements:
  - name: "user-auth-feature"
    status: in_progress
    sub_task_count: 3
    done_count: 1
  - name: "dashboard-redesign"
    status: open
    sub_task_count: 0
    done_count: 0
```

**层级关系：**

```
Requirement（主控管理，需求层面）
  ├── SubTask 1 → 分配给 Worker A → Worker A 的 Change
  ├── SubTask 2 → 分配给 Worker B → Worker B 的 Change
  └── SubTask 3 → 分配给 Worker A → Worker A 的另一个 Change
```

- **Requirement**：主控创建和追踪
- **SubTask**：主控创建，分配给 worker
- **Change**（现有）：worker 实施单元，保持不变
- SubTask 完成 = 对应 Change done → 自动回滚状态到 Requirement
- 所有 SubTask done → Requirement 自动标记 done

**新增 CLI 命令：**

| 命令 | 作用 |
|-----|------|
| `agent-team req create <name>` | 创建需求 |
| `agent-team req split <name>` | 拆分需求为子任务 |
| `agent-team req assign <name> <task-id> <worker>` | 分配子任务 |
| `agent-team req status <name>` | 查看需求及子任务状态 |
| `agent-team req done <name>` | 标记需求完成 |

**不做的事：**
- 测试/验收循环不纳入 Requirement 跟踪
- SubTask 不支持再次拆分（保持两级，不做递归树）

---

### 模块 4：init 职责拆分

**现状问题**：当前 `agent-team init` 混合了安装配置和项目初始化。

**改造：**

| 命令 | 职责 | 执行频率 |
|------|------|---------|
| `agent-team setup`（原 init 改名） | 全局安装级：创建 `.agents/` 目录结构、配置 session backend、检查依赖 | 每台机器一次 |
| `agent-team init`（新语义） | 项目初始化：创建 `.agents/rules/`、生成默认规则文件、创建/更新根目录 provider 文件 | 每个项目一次 |

**Skill 层 init 检查：**

各 skill 执行前检查 `.agents/rules/index.md` 是否存在，不存在则提示运行 `agent-team init`。
可在 Go 的 `PersistentPreRunE` 中对需要 init 的命令统一拦截。

**已有项目兼容：**
- `agent-team init` 幂等执行
- 根目录文件已存在时，只更新 `<!-- agent-team:rules-start -->...<!-- agent-team:rules-end -->` tag 区域
- `agent-team rules sync` 子命令：手动同步 rules → 根目录 provider 文件

---

## 改动文件清单

| 文件 | 改动类型 | 内容 |
|------|---------|------|
| `CLAUDE.md`（新建） | 新建 | Claude 专属 + rules 引用 |
| `AGENTS.md`（新建） | 新建 | Codex 专属 + rules 引用 |
| `GEMINI.md`（修改） | 修改 | Gemini 专属 + rules 引用 |
| `.agents/rules/*.md`（新建） | 新建 | 渐进式规则文件集（7 个文件） |
| `internal/role.go` | 修改 | InjectRolePrompt 精简注入：身份 + index + 技能索引，tag 模式 |
| `internal/task.go` | 修改 | 新增 Requirement + SubTask 数据模型 |
| `internal/task_store.go` | 修改 | Requirement CRUD + index.yaml 管理 |
| `internal/task_lifecycle.go` | 修改 | SubTask→Requirement 状态回滚 |
| `cmd/req_*.go`（新建，5 个文件） | 新建 | req create/split/assign/status/done |
| `cmd/init.go` | 修改 | 拆分为 setup + init 新语义 |
| `cmd/setup.go`（新建） | 新建 | 原 init 逻辑迁移 |
| `internal/init.go` | 修改 | 新增 rules 目录创建 + provider 文件生成 |

---

## 风险与缓解

| 风险 | 缓解 |
|------|------|
| init 改名破坏现有用户流程 | `setup` 命令保留 `init` 作为 alias，deprecation warning |
| 规则驱动的 compact 依赖 AI 自觉性 | 在 task-protocol.md 中用强指令（MUST/ALWAYS），多个规则交叉强化 |
| 两级任务树增加主控认知负担 | `req status` 提供单命令全局视图，index.yaml 避免遍历 |
| 渐进式加载 worker 可能不读规则 | index.md 在主提示词中明确列出触发条件，降低遗漏概率 |

---

## 验证策略

1. **规则系统**：创建 worker → 检查注入内容是否精简 → 验证 tag 更新幂等性
2. **Compact 规则**：观察 worker 会话是否在任务完成后执行 /compact
3. **任务树**：`req create` → `req split` → `req assign` → worker `task done` → 验证状态回滚
4. **init 拆分**：`agent-team setup` + `agent-team init` 分别执行，验证幂等性

---

## 开放问题

1. `req split` 是交互式拆分还是从描述自动生成子任务？（建议：先做手动拆分，后续可加 AI 辅助）
2. 是否需要 `agent-team rules sync` 命令，还是每次 `worker create` 时自动同步？
3. Gemini CLI 无原生 /compact，context-management 规则中如何处理？

---

*Generated: 2026-03-18 · Role: General Strategist · Method: Brainstorming Skill*
