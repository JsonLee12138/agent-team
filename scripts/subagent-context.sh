#!/usr/bin/env bash
# scripts/subagent-context.sh — SubagentStart Hook
# Reads AGENT_TEAM block from parent worktree's CLAUDE.md and outputs to stderr.
set -euo pipefail

INPUT=$(cat)
PARENT_CWD=$(echo "$INPUT" | python3 -c "import sys,json; d=json.load(sys.stdin); print(d.get('parent_cwd', d.get('cwd', '')))" 2>/dev/null || echo "")
[ -z "$PARENT_CWD" ] && exit 0

# Only inject if parent is a worktree
[[ "$PARENT_CWD" != *"/.worktrees/"* ]] && exit 0

CLAUDE_MD="$PARENT_CWD/CLAUDE.md"
[ ! -f "$CLAUDE_MD" ] && exit 0

# Extract AGENT_TEAM block
BLOCK=$(python3 - <<'PYEOF'
import sys, re
content = open(sys.argv[1]).read()
m = re.search(r'<!-- AGENT_TEAM:START -->(.*?)<!-- AGENT_TEAM:END -->', content, re.DOTALL)
if m:
    print(m.group(1).strip())
PYEOF
"$CLAUDE_MD" 2>/dev/null || echo "")

if [ -n "$BLOCK" ]; then
    echo "[subagent-context] Injecting AGENT_TEAM role context from $PARENT_CWD" >&2
    echo "$BLOCK" >&2
fi

exit 0
