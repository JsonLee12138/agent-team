import { Command, Flags, Help } from '@oclif/core'

export default class Role extends Command {
  static description = '角色管理命令'

  static flags = {
    help: Flags.help(),
  }

  async run(): Promise<void> {
    const help = new Help(
      this.config,
      this.config.pjson.oclif?.helpOptions ?? this.config.pjson.helpOptions,
    )

    await help.showHelp(this.id ? [this.id] : [])
  }
}
