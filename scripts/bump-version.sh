#!/usr/bin/env bash
set -euo pipefail

if [ $# -lt 1 ]; then
  echo "Usage: $0 <version>" >&2
  exit 1
fi

VERSION="$1"

# Validate semver format
if [[ ! "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$ ]]; then
  echo "Error: invalid version format: $VERSION (expected x.y.z)" >&2
  exit 1
fi

# Require jq
if ! command -v jq &>/dev/null; then
  echo "Error: jq is required but not found. Install via: brew install jq" >&2
  exit 1
fi

cd "$(git rev-parse --show-toplevel)"

# .claude-plugin/plugin.json
jq --arg v "$VERSION" '.version = $v' .claude-plugin/plugin.json > tmp.json && mv tmp.json .claude-plugin/plugin.json

# .claude-plugin/marketplace.json (metadata.version + plugins[].version)
jq --arg v "$VERSION" '.metadata.version = $v | .plugins[].version = $v' .claude-plugin/marketplace.json > tmp.json && mv tmp.json .claude-plugin/marketplace.json

# gemini-extension.json
jq --arg v "$VERSION" '.version = $v' gemini-extension.json > tmp.json && mv tmp.json gemini-extension.json

echo "Bumped version to $VERSION:"
echo "  .claude-plugin/plugin.json"
echo "  .claude-plugin/marketplace.json"
echo "  gemini-extension.json"
