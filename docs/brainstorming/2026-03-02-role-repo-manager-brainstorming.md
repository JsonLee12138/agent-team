# Role Repo Manager Brainstorming

## Role Used
- general strategist

## Problem Statement
Build a role repository manager in `agent-team` inspired by `vercel-labs/skills`, but focused on role installation and lifecycle management.

The new command namespace must be `role-repo` to avoid conflict with existing `role` commands.

## Goals
- Add a dedicated `role-repo` command group for repository-backed role management.
- Support project-level install by default and global install as an option.
- Use GitHub-based search and discovery.
- Only accept role sources that match strict role path contracts.
- Support role lifecycle commands: search, add, list, remove, check, update.
- Track install state and update metadata with dual lock files.

## Constraints
- Command namespace must be `role-repo`, not `role`.
- Remote role path contract is strict:
  - `<repo>/skills/<role>/references/role.yaml`
  - `<repo>/.agents/teams/<role>/references/role.yaml`
- Global role install target is `~/.agents/roles/`.
- Project install remains default scope.
- Search should use GitHub directly.
- Must support dual lock strategy.

## Assumptions
- Existing local role schema remains `references/role.yaml`-based.
- Existing `agent-team role` and worker flows remain backward compatible.
- GitHub API access may run unauthenticated or authenticated, depending on user environment.

## Candidate Approaches
1. API-only approach
- Search/add/check/update all through GitHub API.
- Pros: no clone cost, fast remote operations.
- Cons: higher implementation complexity, rate-limit handling complexity.

2. Clone-only approach
- Search/add/check/update all through repository clone and local scan.
- Pros: straightforward implementation model.
- Cons: slower, more network and disk overhead.

3. Hybrid approach (recommended)
- Search via GitHub search API.
- Add by source parsing + role discovery + targeted fetch/install (API first, clone fallback if needed).
- Check/update through lock metadata and remote folder hash comparison.
- Pros: good UX and performance balance, close to `vercel-labs/skills` operational model.
- Cons: moderate implementation complexity.

## Recommended Design
Adopt the hybrid model with strict role path filtering and dual lock files.

### Command Surface
- `agent-team role-repo search <query>`
- `agent-team role-repo add <source> [--role <name>...] [-g|--global] [--list] [-y]`
- `agent-team role-repo list [-g|--global]`
- `agent-team role-repo remove [roles...] [-g|--global] [-y]`
- `agent-team role-repo check [-g|--global]`
- `agent-team role-repo update [-g|--global] [-y]`

### Directory Model
- Project install target: `<repo>/.agents/teams/<role>/`
- Global install target: `~/.agents/roles/<role>/`

### Lock Model (Dual Lock Strategy)
- Project lock: `<repo>/roles-lock.json`
- Global lock: `~/.agents/.role-lock.json`
- Entry fields:
  - `name`
  - `source`
  - `sourceType`
  - `sourceUrl`
  - `rolePath`
  - `folderHash`
  - `installedAt`
  - `updatedAt`

### Core Data Flow
1. `search`
- Query GitHub.
- Keep only results under allowed path contracts.

2. `add`
- Parse source.
- Discover candidate roles under allowed paths.
- Optionally filter by `--role`.
- Install into selected scope target.
- Write corresponding lock entries.

3. `list`
- Scan scope install directory.
- Merge display with lock metadata when available.

4. `remove`
- Remove selected role directories in scope.
- Remove corresponding lock entries in scope.

5. `check`
- Read lock entries in selected scope.
- Compare remote `folderHash` against stored `folderHash`.
- Report update candidates and untracked local roles.

6. `update`
- Reinstall roles that have updates.
- Refresh lock metadata.

## Architecture Components
- `cmd/role_repo.go` and subcommand files for `search/add/list/remove/check/update`
- `internal/role_repo_types.go`
- `internal/role_repo_paths.go`
- `internal/role_repo_source.go`
- `internal/role_repo_search.go`
- `internal/role_repo_discovery.go`
- `internal/role_repo_install.go`
- `internal/role_repo_lock.go`
- `internal/role_repo_check_update.go`

## Error Handling Strategy
- Source parse failure: return actionable examples and exit.
- GitHub failure: separate auth/rate-limit (`401/403`) from network errors.
- No valid role paths: explain exact accepted path contracts.
- Install conflict:
  - default: do not overwrite
  - with `-y`: overwrite after local backup
- Lock corruption:
  - recover to empty lock with warning
  - do not block `list`
  - `check/update` skip invalid entries
- Batch operations:
  - per-item `success/failed/skipped` reporting
  - do not fail whole batch on single-item failure

## Validation and Testing Strategy
### Unit Tests
- Source parsing for supported source formats.
- Strict path filtering to only accepted contracts.
- Dual lock read/write behavior and version handling.
- Hash compare logic for update detection.

### Integration Tests
- `add --list`, `add --role`, `list`, `remove`, `check`, `update` flows.
- Scope isolation for same role name across project and global installs.

### Regression Tests
- Verify no behavior regression for existing `agent-team role list`.
- Verify no behavior regression for worker flows (`worker create/open/assign`).

### Failure Injection
- GitHub timeout and `403` scenarios.
- Corrupted lock file handling.
- Read-only target directory handling.

## Risks and Mitigations
- Risk: GitHub API rate limits.
  - Mitigation: token-aware requests, clear guidance on auth setup, retry backoff.
- Risk: inconsistent role naming across repos.
  - Mitigation: enforce folder-derived role name and strict path contract.
- Risk: lock drift from manual file edits.
  - Mitigation: `list` scans filesystem first; `check/update` rely on lock and surface untracked roles.
- Risk: accidental overwrite during updates.
  - Mitigation: default no-overwrite and explicit `-y` semantics with backup.

## Open Questions
- None blocking for v1 design.
- Potential v2 extension: optional source pinning by branch/tag/commit in lock metadata.

