---
name: oclif-patterns
description: Enterprise CLI patterns using oclif framework with TypeScript. Use when building oclif CLIs, creating plugins, implementing commands with flags/args, adding auto-documentation, testing CLI commands, or when user mentions oclif, enterprise CLI, TypeScript CLI, plugin system, or CLI testing.
allowed-tools: Read, Write, Edit, Bash, Glob, Grep
---

# oclif Enterprise CLI Patterns

Provides comprehensive patterns for building production-grade CLIs with oclif framework.

## Core Capabilities

### 1. Command Structure
- Single and multi-command CLIs
- Flag definitions (string, boolean, integer, custom)
- Argument parsing with validation
- Command inheritance and base classes
- Async command execution

### 2. Plugin System
- Installable plugins
- Plugin discovery and loading
- Hook system for extensibility
- Plugin commands and lifecycle

### 3. Auto-Documentation
- Auto-generated help text
- README generation
- Command reference docs
- Flag and argument documentation

### 4. Testing Patterns
- Command unit tests
- Integration testing
- Mock stdin/stdout
- Fixture management
- Test helpers

## Implementation Guide

### Command Creation

**Use templates**:
- `templates/command-basic.ts` - Simple command with flags
- `templates/command-advanced.ts` - Complex command with validation
- `templates/command-async.ts` - Async operations
- `templates/base-command.ts` - Custom base class

**Key patterns**:
1. Import Command from '@oclif/core'
2. Define flags using `Flags` object
3. Define args using `Args` object
4. Implement async run() method
5. Use this.log() for output
6. Use this.error() for errors

### Flag Patterns

**Common flags**:
- String: `Flags.string({ description, required, default })`
- Boolean: `Flags.boolean({ description, allowNo })`
- Integer: `Flags.integer({ description, min, max })`
- Custom: `Flags.custom<T>({ parse: async (input) => T })`
- Multiple: `Flags.string({ multiple: true })`

**Best practices**:
- Always provide clear descriptions
- Use char for common short flags
- Set required vs optional explicitly
- Provide sensible defaults
- Validate in parse function for custom flags

### Argument Patterns

**Definition**:
```typescript
static args = {
  name: Args.string({ description: 'Name', required: true }),
  file: Args.file({ description: 'File path', exists: true })
}
```

**Access in run()**:
```typescript
const { args } = await this.parse(MyCommand)
```

### Plugin Development

**Use templates**:
- `templates/plugin-package.json` - Plugin package.json
- `templates/plugin-command.ts` - Plugin command structure
- `templates/plugin-hooks.ts` - Hook implementations

**Plugin structure**:
```
my-plugin/
├── package.json (oclif configuration)
├── src/
│   ├── commands/ (plugin commands)
│   └── hooks/ (lifecycle hooks)
├── test/ (plugin tests)
└── README.md
```

### Testing Setup

**Use templates**:
- `templates/test-command.ts` - Command test template
- `templates/test-helpers.ts` - Test utilities
- `templates/test-setup.ts` - Test configuration

**Testing approach**:
1. Use @oclif/test for test helpers
2. Mock stdin/stdout with fancy-test
3. Test flag parsing separately
4. Test command execution
5. Test error handling
6. Use fixtures for file operations

### Auto-Documentation

**Generated automatically**:
- Command help via `--help` flag
- README.md with command reference
- Usage examples
- Flag and argument tables

**Use scripts**:
- `scripts/generate-docs.sh` - Generate all documentation
- `scripts/update-readme.sh` - Update README with commands

## Quick Start Examples

### Create Basic Command
```bash
# Use template
./scripts/create-command.sh my-command basic

# Results in: src/commands/my-command.ts
```

### Create Plugin
```bash
# Use template
./scripts/create-plugin.sh my-plugin

# Results in: plugin directory structure
```

### Run Tests
```bash
# Use test helpers
npm test
# or with coverage
npm run test:coverage
```

## Validation Scripts

**Available validators**:
- `scripts/validate-command.sh` - Check command structure
- `scripts/validate-plugin.sh` - Verify plugin structure
- `scripts/validate-tests.sh` - Ensure test coverage

## Templates Reference

### TypeScript Commands
1. `command-basic.ts` - Simple command pattern
2. `command-advanced.ts` - Full-featured command
3. `command-async.ts` - Async/await patterns
4. `base-command.ts` - Custom base class
5. `command-with-config.ts` - Configuration management

### Plugin System
6. `plugin-package.json` - Plugin package.json
7. `plugin-command.ts` - Plugin command
8. `plugin-hooks.ts` - Hook implementations
9. `plugin-manifest.json` - Plugin manifest

### Testing
10. `test-command.ts` - Command unit test
11. `test-helpers.ts` - Test utilities
12. `test-setup.ts` - Test configuration
13. `test-integration.ts` - Integration test

### Configuration
14. `tsconfig.json` - TypeScript config
15. `package.json` - oclif package.json
16. `.eslintrc.json` - ESLint config

## Examples Directory

See `examples/` for complete working examples:
- `examples/basic-cli/` - Simple CLI with commands
- `examples/plugin-cli/` - CLI with plugin support
- `examples/enterprise-cli/` - Full enterprise setup

## Common Patterns

### Error Handling
```typescript
if (!valid) {
  this.error('Invalid input', { exit: 1 })
}
```

### Spinner/Progress
```typescript
const spinner = ux.action.start('Processing')
// ... work
ux.action.stop()
```

### Prompts
```typescript
const answer = await ux.prompt('Continue?')
```

### Table Output
```typescript
ux.table(data, { columns: [...] })
```

## Requirements

- Node.js 18+
- TypeScript 5+
- @oclif/core ^3.0.0
- @oclif/test for testing
- Knowledge of TypeScript decorators (optional but helpful)

## Best Practices

1. **Command Design**: Keep commands focused, single responsibility
2. **Flags**: Use descriptive names, provide help text
3. **Testing**: Test command parsing and execution separately
4. **Documentation**: Let oclif generate docs, keep them updated
5. **Plugins**: Design for extensibility from the start
6. **Error Messages**: Provide actionable error messages
7. **TypeScript**: Use strict mode, define proper types
8. **Async**: Use async/await, handle promises properly

## Advanced Features

### Custom Flag Types
Create reusable custom flag parsers for complex validation.

### Hook System
Implement hooks for: init, prerun, postrun, command_not_found.

### Topic Commands
Organize commands into topics (e.g., `mycli topic:command`).

### Auto-Update
Use @oclif/plugin-update for automatic CLI updates.

### Analytics
Integrate analytics to track command usage.
