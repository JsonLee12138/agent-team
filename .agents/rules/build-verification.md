# Build Verification Rules

## Trigger

Apply this rule before `go build`, before commit, before review handoff, and before reporting task completion.

## Pre-Build Checks

- MUST confirm the target package list and working directory before running build commands.
- MUST ensure required environment variables, generated files, and module metadata are present.
- MUST avoid building from a dirty or partial state when unrelated edits would invalidate results.

## Required Verification Commands

- MUST run `go build ./...` for repository-wide Go changes unless the task scope clearly limits the package set.
- MUST run `go vet ./...` for the affected scope before commit.
- MUST run `go test ./...` for the affected scope; use `./...` for shared or cross-package changes.
- MUST rerun the exact failing build or test command when the task is a fix.

## Pre-Commit Checklist

- ALWAYS confirm the changed files match the task scope.
- ALWAYS review command failures before retrying; NEVER loop without reading output.
- MUST verify that build, vet, and test results are current for the final diff.
- MUST mention any skipped verification explicitly in the completion message with the reason.
