# Enterprise CLI Example

Complete production-ready oclif CLI with all best practices.

## Overview

This example demonstrates:
- Custom base command with common functionality
- Configuration management
- Logging system
- Error handling
- Plugin support
- Comprehensive testing
- CI/CD integration
- Auto-update capability

## Project Structure

```
enterprise-cli/
├── package.json
├── tsconfig.json
├── .eslintrc.json
├── .github/
│   └── workflows/
│       ├── test.yml
│       └── release.yml
├── src/
│   ├── base-command.ts
│   ├── config/
│   │   ├── manager.ts
│   │   └── schema.ts
│   ├── commands/
│   │   ├── deploy.ts
│   │   ├── status.ts
│   │   └── config/
│   │       ├── get.ts
│   │       └── set.ts
│   ├── hooks/
│   │   ├── init.ts
│   │   └── prerun.ts
│   ├── utils/
│   │   ├── logger.ts
│   │   └── error-handler.ts
│   └── index.ts
├── test/
│   ├── commands/
│   ├── helpers/
│   └── fixtures/
└── docs/
    └── commands/
```

## Base Command Implementation

File: `src/base-command.ts`

```typescript
import { Command, Flags } from '@oclif/core'
import { ConfigManager } from './config/manager'
import { Logger } from './utils/logger'
import { ErrorHandler } from './utils/error-handler'

export default abstract class BaseCommand extends Command {
  protected configManager!: ConfigManager
  protected logger!: Logger
  protected errorHandler!: ErrorHandler

  static baseFlags = {
    config: Flags.string({
      char: 'c',
      description: 'Path to config file',
      env: 'CLI_CONFIG',
    }),
    'log-level': Flags.string({
      description: 'Set log level',
      options: ['error', 'warn', 'info', 'debug'],
      default: 'info',
      env: 'LOG_LEVEL',
    }),
    json: Flags.boolean({
      description: 'Output as JSON',
      default: false,
    }),
    'no-color': Flags.boolean({
      description: 'Disable colors',
      default: false,
    }),
  }

  async init(): Promise<void> {
    await super.init()

    // Initialize logger
    const { flags } = await this.parse(this.constructor as typeof BaseCommand)
    this.logger = new Logger(flags['log-level'], !flags['no-color'])

    // Initialize config manager
    this.configManager = new ConfigManager(flags.config)
    await this.configManager.load()

    // Initialize error handler
    this.errorHandler = new ErrorHandler(this.logger)

    // Log initialization
    this.logger.debug(`Initialized ${this.id}`)
  }

  protected async catch(err: Error & { exitCode?: number }): Promise<any> {
    return this.errorHandler.handle(err)
  }

  protected output(data: any, humanMessage?: string): void {
    const { flags } = this.parse(this.constructor as typeof BaseCommand)

    if (flags.json) {
      this.log(JSON.stringify(data, null, 2))
    } else if (humanMessage) {
      this.log(humanMessage)
    } else {
      this.log(JSON.stringify(data, null, 2))
    }
  }
}
```

## Configuration Manager

File: `src/config/manager.ts`

```typescript
import * as fs from 'fs-extra'
import * as path from 'path'
import * as os from 'os'

export class ConfigManager {
  private config: any = {}
  private configPath: string

  constructor(customPath?: string) {
    this.configPath = customPath || this.getDefaultConfigPath()
  }

  async load(): Promise<void> {
    if (await fs.pathExists(this.configPath)) {
      this.config = await fs.readJson(this.configPath)
    } else {
      // Create default config
      this.config = this.getDefaultConfig()
      await this.save()
    }
  }

  async save(): Promise<void> {
    await fs.ensureDir(path.dirname(this.configPath))
    await fs.writeJson(this.configPath, this.config, { spaces: 2 })
  }

  get(key: string, defaultValue?: any): any {
    const keys = key.split('.')
    let value = this.config

    for (const k of keys) {
      if (value && typeof value === 'object' && k in value) {
        value = value[k]
      } else {
        return defaultValue
      }
    }

    return value
  }

  set(key: string, value: any): void {
    const keys = key.split('.')
    const lastKey = keys.pop()!
    let current = this.config

    for (const k of keys) {
      if (!(k in current) || typeof current[k] !== 'object') {
        current[k] = {}
      }
      current = current[k]
    }

    current[lastKey] = value
  }

  private getDefaultConfigPath(): string {
    return path.join(os.homedir(), '.config', 'mycli', 'config.json')
  }

  private getDefaultConfig(): any {
    return {
      version: '1.0.0',
      defaults: {
        region: 'us-east-1',
        timeout: 30000,
      },
    }
  }
}
```

## Logger Implementation

File: `src/utils/logger.ts`

```typescript
import chalk from 'chalk'

export class Logger {
  constructor(
    private level: string = 'info',
    private color: boolean = true
  ) {}

  debug(message: string): void {
    if (this.shouldLog('debug')) {
      this.output('DEBUG', message, chalk.gray)
    }
  }

  info(message: string): void {
    if (this.shouldLog('info')) {
      this.output('INFO', message, chalk.blue)
    }
  }

  warn(message: string): void {
    if (this.shouldLog('warn')) {
      this.output('WARN', message, chalk.yellow)
    }
  }

  error(message: string, error?: Error): void {
    if (this.shouldLog('error')) {
      this.output('ERROR', message, chalk.red)
      if (error && error.stack) {
        console.error(chalk.red(error.stack))
      }
    }
  }

  success(message: string): void {
    if (this.shouldLog('info')) {
      this.output('SUCCESS', message, chalk.green)
    }
  }

  private shouldLog(level: string): boolean {
    const levels = ['error', 'warn', 'info', 'debug']
    const currentIndex = levels.indexOf(this.level)
    const messageIndex = levels.indexOf(level)
    return messageIndex <= currentIndex
  }

  private output(level: string, message: string, colorFn: any): void {
    const timestamp = new Date().toISOString()
    const prefix = this.color ? colorFn(`[${level}]`) : `[${level}]`
    console.log(`${timestamp} ${prefix} ${message}`)
  }
}
```

## Enterprise Deploy Command

File: `src/commands/deploy.ts`

```typescript
import BaseCommand from '../base-command'
import { Flags, Args, ux } from '@oclif/core'

export default class Deploy extends BaseCommand {
  static description = 'Deploy application to environment'

  static examples = [
    '<%= config.bin %> deploy myapp --env production',
    '<%= config.bin %> deploy myapp --env staging --auto-approve',
  ]

  static flags = {
    ...BaseCommand.baseFlags,
    env: Flags.string({
      char: 'e',
      description: 'Environment to deploy to',
      options: ['development', 'staging', 'production'],
      required: true,
    }),
    'auto-approve': Flags.boolean({
      description: 'Skip confirmation prompt',
      default: false,
    }),
    'rollback-on-failure': Flags.boolean({
      description: 'Automatically rollback on failure',
      default: true,
    }),
  }

  static args = {
    app: Args.string({
      description: 'Application name',
      required: true,
    }),
  }

  async run(): Promise<void> {
    const { args, flags } = await this.parse(Deploy)

    this.logger.info(`Starting deployment of ${args.app} to ${flags.env}`)

    // Confirmation prompt (skip in CI or with auto-approve)
    if (!flags['auto-approve'] && !process.env.CI) {
      const confirmed = await ux.confirm(
        `Deploy ${args.app} to ${flags.env}? (y/n)`
      )

      if (!confirmed) {
        this.logger.info('Deployment cancelled')
        return
      }
    }

    // Pre-deployment checks
    ux.action.start('Running pre-deployment checks')
    await this.runPreDeploymentChecks(args.app, flags.env)
    ux.action.stop('passed')

    // Deploy
    ux.action.start(`Deploying ${args.app}`)

    try {
      const result = await this.deploy(args.app, flags.env)
      ux.action.stop('done')

      this.logger.success(`Deployed ${args.app} to ${flags.env}`)
      this.output(result, `Deployment URL: ${result.url}`)
    } catch (error) {
      ux.action.stop('failed')

      if (flags['rollback-on-failure']) {
        this.logger.warn('Deployment failed, rolling back...')
        await this.rollback(args.app, flags.env)
      }

      throw error
    }
  }

  private async runPreDeploymentChecks(
    app: string,
    env: string
  ): Promise<void> {
    // Check credentials
    // Validate app exists
    // Check environment health
    // Verify dependencies
    await new Promise(resolve => setTimeout(resolve, 1000))
  }

  private async deploy(app: string, env: string): Promise<any> {
    // Actual deployment logic
    await new Promise(resolve => setTimeout(resolve, 3000))

    return {
      app,
      env,
      version: '1.2.3',
      url: `https://${app}.${env}.example.com`,
      deployedAt: new Date().toISOString(),
    }
  }

  private async rollback(app: string, env: string): Promise<void> {
    // Rollback logic
    await new Promise(resolve => setTimeout(resolve, 2000))
    this.logger.info('Rollback complete')
  }
}
```

## Testing Setup

File: `test/helpers/test-context.ts`

```typescript
import * as path from 'path'
import * as fs from 'fs-extra'
import { Config } from '@oclif/core'

export class TestContext {
  testDir: string
  config: Config

  constructor() {
    this.testDir = path.join(__dirname, '../fixtures/test-run')
  }

  async setup(): Promise<void> {
    await fs.ensureDir(this.testDir)
    this.config = await Config.load()
  }

  async teardown(): Promise<void> {
    await fs.remove(this.testDir)
  }

  async createConfigFile(config: any): Promise<string> {
    const configPath = path.join(this.testDir, 'config.json')
    await fs.writeJson(configPath, config)
    return configPath
  }
}
```

## Key Enterprise Features

1. **Base Command**: Shared functionality across all commands
2. **Configuration**: Centralized config management
3. **Logging**: Structured logging with levels
4. **Error Handling**: Consistent error handling
5. **Confirmation Prompts**: Interactive confirmations
6. **Rollback**: Automatic rollback on failure
7. **Pre-deployment Checks**: Validation before operations
8. **CI Detection**: Different behavior in CI environments
9. **JSON Output**: Machine-readable output option
10. **Environment Variables**: Config via env vars
