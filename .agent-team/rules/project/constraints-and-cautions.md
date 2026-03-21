# Project Constraints And Cautions

## High-Impact Commands

- `make install` rewires `/usr/local/bin/agent-team` to a symlink of `output/agent-team` and may back up an existing non-symlink binary.
- `make uninstall` removes that symlink and may restore `/usr/local/bin/agent-team.bak`.
- `make release V=<x.y.z>` bumps version files, creates a commit, creates a git tag, and pushes branch and tags.

Treat these as state-changing operations, not routine validation steps.

## Rule And Provider File Mutation

`agent-team init` and `agent-team rules sync` both regenerate project rules under `.agent-team/rules/project/` and update tagged sections in provider files (`CLAUDE.md`, `AGENTS.md`, `GEMINI.md`).

Run them intentionally and review resulting diffs in tracked files.

## Ignored Local Artifacts

Root `.gitignore` excludes `.agent-team/` and `.tmp`.

Do not rely on these paths as versioned deliverables unless ignore rules are intentionally changed.

## Opencode Adapter Runtime Dependency

`adapters/opencode/src/plugin.ts` delegates runtime hook behavior to the `agent-team` binary and validates availability via `agent-team version`.

If the binary is missing from `PATH`, the adapter logs a warning and hook calls are treated as non-fatal.

## Preferred Worker Flow

`worker create` is available, but command text indicates preferred flow is task-first dispatch:

- `agent-team task create ...`
- `agent-team task assign <task-id>`

Use this as default orchestration path unless compatibility behavior is explicitly needed.
