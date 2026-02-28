# agent-team Reference Details

## Directory Layout

### Main Repository

```
agents/
  teams/
    <role-name>/                   <- role skill definition (managed by role-creator)
      SKILL.md
      system.md
      references/role.yaml
  workers/
    <worker-id>/                   <- worker config
      config.yaml                  <- worker_id, role, provider, pane_id, etc.
```

### Worker Worktree

```
.worktrees/<worker-id>/
  .gitignore                       <- excludes .gitignore, .claude/, .codex/, openspec/
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
  openspec/
    specs/                         <- project specifications
    changes/                       <- active changes (managed by OpenSpec)
      <task-slug>/
        .openspec.yaml             <- change metadata
        design.md                  <- brainstorming output (from controller)
        proposal.md                <- work requirements (from controller)
        specs/                     <- delta specs (created by worker)
        tasks.md                   <- task breakdown (created by worker)
    config.yaml                    <- OpenSpec configuration
```

### Worker config.yaml

```yaml
worker_id: frontend-dev-001
role: frontend-dev
default_provider: claude
default_model: ""
pane_id: ""
controller_pane_id: ""
created_at: "2026-02-26T10:00:00Z"
```

## Change Workflow

Changes are managed by OpenSpec. The controller creates a change with design and proposal via `agent-team worker assign`. The worker then proceeds through the OpenSpec workflow:

1. `/opsx:continue` — create remaining artifacts (specs, design, tasks)
2. `/opsx:apply` — implement the tasks
3. `/opsx:verify` — validate implementation
4. Attempt `/openspec archive` for the completed change
5. If `/openspec archive` is unavailable (for example `command not found`), fallback to `/prompts:openspec-archive`
6. Send `agent-team reply-main` AFTER the archive attempt with archive status

## Bidirectional Communication

Workers use `agent-team reply-main` to communicate with main controller.

Task completed (after archive attempt):
```bash
agent-team reply-main "Task completed: <summary>; archive: success via </openspec archive|/prompts:openspec-archive>"
```

Task completed but archive failed (still notify completion):
```bash
agent-team reply-main "Task completed: <summary>; archive failed via </openspec archive|/prompts:openspec-archive>: <error>"
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

The controller's pane ID is saved in `config.yaml` (`controller_pane_id`) when `agent-team worker open` runs.

## Skills Copying

When `agent-team worker open` runs, it:

1. Reads the role's `references/role.yaml` to get the skills list
2. Finds each skill in:
   - `agents/teams/<skill-name>/` (project roles)
   - `skills/<skill-name>/` (project skills)
   - `~/.claude/skills/<skill-name>/` (global skills)
3. Copies the role skill directory and all dependency skills to:
   - `.claude/skills/<skill-name>/`
   - `.codex/skills/<skill-name>/`
4. Both directories are kept in sync (identical content)
