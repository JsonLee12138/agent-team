# 阮一峰科技爱好者周刊 投稿

> **投稿方式**: 在 [ruanyf/weekly](https://github.com/ruanyf/weekly) 仓库提交 Issue
> **投稿分类**: 工具
> **语言**: 中文

---

## 投稿标题

agent-team — 用 Git Worktree 编排多 AI 代理的开发团队

## 投稿正文

[agent-team](https://github.com/JsonLee12138/agent-team) 是一个开源的多 AI 代理编排工具，采用 **角色（Role）+ 工人（Worker）** 模型，让多个 AI 代理（Claude、Gemini、Codex、OpenCode）在隔离的 Git Worktree 中并行工作，互不干扰。

**核心特性：**

- **角色系统**：内置 11 个预定义角色（产品经理、前端架构师、React 开发者、UI 设计师、增长运营等），也可自定义角色
- **隔离工作区**：每个 AI 代理拥有独立的 Git 分支、工作目录和终端会话，彻底避免合并冲突
- **质量门禁**：7 个生命周期钩子，包括"头脑风暴门禁"——要求代理在写代码前必须先提交设计文档
- **多平台支持**：同一套角色可在 Claude Code、Gemini CLI、OpenCode、OpenAI Codex 上运行，无供应商锁定
- **自然语言驱动**：通过对话即可创建工人、分配任务、合并代码

**安装方式：**

```bash
# Homebrew
brew tap JsonLee12138/agent-team && brew install agent-team

# Go
go install github.com/JsonLee12138/agent-team@latest

# AI 原生安装（推荐）
npx skills add JsonLee12138/agent-team -a claude -y
```

**使用场景：**

1. 让前端架构师和后端开发者同时工作，各自在独立分支上开发
2. 用产品经理角色拆解需求，再分发给多个开发者角色执行
3. 同一个任务分别让 Claude 和 Gemini 执行，对比输出质量

Go 编写，MIT 协议，支持中英双语。

---

## 备选短版本（如周刊篇幅有限）

[agent-team](https://github.com/JsonLee12138/agent-team)：一个用 Git Worktree 隔离多个 AI 代理的开源编排工具。内置 11 个专业角色（前端架构师、产品经理等），支持 Claude/Gemini/Codex/OpenCode 四大平台，带有"头脑风暴门禁"等质量管控钩子。Go 编写，中英双语，MIT 协议。

---

## 投稿 Issue 模板

```
标题: 工具 -- agent-team，用 Git Worktree 编排多 AI 代理的开发团队

正文:
https://github.com/JsonLee12138/agent-team

agent-team 是一个多 AI 代理编排工具，让 Claude、Gemini、Codex 等代理在隔离的 Git Worktree 中并行开发。内置 11 个预定义角色，7 个生命周期钩子，支持"头脑风暴门禁"（代理写代码前必须先出设计文档）。Go 编写，MIT 协议。
```
