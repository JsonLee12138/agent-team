#!/usr/bin/env bash
# scripts/task-archive.sh — TaskCompleted Hook
# Archives the active openspec change and notifies main controller.
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

# --- Check openspec availability ---
if ! command -v openspec &>/dev/null; then
    echo "[task-archive] openspec not found, skipping archive." >&2
    agent-team reply-main "Task completed by $WORKER_ID; archive skipped (openspec not found)" 2>/dev/null || true
    exit 0
fi

CHANGES_DIR="$CWD/openspec/changes"
if [ ! -d "$CHANGES_DIR" ]; then
    echo "[task-archive] No openspec/changes/ directory, skipping archive." >&2
    exit 0
fi

# --- Find active change (non-archived subdirectory) ---
ACTIVE=""
for d in "$CHANGES_DIR"/*/; do
    [ -d "$d" ] || continue
    name=$(basename "$d")
    [ "$name" = "archive" ] && continue
    [ -d "$d/archive" ] && continue
    ACTIVE="$name"
    break
done

if [ -z "$ACTIVE" ]; then
    echo "[task-archive] No active change found to archive." >&2
    exit 0
fi

echo "[task-archive] Archiving change: $ACTIVE" >&2

# --- Execute archive ---
if openspec archive --change "$ACTIVE" --dir "$CWD" 2>&1 | sed 's/^/[task-archive] /' >&2; then
    agent-team reply-main "Task completed by $WORKER_ID; archive: success (change: $ACTIVE)" 2>/dev/null || true
else
    agent-team reply-main "Task completed by $WORKER_ID; archive: failed (change: $ACTIVE)" 2>/dev/null || true
fi
