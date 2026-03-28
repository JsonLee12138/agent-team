import { Command, Flags } from '@oclif/core'

export default class {{COMMAND_NAME}} extends Command {
  static description = '{{DESCRIPTION}}'

  static examples = [
    '<%= config.bin %> <%= command.id %> --name World',
    '<%= config.bin %> <%= command.id %> --name "John Doe" --verbose',
  ]

  static flags = {
    name: Flags.string({
      char: 'n',
      description: 'Name to use',
      required: true,
    }),
    verbose: Flags.boolean({
      char: 'v',
      description: 'Show verbose output',
      default: false,
    }),
    force: Flags.boolean({
      char: 'f',
      description: 'Force operation',
      default: false,
    }),
  }

  static args = {}

  async run(): Promise<void> {
    const { flags } = await this.parse({{COMMAND_NAME}})

    if (flags.verbose) {
      this.log('Verbose mode enabled')
    }

    this.log(`Hello, ${flags.name}!`)

    if (flags.force) {
      this.log('Force mode: proceeding without confirmation')
    }
  }
}
