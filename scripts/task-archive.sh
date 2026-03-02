#!/usr/bin/env bash
# scripts/task-archive.sh — TaskCompleted Hook
# Archives the active task change and notifies main controller.
set -euo pipefail

INPUT=$(cat)
CWD=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('cwd', ''))" 2>/dev/null || echo "")
[ -z "$CWD" ] && CWD="$(pwd)"

# --- Worktree detection ---
GIT_COMMON=$(git -C "$CWD" rev-parse --git-common-dir 2>/dev/null || echo "")
[ -z "$GIT_COMMON" ] && exit 0
[[ "$GIT_COMMON" != /* ]] && exit 0
[[ "$CWD" != *"/.worktrees/"* ]] && exit 0

WORKER_ID=$(basename "$CWD")

echo "[task-archive] TaskCompleted in worktree $WORKER_ID" >&2

CHANGES_DIR="$CWD/.tasks/changes"
if [ ! -d "$CHANGES_DIR" ]; then
    echo "[task-archive] No .tasks/changes/ directory, skipping archive." >&2
    exit 0
fi

# --- Find active change (non-archived) ---
ACTIVE=""
for d in "$CHANGES_DIR"/*/; do
    [ -d "$d" ] || continue
    name=$(basename "$d")
    yaml_file="$d/change.yaml"
    [ ! -f "$yaml_file" ] && continue
    STATUS=$(grep '^status:' "$yaml_file" | awk '{print $2}')
    [ "$STATUS" = "archived" ] && continue
    ACTIVE="$name"
    break
done

if [ -z "$ACTIVE" ]; then
    echo "[task-archive] No active change found to archive." >&2
    exit 0
fi

echo "[task-archive] Archiving change: $ACTIVE" >&2

# --- Execute archive ---
if agent-team task archive "$ACTIVE" --dir "$CWD" 2>&1 | sed 's/^/[task-archive] /' >&2; then
    agent-team reply-main "Task completed by $WORKER_ID; verify: passed (change: $ACTIVE)" 2>/dev/null || true
else
    agent-team reply-main "Task completed by $WORKER_ID; verify: failed (change: $ACTIVE)" 2>/dev/null || true
fi
