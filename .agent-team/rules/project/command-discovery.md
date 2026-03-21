# Command Discovery

## When Signals Conflict, Re-Inspect Instead Of Guessing

Use repository-local inspection commands to confirm the current command surface:

- List repo-root make targets:
  - `make -qp 2>/dev/null | awk -F: '/^[a-zA-Z0-9][^$#\t=]*:([^=]|$)/ {print $1}' | sort -u`
- Show CLI command tree from source:
  - `go run . --help`
  - `go run . worker --help`
  - `go run . task --help`
  - `go run . rules --help`
- Read exact script definitions:
  - `cat adapters/opencode/package.json`
  - `cat role-hub/package.json`
- Verify submodule boundary:
  - `cat .gitmodules`

## Missing Information Policy

If a command or workflow is not confirmed by `Makefile`, `go.mod`-root CLI help, or package scripts, do not invent usage. Record the uncertainty and include the exact inspection command needed to verify it.
