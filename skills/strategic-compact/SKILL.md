---
name: strategic-compact
description: >
  Deprecated compatibility shell for historical prompts that referenced strategic-compact.
  Route all new session recovery requests to `context-cleanup` instead of using compact as the primary semantic.
---

# strategic-compact

This skill is deprecated as a primary strategy surface.

## Migration target

- Use `context-cleanup` for session cleanup and file re-anchoring.
- `context-cleanup` is index-first and file-anchored.
- `context-cleanup` is not a synonym for `/compact`.

## Compatibility behavior

- If an old prompt mentions `strategic-compact`, interpret the intent as context stabilization.
- Re-route the work to `context-cleanup` unless the user is explicitly asking about the legacy `agent-team compact` transport command itself.

## Boundary

- Do not describe this skill as the main controller strategy anymore.
- Do not expand or maintain compact-centric recovery policy here.
- Keep this file only so historical references do not break.
