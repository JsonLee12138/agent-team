# Brainstorming: worker create/open Provider Flag Unification

## Role

general strategist

## Problem Statement

`agent-team worker create` and `agent-team worker open` currently accept provider in positional form, while `worker.yaml` already persists provider and model-related state. The desired direction is to make `create` and `open` consistent, use explicit flags for provider selection, and define clear persistence rules for `provider` and `default_model`.

## Goals

- Unify `worker create` and `worker open` around explicit `--provider` and `--model` flags.
- Remove positional provider input from `create` and `open`.
- Keep `worker.yaml` as the persisted source of worker provider/model configuration.
- Ensure `create` writes a default provider when omitted.
- Ensure `open` only updates persisted config when explicit override flags are passed.
- Document that `--new-window` is optional and should not be shown as default behavior.

## Non-Goals

- Do not change `worker assign` provider behavior in this iteration.
- Do not redesign worker session lifecycle beyond provider/model persistence.
- Do not implement backward-compatible deprecation for old positional provider syntax.

## Constraints And Assumptions

- Existing worker config file is `worker.yaml` in the worktree root.
- Existing fields `provider` and `default_model` in `WorkerConfig` are sufficient; no schema expansion is needed.
- This change is intentionally breaking for `worker create <role> codex` and `worker open <id> codex`.
- `create --model` and `open --model` must have the same behavior contract.

## Candidate Approaches

### Approach 1: Documentation-Only Cleanup

Update docs/help text to recommend flags, but keep positional provider behavior in code.

Trade-offs:
- Lowest risk to scripts and existing users.
- Does not actually resolve the CLI inconsistency.
- Leaves implementation and docs partially divergent.

### Approach 2: Explicit Flag Unification For `create` And `open` Only

Change both commands to use explicit `--provider` and `--model`, remove positional provider support, and define persistence rules for `worker.yaml`.

Trade-offs:
- Clean and focused solution for the two relevant commands.
- Matches the desired UX and configuration semantics.
- Introduces a deliberate breaking change for existing positional invocations.

### Approach 3: Full Worker CLI Provider Unification

Apply the same explicit-flag migration to `create`, `open`, and `assign`.

Trade-offs:
- Most globally consistent CLI surface.
- Larger scope than needed for the current problem.
- Adds migration cost where there is no clear user value right now.

## Recommended Design

Choose Approach 2.

Reasoning:
- It directly solves the inconsistency the user cares about.
- It keeps scope controlled.
- It preserves `worker.yaml` as the single persisted configuration source without over-expanding the change set.

## Architecture

Target command forms:

```bash
agent-team worker create <role-name> [--provider <provider>] [--model <model>]
agent-team worker open <worker-id> [--provider <provider>] [--model <model>] [--new-window]
```

Design rules:

- `worker create` no longer accepts positional provider input.
- `worker open` no longer accepts positional provider input.
- `worker assign` remains unchanged.
- `worker.yaml` remains the only persisted worker config file for these values.
- `--new-window` remains optional and is not part of the default usage path.

## Components

### Command Layer

- `cmd/worker_create.go`
  - Change Cobra usage from positional provider to explicit `--provider`.
  - Restrict positional args to `<role-name>` only.
  - Keep `--model`, `--new-window`, and `--fresh`.

- `cmd/worker_open.go`
  - Change Cobra usage from positional provider to explicit `--provider`.
  - Restrict positional args to `<worker-id>` only.
  - Keep `--model` and `--new-window`.
  - Add config update behavior when explicit override flags are passed.

### Config Layer

- `internal/config.go`
  - Reuse existing `Provider` and `DefaultModel` fields.
  - Reuse existing `LoadWorkerConfig` and `Save`.
  - No schema change required.

### Documentation Layer

- `skills/agent-team/SKILL.md`
- `skills/agent-team/references/details.md`
- Any README examples referencing `worker create/open`

All examples should move to flag-based provider syntax.

## Data Flow

### `worker create`

1. Parse `--provider` and `--model`.
2. If `--provider` is omitted, internally set provider to `claude`.
3. Validate provider against supported providers.
4. Create the worktree.
5. Write `worker.yaml` with:
   - `provider`: explicit `--provider` value, or default `claude`
   - `default_model`: explicit `--model` value if provided, otherwise empty
6. Use the resolved provider/model for skill installation and launch command generation.

### `worker open`

1. Load existing `worker.yaml`.
2. If `--provider` is passed:
   - Use it for this session.
   - Update `worker.yaml.provider`.
3. If `--provider` is omitted:
   - Use the existing `worker.yaml.provider`.
   - Do not modify config.
   - If config is missing provider, use `claude` only as an in-memory compatibility fallback for the current launch and do not write it back.
4. If `--model` is passed:
   - Use it for this session.
   - Update `worker.yaml.default_model`.
5. If `--model` is omitted:
   - Do not modify `worker.yaml.default_model`.
6. Spawn the pane and launch the selected provider with the resolved model.

## Error Handling

- Invalid `--provider`
  - Return an error immediately.
  - Help text and error wording should reference the flag form, not positional provider.

- Missing `worker.yaml`
  - Keep current open behavior: fail with a clear worker-not-found/config-not-found error.

- Missing `worker.yaml.provider` during `open`
  - Do not write any automatic repair.
  - Use `claude` for the current launch only as a compatibility fallback.
  - Optional lightweight notice is acceptable.

- Config save failure during `open --provider` or `open --model`
  - Fail fast rather than launching with unsaved state divergence.

- Old positional syntax
  - `worker create <role> codex`
  - `worker open <id> codex`
  - These should fail due to invalid positional argument count after the change.

## Validation And Test Strategy

### Automated Tests

- Add `cmd/worker_create_test.go`
  - Validates only one positional arg is accepted.
  - Validates explicit `--provider` is written to `worker.yaml.provider`.
  - Validates omitted `--provider` defaults to `claude`.
  - Validates explicit `--model` is written to `worker.yaml.default_model`.

- Add `cmd/worker_open_test.go`
  - Validates only one positional arg is accepted.
  - Validates `open --provider` updates `worker.yaml.provider`.
  - Validates omitted `--provider` leaves existing config unchanged.
  - Validates `open --model` updates `worker.yaml.default_model`.
  - Validates omitted `--model` leaves existing config unchanged.
  - Validates old positional provider syntax fails.

### Documentation Checks

- Search for all `worker create` and `worker open` examples in:
  - skill docs
  - references
  - README files
- Replace provider examples with flag-based syntax.
- Ensure `--new-window` is documented as optional, not default.

### Manual Checks

Example manual flows:

```bash
agent-team worker create backend --provider codex --model gpt-5
agent-team worker open backend-001 --provider claude --model sonnet
```

Verify:

- `worker.yaml.provider` is created or updated as designed.
- `worker.yaml.default_model` is created or updated only when an explicit model flag is passed.
- `open` without flags preserves prior config values.

## Risks And Mitigations

- Breaking existing scripts using positional provider syntax
  - Mitigation: update help text and docs everywhere in the same change.

- Runtime/config divergence during `open`
  - Mitigation: require config save success before continuing when explicit override flags are passed.

- Partial documentation migration
  - Mitigation: run repository-wide search for affected examples.

## Open Questions

- None remaining for this design.

## Approved Decisions

- Use explicit `--provider` flags for `worker create` and `worker open`.
- Remove positional provider support from `create` and `open`.
- Do not change `worker assign`.
- `create` defaults `provider` to `claude` and writes it to config.
- `open --provider` updates `worker.yaml.provider`; omitted provider does not rewrite config.
- `create --model` and `open --model` share the same persistence rule: write `default_model` only when the flag is explicitly provided.
- `--new-window` remains optional and should not be shown as default behavior in docs.
