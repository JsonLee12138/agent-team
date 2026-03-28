import { Command, Flags } from '@oclif/core'
import * as fs from 'fs-extra'
import * as path from 'path'
import * as os from 'os'

export default class {{COMMAND_NAME}} extends Command {
  static description = 'Command with configuration file management'

  static examples = [
    '<%= config.bin %> <%= command.id %> --init',
    '<%= config.bin %> <%= command.id %> --set key=value',
    '<%= config.bin %> <%= command.id %> --get key',
    '<%= config.bin %> <%= command.id %> --list',
  ]

  static flags = {
    init: Flags.boolean({
      description: 'Initialize configuration file',
      exclusive: ['set', 'get', 'list'],
    }),
    set: Flags.string({
      description: 'Set configuration value (key=value)',
      multiple: true,
      exclusive: ['init', 'get', 'list'],
    }),
    get: Flags.string({
      description: 'Get configuration value by key',
      exclusive: ['init', 'set', 'list'],
    }),
    list: Flags.boolean({
      description: 'List all configuration values',
      exclusive: ['init', 'set', 'get'],
    }),
    global: Flags.boolean({
      char: 'g',
      description: 'Use global config instead of local',
      default: false,
    }),
  }

  private readonly DEFAULT_CONFIG = {
    version: '1.0.0',
    settings: {
      theme: 'default',
      verbose: false,
      timeout: 30000,
    },
  }

  async run(): Promise<void> {
    const { flags } = await this.parse({{COMMAND_NAME}})
    const configPath = this.getConfigPath(flags.global)

    if (flags.init) {
      await this.initConfig(configPath)
    } else if (flags.set && flags.set.length > 0) {
      await this.setConfig(configPath, flags.set)
    } else if (flags.get) {
      await this.getConfigValue(configPath, flags.get)
    } else if (flags.list) {
      await this.listConfig(configPath)
    } else {
      // Default: show current config location
      this.log(`Config location: ${configPath}`)
      if (await fs.pathExists(configPath)) {
        this.log('Config file exists')
        await this.listConfig(configPath)
      } else {
        this.log('Config file does not exist. Run with --init to create.')
      }
    }
  }

  private getConfigPath(global: boolean): string {
    if (global) {
      return path.join(os.homedir(), '.config', 'mycli', 'config.json')
    } else {
      return path.join(process.cwd(), '.myclirc')
    }
  }

  private async initConfig(configPath: string): Promise<void> {
    if (await fs.pathExists(configPath)) {
      this.error('Config file already exists. Remove it first or use --set to update.', {
        exit: 1,
      })
    }

    // Ensure directory exists
    await fs.ensureDir(path.dirname(configPath))

    // Write default config
    await fs.writeJson(configPath, this.DEFAULT_CONFIG, { spaces: 2 })

    this.log(`✓ Initialized config file at: ${configPath}`)
  }

  private async setConfig(configPath: string, settings: string[]): Promise<void> {
    // Load existing config or use default
    let config = this.DEFAULT_CONFIG
    if (await fs.pathExists(configPath)) {
      config = await fs.readJson(configPath)
    } else {
      await fs.ensureDir(path.dirname(configPath))
    }

    // Parse and set values
    for (const setting of settings) {
      const [key, ...valueParts] = setting.split('=')
      const value = valueParts.join('=')

      if (!value) {
        this.warn(`Skipping invalid setting: ${setting}`)
        continue
      }

      // Support dot notation
      this.setNestedValue(config, key, this.parseValue(value))
      this.log(`✓ Set ${key} = ${value}`)
    }

    // Save config
    await fs.writeJson(configPath, config, { spaces: 2 })
    this.log(`\n✓ Updated config file: ${configPath}`)
  }

  private async getConfigValue(configPath: string, key: string): Promise<void> {
    if (!await fs.pathExists(configPath)) {
      this.error('Config file does not exist. Run with --init first.', { exit: 1 })
    }

    const config = await fs.readJson(configPath)
    const value = this.getNestedValue(config, key)

    if (value === undefined) {
      this.error(`Key not found: ${key}`, { exit: 1 })
    }

    this.log(typeof value === 'object' ? JSON.stringify(value, null, 2) : String(value))
  }

  private async listConfig(configPath: string): Promise<void> {
    if (!await fs.pathExists(configPath)) {
      this.error('Config file does not exist. Run with --init first.', { exit: 1 })
    }

    const config = await fs.readJson(configPath)
    this.log('\nCurrent configuration:')
    this.log(JSON.stringify(config, null, 2))
  }

  private setNestedValue(obj: any, path: string, value: any): void {
    const keys = path.split('.')
    const lastKey = keys.pop()!
    let current = obj

    for (const key of keys) {
      if (!(key in current) || typeof current[key] !== 'object') {
        current[key] = {}
      }
      current = current[key]
    }

    current[lastKey] = value
  }

  private getNestedValue(obj: any, path: string): any {
    const keys = path.split('.')
    let current = obj

    for (const key of keys) {
      if (current && typeof current === 'object' && key in current) {
        current = current[key]
      } else {
        return undefined
      }
    }

    return current
  }

  private parseValue(value: string): any {
    // Try to parse as JSON
    if (value === 'true') return true
    if (value === 'false') return false
    if (value === 'null') return null
    if (/^-?\d+$/.test(value)) return parseInt(value, 10)
    if (/^-?\d+\.\d+$/.test(value)) return parseFloat(value)
    if (value.startsWith('{') || value.startsWith('[')) {
      try {
        return JSON.parse(value)
      } catch {
        return value
      }
    }
    return value
  }
}
