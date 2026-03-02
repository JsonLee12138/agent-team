#!/usr/bin/env bash
# scripts/teammate-idle-notify.sh — TeammateIdle Hook
# Notifies main controller when a worker goes idle.
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

echo "[teammate-idle] Worker $WORKER_ID is idle" >&2

# Notify main controller
agent-team reply-main "Worker idle: $WORKER_ID" 2>/dev/null || true

exit 0
