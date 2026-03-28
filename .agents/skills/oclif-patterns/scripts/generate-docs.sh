#!/usr/bin/env bash

# Generate oclif documentation
# Usage: ./generate-docs.sh

set -e

echo "Generating oclif documentation..."

# Check if oclif is installed
if ! command -v oclif &> /dev/null; then
  echo "Error: oclif CLI not found. Install with: npm install -g oclif"
  exit 1
fi

# Generate manifest
echo "→ Generating command manifest..."
oclif manifest

# Generate README
echo "→ Generating README..."
oclif readme

echo "✓ Documentation generated successfully"
echo ""
echo "Generated files:"
echo "  - oclif.manifest.json (command metadata)"
echo "  - README.md (updated with command reference)"
