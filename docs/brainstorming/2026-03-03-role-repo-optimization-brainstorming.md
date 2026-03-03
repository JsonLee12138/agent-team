# Role-Repo Optimization & Bridge Removal Brainstorming

**Date:** 2026-03-03
**Role:** General Strategist
**Status:** Approved

---

## Problem Statement

Three areas of agent-team need optimization:

1. **Plugin install hooks** — Need to install plugin-bundled roles to global scope, but no provider supports install-time hooks
2. **role-repo interaction UX** — Current CLI flow is basic; needs to align with vercel-labs/skills interaction patterns
3. **OpenCode bridge** — Bridge logic is no longer needed since worker creation already handles skill installation

## Goals

- Provide a smooth onboarding experience via `agent-team init`
- Match vercel-labs/skills UX quality for role-repo commands
- Simplify codebase by removing unnecessary bridge logic

## Constraints & Assumptions

- Claude Code, Gemini CLI, and OpenCode all lack plugin install-time hooks (confirmed via research)
- Claude Code explicitly declined the feature request (Issue #9394, NOT_PLANNED)
- Workers already install skills during creation via `InstallSkillsForWorker()`
- Go CLI uses cobra for command structure

---

## Research Findings

### Provider Install Hook Support

| Provider | Install-time hook? | Status |
|---|---|---|
| Claude Code | No | Explicitly declined (NOT_PLANNED, #9394) |
| Gemini CLI | No | Not implemented; only settings prompts |
| OpenCode | No | `installation.updated` fires on app self-update, not plugin install |

### vercel-labs/skills Interaction Flow

Key patterns from the skills CLI:
- `npx skills add <source>` — supports owner/repo, GitHub URLs, local paths
- Auto-scans `SKILL.md` files to discover skills
- Interactive flow: select skills -> select agents -> select scope -> confirm
- `npx skills find` — fzf-style interactive search with API backend
- `npx skills check` / `update` — lock-file based update checking
- Non-interactive mode via flags (`--skill`, `--agent`, `--global`, `--yes`)

---

## Candidate Approaches

### Approach A: Full Interactive Alignment (Selected)

Full alignment with vercel-labs/skills interaction patterns.

**Pros:** Best UX consistency, aligns with skills ecosystem
**Cons:** Larger change set

### Approach B: Lightweight Enhancement

Only enhance `find` and `add` output without interactive selection.

**Pros:** Smaller change set, quick delivery
**Cons:** Inferior interaction experience

**Decision:** Approach A selected for full UX alignment.

---

## Recommended Design

### 1. `agent-team init` Command

```
Command: agent-team init [--global-only] [--skip-detect]
```

**Execution Flow:**

1. **Detect agent providers**
   - Check: claude, gemini, opencode, codex (via `which` commands)
   - Output: "Detected agents: claude, opencode"

2. **Initialize project `.agents/` structure** (if not exists)
   - Create `.agents/teams/`, `.agents/.gitkeep`
   - Output: "Initialized .agents/ directory"

3. **Install plugin-bundled roles to global**
   - Scan `$CLAUDE_PLUGIN_ROOT/skills/` for `role.yaml` files
   - For each role:
     - If `~/.agents/roles/<name>` exists with same hash → skip
     - If exists with different hash → prompt to update
     - If not exists → copy to `~/.agents/roles/<name>/`
   - Update `~/.agents/.role-lock.json`

4. **Config overview and guidance**
   - Show installed global roles list
   - Show detected agent providers
   - Show next steps guidance

**Implementation Details:**
- Plugin roles path: prefer `$CLAUDE_PLUGIN_ROOT/skills/`, fallback to executable-relative path
- Global install uses existing `GlobalRolesDir()` → `~/.agents/roles/`
- Lock file uses existing `ReadLockFile()`/`WriteLockFile()`
- `--global-only`: skip project structure init
- `--skip-detect`: skip agent provider detection

### 2. role-repo Interaction Refactoring

#### 2.1 `role-repo find` (replaces `search`)

```
Command: agent-team role-repo find [query]
```

**Interactive mode (no args):**
- fzf-style search interface
- Real-time GitHub API search with debouncing
- Select to auto-install via `role-repo add`

**Direct mode (with query):**
- Formatted results with install command hints
- Up to 6 results displayed

#### 2.2 `role-repo add` Interactive Flow

```
Command: agent-team role-repo add <source> [--role <name>] [--global] [--yes]
```

**Full interactive flow (no flags):**
1. Parse source (owner/repo, URL)
2. Scan remote repo for roles (via GitHub API tree)
3. Interactive role selection (checkbox, multi-select)
4. Scope selection: Project (.agents/teams/) or Global (~/.agents/roles/)
5. Confirmation and install
6. Update lock file

**Non-interactive (flags provided):**
```
agent-team role-repo add owner/repo --role frontend --global --yes
```

**Source formats (unchanged):**
- `owner/repo`
- `https://github.com/owner/repo`
- `owner/repo@branch`

#### 2.3 `role-repo list` Enhancement

Enhanced table output:
```
Project roles (.agents/teams/):
  NAME        SOURCE                    UPDATED
  frontend    jsonlee/agent-roles       2026-03-01

Global roles (~/.agents/roles/):
  NAME        SOURCE                    UPDATED
  devops      acme/team-roles           2026-03-02

Total: 2 roles (1 project, 1 global)
```

New: `--json` flag for machine-readable output.

### 3. Bridge Removal

#### Remove:
- `BridgeSkillsForProvider()` function in `internal/skills.go` (~55 lines)
- Bridge invocation in `cmd/hook_session_start.go` (OpenCode provider branch)

#### Preserve:
- `InstallSkillsForWorker()` — worker creation skill install
- `CopySkillsToWorktreeFromPath()` — worktree skill copy
- `findSkillPath()` — 5-layer search
- `runNpxSkillsAdd()` — remote download fallback

---

## File Change Scope

```
New:
  cmd/init.go                      — agent-team init command
  cmd/role_repo_find.go            — refactored find command (replaces search)

Modified:
  cmd/role_repo.go                 — register find, remove search
  cmd/role_repo_add.go             — interactive add flow
  cmd/role_repo_list.go            — enhanced output format
  cmd/hook_session_start.go        — remove bridge call
  internal/skills.go               — remove BridgeSkillsForProvider()
  internal/role_repo_install.go    — interactive selection logic

Potentially Removed:
  cmd/role_repo_search.go          — replaced by find
```

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| fzf interaction requires TTY | Non-TTY environments can't use interactive mode | Detect `isatty()`, fallback to flag mode |
| GitHub API rate limit | find command may be throttled | Use `$GITHUB_TOKEN` for authenticated requests |
| Bridge removal affects OpenCode non-worktree usage | Skills not auto-discovered outside worker | Acceptable; docs guide users to worker mode |
| Repeated init execution | May overwrite user modifications | Hash comparison; skip if identical |

## Testing Strategy

### Unit Tests
- `init` command: provider detection, directory creation, role install logic
- `find` command: search result parsing and formatting
- `add` interactive flow: state machine logic
- Post-bridge-removal session-start hook behavior

### Integration Tests
- `init` → `role-repo add` → `worker create` full flow
- `role-repo find` → select install → verify lock file

### Manual Verification
- Worker creation and skill discovery across providers (claude/opencode/gemini)

---

## Open Questions

None — all design decisions have been confirmed.
