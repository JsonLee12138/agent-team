# Brainstorming: TDD 驱动的任务分配系统 (Spec-Lite)

**日期**: 2026-03-02
**角色**: General Strategist
**状态**: 已批准

---

## 问题陈述与目标

### 当前痛点

当前 `agent-team` 通过 `internal/openspec.go` 深度集成了 [OpenSpec](https://github.com/Fission-AI/OpenSpec)：

- **外部依赖**: 依赖 npm 全局安装 `@fission-ai/openspec@latest`，引入 Node.js 生态
- **工作流偏向**: OpenSpec 设计偏向 "先 spec 后实现"，与 TDD 的 "先测试后实现" 理念冲突
- **过度复杂**: delta spec 同步、多 schema 配置、21+ AI 工具适配器等功能对本项目来说过重
- **不可控**: 外部 CLI 版本更新可能导致兼容性问题

### 目标

1. **去除 OpenSpec 依赖** — 纯 Go 实现，零外部 CLI 依赖
2. **内置 TDD 支持** — 验收测试驱动的任务完成判定
3. **保留 OpenSpec 优点** — Change/Artifact 结构化变更管理
4. **科学的任务管理** — 状态机驱动、可验证、可审计

---

## 约束与假设

- **语言**: Go，与现有 agent-team CLI 一致
- **数据格式**: YAML (元数据) + Markdown (文档)
- **任务层次**: Change/Task 两层结构
- **TDD 模式**: 验收测试驱动 — 任务定义时明确"完成 = 什么测试通过"
- **验证执行**: 系统自动验证（可配置、可跳过）
- **目录命名**: 全新 `.tasks/` 目录，与 OpenSpec 明确断开

---

## 候选方案与权衡

### 方案 A: Spec-Lite — 轻量 Change-Task 引擎 (推荐 ✓)

借鉴 OpenSpec 的 change/artifact 概念，用 Go 原生实现，加入 TDD 验证。

**优点**: 简单、渐进式迁移、与 worker 流程无缝对接
**缺点**: 验收测试定义偏自由格式，约束力度中等

### 方案 B: Task-Contract — 契约式任务系统

每个 task 是 "契约"，包含输入/输出/验证三要素。

**优点**: 高度结构化、对 AI worker 约束力强
**缺点**: 结构较重、简单任务过度设计

### 方案 C: Kanban-TDD — 看板式 + 测试门控

借鉴看板概念，change 在不同 "泳道" 间流转，每次流转需要通过测试门控。

**优点**: 显式 TDD 阶段、可视化进度
**缺点**: 阶段过多、不适用于所有任务类型

### 选择理由

方案 A 在复杂度和功能之间取得最佳平衡：
1. 渐进式迁移 — 目录结构与现有 openspec 相似，迁移成本最低
2. 适当的抽象层级 — 不过度设计也不过于复杂
3. 灵活的 TDD — 验收测试是推荐实践而非强制约束
4. 与 worker 流程自然契合

---

## 推荐设计

### 1. 目录结构

```
.tasks/                              # worktree 根目录下
├── config.yaml                      # 项目级配置
├── changes/
│   └── <timestamp>-<slug>/          # 每个 change 独立目录
│       ├── change.yaml              # 元数据 + tasks + verify 配置
│       ├── proposal.md              # 变更提案（必须）
│       ├── design.md                # 技术设计（可选）
│       └── tests.md                 # 验收测试定义（TDD 核心，可选）
└── archive/                         # 已完成归档
    └── 2026-03-02-<slug>/
```

### 2. 数据模型

**config.yaml** — 项目级配置：

```yaml
version: 1
defaults:
  verify:
    command: "go test ./... -v"
    timeout: 300s
  lifecycle:
    - draft
    - assigned
    - implementing
    - verifying
    - done
```

**change.yaml** — Change 元数据：

```yaml
name: "2026-03-02-add-user-auth"
description: "添加用户认证模块"
status: implementing          # draft | assigned | implementing | verifying | done
created_at: "2026-03-02T10:00:00Z"
assigned_to: "frontend-dev-001"
tasks:
  - id: 1
    title: "定义 User model 和数据库 schema"
    status: done               # pending | in_progress | done | skipped
  - id: 2
    title: "实现 JWT 认证中间件"
    status: in_progress
  - id: 3
    title: "编写验收测试"
    status: pending
verify:
  command: "go test ./auth/... -run TestAcceptance"
  timeout: 120s
  skip: false
```

### 3. Change 生命周期

**状态机：**

```
draft ──→ assigned ──→ implementing ──→ verifying ──→ done ──→ archived
  │                         │               │           │
  └── (取消) ←──────────────┘               │           └── (归档到 archive/)
                                            │
                                     ┌──────┴──────┐
                                     │ verify 失败  │
                                     │ 回退到       │
                                     │ implementing │
                                     └─────────────┘
```

**合法状态转换表：**

| From → To        | 条件                        |
|-------------------|-----------------------------|
| draft → assigned  | assigned_to 已设置          |
| assigned → implementing | Worker 开始工作        |
| implementing → verifying | 全部 tasks done/skipped |
| verifying → done  | verify 命令通过 或 skip=true |
| verifying → implementing | verify 失败 (回退)    |
| done → archived   | 手动执行 task archive       |

### 4. TDD 验证流程

**验证引擎逻辑：**

1. 当所有 tasks 标记为 done/skipped 时，系统自动进入 `verifying` 状态
2. 读取 verify 配置（change.yaml 覆盖 > config.yaml 默认）
3. 如果 `skip: true`，直接标记 `done`
4. 否则执行 verify command，带超时控制
5. exit code 0 → `done`，非 0 → 回退到 `implementing` 并通知 worker

**灵活性保证：**
- verify command 可在 config.yaml 全局配置，也可在 change.yaml 单独覆盖
- `skip: true` 处理纯文档/配置变更等不需要测试的场景
- 没有定义 verify command 时，降级为 worker 自报告

### 5. CLI 命令设计

**新增 task 子命令：**

```
agent-team task create   <description>           # 创建 change（draft 状态）
  --proposal, -p <file>                           # 提案文件
  --design, -d <file>                             # 设计文件
  --verify-cmd <command>                          # 验证命令
  --skip-verify                                   # 跳过验证

agent-team task list     [--status=<status>]      # 列出 changes
agent-team task show     <change-name>            # 查看 change 详情
agent-team task update   <change-name> [fields]   # 更新 change 属性
agent-team task verify   <change-name>            # 手动触发验证
agent-team task archive  <change-name>            # 归档已完成的 change
agent-team task done     <change-name> <task-id>  # 标记单个 task 完成
```

**现有命令修改：**

| 命令 | 变化 |
|------|------|
| `worker create` | 去掉 `openSpecSetup()`，改为初始化 `.tasks/` 目录 |
| `worker assign` | 调用 `task create` + 设置 `assigned_to` + 通知 worker |
| `worker status` | 读取 `.tasks/changes/*/change.yaml` 获取状态 |

### 6. Worker 集成流程

```
[Controller]                              [Worker]
    │                                         │
    ├── task create "添加认证"                 │ (draft)
    │   └── 生成 .tasks/changes/xxx/          │
    │                                         │
    ├── worker assign dev-001 xxx             │ (assigned)
    │   └── 通知: "[New Task] 添加认证         │
    │       Task: .tasks/changes/xxx/          │
    │       请先阅读 proposal.md"              │
    │                                         │
    │                              Worker:     │
    │                              1. 读提案    │
    │                              2. 写验收测试│
    │                              3. 实现功能  │
    │                              4. 更新状态  │
    │                                         │
    │   ←── task done xxx 1                    │ (implementing)
    │   ←── task done xxx 2                    │
    │   ←── task done xxx 3                    │
    │                                         │
    │   ←── reply-main "完成"                  │
    │                                         │
    ├── 自动触发 verify                        │ (verifying)
    │   └── 运行验收测试命令                    │
    │                                         │
    ├── 通过 → done | 失败 → 回退 implementing │
    └──                                       │
```

### 7. Go 代码架构

**新增文件：**

```
internal/
├── task.go           # 核心数据模型 (Change, Task, VerifyConfig, TaskConfig)
├── task_store.go     # 文件系统存储 (CRUD .tasks/ 目录操作)
├── task_verify.go    # 验证引擎 (运行 verify command, 超时, 结果处理)
├── task_lifecycle.go # 状态机 (ValidateTransition, AutoTransition)

cmd/
├── task.go           # task 父命令注册
├── task_create.go    # task create 实现
├── task_list.go      # task list 实现
├── task_show.go      # task show 实现
├── task_verify.go    # task verify 实现
├── task_archive.go   # task archive 实现
├── task_done.go      # task done 实现
```

**删除文件：**

```
internal/openspec.go  # 完全删除
```

**修改文件：**

```
cmd/worker_assign.go  # 替换 OpenSpec 逻辑为 task 系统
cmd/worker_create.go  # 去掉 openSpecSetup，改为 .tasks/ 初始化
cmd/worker_status.go  # 替换 OpenSpec status 解析
```

---

## 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 迁移期间 worker 兼容性中断 | 高 | 先实现新系统，再一次性切换，不做双系统并存 |
| verify command 在不同项目环境差异大 | 中 | config.yaml 支持项目级自定义，change 级覆盖 |
| AI worker 不遵循 TDD 流程 | 中 | 通过 prompt/skill 引导，tests.md 定义验收标准 |
| 任务状态不一致（手动修改 YAML） | 低 | CLI 命令是主要入口，状态机校验非法转换 |

---

## 测试策略

本系统自身采用 TDD 开发：

| 文件 | 类型 | 覆盖内容 |
|------|------|---------|
| `internal/task_test.go` | 单元测试 | Change/Task YAML 序列化/反序列化 |
| `internal/task_store_test.go` | 单元测试 | 文件系统 CRUD（`t.TempDir()`） |
| `internal/task_lifecycle_test.go` | 单元测试 | 状态转换合法性、边界条件 |
| `internal/task_verify_test.go` | 单元测试 | verify 命令执行、超时、失败回退 |
| `cmd/task_*_test.go` | 集成测试 | CLI 命令端到端 |

---

## 开放问题

1. **归档策略**: archive 后的 change 是否保留在 git 历史中，还是 `.tasks/archive/` 也加入 `.gitignore`？
2. **多 worker 共享 change**: 一个 change 是否可以分配给多个 worker？当前设计是 1:1。
3. **hook 集成**: TaskCompleted hook 是否应该自动触发 verify？还是由 worker 显式调用？
