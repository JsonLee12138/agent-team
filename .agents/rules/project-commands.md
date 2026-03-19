# Project Commands Rules

This file is AI-generated for the current project. It is regenerated and may be overwritten by `agent-team init` and `agent-team rules sync`.

AI workers must read this file before running any project command.

## Core Rules

- Use only command entry points that are confirmed by this repository's `Makefile`, `go.mod`, `package.json`, lockfiles, or current project docs.
- Run commands from the correct working directory shown below. Do not assume repo-root execution is valid for subprojects.
- Prefer the documented project entry point over ad hoc substitutes. If the repo exposes `make build`, do not replace it with a guessed raw `go build` unless you have inspected the repo and confirmed the alternative is correct.
- If a command fails, stop and inspect the repository before retrying. Determine the correct command, working directory, prerequisites, or alternative entry point from repo evidence first.
- If rule drift is confirmed, ask the user whether to update `.agents/rules/project-commands.md`.

## Repository Root Commands

Project module: `github.com/JsonLee12138/agent-team`  
Go version declared in `go.mod`: `1.24.2`

Use these commands from the repository root:

| Workflow | Command | Notes |
| --- | --- | --- |
| Build | `make build` | Builds `output/agent-team`. The current `Makefile` uses `$(HOME)/.gvm/gos/go1.24.2/bin/go` via `GORUN`, so Go must be available at that path unless the repo is updated. |
| Test | `make test` | Runs `test ./... -v` through the same `GORUN` path. |
| Lint | `make lint` | Runs `go vet ./...` through the same `GORUN` path. |
| Clean | `make clean` | Removes `output/agent-team`. |
| Install | `make install` | Builds first, then symlinks `/usr/local/bin/agent-team` to the built binary. This modifies `/usr/local/bin`. |
| Uninstall | `make uninstall` | Removes the symlink and restores `/usr/local/bin/agent-team.bak` if present. |
| Migrate | `make migrate` | Builds first, then runs `./output/agent-team migrate`. |
| Plugin Package | `make plugin-pack` | Builds first, then creates `agent-team-plugin.tar.gz` from plugin assets and the built binary. |
| Plugin Test | `make plugin-test` | Builds first, then runs `claude --plugin-dir . --debug`. Requires the `claude` CLI. |
| Release | `make release V=x.y.z` | Requires `V`. Requires a clean git worktree. Runs `scripts/bump-version.sh`, commits, tags, and pushes. Not a routine verification command. |

## `role-hub` Commands

Use these commands from the repository root with `npm --prefix role-hub ...`, or by `cd role-hub` first and then running the equivalent `npm run ...`.

Current repo docs for local setup are in `docs/role-hub-remix-migration.md`, which says to run `npm install` in `role-hub` before local development.

| Workflow | Command | Notes |
| --- | --- | --- |
| Dev | `npm --prefix role-hub run dev` | Starts UnoCSS watch and `remix dev` together. This is the repo's confirmed dev entry point. |
| Build CSS | `npm --prefix role-hub run build:css` | Generates `app/styles/uno.css` with UnoCSS. |
| Build | `npm --prefix role-hub run build` | Runs `build:css` first, then `remix build`. |
| Start | `npm --prefix role-hub run start` | Serves `build/index.js` with `remix-serve`. Use after a successful build. |
| Lint | `npm --prefix role-hub run lint` | Runs `eslint .` inside `role-hub`. |
| Test | `npm --prefix role-hub run test` | Runs `vitest run --passWithNoTests`. |
| E2E | `npm --prefix role-hub run e2e` | Runs `playwright test --pass-with-no-tests`. |

Additional repo-grounded notes for `role-hub`:

- `docs/role-hub-remix-migration.md` documents optional environment variables such as `ROLE_HUB_DB_DIALECT`, `ROLE_HUB_DB_DSN`, `ROLE_HUB_DB_TIMEOUT`, `ROLE_HUB_RATE_LIMIT_RPS`, `ROLE_HUB_RATE_LIMIT_BURST`, `ROLE_HUB_MAX_BODY_BYTES`, `ROLE_HUB_MAX_RESULTS`, and `ROLE_HUB_MAX_INFLIGHT`.
- Both `role-hub/package-lock.json` and `role-hub/pnpm-lock.yaml` exist, but the documented command surface in this repo uses `npm`. Use the documented `npm` entry points unless the repository is changed and verified otherwise.
- If `e2e` fails because of browser or Playwright setup issues, inspect `role-hub/playwright.config.ts`, local dependency state, and project docs before deciding the next step.

## `adapters/opencode` Commands

Use this command from the repository root:

| Workflow | Command | Notes |
| --- | --- | --- |
| Typecheck | `npm --prefix adapters/opencode run typecheck` | Runs `tsc --noEmit`. This is the only confirmed package script in this subproject. |

Additional repo-grounded notes for `adapters/opencode`:

- `adapters/opencode/package.json` does not define `build`, `test`, `lint`, `dev`, or `format` scripts.
- `adapters/opencode/bun.lockb` exists. If dependencies are missing or install behavior is unclear, inspect `adapters/opencode/package.json`, the lockfile, and current repo docs before choosing an install command or package manager.

## No Confirmed Command Detected

The following workflow categories do not have a confirmed repo-local command entry point at the time this file was generated:

- Format: no confirmed repo-wide or subproject `format` command was detected in the root `Makefile`, `role-hub/package.json`, or `adapters/opencode/package.json`.
- Codegen: no confirmed repo-wide or subproject code generation command was detected.
- Root-level dev: no confirmed root `make dev` or equivalent was detected. The confirmed dev entry point is only `role-hub`'s `npm run dev`.

Do not invent commands such as `make fmt`, `go fmt ./...`, `npm run format`, `npm run codegen`, `go generate ./...`, or `make dev` unless you have inspected the repository and confirmed that the command is actually correct for the current state of the project.

## Failure Handling

If any project command fails:

- Inspect the relevant command source before retrying.
- Confirm the correct working directory.
- Confirm prerequisite tools and versions.
- Confirm whether environment variables or generated files are required.
- Confirm whether the repo uses a different entry point for the same workflow.
- Retry only after identifying a repo-grounded reason for the new command or setup.

Minimum inspection targets by area:

- Root Go workflows: inspect `Makefile`, `go.mod`, and the local availability of the `go1.24.2` toolchain path used by `GORUN`.
- `role-hub`: inspect `role-hub/package.json`, lockfiles, `docs/role-hub-remix-migration.md`, and local environment variables.
- `adapters/opencode`: inspect `adapters/opencode/package.json`, `bun.lockb`, and any current docs tied to that adapter.

Do not repeatedly retry the same failing command without new evidence from the repository.

## Rule Drift

Rule drift is confirmed if any of the following becomes true:

- A command listed here no longer exists in `Makefile` or `package.json`.
- A new stable project command entry point is added.
- A documented working directory changes.
- A prerequisite or package-manager expectation materially changes.
- This file conflicts with the current `Makefile`, `go.mod`, `package.json`, lockfiles, or active project docs.

When drift is confirmed, ask the user whether to update `.agents/rules/project-commands.md`.

Because this file is AI-generated and regenerated by `agent-team init` and `agent-team rules sync`, do not assume manual edits will persist.
