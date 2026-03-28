---
name: role-creator
description: 创建和管理 AI 角色定义（JSON 格式）。当用户需要：(1) 创建新角色（如"前端架构师"），(2) 列出可用角色，(3) 查看角色详情，(4) 删除角色，(5) 验证角色，(6) 从全局模板导入角色到本地时触发。
---

# Role Creator

## 概述

Role Creator 用于创建和管理 AI 角色定义。每个角色是一个独立的 JSON 文件，包含职责边界（inScope/outOfScope）、系统提示词（prompt）和关联的 skills。

## 核心概念

- **角色文件**：单一 JSON 文件，包含 name、fileName、version、description、prompt、inScope、outOfScope、skills
- **存储位置**：本地项目 `.otcc/roles/` 和全局插件 `~/.claude/plugins/marketplaces/.otcc/roles/`
- **优先级**：本地项目 > 全局模板库

## 工作流程

### 创建角色标准流程（5 步）

#### Step 1: 标准化输入

- 验证 `fileName` 必须为 kebab-case 格式：`/^[a-z0-9]+(-[a-z0-9]+)*$/`
- 若不符合，提示规范化版本并确认

#### Step 2: AI 生成角色字段

AI 自动生成以下字段（不主动询问用户）：

| 字段 | 说明 |
|------|------|
| description | 角色描述 |
| prompt | 系统提示词 |
| inScope | 职责范围内的工作（逗号分隔） |
| outOfScope | 职责范围外的工作（逗号分隔） |

**提交流审批**：
1. AI 自动生成上述字段
2. 展示给用户审批
3. **用户拒绝或 AI 置信度低** → 调用 `/brainstorming` 技能进行结构化细化
4. 审批通过后进入 Step 3

**Brainstorming 触发条件**：
- 用户拒绝 AI 生成的字段
- AI 置信度低（范围模糊、边界冲突、目标不明确）

#### Step 3: 选择 Skills

1. **AI 推荐**：`find-skills` 查询相关技能作为推荐候选
2. 展示推荐列表，用户选择
3. 询问手动添加额外技能
4. **技能标识符优先级**：
   - 优先使用完整远程标识符（如 `jsonlee12138/prompts@design-patterns-principles`）
   - 仅在无远程匹配时使用本地短名称
5. 合并、去重后确认最终技能列表
6. 技能可为空 → 持久化为 `skills: []`

**回退逻辑**：`find-skills` 不可用时，直接询问手动添加。

#### Step 4: 执行 CLI

使用 `npx otcc role create` 命令创建角色（见下方 CLI 命令参考）

#### Step 5: 验证输出

- 验证 JSON 格式正确（Zod Schema）
- 验证 `fileName` 符合 kebab-case
- 验证 skills 路径可解析性

---

## CLI 命令

所有命令使用 `npx otcc` 前缀：

### otcc role create

创建新角色：

```bash
npx otcc role create --name "前端架构师" --file-name "frontend-architect" --description "负责前端技术选型"
npx otcc role create --interactive    # 交互式创建（引导式流程）
```

**参数**：

| 参数 | 说明 |
|------|------|
| `--name`, `-n` | 角色名称 |
| `--file-name` | 文件名 (kebab-case) |
| `--description`, `-d` | 角色描述 |
| `--prompt`, `-p` | 系统提示词 |
| `--in-scope`, `-i` | 职责范围内的工作（逗号分隔） |
| `--out-of-scope`, `-o` | 职责范围外的工作（逗号分隔） |
| `--skills`, `-s` | 关联的 skills（逗号分隔，格式：`owner/repo@suffix`） |
| `--interactive`, `-I` | 交互式创建模式 |
| `--global`, `-g` | 保存到全局目录 |

### otcc role list

列出所有角色：

```bash
npx otcc role list [--local] [--global]
```

### otcc role show

查看角色详情：

```bash
npx otcc role show <fileName>
```

### otcc role delete

删除角色：

```bash
npx otcc role delete <fileName> [--local] [--global]
```

### otcc role validate

验证角色文件：

```bash
npx otcc role validate <fileName>
```

### otcc role import

从全局模板导入角色到本地：

```bash
npx otcc role import <fileName>
```
