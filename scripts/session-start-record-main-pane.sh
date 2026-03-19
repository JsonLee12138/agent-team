#!/usr/bin/env bash
set -euo pipefail

repo_root="$(git rev-parse --show-toplevel 2>/dev/null || true)"
[ -n "$repo_root" ] || exit 0

common_dir="$(git rev-parse --git-common-dir 2>/dev/null || true)"
[ -n "$common_dir" ] || exit 0

if [ "${common_dir#./}" = ".git" ] || [ "$common_dir" = ".git" ]; then
  project_root="$repo_root"
else
  case "$common_dir" in
    /*) project_root="$(cd "$(dirname "$common_dir")" && pwd)" ;;
    *) project_root="$(cd "$repo_root/$common_dir/.." && pwd)" ;;
  esac
fi

if [ "$repo_root" != "$project_root" ]; then
  exit 0
fi

"${AGENT_TEAM_BIN:-agent-team}" _record-main-pane --root "$project_root" >/dev/null 2>&1 || true
