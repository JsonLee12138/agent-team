---
name: agent-team
description: >
  Compatibility shell for the legacy umbrella agent-team skill.
  Use only when older prompts still reference `agent-team`; route the request to the dedicated scenario skill instead of treating this file as the primary execution surface.
---

# agent-team

This skill is now a navigation shell.

Do not keep adding new behavior here. Prefer the dedicated first-level skills below.

## Primary scenario skills

### Controller / human orchestration

- `task-orchestrator`: task lifecycle actions
- `workflow-orchestrator`: workflow plan governance
- `worker-dispatch`: open worker, reply to worker, targeted worker dispatch
- `project-bootstrap`: `init`, `migrate`
- `rules-maintenance`: `rules sync`
- `skill-maintenance`: `skill check/update/clean`
- `role-repo-manager`: role source management
- `catalog-browser`: read-only catalog browsing
- `worker-inspector`: read-only worker status
- `role-browser`: read-only local role browsing

### Worker-first skills

- `worker-recovery`: recover current assignment from `worker.yaml`
- `worker-reply-main`: send `reply-main` updates back to controller

### Shared / strategy skills

- `task-inspector`: read-only task status and lookup
- `context-cleanup`: clean session context and re-anchor from files with index-first recovery

## Routing rules

- If the request implies a specific lifecycle action, route to the dedicated skill immediately.
- If the request is about context stabilization or resumed work, route to `context-cleanup` or `worker-recovery` instead of describing compact behavior here.
- Keep this file as a compatibility surface for historical prompts only.

## References

- This compatibility shell should not carry the old umbrella reference library anymore.
- Read the dedicated first-level skill documentation instead of any legacy `agent-team/references/*` material.
