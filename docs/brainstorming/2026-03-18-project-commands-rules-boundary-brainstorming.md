# Brainstorming: Rules 生成边界与 `project-commands.md` 收口

日期：2026-03-18
角色：general strategist

## 问题陈述

当前 `agent-team init` 对 `.agents/rules/` 的生成边界不够清晰：

- `debugging.md`、`communication.md`、`context-management.md`、`task-protocol.md`、`worktree.md` 这类平台内置协议，适合稳定模板生成
- `build-verification.md` 这类项目级执行规则，本质上依赖当前仓库的脚本、命令和验证方式，不适合由 CLI 内部脚本逻辑直接硬编码推断

现状把两类规则都放进了 `init` 的同一套生成流程里，容易产生两个问题：

1. 文档与实现边界混淆，看起来像“一整套模板规则”
2. 项目级执行协议被错误地当成产品内置协议处理

用户希望把这两类规则明确拆开：

- 平台协议规则继续走稳定模板
- 项目命令规则改成 AI 基于当前项目上下文生成
- 这份项目规则在主控侧生成，再由 worker 共享使用

## 目标

- 重新定义 `.agents/rules/` 中静态规则与动态规则的边界
- 用 `project-commands.md` 替代现有 `build-verification.md`
- 让 `project-commands.md` 明确表达“当前项目所有脚本与命令的执行方案”
- 保持 worker 共享主控生成的项目规则
- 不继续把 role 的 `system.md` 往拆分独立 rules 文件方向扩展

## 约束与边界

- 本轮只收口设计，不直接实现
- 主控 rules 体系继续保留
- 固定规则仍由 Go 内置模板生成，不迁移到仓库外部模板目录
- `project-commands.md` 由 AI 生成，不考虑“无 AI provider”回退路径
- `init` 与 `rules sync` 都支持更新 `project-commands.md`
- `rules sync` 更新 `project-commands.md` 时，无条件重生成并覆盖
- worker 不单独生成项目命令规则，直接复用主控生成内容
- worker 侧不继续拆 role 的 `system.md` 成更多独立 rules 文件

## 当前实现观察

### 1. 固定规则目前是代码内置文本

当前 `index.md`、`debugging.md`、`communication.md`、`context-management.md`、`task-protocol.md`、`worktree.md` 来自 `internal/init.go` 中的 `defaultRuleFiles`。

这说明它们本质上已经是“产品内置协议规则”。

### 2. `build-verification.md` 当前是动态扫描生成

当前 `build-verification.md` 由 CLI 扫描 `Makefile`、`go.mod`、`package.json` 后生成。

这比静态模板更接近项目语义，但仍然把“如何理解项目命令”硬编码在 CLI 中，扩展性和准确性都受限。

### 3. worker 已支持共享项目根下的拆分 rules

当前 worker 注入逻辑已经支持在 `.agents/rules/` 存在时走 slim mode，并引用 `.agents/rules/index.md`。

因此本轮不需要再为 worker 新设计一套独立 rules 机制。

## 候选方案

### 方案 A：保持现状，只补文档说明

做法：

- 保留 `build-verification.md`
- 保留 CLI 对项目脚本的直接扫描生成
- 只在文档中解释静态规则与动态规则的区别

优点：

- 改动最小
- 兼容性最好

缺点：

- 根本边界没有修正
- `build-verification` 仍然被误放在产品内置协议体系里
- CLI 继续承担项目理解职责

### 方案 B：固定规则模板化，项目命令规则 AI 生成

做法：

- 固定规则继续由 Go 内置模板生成
- 删除 `build-verification.md` 的现有定位
- 新增 `project-commands.md`
- `init` 与 `rules sync` 均由 AI 基于当前项目上下文无条件生成并覆盖 `project-commands.md`
- 失败后由 AI 侦查正确命令，并询问用户是否回写更新该规则文件

优点：

- 产品协议与项目协议边界清晰
- 项目命令规则更贴近真实仓库
- 主控生成一次，worker 直接共享，维护成本低

缺点：

- 引入 AI 生成链路，测试方式需要调整
- `rules sync` 的语义需要重新定义

### 方案 C：项目规则完全由 worker 自治

做法：

- 主控只生成固定规则
- 每个 worker 启动后自行分析项目并生成自己的命令规则

优点：

- worker 自治能力强

缺点：

- 同一项目可能出现多份不一致规则
- 成本高，重复生成
- 与“主控作为项目规则唯一事实源”的目标相反

## 推荐方案

推荐方案 B。

原因：

1. 固定协议和项目命令协议是两类不同问题，必须分层
2. 项目命令规则应由项目上下文驱动，而不是 CLI 里写死扫描逻辑
3. 主控统一生成后共享给 worker，既稳定又避免重复
4. worker 现有 slim mode 已足够承接这套共享规则

## 推荐设计

### 1. 规则体系重新分层

`.agents/rules/` 分为两类：

- 平台内置规则
  - `index.md`
  - `debugging.md`
  - `communication.md`
  - `context-management.md`
  - `task-protocol.md`
  - `worktree.md`
- 项目派生规则
  - `project-commands.md`

其中：

- 平台内置规则由 Go 内置模板稳定生成
- 项目派生规则由 AI 基于当前项目上下文生成

### 2. `build-verification.md` 改名为 `project-commands.md`

新文件语义：

- 不再只表达 build / test / verify
- 表达当前项目所有脚本与命令的执行方案
- 覆盖 `build`、`test`、`lint`、`dev`、`e2e`、`format`、`codegen`、数据库迁移等命令入口

需要明确写进文档的规则包括：

- AI 在运行任何项目命令前，必须先读取 `project-commands.md`
- 如果按规则执行命令失败，AI 不得直接放弃或盲目重试
- AI 必须继续侦查正确命令、正确工作目录、必要前置条件或替代入口
- AI 在确认存在规则偏差后，需要询问用户是否更新 `project-commands.md`

### 3. `index.md` 纳入 `project-commands.md`

`index.md` 需要新增一条一等入口规则：

- `project-commands.md`: before running any project command

这样主控与 worker 在进入命令执行阶段时，都能从索引层面得到一致指引。

### 4. `init` 行为

`agent-team init` 应分成三类动作：

1. 初始化项目结构
2. 生成固定规则
3. 生成 `project-commands.md`

其中第三步不再依赖 CLI 的静态扫描拼文案，而是：

- 收集当前项目上下文
- 由主控当前使用的 provider 驱动 AI 生成
- 无条件覆盖 `project-commands.md`

### 5. `rules sync` 行为

`agent-team rules sync` 的语义需要更新为：

- 同步固定 rules 到 provider 文件引用
- 重新生成 `project-commands.md`
- 覆盖已有内容

这里不再使用“检测是否变化再决定是否生成”的策略，而是无条件重生成并覆盖。

### 6. worker 继承方式

worker 不需要独立生成项目命令规则。

worker 侧策略为：

- 继续使用固定 role `system.md`
- 不继续把 role 的 `system.md` 拆成独立 rules 文件
- 继续共享项目根下 `.agents/rules/`
- 通过现有 slim mode 读取 `.agents/rules/index.md`
- 按索引规则加载 `project-commands.md`

因此可以把 worker 设计表述为：

- worker 共享主控生成的项目规则
- worker 不拥有独立的项目命令规则事实源

## 数据流建议

推荐数据流如下：

1. 主控执行 `agent-team init`
2. 系统生成固定 rules
3. 系统调用当前 provider 生成 `project-commands.md`
4. 主控 provider 文件更新 rules 引用
5. worker 创建或打开时，注入固定 role prompt
6. worker 通过项目根 `.agents/rules/index.md` 感知项目规则
7. worker 在运行任何项目命令前读取 `project-commands.md`

## 风险与缓解

### 风险 1：AI 生成质量不稳定

影响：

- `project-commands.md` 可能遗漏命令或写错执行入口

缓解：

- 明确其为“可修正的项目命令事实文档”
- 运行失败后要求 AI 继续侦查
- 在侦查成功后询问用户是否回写更新规则

### 风险 2：用户误以为固定规则也会被 AI 改写

影响：

- 用户对规则来源失去信心

缓解：

- 文档中明确区分“固定规则”和“项目派生规则”
- 在文件头或命令输出中明确说明 `project-commands.md` 是 AI 生成内容

### 风险 3：worker 与主控对规则理解不一致

影响：

- worker 仍按旧习惯直接执行命令

缓解：

- 保持 worker 继续共享主控 `.agents/rules/`
- 在 `index.md` 中把 `project-commands.md` 提升为一等规则入口

## 验证策略

设计落地后应验证：

1. `init` 后是否生成：
   - 固定 rules
   - `project-commands.md`
   - provider 文件引用
2. `index.md` 是否包含：
   - `project-commands.md: before running any project command`
3. `rules sync` 是否会：
   - 无条件重生成并覆盖 `project-commands.md`
4. worker 在 slim mode 下是否仍能：
   - 读取 `.agents/rules/index.md`
   - 通过索引加载 `project-commands.md`
5. 当命令执行失败时，AI 是否按协议：
   - 继续侦查正确命令
   - 询问用户是否更新该规则文件

## 实施建议

建议后续实施顺序：

1. 先重命名规则概念：`build-verification.md` -> `project-commands.md`
2. 再重写 `index.md` 对该规则的触发描述
3. 再调整 `init` / `rules sync` 的生成边界
4. 最后补充主控与 worker 的引用文档说明

## 开放问题

本轮已确认的设计决策如下，无新增开放问题：

- 固定规则继续使用 Go 内置模板
- 项目命令规则由 AI 生成
- 不考虑无 AI provider 回退
- `rules sync` 无条件重生成并覆盖
- worker 共享主控规则，但不继续拆 role `system.md`

## 最终结论

最合适的方案是把 `.agents/rules/` 明确拆成两层：

- 一层是 `agent-team` 的平台内置协议规则，由 Go 模板稳定生成
- 一层是当前项目的命令执行规则，由 AI 生成 `project-commands.md`

同时：

- `project-commands.md` 进入 `index.md` 成为一等规则项
- `init` 与 `rules sync` 都负责生成并覆盖它
- worker 不自行生成该规则，只共享主控产出的项目级规则
- role 的 `system.md` 不再继续向拆分独立 rules 文件方向演进

这样可以把“平台协议”和“项目命令事实”彻底分层，减少边界混乱，并让后续规则维护更贴近真实项目。
