#!/usr/bin/env bash

# Validate test coverage for oclif CLI
# Usage: ./validate-tests.sh

set -e

echo "Validating test coverage..."

# Check if test directory exists
if [ ! -d "test" ]; then
  echo "✗ Missing test directory"
  exit 1
fi
echo "✓ Has test directory"

# Count test files
TEST_COUNT=$(find test -name "*.test.ts" | wc -l)
if [ "$TEST_COUNT" -eq 0 ]; then
  echo "✗ No test files found"
  exit 1
fi
echo "✓ Has $TEST_COUNT test file(s)"

# Count command files
if [ -d "src/commands" ]; then
  COMMAND_COUNT=$(find src/commands -name "*.ts" | wc -l)
  echo "  $COMMAND_COUNT command file(s) in src/commands"

  # Check if each command has tests
  for cmd in src/commands/*.ts; do
    CMD_NAME=$(basename "$cmd" .ts)
    if [ ! -f "test/commands/$CMD_NAME.test.ts" ]; then
      echo "⚠ Warning: No test for command: $CMD_NAME"
    fi
  done
fi

# Run tests if available
if command -v npm &> /dev/null && grep -q '"test"' package.json; then
  echo ""
  echo "→ Running tests..."
  if npm test; then
    echo "✓ All tests passed"
  else
    echo "✗ Some tests failed"
    exit 1
  fi
fi

# Check for test coverage script
if grep -q '"test:coverage"' package.json; then
  echo ""
  echo "→ Checking test coverage..."
  npm run test:coverage
fi

echo ""
echo "✓ Test validation complete"
