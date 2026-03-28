#!/usr/bin/env bash

# Validate oclif plugin structure
# Usage: ./validate-plugin.sh <plugin-directory>

set -e

PLUGIN_DIR="${1:-.}"

if [ ! -d "$PLUGIN_DIR" ]; then
  echo "Error: Plugin directory not found: $PLUGIN_DIR"
  exit 1
fi

echo "Validating plugin: $PLUGIN_DIR"

# Check for package.json
if [ ! -f "$PLUGIN_DIR/package.json" ]; then
  echo "✗ Missing package.json"
  exit 1
fi
echo "✓ Has package.json"

# Check for oclif configuration in package.json
if ! grep -q '"oclif"' "$PLUGIN_DIR/package.json"; then
  echo "✗ Missing oclif configuration in package.json"
  exit 1
fi
echo "✓ Has oclif configuration"

# Check for commands directory
if [ ! -d "$PLUGIN_DIR/src/commands" ]; then
  echo "⚠ Warning: Missing src/commands directory"
else
  echo "✓ Has src/commands directory"

  # Check for at least one command
  COMMAND_COUNT=$(find "$PLUGIN_DIR/src/commands" -name "*.ts" | wc -l)
  if [ "$COMMAND_COUNT" -eq 0 ]; then
    echo "⚠ Warning: No commands found"
  else
    echo "✓ Has $COMMAND_COUNT command(s)"
  fi
fi

# Check for hooks directory (optional)
if [ -d "$PLUGIN_DIR/src/hooks" ]; then
  echo "✓ Has src/hooks directory"
  HOOK_COUNT=$(find "$PLUGIN_DIR/src/hooks" -name "*.ts" | wc -l)
  echo "  ($HOOK_COUNT hook(s))"
fi

# Check for test directory
if [ ! -d "$PLUGIN_DIR/test" ]; then
  echo "⚠ Warning: Missing test directory"
else
  echo "✓ Has test directory"

  # Check for test files
  TEST_COUNT=$(find "$PLUGIN_DIR/test" -name "*.test.ts" | wc -l)
  if [ "$TEST_COUNT" -eq 0 ]; then
    echo "⚠ Warning: No test files found"
  else
    echo "✓ Has $TEST_COUNT test file(s)"
  fi
fi

# Check for TypeScript config
if [ ! -f "$PLUGIN_DIR/tsconfig.json" ]; then
  echo "⚠ Warning: Missing tsconfig.json"
else
  echo "✓ Has tsconfig.json"
fi

# Check for README
if [ ! -f "$PLUGIN_DIR/README.md" ]; then
  echo "⚠ Warning: Missing README.md"
else
  echo "✓ Has README.md"
fi

# Check dependencies in package.json
if grep -q '"@oclif/core"' "$PLUGIN_DIR/package.json"; then
  echo "✓ Has @oclif/core dependency"
else
  echo "✗ Missing @oclif/core dependency"
fi

# Check if plugin can be built
if [ -f "$PLUGIN_DIR/package.json" ]; then
  if grep -q '"build"' "$PLUGIN_DIR/package.json"; then
    echo "✓ Has build script"
  else
    echo "⚠ Warning: Missing build script"
  fi
fi

echo ""
echo "✓ Plugin validation complete"
