#!/usr/bin/env bash

# Create oclif command from template
# Usage: ./create-command.sh <command-name> <template-type>
# Template types: basic, advanced, async

set -e

COMMAND_NAME="$1"
TEMPLATE_TYPE="${2:-basic}"

if [ -z "$COMMAND_NAME" ]; then
  echo "Error: Command name is required"
  echo "Usage: ./create-command.sh <command-name> <template-type>"
  echo "Template types: basic, advanced, async"
  exit 1
fi

# Validate template type
if [[ ! "$TEMPLATE_TYPE" =~ ^(basic|advanced|async)$ ]]; then
  echo "Error: Invalid template type. Must be: basic, advanced, or async"
  exit 1
fi

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMPLATE_DIR="$(dirname "$SCRIPT_DIR")/templates"

# Determine output directory
if [ -d "src/commands" ]; then
  OUTPUT_DIR="src/commands"
elif [ -d "commands" ]; then
  OUTPUT_DIR="commands"
else
  echo "Error: Cannot find commands directory. Are you in the CLI project root?"
  exit 1
fi

# Convert command name to proper format
# e.g., "my-command" -> "MyCommand"
COMMAND_CLASS=$(echo "$COMMAND_NAME" | sed -r 's/(^|-)([a-z])/\U\2/g')

# Determine output file path
OUTPUT_FILE="$OUTPUT_DIR/${COMMAND_NAME}.ts"

# Check if file already exists
if [ -f "$OUTPUT_FILE" ]; then
  echo "Error: Command file already exists: $OUTPUT_FILE"
  exit 1
fi

# Select template file
TEMPLATE_FILE="$TEMPLATE_DIR/command-${TEMPLATE_TYPE}.ts"

if [ ! -f "$TEMPLATE_FILE" ]; then
  echo "Error: Template file not found: $TEMPLATE_FILE"
  exit 1
fi

# Create command file from template
echo "Creating command from template: $TEMPLATE_TYPE"
cp "$TEMPLATE_FILE" "$OUTPUT_FILE"

# Replace placeholders
sed -i "s/{{COMMAND_NAME}}/$COMMAND_CLASS/g" "$OUTPUT_FILE"
sed -i "s/{{DESCRIPTION}}/Command description for $COMMAND_NAME/g" "$OUTPUT_FILE"

echo "✓ Created command: $OUTPUT_FILE"
echo ""
echo "Next steps:"
echo "  1. Edit the command: $OUTPUT_FILE"
echo "  2. Update the description and flags"
echo "  3. Implement the run() method"
echo "  4. Build: npm run build"
echo "  5. Test: npm test"
echo ""
echo "Run the command:"
echo "  ./bin/run.js $COMMAND_NAME --help"
