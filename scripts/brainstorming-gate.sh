#!/usr/bin/env bash
# scripts/brainstorming-gate.sh — PreToolUse(Write|Edit) Hook
# Warns when an active change has no design.md (or design.md is empty).
# All 5 conditions must be true to emit a warning.
set -euo pipefail

INPUT=$(cat)
CWD=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('cwd', ''))" 2>/dev/null || echo "")
[ -z "$CWD" ] && CWD="$(pwd)"

# 1. Must be in a worktree
GIT_COMMON=$(git -C "$CWD" rev-parse --git-common-dir 2>/dev/null || echo "")
[ -z "$GIT_COMMON" ] && exit 0
[[ "$GIT_COMMON" != /* ]] && exit 0
[[ "$CWD" != *"/.worktrees/"* ]] && exit 0

# 2. .tasks/ must exist
[ ! -d "$CWD/.tasks" ] && exit 0

# 3. .tasks/changes/ must exist
CHANGES_DIR="$CWD/.tasks/changes"
[ ! -d "$CHANGES_DIR" ] && exit 0

# 4. Find active (non-archived) change
ACTIVE=""
for d in "$CHANGES_DIR"/*/; do
    [ -d "$d" ] || continue
    name=$(basename "$d")
    yaml_file="$d/change.yaml"
    [ ! -f "$yaml_file" ] && continue
    STATUS=$(grep '^status:' "$yaml_file" | awk '{print $2}')
    [ "$STATUS" = "archived" ] && continue
    ACTIVE="$name"
    ACTIVE_DIR="$d"
    break
done
[ -z "$ACTIVE" ] && exit 0

# 5. design.md missing or empty
DESIGN_MD="${ACTIVE_DIR}design.md"
if [ ! -f "$DESIGN_MD" ] || [ ! -s "$DESIGN_MD" ]; then
    echo "[brainstorming-gate] Warning: Active change '$ACTIVE' has no design.md" >&2
    echo "[brainstorming-gate] Consider running /brainstorming to define requirements before editing." >&2
fi

exit 0
