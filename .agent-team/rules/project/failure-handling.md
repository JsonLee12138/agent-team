# Failure Handling

## Initialization Gate For Most CLI Commands

`cmd/root.go` enforces initialization for most commands on non-`team/*` branches.

- Symptom: `.agent-team/rules/ not found. Run 'agent-team init' first ...`
- Action: run `agent-team init` (or `go run . init`) from repo root.
- Non-interactive behavior: with `AGENT_TEAM_NONINTERACTIVE=1`, the command fails instead of prompting.

`init` and `setup` are explicitly exempt from this pre-check.

## Rules Sync Preconditions

`agent-team rules sync` requires an existing rules directory.

- Symptom: `.agent-team/rules/ not found. Run 'agent-team init' first`
- Action: run `agent-team init` first, then `agent-team rules sync`.

## Worker Creation Warning

`worker create` warns when project rules are missing:

- Warning: run `agent-team rules sync` to generate `.agent-team/rules/project/` before command-heavy work.
- Action: sync rules before assigning large or command-intensive tasks.

## Migration Safety Checks

`agent-team migrate` handles only `agents/ -> .agents/`.

- If `agents/` is absent: command exits successfully with "Nothing to migrate".
- If `.agents/` already exists: command fails to avoid data loss.

## Make Target Failure Modes

- `make release V=<x.y.z>` fails when `V` is missing, working tree is dirty, or tag already exists.
- `scripts/bump-version.sh` also requires `jq`; release flow fails without it.
- `make plugin-pack` currently fails in this checkout because `hooks/` is referenced by `tar` but does not exist at repo root.
- `make plugin-test` requires `claude` CLI availability.

## Restricted Environment Build Cache

In restricted environments, Go build/test may fail on default cache paths.

- Symptom: `open .../Library/Caches/go-build/...: operation not permitted`
- Action: set a writable cache path, for example `GOCACHE=$(pwd)/.tmp/gocache`.

## Node Test Semantics

In `role-hub` scripts:

- `npm --prefix role-hub run test` uses `vitest run --passWithNoTests`
- `npm --prefix role-hub run e2e` uses `playwright test --pass-with-no-tests`

A passing status can mean "no tests found". Check output counts before treating it as full validation.
