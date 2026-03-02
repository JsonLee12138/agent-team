#!/usr/bin/env bash
# scripts/session-init.sh — SessionStart Hook
# Reads stdin JSON, detects worktree, injects role prompt via agent-team CLI.
set -euo pipefail

INPUT=$(cat)
CWD=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('cwd', ''))" 2>/dev/null || echo "")
[ -z "$CWD" ] && CWD="$(pwd)"

# --- Worktree detection ---
GIT_COMMON=$(git -C "$CWD" rev-parse --git-common-dir 2>/dev/null || echo "")
[ -z "$GIT_COMMON" ] && exit 0

# Non-absolute path = main repo, not a worktree
[[ "$GIT_COMMON" != /* ]] && exit 0
# Path must contain /.worktrees/
[[ "$CWD" != *"/.worktrees/"* ]] && exit 0

WORKER_ID=$(basename "$CWD")
MAIN_ROOT=$(dirname "$(dirname "$GIT_COMMON")")  # .git/worktrees → .git → root

echo "[session-init] Detected worktree: $CWD (worker: $WORKER_ID, root: $MAIN_ROOT)" >&2

# --- Locate config.yaml ---
CONFIG_PATH=""
for agents_dir in ".agents" "agents"; do
    candidate="$MAIN_ROOT/$agents_dir/workers/$WORKER_ID/config.yaml"
    if [ -f "$candidate" ]; then
        CONFIG_PATH="$candidate"
        break
    fi
done

if [ -z "$CONFIG_PATH" ]; then
    echo "[session-init] Warning: config.yaml not found for worker $WORKER_ID, skipping inject." >&2
    exit 0
fi

# --- Extract role from config.yaml ---
ROLE=$(python3 -c "
import sys, yaml
with open('$CONFIG_PATH') as f:
    d = yaml.safe_load(f)
print(d.get('role', ''))
" 2>/dev/null || grep -E '^role:' "$CONFIG_PATH" | awk '{print $2}' || echo "")

if [ -z "$ROLE" ]; then
    echo "[session-init] Warning: could not determine role from $CONFIG_PATH, skipping inject." >&2
    exit 0
fi

echo "[session-init] Injecting role prompt for role=$ROLE worker=$WORKER_ID" >&2

# --- Inject role prompt via CLI ---
agent-team _inject-role-prompt \
    --worktree "$CWD" \
    --worker-id "$WORKER_ID" \
    --role "$ROLE" \
    --root "$MAIN_ROOT" \
    2>&1 | sed 's/^/[session-init] /' >&2 || true
