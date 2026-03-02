#!/usr/bin/env bash
# scripts/stop-guard.sh — Stop Hook
# Warns if there are unarchived openspec changes in the worktree.
set -euo pipefail

INPUT=$(cat)
CWD=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('cwd', ''))" 2>/dev/null || echo "")
[ -z "$CWD" ] && CWD="$(pwd)"

# --- Worktree detection ---
GIT_COMMON=$(git -C "$CWD" rev-parse --git-common-dir 2>/dev/null || echo "")
[ -z "$GIT_COMMON" ] && exit 0
[[ "$GIT_COMMON" != /* ]] && exit 0
[[ "$CWD" != *"/.worktrees/"* ]] && exit 0

CHANGES_DIR="$CWD/openspec/changes"
[ ! -d "$CHANGES_DIR" ] && exit 0

# --- Check for unarchived changes ---
for d in "$CHANGES_DIR"/*/; do
    [ -d "$d" ] || continue
    name=$(basename "$d")
    [ "$name" = "archive" ] && continue
    [ -d "$d/archive" ] && continue
    echo "[stop-guard] Warning: unarchived change '$name' exists in $CWD" >&2
    echo "[stop-guard] Run 'openspec archive --change $name' before stopping." >&2
done

exit 0
