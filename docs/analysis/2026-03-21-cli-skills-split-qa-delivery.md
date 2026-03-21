# CLI Skills Split QA 设计交付

日期：2026-03-21

## 1. 交付目标

本文基于 `docs/brainstorming/2026-03-21-cli-skills-split-brainstorming.md`，给出面向后续实现阶段的 QA 设计交付，用于验证新的独立 skill 体系是否满足以下目标：

1. 按用户场景而不是按命令树暴露能力。
2. 每个 skill 都有明确受众、触发词、CLI 绑定与上下文入口。
3. `context-cleanup` 与现有 `/compact` 语义彻底分离。
4. 所有上下文恢复都遵守“索引优先，再按需读正文”。
5. skill 设计与当前仓库中的真实 CLI 能力保持对齐，避免文档/协议漂移。

---

## 2. QA 范围

### 2.1 本次覆盖

本次 QA 交付覆盖两层：

1. **设计层 QA**
   - 检查 14 个 skill 的职责边界、触发语义、CLI 绑定、上下文入口是否完整且互不冲突。
   - 检查设计是否与当前仓库中的真实命令面一致。
   - 为后续实现准备验收矩阵、测试分层与质量门禁。

2. **实现前 QA 基线**
   - 定义最小测试资产结构。
   - 定义实现阶段必须补齐的测试集。
   - 定义高风险区与阻断条件。

### 2.2 本次不覆盖

1. 不直接实现任何 skill。
2. 不直接修改现有 `skills/` 目录内容。
3. 不执行端到端路由自动化测试；这里只定义未来执行方案。

---

## 3. CLI 基线核对结论

为避免 skill 设计与真实 CLI 脱节，先以当前代码为基线确认命令面。

### 3.1 当前已确认的主命令族

- `task`: `create/list/show/assign/done/archive`
- `workflow plan`: `generate/approve/activate/close`
- `worker`: `create/open/close/assign/status/merge/delete`
- `reply`
- `reply-main`
- `init`
- `setup`
- `migrate`
- `rules sync`
- `skill`: `check/update/clean`
- `role-repo`: `find/add/list/remove/check/update`
- `catalog`: `list/search/show/repo/normalize/stats/discover/serve`
- `role list/create`
- `compact`

### 3.2 对 skill 设计的 QA 含义

1. brainstorming 文档采用“场景优先”拆分是合理的，但实现时必须维护一份**场景 skill -> CLI 子命令**的映射快照，避免后续命令演进造成 skill 文档漂移。
2. `catalog-browser` 当前设计只绑定 `search/show/list/repo/stats`，这是一个合理的收敛子集；QA 需验证未纳入 skill 的 `normalize/discover/serve` 是否应继续保留在 human-only 运维路径，而非被误纳入浏览型 skill。
3. 当前仓库已有 `compact` 命令，因此 `context-cleanup` 的验收不能只看“能否恢复上下文”，还必须验证它**不会退化为沿用 compact 语义**。

---

## 4. QA 总体策略

采用三层验证：

### Layer A — 设计完整性验证

目标：确认 14 个 skill 的设计本身可落地。

重点检查：
- 受众是否唯一明确。
- 触发词是否能覆盖主场景。
- CLI 绑定是否指向真实存在的命令。
- 必需入口是否明确且满足索引优先。
- skill 之间是否存在职责重叠。

### Layer B — skill 骨架验收验证

目标：当最小 `SKILL.md` 落地后，验证每个 skill 的说明是否足够稳定支持正确触发。

重点检查：
- skill 描述是否能将自然语言意图收敛到正确命令。
- 是否包含明确的 Required entry。
- 是否清楚描述“什么时候不该由我处理”。
- 是否出现跨 skill 的文案漂移。

### Layer C — 路由与上下文集成验证

目标：在实际运行中验证 skill 路由、文件上下文恢复、命令绑定和边界保护。

重点检查：
- controller / worker / human prompt 是否命中正确 skill。
- 上下文类 skill 是否总是先读索引入口。
- 当 prompt 含糊时，是否会请求澄清而不是误调用错误 CLI。
- `context-cleanup` 是否只做会话清理 + 文件重锚，不复用 compact 语义。

---

## 5. 14 个 skill 验收矩阵

> 验收标准中的“P0”表示未通过即阻断设计落地；“P1”表示建议实现阶段补齐；“P2”表示优化项。

| Skill | 受众 | 核心绑定 | 必需入口 | 关键验收点 | 优先级 |
| --- | --- | --- | --- | --- | --- |
| task-orchestrator | controller, human | `agent-team task create/list/show/assign/done/archive` | `.agents/rules/index.md` | 能覆盖任务生命周期；不会把 worker-only 回复流误收进来 | P0 |
| workflow-orchestrator | controller, human | `agent-team workflow plan generate/approve/activate/close` | `.agents/rules/index.md` | 仅覆盖治理 plan，不误承诺不存在的自动执行器 | P0 |
| worker-dispatch | controller, human | `agent-team worker open/status`, `agent-team reply` | `.agents/rules/index.md` | 能区分“开 worker”“看 worker”“给 worker 回复”三类意图 | P0 |
| worker-recovery | worker | 读工件，必要时 `agent-team task show` | `worker.yaml` | 恢复链路必须 `worker.yaml -> task.yaml -> context.md -> 引用材料` | P0 |
| worker-reply-main | worker | `agent-team reply-main` | `worker.yaml` | 只承担 worker -> main 汇报，不承担 controller -> worker 回复 | P0 |
| project-bootstrap | human, controller | `agent-team init/setup/migrate` | `.agents/rules/index.md` | 能区分项目接入、全局 setup、目录迁移 | P0 |
| rules-maintenance | human, controller | `agent-team rules sync` | `.agents/rules/index.md` | 只负责规则同步，不误吞 skill 维护场景 | P1 |
| skill-maintenance | human, controller | `agent-team skill check/update/clean` | `.agents/rules/index.md` | 只负责 skill cache 维护，不误处理 role/规则问题 | P1 |
| role-repo-manager | human, controller | `agent-team role-repo find/add/list/check/update/remove` | `.agents/rules/index.md` | 能区分 role 来源管理与 catalog 浏览 | P0 |
| catalog-browser | human, controller | `agent-team catalog search/show/list/repo/stats` | `.agents/rules/index.md` | 只暴露浏览子集；不误包含 ingest/serve/normalize 运维操作 | P0 |
| task-inspector | controller, worker, human | `agent-team task list/show` | controller/human 读 `.agents/rules/index.md`；worker 读 `worker.yaml` | 只读定位稳定；不同受众入口正确分流 | P0 |
| worker-inspector | controller, human | `agent-team worker status` | `.agents/rules/index.md` | 只读 worker 查看，不误进入 dispatch/回复动作 | P1 |
| role-browser | controller, worker, human | `agent-team role list` | `.agents/rules/index.md` | 仅浏览本地 role；不误扩展到 role create / role-repo | P1 |
| context-cleanup | controller, worker | 策略 skill；不等价于 `compact` | controller 读 `.agents/rules/index.md`；worker 读 `worker.yaml` | 强制索引优先；禁止 compact 语义；不默认全量扫正文 | P0 |

---

## 6. 核心测试套件设计

## 6.1 设计验证套件（DV）

### DV-01：Skill 总量与分类完整性
- **目标**：验证 inventory 固定为 14 个 skill，且完整分布在 controller-first / worker-first / human-first / shared / strategy 五类。
- **通过标准**：数量、分类、命名与 brainstorming 一致。
- **失败后果**：设计基线不稳定，阻断进入实现。
- **优先级**：P0

### DV-02：每个 skill 都具备最小四元信息
- **检查项**：Audience / Triggers / CLI / Required entry。
- **通过标准**：14 个 skill 全量具备。
- **优先级**：P0

### DV-03：CLI 绑定全部可追溯到真实命令
- **目标**：skill 文档中列出的所有 CLI 均能在当前仓库代码中找到对应命令面。
- **通过标准**：无虚构命令、无过期子命令、无缺失关键命令。
- **优先级**：P0

### DV-04：skill 边界无高重叠
- **重点组合**：
  - `worker-dispatch` vs `worker-inspector`
  - `role-repo-manager` vs `catalog-browser`
  - `task-orchestrator` vs `task-inspector`
  - `worker-recovery` vs `context-cleanup`
- **通过标准**：每组至少存在明确主受众或主动作差异。
- **优先级**：P0

### DV-05：上下文恢复遵守索引优先
- **目标**：所有涉及恢复/重锚的 skill 必须先读入口索引。
- **通过标准**：
  - controller 场景先读 `.agents/rules/index.md`
  - worker 场景先读 `worker.yaml`
- **优先级**：P0

### DV-06：`context-cleanup` 与 compact 语义分离
- **目标**：验证设计上没有把 `context-cleanup` 定义成 `/compact` 同义词。
- **通过标准**：文档中明确强调“清理会话上下文 + 文件重锚”，而非压缩上下文。
- **优先级**：P0

---

## 6.2 路由验收套件（RT）

该套件用于后续 skill 落地后，以典型 prompt 进行抽样路由验证。

| 编号 | 样例 prompt | 期望 skill | 验收点 | 优先级 |
| --- | --- | --- | --- | --- |
| RT-01 | “帮我创建一个任务并派给 frontend worker” | `task-orchestrator` | 正确落到 task 生命周期入口 | P0 |
| RT-02 | “生成 workflow plan，等我审批后激活” | `workflow-orchestrator` | 只绑定 plan 治理命令 | P0 |
| RT-03 | “打开 qa worker，并告诉他先看 task.yaml” | `worker-dispatch` | 不误落入 worker-recovery | P0 |
| RT-04 | “我在 worker 里，恢复当前任务继续做” | `worker-recovery` | 先读 `worker.yaml` | P0 |
| RT-05 | “我做完了，向 main 汇报并请求验收” | `worker-reply-main` | 仅走 `reply-main` | P0 |
| RT-06 | “给这个仓库初始化 agent-team” | `project-bootstrap` | 能区分 init / setup / migrate | P0 |
| RT-07 | “同步一下规则，感觉规则漂了” | `rules-maintenance` | 只走 `rules sync` | P1 |
| RT-08 | “检查并更新 skill cache” | `skill-maintenance` | 只走 skill 维护命令 | P1 |
| RT-09 | “帮我找一个 backend role repo 并加到项目里” | `role-repo-manager` | 不误落到 catalog-browser | P0 |
| RT-10 | “搜索 catalog 里有哪些 product roles” | `catalog-browser` | 只读浏览链路正确 | P0 |
| RT-11 | “看下 task 状态” | `task-inspector` | 模糊场景落到只读查看是合理默认值 | P0 |
| RT-12 | “看下有哪些 worker 在线” | `worker-inspector` | 只读，不做打开/派发 | P1 |
| RT-13 | “列出本地 roles” | `role-browser` | 只调用 `role list` | P1 |
| RT-14 | “会话乱了，帮我清理并重新锚定” | `context-cleanup` | 明确非 compact 语义 | P0 |

---

## 6.3 负向与冲突测试套件（NG）

### NG-01：模糊 prompt 不应误执行高影响命令
- **样例**：`“看看 worker”`
- **期望**：优先命中 `worker-inspector` 或请求澄清，不应直接 `worker open`。
- **优先级**：P0

### NG-02：worker 身份下不应错误进入 controller skill
- **样例**：worker 会话中输入 `“approve plan”`
- **期望**：拒绝直接执行，提示这是 controller/human 场景。
- **优先级**：P0

### NG-03：controller 不应用 `worker-recovery` 恢复 worker 上下文
- **样例**：主控会话里说 `“恢复 worker 当前任务”`
- **期望**：命中 `worker-dispatch` 或请求澄清，而不是假装读取 `worker.yaml`。
- **优先级**：P0

### NG-04：`catalog-browser` 不应暴露写操作
- **样例**：`“把这个 repo 加到 role sources”`
- **期望**：转交 `role-repo-manager`。
- **优先级**：P0

### NG-05：`task-inspector` 不应吞掉 task 变更意图
- **样例**：`“完成任务并归档”`
- **期望**：转交 `task-orchestrator`。
- **优先级**：P0

### NG-06：`context-cleanup` 不应默认全量重读所有正文
- **期望**：若未先走索引入口，则判失败。
- **优先级**：P0

---

## 7. `context-cleanup` 专项 QA 方案

这是整个设计里风险最高、最容易语义回退的部分，需要单独设 Hard Gate。

### 7.1 专项验收目标

1. 它清理的是**会话上下文**，不是文件内容。
2. 它恢复的是**文件锚点**，不是继续依赖会话记忆。
3. 它必须执行**索引优先**恢复。
4. 它不能被实现成 `compact` 的重命名包装。

### 7.2 专项测试用例

#### CC-01：controller 索引优先恢复
- 步骤：触发 `context-cleanup`（controller 侧）
- 预期：先读 `.agents/rules/index.md`，再读取命中的规则正文，再读取当前 workflow/task 工件。
- 优先级：P0

#### CC-02：worker 索引优先恢复
- 步骤：触发 `context-cleanup`（worker 侧）
- 预期：先读 `worker.yaml`，再读 `task.yaml`，必要时才读 `context.md` 与补充材料。
- 优先级：P0

#### CC-03：禁止跳过索引直读正文
- 预期：实现中若直接打开 `context.md` 或规则正文而未先读索引，判失败。
- 优先级：P0

#### CC-04：禁止默认全量扫描所有上下文文件
- 预期：只按索引命中结果展开需要的正文。
- 优先级：P0

#### CC-05：禁止 compact 语义文案
- 预期：SKILL.md、引用文档、触发提示中不再以“压缩上下文”为主语义。
- 优先级：P0

#### CC-06：phase transition 触发稳定
- 样例：controller 完成任务分发后进入 review phase。
- 预期：能识别“换阶段”而触发清理/重锚策略。
- 优先级：P1

#### CC-07：manual trigger 稳定
- 样例：`“会话乱了，重新锚定一下”`
- 预期：稳定命中 `context-cleanup`。
- 优先级：P1

#### CC-08：恢复载荷最小化
- 预期：恢复时只保留 Goal / Phase / Constraints / Done / Next / Risks 所需最小状态，不做无界展开。
- 优先级：P1

### 7.3 `context-cleanup` 上线 Hard Gates

以下任一不满足，则阻断发布：

1. controller 路径未先读 `.agents/rules/index.md`。
2. worker 路径未先读 `worker.yaml`。
3. 文档或实现仍以 `/compact` 为默认语义。
4. 恢复流程默认全量扫描正文文件。
5. skill 无法区分 controller 与 worker 两条入口。

---

## 8. 建议的最小测试资产

后续实现阶段建议按以下最小资产落地：

### 8.1 文档资产

1. `docs/testing/cli-skills-routing-matrix.md`
   - 存放 RT / NG 用例与样例 prompt。
2. `docs/testing/context-cleanup-checklist.md`
   - 存放 CC 专项验收条目。
3. `docs/testing/skill-cli-binding-baseline.md`
   - 固定记录 skill -> CLI 映射，便于发现漂移。

### 8.2 自动化资产

1. 路由快照测试
   - 用固定 prompt 集验证 skill 选择结果。
2. 命令绑定快照测试
   - 用真实命令清单比对 skill 文档中的绑定项。
3. 上下文入口测试
   - 验证 context 类 skill 首次读取目标是否为索引文件。

### 8.3 人工审查资产

1. SKILL.md 审查清单
2. 触发词歧义清单
3. compact 语义回退检查清单

---

## 9. 质量门禁（Quality Gates）

## Gate A — 设计冻结前

- 14 个 skill 名单冻结。
- 每个 skill 都有 Audience / Triggers / CLI / Required entry。
- `context-cleanup` 已明确声明非 compact 语义。
- 所有 context 恢复路径都写明索引优先。

**结论：缺一不可。**

## Gate B — 第一批 SKILL.md 落地后

- 至少抽样覆盖 controller / worker / human 三类 prompt。
- 每个已落地 skill 都能命中一个稳定主场景。
- 未发现高重叠边界冲突。

**通过标准：P0 用例 100% 通过。**

## Gate C — 全量 skill 实现后

- RT 套件通过率 100%。
- NG 套件无 P0 误路由。
- CC 套件全部通过。
- skill -> CLI 映射与代码现状一致。

**通过标准：P0 100%，P1 ≥ 90%。**

---

## 10. 主要风险与 QA 缓解

### 风险 1：skill 文档与真实命令持续漂移
**缓解：** 维护独立的 skill-cli baseline，并在每次 CLI 变更后回归比对。

### 风险 2：场景 skill 重新退化为命令树包装
**缓解：** 路由测试以自然语言 prompt 为中心，而不是以命令帮助文档为中心。

### 风险 3：`context-cleanup` 在实现期被偷换为 compact transport
**缓解：** 将 CC-01~05 列为阻断用例，单独审查。

### 风险 4：shared skill 抢占 orchestrator skill 的主场景
**缓解：** 对 `*-inspector` 明确限定为只读；对变更类 prompt 强制路由到 orchestrator。

### 风险 5：worker/controller 身份边界被弱化
**缓解：** 所有 worker-first skill 必须依赖 `worker.yaml`，controller-first skill 必须依赖 `.agents/rules/index.md`。

---

## 11. 实施顺序建议（QA 视角）

建议分两波验收，降低回归面。

### Wave 1：P0 主路径
1. `task-orchestrator`
2. `workflow-orchestrator`
3. `worker-dispatch`
4. `worker-recovery`
5. `worker-reply-main`
6. `context-cleanup`
7. `task-inspector`
8. `role-repo-manager`
9. `catalog-browser`

**原因：** 先覆盖 controller/worker 主工作流与最高风险的上下文恢复语义。

### Wave 2：维护与浏览补齐
1. `project-bootstrap`
2. `rules-maintenance`
3. `skill-maintenance`
4. `worker-inspector`
5. `role-browser`

**原因：** 这些能力相对低风险，更适合在主链路稳定后补齐。

---

## 12. QA 结论

当前 brainstorming 方案从 QA 角度**可以进入实现阶段**，但前提是后续落地必须满足以下三条硬约束：

1. **所有上下文恢复都必须索引优先。**
2. **`context-cleanup` 必须与 compact 语义彻底切断。**
3. **skill 文档中的 CLI 绑定必须持续对齐当前仓库真实命令面。**

若后续实现严格执行上述约束，则这套 14-skill 拆分方案具备可测试性、可验收性与可演进性；反之，最可能出现的问题将是边界重叠、命令漂移，以及上下文恢复重新退化回会话记忆驱动。
