# agent-team 优化 TODO 方案

日期：2026-03-21

## 目标

把当前 `agent-team` 的优化方向整理成一份**可执行的 todo 方案**，用于后续按阶段确认、脑暴和实现。

当前原则：

- 先修正文档与真实实现不一致的问题
- 先补 task 工件闭环
- roadmap 明确需要，但先作为未来规划层能力保留
- phase / milestone 先定义概念，暂不直接并入当前核心执行模型

---

## 一、概念定义（作为后续规划层术语基线）

### roadmap 是什么

`roadmap` 是路线图，用来回答：

- 现在做什么
- 下一步做什么
- 后面准备做什么
- 哪些是未来方向，不是当前承诺

建议定位：

- 用于组织 `Now / Next / Later`
- 用于组织多个 milestone
- 后续更适合作为独立 skill 产出的规划工件

### milestone 是什么

`milestone` 是阶段性可交付目标，用来回答：

- 这一大轮完成后，用户能得到什么
- 哪个较完整的能力集已经完成
- 这一轮的 definition of done 是什么

建议定位：

- 一个 milestone 可以包含多个 phase
- milestone 负责阶段性交付说明

### phase 是什么

`phase` 是 milestone 内部推进阶段，用来回答：

- 为了完成 milestone，需要先做哪一步，再做哪一步
- 哪几个阶段分别解决哪些子问题

建议定位：

- 一个 phase 可以包含多个 task
- phase 是阶段分解，不替代 task

### task 是什么

`task` 是当前系统中的最小执行单元。

当前推荐保持：

- task-first
- single-task worker lifecycle
- worker 只绑定一个 task

### verification 是什么

`verification` 是 task 的验收工件，用来记录：

- task 的验收标准
- 实际检查过程
- 当前结果
- 遗留问题与风险

建议新增为 task 的标准工件之一。

---

## 二、推荐层级关系（未来规划层）

```text
roadmap
  └── milestone
        └── phase
              └── task
                    └── verification
```

当前立即落地的核心层：

- `task`
- `verification`

未来规划层：

- `roadmap`
- `milestone`
- `phase`

---

## 三、总 TODO 清单

## P0：当前必须推进

### TODO-001：修正文档与 README，使其与当前实现一致

**目标**

让 README 和相关文档反映当前真实能力，而不是旧版本叙事。

**范围**

- README 主叙事
- 目录结构说明
- 主流程说明
- context-cleanup 语义说明
- 主命令 / 兼容命令说明

**需要覆盖的内容**

- `.agent-team` 是当前主目录与主语义
- `task-first + single-task worker lifecycle` 是当前主流程
- 当前推荐主流程：
  - task create
  - task assign
  - worker deliver
  - task done
  - archive / merge / cleanup
- `context-cleanup` 是文件重锚恢复，不是简单 `/compact`
- 历史或兼容能力需要单独标明，不应继续作为主叙事

**完成标准**

- README 不再误导用户走旧路径
- 新用户能从 README 直接理解当前推荐工作流
- AI 读取 README 时不会优先得到过时目录或流程信息

**当前状态**

- [ ] 未开始

---

### TODO-002：在文档中增加规划概念定义章节

**目标**

先统一 `roadmap / milestone / phase / task / verification` 概念，为后续脑暴和规划扩展打基础。

**范围**

- README 中新增概念章节，或
- 新增专门文档并在 README 中链接

**需要覆盖的内容**

- roadmap 是路线图，不是当前执行单元
- milestone 是阶段性可交付目标
- phase 是 milestone 内部推进阶段
- task 是最小执行单元
- verification 是 task 验收工件

**完成标准**

- 后续讨论规划层时，不需要反复重新解释术语
- roadmap / milestone / phase 不会被误解为当前必须立即实现的执行层

**当前状态**

- [ ] 未开始

---

### TODO-003：为 task 增加 `verification.md`

**目标**

把 verification 纳入 task 标准工件，补齐 task 闭环能力。

**范围**

- task 创建时的工件结构
- task 文档模板
- task 验收信息记录方式

**建议模板字段**

- `Acceptance Criteria`
- `Checks Performed`
- `Result`
- `Issues`
- `Verified By`
- `Verified At`

**完成标准**

- 每个 task 都可以有标准化验收工件
- 验收信息不再只能散落在聊天或临时回复里
- 后续可以基于该工件扩展 gate / 状态展示 / archive 审核

**当前状态**

- [ ] 未开始

---

## P1：短期增强

### TODO-004：让 task 状态流转感知 `verification.md`

**目标**

让 verification 从“可选附加文档”升级为“状态流转可感知工件”。

**建议分阶段推进**

第一阶段：

- `task done` / `archive` 前检查 `verification.md` 是否存在
- 缺失时 warning，不阻断

第二阶段：

- 增加可配置 gate
- strict 模式下，没有 verification 不允许 done / archive

**完成标准**

- task 完成和归档流程能感知 verification
- 用户逐步形成“做完任务就补验收”的默认习惯

**当前状态**

- [ ] 未开始

---

### TODO-005：在状态视图中展示 verification 状态

**目标**

让 controller 或用户能在状态面板中看到 task 的验收状态。

**建议展示内容**

- task 是否存在 verification
- verification 结果：pass / partial / fail
- 当前是否满足 archive 条件

**完成标准**

- task 不再只是“状态记录”，而是“交付记录”
- controller 能更直观地判断哪些任务可以收尾

**当前状态**

- [ ] 未开始

---

## P2：中期规划层

### TODO-006：脑暴并规划 roadmap / milestone / phase 层

**目标**

先把 `roadmap / milestone / phase` 作为一个整体规划主题进行脑暴和边界确认，再决定是否拆分成多个独立能力或 skill。

**为什么先合并为一个 todo**

- 这三层概念强相关
- 现在直接拆成三个实现 todo 容易过早承诺结构
- 当前更需要先确认边界、职责和与 task 的关系

**脑暴需要回答的问题**

- roadmap 是否应以独立 skill 形式引入
- milestone 是否需要单独工件层
- phase 是否真的需要，还是可以先停留在规划语义
- roadmap / milestone / phase 与 task 的映射关系应该如何设计
- 这些能力何时进入实现，哪些只停留在文档与规划层

**候选输出**

- 一份 roadmap 规划草案
- 是否引入 milestone 的结论
- 是否引入 phase 的结论
- 若决定实现，再拆分为后续独立 todo

**完成标准**

- 对 roadmap / milestone / phase 有统一边界定义
- 明确哪些内容进入后续实现，哪些暂不做
- 如需实现，再生成新的拆分 todo，而不是现在直接承诺实现

**当前状态**

- [ ] 暂不启动，需先脑暴

---

## 四、推荐推进顺序

### 第 1 步：立即推进

- [ ] TODO-001 修正文档与 README
- [ ] TODO-002 增加规划概念定义章节
- [ ] TODO-003 为 task 增加 `verification.md`

### 第 2 步：短期增强

- [ ] TODO-004 task 状态流转感知 verification
- [ ] TODO-005 状态视图展示 verification 状态

### 第 3 步：中期规划层

- [ ] TODO-006 脑暴并规划 roadmap / milestone / phase 层

---

## 五、当前明确结论

- 现在最重要的是：**文档同步 + `verification.md`**
- roadmap 明确需要，但属于未来规划层
- phase / milestone 也可能需要，但当前先定义概念，不急于实现
- 下一步应该先围绕 P0 项目做计划确认与脑暴，再进入代码实现

---

## 六、下一步脑暴入口

在本 todo 方案基础上，后续建议优先脑暴两个主题：

### 脑暴主题 A：README / 文档如何重写

聚焦：

- 首页叙事如何改
- 当前主流程如何讲清楚
- 历史能力如何降级为兼容说明
- 规划层概念如何放入文档而不让系统显得过重

### 脑暴主题 B：`verification.md` 如何接入 task 工件体系

聚焦：

- 模板字段怎么设计
- 创建时机怎么定
- `done/archive` 如何感知它
- warning 与 strict 怎么演进
- 状态视图如何展示它
