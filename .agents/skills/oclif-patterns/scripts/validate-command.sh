#!/usr/bin/env bash

# Validate oclif command structure
# Usage: ./validate-command.sh <command-file>

set -e

COMMAND_FILE="$1"

if [ -z "$COMMAND_FILE" ]; then
  echo "Error: Command file path is required"
  echo "Usage: ./validate-command.sh <command-file>"
  exit 1
fi

if [ ! -f "$COMMAND_FILE" ]; then
  echo "Error: Command file not found: $COMMAND_FILE"
  exit 1
fi

echo "Validating command: $COMMAND_FILE"

# Check for required imports
if ! grep -q "from '@oclif/core'" "$COMMAND_FILE"; then
  echo "✗ Missing import from @oclif/core"
  exit 1
fi
echo "✓ Has @oclif/core import"

# Check for Command class
if ! grep -q "extends Command" "$COMMAND_FILE"; then
  echo "✗ Missing 'extends Command'"
  exit 1
fi
echo "✓ Extends Command class"

# Check for description
if ! grep -q "static description" "$COMMAND_FILE"; then
  echo "⚠ Warning: Missing static description"
else
  echo "✓ Has static description"
fi

# Check for examples
if ! grep -q "static examples" "$COMMAND_FILE"; then
  echo "⚠ Warning: Missing static examples"
else
  echo "✓ Has static examples"
fi

# Check for run method
if ! grep -q "async run()" "$COMMAND_FILE"; then
  echo "✗ Missing async run() method"
  exit 1
fi
echo "✓ Has async run() method"

# Check for proper flag access
if grep -q "this.parse(" "$COMMAND_FILE"; then
  echo "✓ Properly parses flags"
else
  echo "⚠ Warning: May not be parsing flags correctly"
fi

# Check TypeScript
if command -v tsc &> /dev/null; then
  echo "→ Checking TypeScript compilation..."
  if tsc --noEmit "$COMMAND_FILE" 2>/dev/null; then
    echo "✓ TypeScript compilation successful"
  else
    echo "⚠ Warning: TypeScript compilation has issues"
  fi
fi

echo ""
echo "✓ Command validation complete"
