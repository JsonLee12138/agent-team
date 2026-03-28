#!/usr/bin/env bash

# Create oclif plugin structure
# Usage: ./create-plugin.sh <plugin-name>

set -e

PLUGIN_NAME="$1"

if [ -z "$PLUGIN_NAME" ]; then
  echo "Error: Plugin name is required"
  echo "Usage: ./create-plugin.sh <plugin-name>"
  exit 1
fi

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEMPLATE_DIR="$(dirname "$SCRIPT_DIR")/templates"

# Determine output directory
PLUGIN_DIR="plugin-$PLUGIN_NAME"

if [ -d "$PLUGIN_DIR" ]; then
  echo "Error: Plugin directory already exists: $PLUGIN_DIR"
  exit 1
fi

echo "Creating plugin: $PLUGIN_NAME"

# Create directory structure
mkdir -p "$PLUGIN_DIR/src/commands"
mkdir -p "$PLUGIN_DIR/src/hooks"
mkdir -p "$PLUGIN_DIR/test"

# Copy package.json template
cp "$TEMPLATE_DIR/plugin-package.json" "$PLUGIN_DIR/package.json"

# Replace placeholders in package.json
sed -i "s/{{PLUGIN_NAME}}/$PLUGIN_NAME/g" "$PLUGIN_DIR/package.json"
sed -i "s/{{DESCRIPTION}}/Plugin for $PLUGIN_NAME/g" "$PLUGIN_DIR/package.json"
sed -i "s/{{AUTHOR}}/Your Name/g" "$PLUGIN_DIR/package.json"
sed -i "s/{{GITHUB_USER}}/yourusername/g" "$PLUGIN_DIR/package.json"

# Copy TypeScript config
cp "$TEMPLATE_DIR/tsconfig.json" "$PLUGIN_DIR/tsconfig.json"

# Copy ESLint config
cp "$TEMPLATE_DIR/.eslintrc.json" "$PLUGIN_DIR/.eslintrc.json"

# Create plugin command
cp "$TEMPLATE_DIR/plugin-command.ts" "$PLUGIN_DIR/src/commands/$PLUGIN_NAME.ts"
sed -i "s/{{PLUGIN_NAME}}/$PLUGIN_NAME/g" "$PLUGIN_DIR/src/commands/$PLUGIN_NAME.ts"
sed -i "s/{{COMMAND_NAME}}/main/g" "$PLUGIN_DIR/src/commands/$PLUGIN_NAME.ts"
sed -i "s/{{COMMAND_CLASS}}/Main/g" "$PLUGIN_DIR/src/commands/$PLUGIN_NAME.ts"
sed -i "s/{{DESCRIPTION}}/Main plugin command/g" "$PLUGIN_DIR/src/commands/$PLUGIN_NAME.ts"

# Create hooks
cp "$TEMPLATE_DIR/plugin-hooks.ts" "$PLUGIN_DIR/src/hooks/init.ts"
sed -i "s/{{PLUGIN_NAME}}/$PLUGIN_NAME/g" "$PLUGIN_DIR/src/hooks/init.ts"

# Create index.ts
cat > "$PLUGIN_DIR/src/index.ts" << EOF
export { default as init } from './hooks/init'
EOF

# Create README
cat > "$PLUGIN_DIR/README.md" << EOF
# @mycli/plugin-$PLUGIN_NAME

Plugin for mycli: $PLUGIN_NAME

## Installation

\`\`\`bash
mycli plugins:install @mycli/plugin-$PLUGIN_NAME
\`\`\`

## Usage

\`\`\`bash
mycli $PLUGIN_NAME:main --help
\`\`\`

## Commands

<!-- commands -->
<!-- commandsstop -->

## Development

\`\`\`bash
npm install
npm run build
npm test
\`\`\`

## License

MIT
EOF

# Create .gitignore
cat > "$PLUGIN_DIR/.gitignore" << EOF
*-debug.log
*-error.log
*.tgz
.DS_Store
/.nyc_output
/dist
/lib
/package-lock.json
/tmp
node_modules
oclif.manifest.json
tsconfig.tsbuildinfo
EOF

echo "✓ Created plugin structure: $PLUGIN_DIR"
echo ""
echo "Next steps:"
echo "  cd $PLUGIN_DIR"
echo "  npm install"
echo "  npm run build"
echo "  npm test"
echo ""
echo "To install plugin locally:"
echo "  mycli plugins:link $PWD/$PLUGIN_DIR"
