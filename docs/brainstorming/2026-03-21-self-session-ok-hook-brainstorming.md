# Self-session `ok` Hook Brainstorming

- Date: 2026-03-21
- Role used: general strategist

## Problem statement and goals

Validate the smallest possible experiment for Claude Code hooks to inject a fixed message `ok` back into the **current Claude session**.

Goals:
- Prove self-message injection works in the current session
- Minimize blast radius and rollback cost
- Keep the first experiment local and easy to observe

Non-goals:
- Cross-session communication
- Message routing or broker design
- Production-ready plugin packaging in the first step

## Constraints and assumptions

- The experiment should target the **current session**, not another Claude instance.
- The injected content should be plain text: `ok`.
- The first step should be local-only and reversible.
- Later, the same behavior may be promoted into a plugin-level hook.
- Hook output should stay minimal to avoid polluting context with unrelated shell output.

## Candidate approaches

### Approach A — Local project hook in `.claude/settings.local.json` (Recommended)
Use a `UserPromptSubmit` hook so that each time the user submits a prompt, the hook emits `ok`, which Claude Code injects into the current session context.

**Pros**
- Smallest experiment
- Local-only and safe to revert
- Easiest to verify visually

**Cons**
- Not shareable by default
- Not yet productized

### Approach B — Shared project hook in `.claude/settings.json`
Use the same event and behavior, but store the hook in the repo-level shared settings.

**Pros**
- Shareable with collaborators
- Closer to a project convention

**Cons**
- Too broad for an experiment
- Risks surprising other users of the repo

### Approach C — Plugin-level hook in `hooks/hooks.json`
Move the same behavior into the plugin layer, ideally by calling a dedicated script.

**Pros**
- Productizable and versionable
- Good base for future session messaging features

**Cons**
- More moving parts than needed for the first validation
- Slightly higher debugging and packaging overhead

## Recommended design

### Phase 1: Local validation
Adopt **Approach A** first.

- Trigger: `UserPromptSubmit`
- Scope: current project, local machine only
- Injected payload: plain text `ok`
- Success condition: after a new prompt submission, the current Claude session behavior confirms `ok` was injected into context

### Phase 2: Plugin packaging
After Phase 1 is confirmed, promote the same behavior to **Approach C**.

- Keep the same trigger and payload
- Move behavior into plugin-managed hook configuration
- Prefer a dedicated script entrypoint for maintainability and future extension

## Architecture

### Phase 1 architecture
- User submits prompt
- `UserPromptSubmit` hook fires
- Hook emits plain text `ok`
- Claude Code appends that output to the current session context

### Phase 2 architecture
- User submits prompt
- Plugin-defined `UserPromptSubmit` hook fires
- Hook delegates to a script
- Script emits plain text `ok`
- Claude Code appends that output to the current session context

## Components

### Phase 1
- Local project settings file
- One hook entry bound to `UserPromptSubmit`
- No external broker, queue, or cross-session state

### Phase 2
- Plugin hook configuration
- One dedicated script for output generation
- Optional future extension points for session metadata or structured messages

## Data flow

1. User submits a prompt
2. `UserPromptSubmit` event occurs
3. Hook runs
4. Hook writes `ok` to stdout
5. Claude Code injects that stdout into the current session context
6. The session continues with the injected context available

## Error handling and risks

### Risk 1: Shell/profile noise contaminates output
If the shell environment prints extra text, the injected content may no longer be clean `ok`.

**Mitigation:** keep the command minimal and avoid any extra output.

### Risk 2: Repeated injection on every prompt becomes noisy
The experiment uses a high-frequency event.

**Mitigation:** accept this for validation, then tighten behavior later if needed.

### Risk 3: Shared config surprises other users
If placed in shared settings too early, the behavior affects more than the experiment target.

**Mitigation:** keep Phase 1 in `.claude/settings.local.json` only.

## Validation and test strategy

### Acceptance criteria
- The hook runs on `UserPromptSubmit`
- The emitted content is exactly `ok`
- The injected text reaches the current Claude session context
- The experiment remains local and easy to remove

### Manual validation
- Submit a new prompt after enabling the hook
- Observe whether the current session reflects the injected `ok`
- Confirm no extra unintended output is injected

### Promotion criteria for plugin version
- Local experiment is stable
- The behavior is useful beyond one machine/session
- A script entrypoint is preferred for maintainability

## Open questions

- None for the minimal validation scope

## Decision summary

Proceed in two stages:
1. Implement the minimal local experiment in `.claude/settings.local.json`
2. After successful validation, provide the plugin `hooks/hooks.json` variant as the next step
