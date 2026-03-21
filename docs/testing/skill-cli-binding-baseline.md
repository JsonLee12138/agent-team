# Skill CLI Binding Baseline

## Controller / human

- `task-orchestrator` -> `agent-team task create/list/show/assign/done/archive`
- `workflow-orchestrator` -> `agent-team workflow plan generate/approve/activate/close`
- `worker-dispatch` -> `agent-team worker open/status`, `agent-team reply`
- `project-bootstrap` -> `agent-team init/setup/migrate`
- `rules-maintenance` -> `agent-team rules sync`
- `skill-maintenance` -> `agent-team skill check/update/clean`
- `role-repo-manager` -> `agent-team role-repo find/add/list/check/update/remove`
- `catalog-browser` -> `agent-team catalog search/show/list/repo/stats`
- `worker-inspector` -> `agent-team worker status`
- `role-browser` -> `agent-team role list`

## Worker

- `worker-recovery` -> artifact reads, optional `agent-team task show`
- `worker-reply-main` -> `agent-team reply-main`

## Shared / strategy

- `task-inspector` -> `agent-team task list/show`
- `context-cleanup` -> `agent-team context-cleanup` (index-first recovery guidance), no `compact` synonym

## Compatibility shells

- `agent-team` -> navigation shell only
- `strategic-compact` -> deprecated shell redirecting to `context-cleanup`
