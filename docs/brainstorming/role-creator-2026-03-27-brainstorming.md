# role-creator Skill 设计方案

**日期**: 2026-03-27
**角色**: general strategist

---

## 1. 问题陈述

当前 Claude Code 的 agent-team 缺乏角色创建机制。用户希望能够定义 AI 角色（如"前端架构师"），明确其任务边界、所需技能（skills），并能跨项目复用。

核心需求：
- 定义角色的职责边界（inScope / outOfScope）
- 为角色绑定所需的 skills（含路径和描述）
- 支持本地项目和全局插件两个存储位置
- 角色即 JSON，通过 skill/prompt 读取使用

---

## 2. 目标

设计一个 `role-creator` skill，实现：
1. 创建角色定义（JSON 格式，Zod 校验）
2. 管理角色生命周期（创建、查询、列出、删除）
3. 支持本地和全局两种存储位置
4. CLI + 对话混合模式

---

## 3. 数据模型

### 角色定义文件结构

```json
{
  "name": "前端架构师",
  "fileName": "frontend-architect",
  "version": "1.0.0",
  "description": "负责前端技术选型、架构设计、代码审查",
  "prompt": "你是一位资深前端架构师，擅长 React 生态...",
  "inScope": [
    "技术方案评审",
    "前端架构设计",
    "性能优化指导"
  ],
  "outOfScope": [
    "后端 API 设计",
    "数据库选型",
    "运维部署"
  ],
  "skills": [
    {
      "path": "pexoai/pexo-skills@pexoai-agent",
      "name": "pexoai-agent",
      "description": "前端架构设计助手，擅长 React 生态"
    }
  ]
}
```

### 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| name | string | 展示名（中文） |
| fileName | string | 文件名（kebab-case） |
| version | string | 版本号，语义化版本 |
| description | string | 角色简介 |
| prompt | string | 系统提示词，AI 直接使用 |
| inScope | string[] | 职责范围内的工作 |
| outOfScope | string[] | 职责范围外的工作 |
| skills | Skill[] | 关联的 skills 列表 |

### Skill 结构

| 字段 | 类型 | 说明 |
|------|------|------|
| path | string | skills 完整路径，格式：`owner/repo@suffix` |
| name | string | skill 展示名 |
| description | string | skill 描述，AI 据此判断何时使用 |

### Zod Schema

```typescript
import { z } from 'zod'

export const SkillSchema = z.object({
  path: z.string(),
  name: z.string(),
  description: z.string()
})

export const RoleSchema = z.object({
  name: z.string(),
  fileName: z.string().regex(/^[a-z0-9]+(-[a-z0-9]+)*$/),
  version: z.string(),
  description: z.string(),
  prompt: z.string(),
  inScope: z.array(z.string()),
  outOfScope: z.array(z.string()),
  skills: z.array(SkillSchema)
})
```

---

## 4. 目录结构

### 全局模板库
```
~/.claude/plugins/marketplaces/agent-team/roles/
├── frontend-architect.json
└── backend-architect.json
```

### 本地项目角色
```
<project>/.claude/roles/
└── frontend-architect.json
```

**优先级**：本地项目 > 全局模板库

---

## 5. 使用方式

### 读取角色

任意 skill 或 prompt 都可以读取角色 JSON：
- 读取 `prompt` 字段作为系统提示词
- 读取 `skills` 字段获取关联的 skills
- 读取 `inScope` / `outOfScope` 明确职责边界

### Skills 搜索顺序

1. **本地项目** `.claude/skills/`
2. **全局插件目录** `~/.claude/plugins/*/skills/`
3. **网络搜索** 通过 `find-skills` 工具搜索远程 skills

---

## 6. 交互方式

### 混合模式

| 场景 | 交互方式 |
|------|----------|
| 简单需求 | CLI 指令，一句话创建 |
| 复杂需求 | 对话引导，逐步确认 |

**CLI 指令示例**：
```bash
role-creator create --name "前端架构师" --desc "..." --skills skill1 skill2
role-creator create --interactive
```

**对话引导流程**：
1. 询问角色名称（自动生成 fileName）
2. 询问角色描述 description
3. 询问系统提示词 prompt（或自动生成）
4. 询问 inScope / outOfScope
5. 询问需要的 skills（自动搜索推荐）
6. 确认并保存

---

## 7. 核心命令

| 命令 | 说明 |
|------|------|
| `role-creator create` | 创建新角色（CLI 或对话模式） |
| `role-creator list` | 列出所有可用角色 |
| `role-creator show <fileName>` | 查看角色详情 |
| `role-creator delete <fileName>` | 删除角色 |
| `role-creator import <fileName>` | 从全局模板导入到本地 |

---

## 8. 实现要点

### 单一文件原则
每个角色就是一个 JSON 文件，比 agent-team 的 3 文件方案更简洁灵活。

### Zod 类型校验
使用 Zod 做运行时类型校验，确保数据合法性。

### 职责边界分离
`inScope` / `outOfScope` 明确角色职责范围，AI 不会超范围工作。

### 不支持继承
每个角色独立完整定义，避免复杂的继承链。

---

## 9. 验证策略

1. 创建角色后验证 JSON 格式正确（Zod 校验）
2. 验证 fileName 符合 kebab-case
3. 测试本地/全局优先级是否正确
4. 测试 skills 路径可解析性

---

## 10. 后续步骤

实现 `role-creator` skill，主要工作：
1. 设计 skill 目录结构和入口文件
2. 实现 Zod schema 定义
3. 实现角色 CRUD 操作
4. 实现对话引导流程
5. 添加测试用例
