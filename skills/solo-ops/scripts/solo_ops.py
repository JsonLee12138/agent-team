#!/usr/bin/env python3
"""solo-ops - AI team role manager

Manages AI assistant roles with git worktrees and WezTerm tabs.
Run: python /path/to/solo_ops.py <command>
"""

import sys
import os
import re
import subprocess
import time
from datetime import datetime, timezone
from pathlib import Path


# ─── helpers ─────────────────────────────────────────────────────────────────

SUPPORTED_PROVIDERS = {"claude", "codex", "opencode"}


def get_session_backend():
    backend = os.environ.get("SOLO_OPS_BACKEND", "wezterm").strip().lower()
    if backend == "tmux":
        return "tmux"
    return "wezterm"


def parse_provider_and_model(args):
    provider = ""
    model = ""
    i = 0
    while i < len(args):
        token = args[i]
        if token in ("-m", "--model"):
            if i + 1 >= len(args):
                raise ValueError("missing model value")
            model = args[i + 1]
            i += 2
            continue
        if token in SUPPORTED_PROVIDERS and not provider:
            provider = token
            i += 1
            continue
        raise ValueError(f"unsupported option: {token}")
    return provider, model


def build_launch_cmd(provider, model):
    base_map = {
        "claude": "claude --dangerously-skip-permissions",
        "codex": "codex --dangerously-bypass-approvals-and-sandbox",
        "opencode": "opencode",
    }
    base = base_map.get(provider or "claude", base_map["claude"])
    if model:
        return f"{base} --model {model}"
    return base


def resolve_model(cli_model, config_model):
    """Resolve model: CLI argument takes precedence over config."""
    return cli_model or config_model or ""


def find_git_root():
    result = subprocess.run(
        ['git', 'rev-parse', '--show-toplevel'],
        capture_output=True, text=True
    )
    if result.returncode != 0:
        print("Error: not in a git repository", file=sys.stderr)
        sys.exit(1)
    return result.stdout.strip()


def find_wt_base(root):
    if Path(root, '.worktrees').is_dir():
        return '.worktrees'
    elif Path(root, 'worktrees').is_dir():
        return 'worktrees'
    return '.worktrees'


def list_roles(root, wt_base):
    base = Path(root, wt_base)
    if not base.is_dir():
        return []
    roles = []
    for d in sorted(base.iterdir()):
        if not d.is_dir():
            continue
        name = d.name
        if Path(d, 'agents', 'teams', name, 'config.yaml').is_file():
            roles.append(name)
    return roles


def cfg_get(filepath, key):
    try:
        with open(filepath) as f:
            for line in f:
                if line.startswith(f'{key}:'):
                    return line[len(key) + 1:].strip().strip('"')
    except FileNotFoundError:
        pass
    return ''


def cfg_set(filepath, key, value):
    filepath = Path(filepath)
    content = filepath.read_text() if filepath.exists() else ''
    lines = content.splitlines(keepends=True)
    pattern = re.compile(f'^{re.escape(key)}:.*')
    new_line = f'{key}: {value}\n'
    found = False
    new_lines = []
    for line in lines:
        if pattern.match(line):
            new_lines.append(new_line)
            found = True
        else:
            new_lines.append(line)
    if not found:
        new_lines.append(new_line)
    filepath.write_text(''.join(new_lines))


def pane_alive(pane_id):
    if not pane_id:
        return False
    backend = get_session_backend()
    if backend == "tmux":
        result = subprocess.run(
            ['tmux', 'list-panes', '-a', '-F', '#{pane_id}'],
            capture_output=True, text=True
        )
        if result.returncode != 0:
            return False
        return str(pane_id) in {line.strip() for line in result.stdout.splitlines()}

    result = subprocess.run(
        ['wezterm', 'cli', 'list'],
        capture_output=True, text=True
    )
    if result.returncode != 0:
        return False
    for line in result.stdout.splitlines()[1:]:  # skip header
        parts = line.split()
        if len(parts) >= 3 and parts[2] == str(pane_id):
            return True
    return False


def pane_send(pane_id, text):
    """Send text + Enter to a WezTerm pane via stdin pipe."""
    backend = get_session_backend()
    if backend == "tmux":
        subprocess.run(
            ['tmux', 'send-keys', '-t', str(pane_id), '-l', text],
            capture_output=True
        )
        subprocess.run(
            ['tmux', 'send-keys', '-t', str(pane_id), 'Enter'],
            capture_output=True
        )
        return

    subprocess.run(
        ['wezterm', 'cli', 'send-text', '--pane-id', str(pane_id), '--no-paste'],
        input=text.encode(),
        capture_output=True
    )
    time.sleep(0.1)
    subprocess.run(
        ['wezterm', 'cli', 'send-text', '--pane-id', str(pane_id), '--no-paste'],
        input=b'\r',
        capture_output=True
    )


# ─── commands ────────────────────────────────────────────────────────────────

def cmd_create(name):
    if not name:
        print("Usage: solo-ops create <name>", file=sys.stderr)
        sys.exit(1)

    root = find_git_root()
    wt_base = find_wt_base(root)
    wt_path = Path(root, wt_base, name)
    branch = f'team/{name}'

    if wt_path.is_dir():
        print(f"Error: role '{name}' already exists at {wt_path}", file=sys.stderr)
        sys.exit(1)

    print(f"Creating role '{name}'...")
    subprocess.run(['git', 'worktree', 'add', str(wt_path), '-b', branch],
                   cwd=root, check=True)

    teams_dir = wt_path / 'agents' / 'teams' / name
    (teams_dir / 'tasks' / 'pending').mkdir(parents=True)
    (teams_dir / 'tasks' / 'done').mkdir(parents=True)

    now = datetime.now(timezone.utc).strftime('%Y-%m-%dT%H:%M:%SZ')
    (teams_dir / 'config.yaml').write_text(
        f'name: {name}\n'
        f'description: ""\n'
        f'default_provider: claude\n'
        f'default_model: ""\n'
        f'created_at: {now}\n'
        f'pane_id: ""\n'
    )

    (teams_dir / 'prompt.md').write_text(
        f'# Role: {name}\n\n'
        '## Description\n'
        "Describe this role's responsibilities here.\n\n"
        '## Expertise\n'
        '- List key areas of expertise\n\n'
        '## Behavior\n'
        '- How this role approaches tasks\n'
        '- Communication style and boundaries\n\n'
        '## Communication Protocol\n\n'
        'When you need clarification or have a question for the main controller, use:\n\n'
        '```bash\n'
        f'ask claude "{name}: <your question here>"\n'
        '```\n\n'
        'Wait for the main controller to reply. Replies will appear as:\n'
        '`[Main Controller Reply]`\n\n'
        'Do NOT proceed on blocked tasks until you receive a reply.\n'
    )

    print(f"✓ Created role '{name}' at {wt_path}")
    print(f"  → Edit {teams_dir}/prompt.md to define the role")
    print(f"  → Edit {teams_dir}/config.yaml to set default_provider")


def cmd_delete(name):
    if not name:
        print("Usage: solo-ops delete <name>", file=sys.stderr)
        sys.exit(1)

    root = find_git_root()
    wt_base = find_wt_base(root)
    wt_path = Path(root, wt_base, name)
    config = wt_path / 'agents' / 'teams' / name / 'config.yaml'

    if not wt_path.is_dir():
        print(f"Error: role '{name}' not found", file=sys.stderr)
        sys.exit(1)

    print(f"Deleting role '{name}'...")
    pane_id = cfg_get(str(config), 'pane_id')
    if pane_alive(pane_id):
        if get_session_backend() == "tmux":
            close_result = subprocess.run(
                ['tmux', 'kill-pane', '-t', str(pane_id)],
                capture_output=True
            )
        else:
            close_result = subprocess.run(
                ['wezterm', 'cli', 'kill-pane', '--pane-id', str(pane_id)],
                capture_output=True
            )
        if close_result.returncode != 0:
            print(
                f"Warning: failed to close pane {pane_id}; continuing delete",
                file=sys.stderr
            )

    result = subprocess.run(
        ['git', 'worktree', 'remove', str(wt_path), '--force'],
        cwd=root, capture_output=True
    )
    if result.returncode != 0:
        import shutil
        try:
            shutil.rmtree(wt_path)
        except FileNotFoundError:
            # git worktree remove may have already deleted the directory
            pass

    subprocess.run(['git', 'branch', '-D', f'team/{name}'],
                   cwd=root, capture_output=True)
    print(f"✓ Deleted role '{name}'")


def cmd_open(name, provider='', model=''):
    if not name:
        print("Usage: solo-ops open <name> [claude|codex|opencode] [--model <model>]", file=sys.stderr)
        sys.exit(1)

    root = find_git_root()
    wt_base = find_wt_base(root)
    wt_path = Path(root, wt_base, name)
    teams_dir = wt_path / 'agents' / 'teams' / name
    config = teams_dir / 'config.yaml'

    if not teams_dir.is_dir():
        print(f"Error: role '{name}' not found", file=sys.stderr)
        sys.exit(1)

    if not provider:
        provider = cfg_get(str(config), 'default_provider') or 'claude'

    pane_id = cfg_get(str(config), 'pane_id')
    if pane_alive(pane_id):
        print(f"Role '{name}' is already running (pane {pane_id})")
        return

    # Generate CLAUDE.md from prompt.md so the AI reads the role on startup
    prompt_md = teams_dir / 'prompt.md'
    claude_md = wt_path / 'CLAUDE.md'
    if prompt_md.exists():
        claude_md.write_text(prompt_md.read_text())

    # Append git worktree context so the AI knows where and how to commit
    with open(claude_md, 'a') as f:
        f.write(
            '\n## Development Environment\n\n'
            'You are working in an **isolated git worktree**. All development MUST happen here:\n\n'
            f'- **Working directory**: `{wt_path}`\n'
            f'- **Git branch**: `team/{name}` (your dedicated branch)\n'
            f'- **Main project root**: `{root}`\n\n'
            '### Git Rules\n\n'
            f'- All changes and commits go to the `team/{name}` branch — this is already checked out\n'
            '- **Never** run `git checkout`, `git switch`, or change branches\n'
            '- **Never** merge or rebase from within this worktree\n'
            '- Commit regularly with clear messages as you complete work\n'
            '- When your task is fully done, move its file from `tasks/pending/` to `tasks/done/`\n\n'
            'The main controller will merge your branch back to main when ready.\n'
        )

    launch_cmd = build_launch_cmd(provider, model)

    backend = get_session_backend()
    current_pane = os.environ.get('WEZTERM_PANE', '')

    if backend == "tmux":
        result = subprocess.run(
            ['tmux', 'new-session', '-d', '-P', '-F', '#{pane_id}', '-c', str(wt_path)],
            capture_output=True, text=True
        )
    else:
        # Spawn a new tab with an interactive shell (no command = WezTerm opens default shell)
        # This avoids the issue where `zsh -c "claude"` causes TUI apps to exit immediately
        result = subprocess.run(
            ['wezterm', 'cli', 'spawn', '--cwd', str(wt_path)],
            capture_output=True, text=True
        )

    if result.returncode != 0 or not result.stdout.strip():
        print(f"✗ Failed to open {backend} session for '{name}'", file=sys.stderr)
        sys.exit(1)

    new_pane_id = result.stdout.strip()

    if backend == "tmux":
        subprocess.run(
            ['tmux', 'rename-window', '-t', str(new_pane_id), name],
            capture_output=True
        )
    else:
        # Set tab title
        subprocess.run(
            ['wezterm', 'cli', 'set-tab-title', '--pane-id', new_pane_id, name],
            capture_output=True
        )

        # Return focus to the caller's pane — don't steal focus
        if current_pane:
            subprocess.run(
                ['wezterm', 'cli', 'activate-pane', '--pane-id', current_pane],
                capture_output=True
            )

    cfg_set(str(config), 'pane_id', new_pane_id)

    # Wait for the interactive shell to fully initialize (zsh + plugins), then launch AI
    print("  Waiting for shell to initialize...")
    time.sleep(2)
    pane_send(new_pane_id, launch_cmd)

    print(f"✓ Opened role '{name}' ({provider}) in {backend} [pane {new_pane_id}]")


def cmd_open_all(provider='', model=''):
    root = find_git_root()
    wt_base = find_wt_base(root)
    roles = list_roles(root, wt_base)
    if not roles:
        print("No roles found. Create one with: solo-ops create <name>", file=sys.stderr)
        sys.exit(1)
    for role in roles:
        cmd_open(role, provider, model)


def cmd_assign(name, task, provider='', model=''):
    if not name or not task:
        print('Usage: solo-ops assign <name> "<task description>" [claude|codex|opencode] [--model <model>]',
              file=sys.stderr)
        sys.exit(1)

    root = find_git_root()
    wt_base = find_wt_base(root)
    teams_dir = Path(root, wt_base, name, 'agents', 'teams', name)
    config = teams_dir / 'config.yaml'

    if not teams_dir.is_dir():
        print(f"Error: role '{name}' not found", file=sys.stderr)
        sys.exit(1)

    ts = datetime.now().strftime('%Y-%m-%d-%H-%M-%S')
    slug = re.sub(r'[^a-z0-9]+', '-', task.lower()).strip('-')[:50] or 'task'
    task_file = teams_dir / 'tasks' / 'pending' / f'{ts}-{slug}.md'

    now_utc = datetime.now(timezone.utc).strftime('%Y-%m-%dT%H:%M:%SZ')
    task_file.write_text(
        f'# Task: {task}\n\n'
        f'Assigned: {now_utc}\n'
        f'Status: pending\n\n'
        '## Description\n\n'
        f'{task}\n\n'
        '## Notes\n\n'
        '_Add implementation notes here_\n'
    )
    print(f"✓ Task file: {task_file}")

    pane_id = cfg_get(str(config), 'pane_id')
    if not pane_alive(pane_id):
        print(f"Role '{name}' is not running, opening session first...")
        cmd_open(name, provider, model)
        pane_id = cfg_get(str(config), 'pane_id')
        print("  Waiting for AI to initialize...")
        time.sleep(3)

    task_rel = f'agents/teams/{name}/tasks/pending/{task_file.name}'
    msg = (
        f'New task assigned: {task}\n'
        f'Please read the task file at: {task_rel}\n'
        f'When complete, move it to agents/teams/{name}/tasks/done/'
    )
    pane_send(pane_id, msg)
    print(f"✓ Assigned to '{name}': {task}")


def cmd_status():
    root = find_git_root()
    wt_base = find_wt_base(root)
    roles = list_roles(root, wt_base)
    if not roles:
        print("No roles found. Create one with: solo-ops create <name>")
        return

    print(f"{'Role':<16} {'Status':<24} {'Pending Tasks'}")
    print(f"{'─' * 16} {'─' * 24} {'─' * 13}")
    for role in roles:
        config = Path(root, wt_base, role, 'agents', 'teams', role, 'config.yaml')
        pending_dir = Path(root, wt_base, role, 'agents', 'teams', role, 'tasks', 'pending')
        pane_id = cfg_get(str(config), 'pane_id')
        status = f'✓ running [p:{pane_id}]' if pane_alive(pane_id) else '✗ offline'
        count = len(list(pending_dir.glob('*.md'))) if pending_dir.is_dir() else 0
        print(f"{role:<16} {status:<24} {count}")


def cmd_reply(name, answer):
    if not name or not answer:
        print('Usage: solo-ops reply <name> "<answer>"', file=sys.stderr)
        sys.exit(1)

    root = find_git_root()
    wt_base = find_wt_base(root)
    config = Path(root, wt_base, name, 'agents', 'teams', name, 'config.yaml')

    if not config.is_file():
        print(f"Error: role '{name}' not found", file=sys.stderr)
        sys.exit(1)

    pane_id = cfg_get(str(config), 'pane_id')
    if not pane_alive(pane_id):
        print(f"Error: role '{name}' is not running", file=sys.stderr)
        sys.exit(1)

    pane_send(pane_id, f'[Main Controller Reply] {answer}')
    print(f"✓ Replied to '{name}'")


def cmd_merge(name):
    if not name:
        print("Usage: solo-ops merge <name>", file=sys.stderr)
        sys.exit(1)

    root = find_git_root()
    wt_base = find_wt_base(root)
    wt_path = Path(root, wt_base, name)
    branch = f'team/{name}'

    if not wt_path.is_dir():
        print(f"Error: role '{name}' not found", file=sys.stderr)
        sys.exit(1)

    result = subprocess.run(
        ['git', 'symbolic-ref', '--short', 'HEAD'],
        cwd=root, capture_output=True, text=True
    )
    main_branch = result.stdout.strip() if result.returncode == 0 else 'main'

    print(f"Merging branch '{branch}' into '{main_branch}'...")
    subprocess.run(
        ['git', 'merge', branch, '--no-ff', '-m',
         f"merge: integrate work from team role '{name}'"],
        cwd=root, check=True
    )
    print(f"✓ Merged '{name}' into {main_branch}")
    print(f"  → Run 'solo-ops delete {name}' to remove the worktree when done")


def cmd_install():
    """Install skill to ~/.claude/skills/solo-ops/ and create ~/.local/bin/solo-ops symlink."""
    script_path = Path(__file__).resolve()
    skill_root = script_path.parent.parent
    skill_target = Path.home() / '.claude' / 'skills' / 'solo-ops'

    (skill_target / 'scripts').mkdir(parents=True, exist_ok=True)

    import shutil
    target_script = skill_target / 'scripts' / 'solo_ops.py'
    if script_path != target_script:
        shutil.copy2(script_path, target_script)
    target_script.chmod(0o755)

    skill_md = skill_root / 'SKILL.md'
    if skill_md.exists():
        shutil.copy2(skill_md, skill_target / 'SKILL.md')

    print(f"✓ Installed skill: {skill_target}")

    # Create ~/.local/bin/solo-ops symlink for convenience
    bin_dir = Path.home() / '.local' / 'bin'
    bin_dir.mkdir(parents=True, exist_ok=True)
    bin_link = bin_dir / 'solo-ops'
    if bin_link.exists() or bin_link.is_symlink():
        bin_link.unlink()
    bin_link.symlink_to(target_script)
    print(f"✓ Symlinked: {bin_link} -> {target_script}")
    print(f"\nRun directly (no PATH needed):")
    print(f"  python {target_script} <command>")


# ─── dispatch ────────────────────────────────────────────────────────────────

def main_for_test(args):
    """Test dispatcher that takes explicit args instead of sys.argv."""
    cmd = args[0] if args else 'help'
    rest = args[1:]

    if cmd == 'create':
        cmd_create(rest[0] if rest else '')
    elif cmd == 'delete':
        cmd_delete(rest[0] if rest else '')
    elif cmd == 'open':
        if len(rest) > 0:
            name = rest[0]
            provider, model = parse_provider_and_model(rest[1:])
            cmd_open(name, provider, model)
    elif cmd == 'open-all':
        provider, model = parse_provider_and_model(rest)
        cmd_open_all(provider, model)
    elif cmd == 'assign':
        if len(rest) > 1:
            name = rest[0]
            task = rest[1]
            provider, model = parse_provider_and_model(rest[2:])
            cmd_assign(name, task, provider, model)
    elif cmd == 'reply':
        cmd_reply(rest[0] if rest else '', rest[1] if len(rest) > 1 else '')
    elif cmd == 'status':
        cmd_status()
    elif cmd == 'merge':
        cmd_merge(rest[0] if rest else '')
    elif cmd in ('help', ''):
        print(HELP_TEXT)
    else:
        print(f"Unknown command: {cmd}. Run 'solo-ops help'", file=sys.stderr)


HELP_TEXT = """\
solo-ops — AI team role manager

  create <name>                          Create role + git worktree (branch: team/<name>)
  delete <name>                          Remove role + worktree
  open <name> [provider] [--model <m>]   Open role session (provider: claude|codex|opencode)
  open-all [provider] [--model <m>]      Open all role sessions
  assign <name> "<task>" [provider] [--model <m>]   Write task file + notify session
  reply <name> "<answer>"                Send a reply to a role's running session
  status                                 Show all roles, running state, pending task count
  merge <name>                           Merge team/<name> branch back to current branch

Providers: claude, codex, opencode (default: claude)
Model flag: --model <model-identifier>  Specify AI model (e.g., claude-sonnet-4-6, openai/gpt-5)

Roles live in: .worktrees/<name>/agents/teams/<name>/

Run as:
  python3 <skill-base-dir>/scripts/solo_ops.py <command>

Tmux backend:
  SOLO_OPS_BACKEND=tmux python3 <skill-base-dir>/scripts/solo_ops.py <command>
"""


def main():
    args = sys.argv[1:]
    main_for_test(args)


if __name__ == '__main__':
    main()
