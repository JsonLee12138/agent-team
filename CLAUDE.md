# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Type

本项目是一个 Claude Plugin 项目，主要以 **提示词工程** 为主，**CLI 为辅**。

## Skill 规则

- **写提示词（Prompt）时**：必须调用 `prompt-engineering-patterns` skill，遵循其中的最佳实践
- **写代码时, 架构设计时**：必须遵循 `design-patterns-principles` skill 的设计模式和原则
- **写cli**: 必须使用 `oclif-patterns` skill, 必须参考 `./docs/cli-development-spec.md` 规范

## Project Status

TypeScript + Bun CLI 项目，已完成基础搭建。

## 项目结构

```
otcc/
├── bin/              # 编译输出目录 (npx 入口)
├── src/              # 源代码
│   └── index.ts      # CLI 入口
├── package.json
└── tsconfig.json
```

## Commands

| 命令 | 说明 |
|------|------|
| `bun run dev` | 开发模式运行 |
| `bun run build` | 构建到 bin/ |
| `bun run typecheck` | TypeScript 类型检查 |
| `npx .` | 直接运行当前项目 |
