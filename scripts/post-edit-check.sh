#!/usr/bin/env bash
# scripts/post-edit-check.sh — PostToolUse(Write|Edit) Hook
# Runs quality_checks defined in role.yaml with 15s timeout per check.
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
MAIN_ROOT=$(dirname "$(dirname "$GIT_COMMON")")

# --- Locate role.yaml ---
ROLE_YAML=""
for agents_dir in ".agents" "agents"; do
    # Read role from worker config first
    for config_candidate in "$MAIN_ROOT/$agents_dir/workers/$WORKER_ID/config.yaml"; do
        [ -f "$config_candidate" ] || continue
        ROLE=$(python3 -c "
import sys, yaml
with open('$config_candidate') as f:
    d = yaml.safe_load(f)
print(d.get('role', ''))
" 2>/dev/null || echo "")
        if [ -n "$ROLE" ]; then
            candidate="$MAIN_ROOT/$agents_dir/teams/$ROLE/references/role.yaml"
            if [ -f "$candidate" ]; then
                ROLE_YAML="$candidate"
                break 2
            fi
        fi
    done
done

if [ -z "$ROLE_YAML" ]; then
    exit 0
fi

# --- Parse quality_checks from role.yaml ---
CHECKS=$(python3 - <<'PYEOF'
import sys, yaml
role_yaml = sys.argv[1]
try:
    with open(role_yaml) as f:
        d = yaml.safe_load(f)
    checks = d.get('quality_checks', {})
    if isinstance(checks, dict):
        for name, cmd in checks.items():
            if cmd and cmd.strip():
                print(f"{name}\t{cmd.strip()}")
except Exception as e:
    print(f"Error: {e}", file=sys.stderr)
PYEOF
"$ROLE_YAML" 2>/dev/null || echo "")

[ -z "$CHECKS" ] && exit 0

echo "[post-edit-check] Running quality checks for worker $WORKER_ID" >&2

# --- Run each check with 15s timeout ---
while IFS=$'\t' read -r CHECK_NAME CHECK_CMD; do
    [ -z "$CHECK_CMD" ] && continue
    echo "[post-edit-check] Running: $CHECK_NAME → $CHECK_CMD" >&2
    if timeout 15 bash -c "cd '$CWD' && $CHECK_CMD" 2>&1 | sed "s/^/[post-edit-check:$CHECK_NAME] /" >&2; then
        echo "[post-edit-check] $CHECK_NAME: PASS" >&2
    else
        echo "[post-edit-check] $CHECK_NAME: FAIL (exit $?)" >&2
    fi
done <<< "$CHECKS"

exit 0
