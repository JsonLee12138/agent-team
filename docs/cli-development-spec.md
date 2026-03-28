# OTCC CLI 开发规范

基于当前 `role` 模块实现整理，适用于本项目后续 CLI 命令开发。

## 1. 总体原则

- CLI 基于 `@oclif/core` 开发。
- 命令按主题拆分到 `src/commands/<topic>/` 目录。
- 一个文件只实现一个命令类，默认导出一个 `Command` 子类。
- 优先保持命令职责单一，不在一个命令中混入过多业务分支。
- CLI 负责参数解析、交互、输出；业务规则下沉到 `core` 层。

## 2. 目录约定

以当前 `role` 模块为基准：

```text
src/
├── index.ts                 # CLI 入口
├── commands/
│   └── role/
│       ├── index.ts         # topic 入口
│       ├── create.ts        # 子命令
│       ├── list.ts
│       ├── show.ts
│       ├── validate.ts
│       ├── delete.ts
│       └── import.ts
├── prompts/
│   └── role.ts              # 交互式提问
└── core/
    └── role/
        └── schema.ts        # schema、路径、读写、校验等业务能力
```

约定：

- `commands/`：只放命令入口和命令编排。
- `prompts/`：放交互式输入逻辑。
- `core/`：放可复用的领域逻辑、schema、文件系统操作、校验逻辑。

## 3. 命名规范

### 3.1 命令文件命名

- 文件名使用 kebab-case 或简单英文动词，和命令名一致。
- 例如：`create.ts`、`list.ts`、`show.ts`。

### 3.2 类名命名

- 使用 `主题 + 动作` 的 PascalCase 命名。
- 例如：`RoleCreate`、`RoleList`、`RoleShow`。

### 3.3 topic 入口命名

- `src/commands/role/index.ts` 这类文件作为 topic 占位入口。
- 作用是显示该主题的 help，而不是承载业务逻辑。

## 4. 命令结构规范

每个命令文件建议遵循以下结构：

1. import
2. `export default class Xxx extends Command`
3. `static description`
4. `static examples`
5. `static flags`
6. `static args`
7. `async run()`

参考模式：

```ts
export default class XxxCommand extends Command {
  static description = '命令描述'

  static examples = [
    '$ otcc topic action',
  ]

  static flags = {
    verbose: Flags.boolean({ char: 'v', description: '显示详细信息' }),
  }

  static args = {
    name: Args.string({ required: true, description: '名称' }),
  }

  async run(): Promise<void> {
    const { flags, args } = await this.parse(XxxCommand)
    // 执行逻辑
  }
}
```

## 5. Flags / Args 规范

### 5.1 flags

- 所有 flag 都必须写 `description`。
- 常用 flag 应提供 `char` 短参数。
- 布尔开关统一使用 `Flags.boolean()`。
- 字符串输入统一使用 `Flags.string()`。
- 不要定义无实际用途的 flag。

参考现状：

- `role create` 使用 `--name`、`--description`、`--prompt`、`--interactive`、`--global`、`--local`
- `role list` 使用 `--global`、`--local`

### 5.2 args

- 位置参数用于命令的核心对象标识。
- 如查看、删除、验证单个资源时，优先用 args。
- 参数必须带 `description`，必要时标记 `required: true`。

例如：

- `role show <name>`
- `role delete <name>`
- `role validate <name>`

### 5.3 flag 与 arg 的使用边界

- “资源标识”优先用 arg。
- “行为修饰项”优先用 flag。
- 同一语义不要重复提供过多入口，除非确实为了兼容交互体验。

## 6. 交互式命令规范

当命令存在较多输入项时，支持交互式模式。

### 6.1 适用场景

- 输入字段较多。
- 用户不容易一次性记住全部 flags。
- 存在默认值、引导式填写需求。

### 6.2 实现方式

- 统一把交互逻辑放到 `src/prompts/<topic>.ts`。
- 命令文件中只负责判断是否进入交互模式。
- 使用 `inquirer` 实现输入、确认等交互。

当前 `role create` 的模式可作为标准：

- 显式传入 `--interactive`
- 或未提供核心参数时自动进入交互模式

### 6.3 交互函数拆分

交互函数按字段拆分为小函数：

- `promptRoleName()`
- `promptFileName()`
- `promptDescription()`
- `promptVersion()`
- `promptPrompt()`
- `promptList()`
- `promptSkills()`

规范要求：

- 一类问题对应一个独立函数。
- 函数返回值类型明确。
- 聚合入口函数统一返回结构化对象，如 `promptRoleInteractive()`。

## 7. 业务逻辑分层规范

CLI 层不要直接承载复杂领域逻辑。

### 7.1 Command 层职责

- 解析 flags / args
- 决定调用哪段业务逻辑
- 控制交互流程
- 输出结果和错误

### 7.2 Core 层职责

- schema 定义
- 数据校验
- 路径解析
- 文件读写封装
- 业务模型转换

当前 `role` 模块已经体现出该分层：

- `commands/role/*.ts`：命令入口
- `prompts/role.ts`：交互输入
- `core/role/schema`：路径、校验、读取、转换

后续新模块应保持同样结构。

## 8. 输出规范

### 8.1 成功输出

- 输出简洁、可读。
- 对用户最重要的信息放在首行。
- 可以使用简短符号增强识别，但不要过度。

当前示例：

- `✅ 角色已保存: ...`
- `✅ 已导入: ...`
- `✅ 角色验证通过: ...`

### 8.2 空结果输出

- 不要静默返回。
- 明确告诉用户未找到内容。

例如：

- `没有找到任何角色`

### 8.3 列表输出

- 列表类命令统一先给标题，再逐项输出。
- 每个条目保持固定字段顺序。

当前 `role list` 顺序可参考：

1. 名称 / fileName
2. 描述
3. 路径
4. skills 数量

### 8.4 错误输出

- 统一使用 `this.error()` 抛出错误。
- 错误信息必须可执行、可定位，避免空泛描述。

例如：

- `未找到角色: ${args.name}`
- `角色数据验证失败: ${e}`

## 9. 文件系统操作规范

涉及本地文件时遵循以下规则：

- 写入前先确保目录存在，例如 `mkdir(dir, { recursive: true })`
- 覆盖已有文件前必须明确确认
- 删除前必须先检查目标是否存在
- 本地与全局目录的优先级要一致且可预测

当前 `role` 模块约定：

- 默认优先本地目录
- `--global` 显式操作全局目录
- 本地 / 全局双目录行为要清晰，不要隐式修改两个位置

## 10. 校验规范

- 所有可持久化数据必须做 schema 校验。
- 校验逻辑统一放 `core` 层，不放在命令层手写分散判断。
- 校验失败时要直接中止命令。

当前标准做法：

- `validateRole()` 负责运行时校验
- `parseSkillInput()` 负责输入格式转换
- `kebabCase()` 负责标准化文件名

## 11. Topic 设计规范

每个主题命令都应提供 topic 入口，如 `role`。

要求：

- topic 入口只负责展示 help
- 不在 topic 入口中处理实际业务
- 便于用户执行 `otcc role` 时看到完整帮助

`src/commands/role/index.ts` 可作为标准实现。

## 12. 新命令开发流程

后续新增 CLI 命令时，建议按以下步骤实现：

1. 确认命令属于哪个 topic
2. 在 `src/commands/<topic>/` 下新增命令文件
3. 明确 `description`、`examples`、`flags`、`args`
4. 如需交互，新增 `src/prompts/<topic>.ts` 中对应函数
5. 将 schema、路径、读写、校验逻辑放入 `src/core/<topic>/`
6. 在 `run()` 中只做编排，不堆叠业务细节
7. 补充成功、空结果、错误三类输出
8. 至少执行一次 `bun run typecheck`

## 13. 针对当前项目的补充约束

结合本项目现状，新增 CLI 命令还应遵守：

- 保持 TypeScript ESM 风格，使用 `import`
- 保持与现有代码一致的简洁风格，不引入不必要抽象
- 优先复用已有 `core` 能力，避免命令间复制文件读写逻辑
- 优先新增子命令，不要把多个动作硬塞进单一命令
- 若命令以提示词工程为核心，CLI 只负责组织输入输出，不负责承载大段 prompt 内容的业务判断

## 14. 推荐模板

新增命令时，可直接参考下面骨架：

```ts
import { Command, Flags, Args } from '@oclif/core'
import { someDomainAction } from '../../core/<topic>/xxx'

export default class TopicAction extends Command {
  static description = '命令说明'

  static examples = [
    '$ otcc <topic> <action>',
  ]

  static flags = {
    example: Flags.string({ char: 'e', description: '示例参数' }),
  }

  static args = {
    name: Args.string({ required: true, description: '资源名称' }),
  }

  async run(): Promise<void> {
    const { flags, args } = await this.parse(TopicAction)

    const result = await someDomainAction({
      name: args.name,
      example: flags.example,
    })

    this.log(result)
  }
}
```

## 15. 现有 role 模块可作为基线的点

建议后续 CLI 全部对齐以下基线：

- `index.ts` 只做 topic help
- `create` 负责创建，支持 flags + interactive 双模式
- `list` 负责列表浏览
- `show` 负责单项详情
- `validate` 负责结构校验
- `delete` 负责删除
- `import` 负责从全局到本地的同步

这套结构已经形成一个比较清晰的 CRUD + validate + import 模式，后续其它主题命令可直接复用该组织方式。
