#!/usr/bin/env bash
# scripts/stop-guard.sh — Stop Hook
# Warns if there are incomplete task changes in the worktree.
set -euo pipefail

INPUT=$(cat)
CWD=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('cwd', ''))" 2>/dev/null || echo "")
[ -z "$CWD" ] && CWD="$(pwd)"

# --- Worktree detection ---
GIT_COMMON=$(git -C "$CWD" rev-parse --git-common-dir 2>/dev/null || echo "")
[ -z "$GIT_COMMON" ] && exit 0
[[ "$GIT_COMMON" != /* ]] && exit 0
[[ "$CWD" != *"/.worktrees/"* ]] && exit 0

CHANGES_DIR="$CWD/.tasks/changes"
[ ! -d "$CHANGES_DIR" ] && exit 0

# --- Check for incomplete changes ---
for d in "$CHANGES_DIR"/*/; do
    [ -d "$d" ] || continue
    name=$(basename "$d")
    yaml_file="$d/change.yaml"
    [ ! -f "$yaml_file" ] && continue
    STATUS=$(grep '^status:' "$yaml_file" | awk '{print $2}')
    [ "$STATUS" = "archived" ] && continue
    echo "[stop-guard] Warning: incomplete change '$name' exists in $CWD" >&2
    echo "[stop-guard] Run 'agent-team task archive <worker-id> $name' before stopping." >&2
done

exit 0
