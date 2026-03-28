import { Command, Flags } from '@oclif/core'

/**
 * Plugin command that extends the main CLI
 *
 * This command will be available as:
 *   mycli {{PLUGIN_NAME}}:{{COMMAND_NAME}}
 */
export default class {{COMMAND_CLASS}} extends Command {
  static description = 'Plugin command: {{DESCRIPTION}}'

  static examples = [
    '<%= config.bin %> {{PLUGIN_NAME}}:{{COMMAND_NAME}} --option value',
  ]

  static flags = {
    option: Flags.string({
      char: 'o',
      description: 'Plugin-specific option',
    }),
    verbose: Flags.boolean({
      char: 'v',
      description: 'Verbose output',
      default: false,
    }),
  }

  static args = {}

  async run(): Promise<void> {
    const { flags } = await this.parse({{COMMAND_CLASS}})

    if (flags.verbose) {
      this.log('Running plugin command...')
    }

    // Plugin-specific logic here
    this.log(`Plugin {{PLUGIN_NAME}} executing: ${this.id}`)

    if (flags.option) {
      this.log(`Option value: ${flags.option}`)
    }

    // Access main CLI config if needed
    const cliConfig = this.config
    this.log(`CLI version: ${cliConfig.version}`)

    // You can also access other plugins
    const plugins = cliConfig.plugins
    if (flags.verbose) {
      this.log(`Loaded plugins: ${plugins.map(p => p.name).join(', ')}`)
    }
  }
}
