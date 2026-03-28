# Plugin CLI Example

Complete example of an oclif CLI with plugin support.

## Overview

This example shows:
- Main CLI with core commands
- Plugin system for extensibility
- Plugin installation and management
- Shared hooks between CLI and plugins

## Main CLI Structure

```
mycli/
├── package.json
├── src/
│   ├── commands/
│   │   └── core.ts
│   └── hooks/
│       └── init.ts
└── plugins/
    └── plugin-deploy/
        ├── package.json
        └── src/
            └── commands/
                └── deploy.ts
```

## Step 1: Create Main CLI

### Main CLI package.json

```json
{
  "name": "mycli",
  "version": "1.0.0",
  "oclif": {
    "bin": "mycli",
    "commands": "./lib/commands",
    "plugins": [
      "@oclif/plugin-help",
      "@oclif/plugin-plugins"
    ],
    "hooks": {
      "init": "./lib/hooks/init"
    }
  },
  "dependencies": {
    "@oclif/core": "^3.0.0",
    "@oclif/plugin-help": "^6.0.0",
    "@oclif/plugin-plugins": "^4.0.0"
  }
}
```

### Core Command

File: `src/commands/core.ts`

```typescript
import { Command, Flags } from '@oclif/core'

export default class Core extends Command {
  static description = 'Core CLI functionality'

  static flags = {
    version: Flags.boolean({
      char: 'v',
      description: 'Show CLI version',
    }),
  }

  async run(): Promise<void> {
    const { flags } = await this.parse(Core)

    if (flags.version) {
      this.log(`Version: ${this.config.version}`)
      return
    }

    this.log('Core CLI is running')

    // List installed plugins
    const plugins = this.config.plugins
    this.log(`\nInstalled plugins: ${plugins.length}`)
    plugins.forEach(p => {
      this.log(`  - ${p.name} (${p.version})`)
    })
  }
}
```

### Init Hook

File: `src/hooks/init.ts`

```typescript
import { Hook } from '@oclif/core'

const hook: Hook<'init'> = async function (opts) {
  // Initialize CLI
  this.debug('Initializing mycli...')

  // Check for updates, load config, etc.
}

export default hook
```

## Step 2: Create Plugin

### Plugin package.json

```json
{
  "name": "@mycli/plugin-deploy",
  "version": "1.0.0",
  "description": "Deployment plugin for mycli",
  "oclif": {
    "bin": "mycli",
    "commands": "./lib/commands",
    "topics": {
      "deploy": {
        "description": "Deployment commands"
      }
    }
  },
  "dependencies": {
    "@oclif/core": "^3.0.0"
  }
}
```

### Deploy Command

File: `plugins/plugin-deploy/src/commands/deploy.ts`

```typescript
import { Command, Flags } from '@oclif/core'

export default class Deploy extends Command {
  static description = 'Deploy application'

  static examples = [
    '<%= config.bin %> deploy --env production',
  ]

  static flags = {
    env: Flags.string({
      char: 'e',
      description: 'Environment to deploy to',
      options: ['development', 'staging', 'production'],
      required: true,
    }),
    force: Flags.boolean({
      char: 'f',
      description: 'Force deployment',
      default: false,
    }),
  }

  async run(): Promise<void> {
    const { flags } = await this.parse(Deploy)

    this.log(`Deploying to ${flags.env}...`)

    if (flags.force) {
      this.log('Force deployment enabled')
    }

    // Deployment logic here
    this.log('✓ Deployment successful')
  }
}
```

## Step 3: Build and Link Plugin

```bash
# Build main CLI
cd mycli
npm run build

# Build plugin
cd plugins/plugin-deploy
npm run build

# Link plugin to main CLI
cd ../../
mycli plugins:link ./plugins/plugin-deploy
```

## Step 4: Use Plugin Commands

```bash
# List plugins
mycli plugins

# Use plugin command
mycli deploy --env production

# Get help for plugin command
mycli deploy --help
```

## Step 5: Install Plugin from npm

### Publish Plugin

```bash
cd plugins/plugin-deploy
npm publish
```

### Install Plugin

```bash
mycli plugins:install @mycli/plugin-deploy
```

## Plugin Management Commands

```bash
# List installed plugins
mycli plugins

# Install plugin
mycli plugins:install @mycli/plugin-name

# Update plugin
mycli plugins:update @mycli/plugin-name

# Uninstall plugin
mycli plugins:uninstall @mycli/plugin-name

# Link local plugin (development)
mycli plugins:link /path/to/plugin
```

## Advanced: Plugin with Hooks

File: `plugins/plugin-deploy/src/hooks/prerun.ts`

```typescript
import { Hook } from '@oclif/core'

const hook: Hook<'prerun'> = async function (opts) {
  // Check deployment prerequisites
  if (opts.Command.id === 'deploy') {
    this.log('Checking deployment prerequisites...')

    // Check environment, credentials, etc.
  }
}

export default hook
```

Register in plugin package.json:

```json
{
  "oclif": {
    "hooks": {
      "prerun": "./lib/hooks/prerun"
    }
  }
}
```

## Key Concepts Demonstrated

1. **Plugin System**: @oclif/plugin-plugins integration
2. **Plugin Discovery**: Automatic command loading from plugins
3. **Plugin Management**: Install, update, uninstall commands
4. **Local Development**: plugins:link for local plugin development
5. **Hooks**: Shared hooks between main CLI and plugins
6. **Topic Commands**: Organized plugin commands (deploy:*)
7. **Plugin Metadata**: Package.json oclif configuration
8. **Plugin Distribution**: Publishing to npm

## Testing Plugins

File: `plugins/plugin-deploy/test/commands/deploy.test.ts`

```typescript
import { expect, test } from '@oclif/test'

describe('deploy', () => {
  test
    .stdout()
    .command(['deploy', '--env', 'production'])
    .it('deploys to production', ctx => {
      expect(ctx.stdout).to.contain('Deploying to production')
      expect(ctx.stdout).to.contain('successful')
    })

  test
    .command(['deploy'])
    .catch(error => {
      expect(error.message).to.contain('Missing required flag')
    })
    .it('requires env flag')
})
```
