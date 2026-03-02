# OpenCode / Codex Skills 扩展方案调研

更新时间：2026-03-02（美国时区）

## 1) OpenCode：插件能力概览

OpenCode 支持通过插件扩展运行时行为，主要能力如下：

- 插件加载来源
  - 本地目录：`.opencode/plugins/`、`~/.config/opencode/plugins/`
  - npm 包：在 `opencode.json` 的 `plugin` 字段声明
- 插件开发
  - JavaScript / TypeScript
  - TypeScript 可使用 `@opencode-ai/plugin` 类型
- 可用能力（事件/钩子）
  - 事件订阅：`session.*`、`file.*`、`message.*`、`permission.*`、`lsp.*`、`tui.*` 等
  - 工具拦截：`tool.execute.before`、`tool.execute.after`
  - Shell 注入：`shell.env`
  - 自定义工具：`tool(...)` + schema + execute
  - 结构化日志：`client.app.log()`
- 依赖管理
  - 在 `.opencode/package.json` 声明依赖，启动时自动 `bun install`

结论：OpenCode 的插件机制足够做“skills 目录桥接/同步器”。

---

## 2) OpenCode：通过插件扩展 skills 目录（方案）

### 2.1 约束与思路

- OpenCode skills 有固定扫描目录（`.opencode/skills`、`~/.config/opencode/skills`、`.claude/skills`、`.agents/skills` 等）。
- 配置里没有单独的 `skillsDir` 列表字段。
- 因此可行方案是：**插件把外部 skills 映射（symlink）或同步到官方扫描目录**。

建议采用：

- 外部目录：`/path/to/external-skills/<skill-name>/SKILL.md`
- 目标目录：`.opencode/skills/`
- 映射方式：符号链接（symlink），不复制内容，便于统一维护。

### 2.2 目录结构约定

```text
/path/to/external-skills/
  skill-a/
    SKILL.md
  skill-b/
    SKILL.md
```

### 2.3 插件示例（skills bridge）

文件：`.opencode/plugins/skills-bridge.ts`

```ts
import { promises as fs } from "node:fs"
import path from "node:path"
import type { Plugin } from "@opencode-ai/plugin"

// 通过环境变量传入外部 skills 根目录，支持多个目录（冒号分隔）
// 例如：OPENCODE_EXTRA_SKILLS_DIRS="/opt/skills:/Users/me/skills"
export const SkillsBridgePlugin: Plugin = async ({ directory, client }) => {
  const raw = process.env.OPENCODE_EXTRA_SKILLS_DIRS ?? ""
  const roots = raw.split(":").map((s) => s.trim()).filter(Boolean)
  const targetRoot = path.join(directory, ".opencode", "skills")

  await fs.mkdir(targetRoot, { recursive: true })

  for (const root of roots) {
    let entries: Array<import("node:fs").Dirent> = []
    try {
      entries = await fs.readdir(root, { withFileTypes: true })
    } catch {
      continue
    }

    for (const entry of entries) {
      if (!entry.isDirectory()) continue
      const src = path.join(root, entry.name)
      const skillFile = path.join(src, "SKILL.md")
      try {
        await fs.access(skillFile)
      } catch {
        continue
      }

      const dest = path.join(targetRoot, entry.name)
      try {
        await fs.lstat(dest)
        continue
      } catch {
        // dest 不存在，继续创建 symlink
      }

      try {
        await fs.symlink(src, dest, "dir")
      } catch (err) {
        await client.app.log({
          body: {
            service: "skills-bridge",
            level: "warn",
            message: "failed to link skill",
            extra: { src, dest, err: String(err) },
          },
        })
      }
    }
  }

  return {}
}
```

`opencode.json`：

```json
{
  "$schema": "https://opencode.ai/config.json",
  "plugin": ["./.opencode/plugins/skills-bridge.ts"]
}
```

启动前设置环境变量：

```bash
export OPENCODE_EXTRA_SKILLS_DIRS="/path/to/external-skills"
opencode
```

### 2.4 注意事项

- 若 skills 在启动后才被映射，可能需要重启 OpenCode 才稳定可见。
- skill 名字冲突时，建议统一命名规范（如前缀 `team-`）。
- 优先用 symlink 而非复制，避免内容漂移。
- 如果你不依赖插件生命周期，也可以直接用启动脚本预先建立 symlink（更确定）。

---

## 3) Codex：扩展 skills 目录方案

Codex 没有 OpenCode 那种插件 hook 体系；扩展 skills 通常用以下两种方式：

### 3.1 方案 A（推荐）：symlink 到官方扫描目录

Codex 原生会扫描：

- 仓库链路上的 `.agents/skills`（从当前目录向上到 repo root）
- 用户目录：`$HOME/.agents/skills`
- 管理员目录：`/etc/codex/skills`

并且官方说明支持 **symlinked skill folders**。  
所以最稳妥做法是把外部 skills 目录软链进 `.agents/skills`。

示例：

```bash
mkdir -p .agents/skills
ln -s /path/to/external-skills/my-skill .agents/skills/my-skill
```

### 3.2 方案 B：`~/.codex/config.toml` 用 `[[skills.config]]`

可为 skill 做显式配置（启用/禁用与路径）：

```toml
[[skills.config]]
path = "/path/to/skill-folder"
enabled = true
```

补充说明：

- Config Reference 写的是 `path` 指向“包含 `SKILL.md` 的 skill 目录”。
- Skills 页示例曾出现 `path = "/path/to/skill/SKILL.md"`。
- 实际落地建议优先用“目录路径”写法，并在本机验证一次。

---

## 4) 落地建议（OpenCode + Codex 共存）

如果你希望“一份外部 skills，两个客户端共用”，建议：

1. 统一维护外部目录：`/path/to/external-skills/*/SKILL.md`
2. Codex：直接 symlink 到 `.agents/skills`
3. OpenCode：
   - 简单场景：同样 symlink 到 `.opencode/skills`
   - 需要自动化场景：使用上面的 `skills-bridge` 插件自动映射
4. 建立 CI/脚本校验：每个 skill 必须包含 `SKILL.md` 且目录名与 skill 名一致

---

## 参考文档

- OpenCode Plugins: https://opencode.ai/docs/plugins/
- OpenCode Config: https://opencode.ai/docs/config/
- OpenCode Skills: https://opencode.ai/docs/skills/
- Codex Skills: https://developers.openai.com/codex/skills
- Codex Config Reference: https://developers.openai.com/codex/config-reference
