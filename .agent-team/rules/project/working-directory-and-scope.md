# Working Directory And Scope

## Repository Root Expectations

Use repository root (`github.com/JsonLee12138/agent-team` module root) for:

- all `make ...` commands
- all `go run . ...` commands
- all `agent-team ...` commands executed from source or built binary

## Nested Node Package Expectations

From repository root, prefer `npm --prefix` so command scope is explicit:

- `npm --prefix adapters/opencode run typecheck`
- `npm --prefix role-hub run <script>`

Equivalent `cd` usage is valid when needed:

- `cd adapters/opencode && npm run typecheck`
- `cd role-hub && npm run dev|build|test|e2e|lint|start`

## Submodule Boundary

`role-hub` is declared as a git submodule in `.gitmodules` and has its own `.git` directory. Confirm whether your change belongs to the parent repo or the `role-hub` repository before committing.

## Non-Project Cache Boundary

Do not use `.tmp/gomodcache/**/Makefile` or `.tmp/gomodcache/**/go.mod` as project execution context. Those files are dependency cache artifacts, not this repository's primary command surface.
