# Context Cleanup Checklist

## Hard gates

- [ ] controller 路径先读 `.agent-team/rules/index.md`
- [ ] worker 路径先读 `worker.yaml`
- [ ] controller 只展开命中的规则正文与当前 workflow/task 工件
- [ ] worker 只按 `worker.yaml -> task.yaml -> context.md -> references` 顺序按需展开
- [ ] 文档主语义是 context-cleanup / index-first recovery
- [ ] 不默认全量扫描所有上下文正文

## 文案检查

- [ ] `skills/context-cleanup/SKILL.md` 明确声明其为独立 context-cleanup 入口
- [ ] `.agent-team/rules/core/context-management.md` 使用 context-cleanup / index-first 语义
- [ ] provider 注入模板指向 context-cleanup 规则
