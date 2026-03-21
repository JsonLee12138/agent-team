# Context Cleanup Checklist

## Hard gates

- [ ] controller 路径先读 `.agents/rules/index.md`
- [ ] worker 路径先读 `worker.yaml`
- [ ] controller 只展开命中的规则正文与当前 workflow/task 工件
- [ ] worker 只按 `worker.yaml -> task.yaml -> context.md -> references` 顺序按需展开
- [ ] 文档主语义不是 `/compact`
- [ ] 不默认全量扫描所有上下文正文

## 文案检查

- [ ] `skills/context-cleanup/SKILL.md` 明确声明不是 `/compact` 同义词
- [ ] `.agents/rules/context-management.md` 使用 context-cleanup / index-first 语义
- [ ] provider 注入模板不再要求“必须 `/compact`”
- [ ] `skills/strategic-compact/SKILL.md` 仅作为兼容壳并迁移到 `context-cleanup`
