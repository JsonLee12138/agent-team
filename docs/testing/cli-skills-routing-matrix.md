# CLI Skills Routing Matrix

基线来源：

- `docs/brainstorming/2026-03-21-cli-skills-split-brainstorming.md`
- `docs/analysis/2026-03-21-cli-skills-split-qa-delivery.md`

## P0 路由样例

| Prompt | Expected skill |
| --- | --- |
| 帮我创建一个任务并派给 frontend worker | `task-orchestrator` |
| 生成 workflow plan，等我审批后激活 | `workflow-orchestrator` |
| 打开 qa worker，并告诉他先看 task.yaml | `worker-dispatch` |
| 我在 worker 里，恢复当前任务继续做 | `worker-recovery` |
| 我做完了，向 main 汇报并请求验收 | `worker-reply-main` |
| 帮我找一个 backend role repo 并加到项目里 | `role-repo-manager` |
| 搜索 catalog 里有哪些 product roles | `catalog-browser` |
| 看下 task 状态 | `task-inspector` |
| 会话乱了，帮我清理并重新锚定 | `context-cleanup` |

## 负向边界

| Prompt | Expected behavior |
| --- | --- |
| 看看 worker | 命中 `worker-inspector` 或请求澄清；不能直接 `worker open` |
| 完成任务并归档 | 命中 `task-orchestrator`，不能落到 `task-inspector` |
| 把这个 repo 加到 role sources | 命中 `role-repo-manager`，不能落到 `catalog-browser` |
| 恢复 worker 当前任务（controller 会话） | 命中 `worker-dispatch` 或请求澄清；不能假装读 `worker.yaml` |
