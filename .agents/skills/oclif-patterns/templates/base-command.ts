import { Command, Flags } from '@oclif/core'
import * as fs from 'fs-extra'
import * as path from 'path'

/**
 * Base command class with common functionality for all commands
 * Extend this class instead of Command for consistent behavior
 */
export default abstract class BaseCommand extends Command {
  // Global flags available to all commands
  static baseFlags = {
    config: Flags.string({
      char: 'c',
      description: 'Path to config file',
      env: 'CLI_CONFIG',
    }),
    'log-level': Flags.string({
      description: 'Log level',
      options: ['error', 'warn', 'info', 'debug'],
      default: 'info',
    }),
    json: Flags.boolean({
      description: 'Output as JSON',
      default: false,
    }),
  }

  protected config_: any = null

  /**
   * Initialize command - load config, setup logging
   */
  async init(): Promise<void> {
    await super.init()
    await this.loadConfig()
    this.setupLogging()
  }

  /**
   * Load configuration file
   */
  private async loadConfig(): Promise<void> {
    const { flags } = await this.parse(this.constructor as typeof BaseCommand)

    if (flags.config) {
      if (!await fs.pathExists(flags.config)) {
        this.error(`Config file not found: ${flags.config}`, { exit: 1 })
      }

      try {
        const content = await fs.readFile(flags.config, 'utf-8')
        this.config_ = JSON.parse(content)
        this.debug(`Loaded config from ${flags.config}`)
      } catch (error) {
        this.error(`Failed to parse config file: ${error instanceof Error ? error.message : 'Unknown error'}`, {
          exit: 1,
        })
      }
    } else {
      // Try to load default config locations
      const defaultLocations = [
        path.join(process.cwd(), '.clirc'),
        path.join(process.cwd(), '.cli.json'),
        path.join(this.config.home, '.config', 'cli', 'config.json'),
      ]

      for (const location of defaultLocations) {
        if (await fs.pathExists(location)) {
          try {
            const content = await fs.readFile(location, 'utf-8')
            this.config_ = JSON.parse(content)
            this.debug(`Loaded config from ${location}`)
            break
          } catch {
            // Ignore parse errors for default configs
          }
        }
      }
    }
  }

  /**
   * Setup logging based on log-level flag
   */
  private setupLogging(): void {
    // Implementation would setup actual logging library
    this.debug('Logging initialized')
  }

  /**
   * Get config value with dot notation support
   */
  protected getConfig(key: string, defaultValue?: any): any {
    if (!this.config_) return defaultValue

    const keys = key.split('.')
    let value = this.config_

    for (const k of keys) {
      if (value && typeof value === 'object' && k in value) {
        value = value[k]
      } else {
        return defaultValue
      }
    }

    return value
  }

  /**
   * Output data respecting --json flag
   */
  protected output(data: any, message?: string): void {
    const { flags } = this.parse(this.constructor as typeof BaseCommand)

    if (flags.json) {
      this.log(JSON.stringify(data, null, 2))
    } else if (message) {
      this.log(message)
    } else if (typeof data === 'string') {
      this.log(data)
    } else {
      this.log(JSON.stringify(data, null, 2))
    }
  }

  /**
   * Enhanced error handling
   */
  protected handleError(error: Error, context?: string): never {
    const message = context ? `${context}: ${error.message}` : error.message

    if (this.config.debug) {
      this.error(error.stack || message, { exit: 1 })
    } else {
      this.error(message, { exit: 1 })
    }
  }

  /**
   * Log only if verbose/debug mode
   */
  protected debug(message: string): void {
    const { flags } = this.parse(this.constructor as typeof BaseCommand)

    if (flags['log-level'] === 'debug') {
      this.log(`[DEBUG] ${message}`)
    }
  }

  /**
   * Validate required environment variables
   */
  protected requireEnv(vars: string[]): void {
    const missing = vars.filter(v => !process.env[v])

    if (missing.length > 0) {
      this.error(
        `Missing required environment variables: ${missing.join(', ')}`,
        { exit: 1 }
      )
    }
  }

  /**
   * Check if running in CI environment
   */
  protected isCI(): boolean {
    return Boolean(process.env.CI)
  }

  /**
   * Prompt user unless in CI or non-interactive mode
   */
  protected async prompt(message: string, defaultValue?: string): Promise<string> {
    if (this.isCI()) {
      return defaultValue || ''
    }

    const { default: inquirer } = await import('inquirer')
    const { value } = await inquirer.prompt([
      {
        type: 'input',
        name: 'value',
        message,
        default: defaultValue,
      },
    ])

    return value
  }
}
