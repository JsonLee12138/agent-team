# agent-team Reference Details

## Role directory layout

```
.worktrees/<name>/
  CLAUDE.md                          <- auto-generated from prompt.md on open
  agents/teams/<name>/
    config.yaml                      <- name, default_provider, default_model, pane_id, controller_pane_id
    prompt.md                        <- role system prompt (edit manually)
  openspec/
    specs/                           <- project specifications
    changes/                         <- active changes (managed by OpenSpec)
      <change-name>/
        .openspec.yaml               <- change metadata
        proposal.md                  <- brainstorming output from controller
        specs/                       <- delta specs (created by role)
        design.md                    <- design artifact (created by role)
        tasks.md                     <- task breakdown (created by role)
    config.yaml                      <- OpenSpec configuration
```

## Change workflow

Changes are managed by OpenSpec. The controller creates a change with a proposal via `agent-team assign`. The role then proceeds through the OpenSpec workflow:

1. `/opsx:continue` — create remaining artifacts (specs, design, tasks)
2. `/opsx:apply` — implement the tasks
3. `/opsx:verify` — validate implementation matches the design

## Bidirectional communication

Role asks a question to the main controller:
```bash
agent-team reply-main "<question>"
```
Message appears in the controller's terminal as `[Role: <name>] <question>`.

Main controller replies:
```bash
agent-team reply <rolename> "<answer>"
```
Reply appears in the role's terminal tab as `[Main Controller Reply]`.

The role AI must NOT proceed on blocked tasks until it receives a reply. The `prompt.md` template includes this communication protocol automatically.

The controller's pane ID is saved in `config.yaml` (`controller_pane_id`) when `agent-team open` runs, read from `$WEZTERM_PANE` or `$TMUX_PANE`.
