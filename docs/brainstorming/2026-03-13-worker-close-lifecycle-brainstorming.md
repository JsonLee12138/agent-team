# Worker Close And Lifecycle Brainstorming

## Role

general strategist

## Problem Statement And Goals

The current worker lifecycle is missing a dedicated `close` command. The CLI only exposes `open` and `delete`, while the skill documentation already describes a "close session without deleting the worker" capability. At the same time, the current lifecycle semantics are inconsistent:

- `worker create` opens and launches a session immediately
- `worker merge` implicitly closes a session
- `worker delete` contains its own pane-closing logic

The goal is to make worker lifecycle behavior explicit and consistent:

- Add `agent-team worker close <worker-id>`
- Make `worker open` the only command that starts a terminal session and launches the provider CLI
- Make `worker create` prepare the worker only
- Remove implicit close behavior from `worker merge`
- Make `worker delete` reuse the new close flow

## Constraints And Assumptions

- The existing terminal backend abstraction in `internal/session.go` is sufficient. No new backend API is needed.
- `worker close` must be idempotent.
- `worker close` clears only `PaneID`.
- `ControllerPaneID`, `Provider`, and `DefaultModel` must be preserved across close/open cycles.
- `worker open` must wait 3 seconds after spawning the terminal pane before sending `codex`, `claude`, `gemini`, or `opencode`.
- `worker create` should still sync skills and inject role prompt/context files.
- If skill sync fails during `worker create`, worker creation still succeeds, but the CLI must warn the user.
- Role or skill binding fixes must not be auto-handled by scripts. The user should be prompted via `skills/agent-team/SKILL.md` guidance whether role skill bindings need adjustment.

## Candidate Approaches

### Approach 1: Shared close helper in `cmd` layer

Create a shared session-close helper in the `cmd` package and reuse it from `worker close` and `worker delete`.

Pros:

- Minimal change aligned with current command structure
- Easy to apply consistently across `close` and `delete`
- Easy to test with current command-level mocks

Cons:

- Lifecycle logic remains in CLI layer rather than moving deeper into domain helpers

### Approach 2: Move close lifecycle logic into `internal`

Add worker lifecycle helpers in `internal` for loading config, closing pane, and persisting `PaneID` cleanup.

Pros:

- Cleaner separation if more lifecycle actions are added later
- Potentially more reusable outside CLI commands

Cons:

- More abstraction than needed for a small feature
- Higher refactor surface for little immediate benefit

### Approach 3: Add `worker close`, reuse `RunWorkerClose` directly from `delete`

Implement a new `worker close` command and call the command method from `delete` without extracting a shared helper.

Pros:

- Straightforward implementation path

Cons:

- Tighter coupling between command entrypoints and internal reuse
- Higher risk of duplicated output or error handling drift

## Recommended Design

Choose Approach 1.

This keeps the change small, fits the current `cmd/RunWorkerXxx` structure, and lets `worker close` and `worker delete` share a single source of truth for session shutdown semantics without introducing unnecessary abstraction.

## Architecture

### Command Responsibilities

- `agent-team worker create <role-name>`
  Creates the worker worktree, branch, config, tasks, skills sync, and role prompt/context files. Does not open a session.
- `agent-team worker open <worker-id>`
  Opens the terminal pane, waits 3 seconds, then launches the provider CLI. This becomes the only runtime start command.
- `agent-team worker close <worker-id>`
  Closes an existing session without deleting the worker. Idempotent.
- `agent-team worker delete <worker-id>`
  Reuses the shared close flow, then removes worktree and branch.
- `agent-team worker merge <worker-id>`
  Merges code only. No implicit session shutdown.

### Components To Change

- `cmd/worker.go`
  Register the new `close` subcommand.
- `cmd/worker_close.go`
  Add the new CLI entrypoint.
- `cmd` package shared helper
  Centralize session close behavior for `close` and `delete`.
- `cmd/worker_create.go`
  Remove pane spawn and provider launch behavior. Keep skills sync and role prompt injection.
- `cmd/worker_open.go`
  Keep pane spawn and provider launch behavior, but enforce 3 second delay.
- `cmd/worker_delete.go`
  Reuse shared close behavior before deleting the worker.
- `cmd/worker_merge.go`
  Remove session close logic.
- `README.md`
  Update worker lifecycle documentation.
- `README.zh.md`
  Update Chinese worker lifecycle documentation.
- `skills/agent-team/SKILL.md`
  Update lifecycle semantics and add guidance for skill sync failure follow-up with the user.

## Data Flow

### `worker create`

1. Resolve role and compute next worker ID.
2. Create worktree and branch.
3. Write `worker.yaml`.
4. Initialize `.tasks/`.
5. Sync role skills into the worktree.
6. Inject role prompt and context files.
7. Return success without opening a pane.
8. Leave `PaneID` empty.

### `worker open`

1. Load `worker.yaml`.
2. If `PaneID` points to a live pane, return `already running`.
3. Spawn a new pane.
4. Save the new `PaneID` and current `ControllerPaneID`.
5. Wait 3 seconds.
6. Send the provider launch command.
7. Return success.

### `worker close`

1. Load `worker.yaml`.
2. If the worker or config does not exist, return an error.
3. If `PaneID` is empty, succeed idempotently.
4. If `PaneID` is set but the pane is already offline, succeed idempotently.
5. If the pane is alive, kill it.
6. On successful completion, clear `PaneID` and save config.
7. Preserve `ControllerPaneID`, `Provider`, and `DefaultModel`.

### `worker delete`

1. Validate that the worker exists.
2. Run the shared close flow.
3. Remove the worktree.
4. Delete the worker branch.
5. Return success.

### `worker merge`

1. Validate that the worker exists.
2. Merge `team/<worker-id>` into the current branch.
3. Do not touch session state.

## Error Handling

- `worker close`
  Returns an error when the worker or config does not exist.
- `worker close`
  Treats empty `PaneID` or already-dead pane as success and normalizes state by clearing `PaneID`.
- `worker close`
  If `KillPane` fails for a live pane, return an error and do not clear `PaneID`.
- `worker delete`
  Stops if the shared close flow fails, so the system does not delete the worktree while a live session may still exist.
- `worker open`
  If pane creation fails, do not persist a new `PaneID`.
- `worker open`
  If provider launch send fails after pane creation, keep the persisted `PaneID` because the pane exists and can still be handled manually.
- `worker create`
  If skills sync fails, creation still succeeds but emits a warning.
- `worker create`
  On skills sync failure, the CLI should not try to auto-edit role or skill bindings.
- `skills/agent-team/SKILL.md`
  Should instruct the controller to ask the user whether role skill bindings need adjustment when skills sync warnings appear.
- `worker create`
  Role prompt injection remains required. If that fails, creation should fail because the worker is incomplete.

## Validation And Test Strategy

Add or update tests for:

- `worker close` closes a live pane and clears `PaneID`
- `worker close` succeeds when pane is already offline and clears `PaneID`
- `worker close` succeeds when `PaneID` is already empty
- `worker delete` reuses the shared close flow before removing the worktree
- `worker merge` no longer calls `KillPane`
- `worker create` no longer spawns a pane or launches a provider
- `worker create` still performs skills sync and role prompt injection
- `worker create` warns but succeeds when skills sync fails
- `worker open` remains the only launch path and waits 3 seconds before sending the provider command
- README and skill docs reflect the updated lifecycle semantics

## Risks And Mitigations

- Risk: Behavior changes may surprise users who relied on `worker create` auto-opening a session.
  Mitigation: Update README, Chinese README, and skill docs to state that `create` prepares and `open` starts.

- Risk: Session state can drift if `PaneID` is not normalized on offline panes.
  Mitigation: Make `close` explicitly idempotent and always clear stale `PaneID`.

- Risk: `worker delete` could become partially destructive if close and delete are not coordinated.
  Mitigation: Stop delete when live-pane shutdown fails.

- Risk: Skills sync warnings may be ignored, producing confusing worker behavior later.
  Mitigation: Surface a clear warning and update `skills/agent-team/SKILL.md` so the controller asks the user whether role skill bindings should be adjusted.

## Open Questions

No open design questions remain for this change set.
