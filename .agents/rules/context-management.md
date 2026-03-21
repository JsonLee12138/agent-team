# Context Management Rules

## Trigger

Apply this rule whenever context grows, the task changes phase, or resumed work needs a recovery anchor.

## Context-Cleanup Triggers

1. MUST enter context cleanup before starting a new logical phase after finishing the current one.
2. MUST enter context cleanup before reading or pasting large outputs, logs, replies, or diffs that may displace the current working context.
3. MUST enter context cleanup when the active thread can no longer hold the task goal, constraints, and next actions clearly.
4. MUST enter context cleanup before resumed work after a long pause, restart, handoff, or provider switch.

## Required Recovery Model

- Context cleanup resets session context; it MUST NOT rewrite, compress, or discard file artifacts.
- Context cleanup is NOT a synonym for `/compact`.
- Controller/main MUST read `.agents/rules/index.md` first, then open only the matching rule files, then the current workflow/task artifacts.
- Worker MUST read `worker.yaml` first, then `task.yaml`, and only then read `context.md` or referenced materials when needed.
- NEVER jump directly to rule bodies, `context.md`, or other detail files before reading the required entry file.
- NEVER default to scanning every context file; expand only what the index or current task explicitly requires.

## Provider Handling

- Claude, Codex, Gemini, and other providers MUST follow the same context-cleanup and index-first recovery strategy.
- Provider-specific prompt injections MUST point to this rule instead of requiring `/compact`.
