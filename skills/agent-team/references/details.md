# agent-team Reference Details

## Directory Layout

### Main Repository

```
.agents/
  teams/
    <role-name>/                   <- role skill definition (managed by role-creator)
      SKILL.md
      system.md
      references/role.yaml
```

### Worker Worktree

```
.worktrees/<worker-id>/
  .gitignore                       <- excludes .gitignore, .claude/, .codex/, .gemini/, .opencode/, .tasks/, worker.yaml
  worker.yaml                      <- worker instance config (excluded from git)
  .claude/
    skills/                        <- dynamically copied on open
      <role-skill>/
      <dependency-skill-1>/
  .codex/
    skills/                        <- mirrored from .claude/skills/
      <role-skill>/
      <dependency-skill-1>/
  CLAUDE.md                        <- generated from role system.md
  AGENTS.md                        <- generated from role system.md
  GEMINI.md                        <- generated from role system.md
  .tasks/
    config.yaml                    <- verify defaults (command, timeout), lifecycle
    changes/                       <- active changes (managed by task system)
      <timestamp>-<slug>/
        change.yaml                <- status + task list + verify config
        proposal.md                <- work requirements (from controller)
        design.md                  <- architecture decisions (optional)
        tests.md                   <- acceptance criteria (written by worker, TDD)
```

### worker.yaml

```yaml
worker_id: frontend-dev-001
role: frontend-dev
role_scope: ""                     # "project" | "global", omitted when empty
role_path: ""                      # absolute path for global roles, omitted when empty
provider: claude                   # worker create defaults to claude when --provider is omitted
default_model: ""                  # persisted model reused by worker open when --model is omitted
main_session_id: ""
pane_id: ""
controller_pane_id: ""
created_at: "2026-02-26T10:00:00Z"
```

## Change Workflow

Changes are managed by the built-in task system. The controller creates a change via `agent-team worker assign` with optional `--proposal`, `--design`, and `--verify-cmd` flags. The worker then follows the TDD cycle defined in [worker-workflow.md](worker-workflow.md).

## Bidirectional Communication

Workers use `agent-team reply-main` to communicate with main controller.

Task completed (after verify):
```bash
agent-team reply-main "Task completed: <summary>; verify: passed"
```

Task completed but verify failed (still notify completion):
```bash
agent-team reply-main "Task completed: <summary>; verify: failed — <reason>"
```

Task completed but verify skipped:
```bash
agent-team reply-main "Task completed: <summary>; verify: skipped"
```

Blocked / needs decision on options:
```bash
agent-team reply-main "Need decision: <problem or options>"
```

Worker asks a question to the main controller:
```bash
agent-team reply-main "<question>"
```
Message appears in the controller's terminal as `[Worker: <worker-id>] <question>`.

Main controller replies:
```bash
agent-team reply <worker-id> "<answer>"
```
Reply appears in the worker's terminal as `[Main Controller Reply]`.

The worker AI must NOT proceed on blocked tasks until it receives a reply.

The controller's pane ID is saved in `worker.yaml` (`controller_pane_id`) when `agent-team worker open` runs.

`agent-team worker open <worker-id> [--provider <provider>] [--model <model>] [--new-window]`

- `--provider` updates `worker.yaml.provider` only when explicitly passed.
- `--model` updates `worker.yaml.default_model` only when explicitly passed.
- Omitting either flag reuses the persisted value already stored in `worker.yaml`.
- If `worker.yaml.provider` is missing, `open` uses `claude` for that launch only and does not write it back.
- `--new-window` is optional; the default behavior is to open in the current terminal workspace.

## Skills Copying

When `agent-team worker open` runs, it:

1. Reads the role's `references/role.yaml` to get the skills list
2. Finds each skill in:
   - Plugin built-in `$CLAUDE_PLUGIN_ROOT/skills/` (highest priority)
   - `.agents/teams/<skill-name>/` (project roles)
   - `skills/<skill-name>/` (project skills)
   - `.claude/skills/<skill-name>/` (local cached)
   - `~/.claude/skills/<skill-name>/` (global skills)
3. Copies the role skill directory and all dependency skills to:
   - `.claude/skills/<skill-name>/`
   - `.codex/skills/<skill-name>/`
   - `.opencode/skills/<skill-name>/`
   - `.gemini/skills/<skill-name>/`
4. All four directories are kept in sync (identical content)
