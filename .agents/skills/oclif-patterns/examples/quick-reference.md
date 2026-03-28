# oclif Patterns Quick Reference

Fast lookup for common oclif patterns and commands.

## Command Creation

### Basic Command
```typescript
import { Command, Flags, Args } from '@oclif/core'

export default class MyCommand extends Command {
  static description = 'Description'

  static flags = {
    name: Flags.string({ char: 'n', required: true }),
  }

  async run(): Promise<void> {
    const { flags } = await this.parse(MyCommand)
    this.log(`Hello ${flags.name}`)
  }
}
```

### Using Scripts
```bash
# Create command from template
./scripts/create-command.sh my-command basic

# Create advanced command
./scripts/create-command.sh deploy advanced

# Create async command
./scripts/create-command.sh fetch async
```

## Flag Patterns

### String Flag
```typescript
name: Flags.string({
  char: 'n',
  description: 'Name',
  required: true,
  default: 'World',
})
```

### Boolean Flag
```typescript
verbose: Flags.boolean({
  char: 'v',
  description: 'Verbose output',
  default: false,
  allowNo: true,  // Enables --no-verbose
})
```

### Integer Flag
```typescript
port: Flags.integer({
  char: 'p',
  description: 'Port number',
  min: 1024,
  max: 65535,
  default: 3000,
})
```

### Option Flag (Enum)
```typescript
env: Flags.string({
  char: 'e',
  description: 'Environment',
  options: ['dev', 'staging', 'prod'],
  required: true,
})
```

### Multiple Values
```typescript
tags: Flags.string({
  char: 't',
  description: 'Tags',
  multiple: true,
})
// Usage: --tags=foo --tags=bar
```

### Custom Flag
```typescript
date: Flags.custom<Date>({
  parse: async (input) => new Date(input),
})
```

## Argument Patterns

### Required Argument
```typescript
static args = {
  file: Args.string({
    description: 'File path',
    required: true,
  }),
}
```

### File Argument
```typescript
static args = {
  file: Args.file({
    description: 'Input file',
    exists: true,  // Validates file exists
  }),
}
```

### Directory Argument
```typescript
static args = {
  dir: Args.directory({
    description: 'Target directory',
    exists: true,
  }),
}
```

## Output Patterns

### Simple Log
```typescript
this.log('Message')
```

### Error with Exit
```typescript
this.error('Error message', { exit: 1 })
```

### Warning
```typescript
this.warn('Warning message')
```

### Spinner
```typescript
import { ux } from '@oclif/core'

ux.action.start('Processing')
// ... work
ux.action.stop('done')
```

### Progress Bar
```typescript
import { ux } from '@oclif/core'

const total = 100
ux.progress.start(total)
for (let i = 0; i < total; i++) {
  ux.progress.update(i)
}
ux.progress.stop()
```

### Table Output
```typescript
import { ux } from '@oclif/core'

ux.table(data, {
  id: {},
  name: {},
  status: { extended: true },
})
```

### Prompt
```typescript
import { ux } from '@oclif/core'

const name = await ux.prompt('What is your name?')
const password = await ux.prompt('Password', { type: 'hide' })
const confirmed = await ux.confirm('Continue? (y/n)')
```

## Testing Patterns

### Basic Test
```typescript
import { expect, test } from '@oclif/test'

test
  .stdout()
  .command(['mycommand', '--name', 'Test'])
  .it('runs command', ctx => {
    expect(ctx.stdout).to.contain('Test')
  })
```

### Test with Error
```typescript
test
  .command(['mycommand'])
  .catch(error => {
    expect(error.message).to.contain('Missing')
  })
  .it('fails without flags')
```

### Test with Mock
```typescript
test
  .nock('https://api.example.com', api =>
    api.get('/data').reply(200, { result: 'success' })
  )
  .stdout()
  .command(['mycommand'])
  .it('handles API call', ctx => {
    expect(ctx.stdout).to.contain('success')
  })
```

### Test with Environment
```typescript
test
  .env({ API_KEY: 'test-key' })
  .stdout()
  .command(['mycommand'])
  .it('reads from env')
```

## Plugin Patterns

### Create Plugin
```bash
./scripts/create-plugin.sh my-plugin
```

### Link Plugin
```bash
mycli plugins:link ./plugin-my-plugin
```

### Install Plugin
```bash
mycli plugins:install @mycli/plugin-name
```

## Hook Patterns

### Init Hook
```typescript
import { Hook } from '@oclif/core'

const hook: Hook<'init'> = async function (opts) {
  // Runs before any command
}

export default hook
```

### Prerun Hook
```typescript
const hook: Hook<'prerun'> = async function (opts) {
  const { Command, argv } = opts
  // Runs before each command
}
```

## Common Commands

### Generate Documentation
```bash
npm run prepack
# Generates oclif.manifest.json and updates README.md
```

### Build
```bash
npm run build
```

### Test
```bash
npm test
npm run test:coverage
```

### Lint
```bash
npm run lint
npm run lint:fix
```

## Validation

### Validate Command
```bash
./scripts/validate-command.sh src/commands/mycommand.ts
```

### Validate Plugin
```bash
./scripts/validate-plugin.sh ./my-plugin
```

### Validate Tests
```bash
./scripts/validate-tests.sh
```

## Configuration Patterns

### Read Config
```typescript
const configPath = path.join(this.config.home, '.myclirc')
const config = await fs.readJson(configPath)
```

### Write Config
```typescript
await fs.writeJson(configPath, config, { spaces: 2 })
```

### Environment Variables
```typescript
const apiKey = process.env.API_KEY || this.error('API_KEY required')
```

## Error Handling

### Try-Catch
```typescript
try {
  await riskyOperation()
} catch (error) {
  this.error(`Operation failed: ${error.message}`, { exit: 1 })
}
```

### Custom Error
```typescript
if (!valid) {
  this.error('Invalid input', {
    exit: 1,
    suggestions: ['Try --help for usage']
  })
}
```

## Async Patterns

### Concurrent Operations
```typescript
const results = await Promise.all([
  operation1(),
  operation2(),
  operation3(),
])
```

### Sequential Operations
```typescript
for (const item of items) {
  await processItem(item)
}
```

### With Timeout
```typescript
const controller = new AbortController()
const timeout = setTimeout(() => controller.abort(), 5000)

try {
  const response = await fetch(url, { signal: controller.signal })
} finally {
  clearTimeout(timeout)
}
```

## Best Practices

1. Always provide clear descriptions for flags and commands
2. Use char flags for common options (e.g., -v for verbose)
3. Validate inputs early in the run() method
4. Use ux.action.start/stop for long operations
5. Handle errors gracefully with helpful messages
6. Test both success and failure cases
7. Generate documentation with oclif manifest
8. Use TypeScript strict mode
9. Follow naming conventions (kebab-case for commands)
10. Keep commands focused and single-purpose
