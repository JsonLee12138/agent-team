import { Command, Flags, Args } from '@oclif/core'
import { access, rm } from 'node:fs/promises'
import { getGlobalRolesDir, getLocalRolesDir } from '../../core/role/schema'

export default class RoleDelete extends Command {
  static description = '删除角色'

  static examples = [
    '$ otcc role delete 前端架构师',
    '$ otcc role delete 前端架构师 --global',
  ]

  static flags = {
    global: Flags.boolean({ char: 'g', description: '删除全局角色' }),
    local: Flags.boolean({ char: 'l', description: '删除本地角色' }),
  }

  static args = {
    name: Args.string({ required: true, description: '角色文件名' }),
  }

  async run(): Promise<void> {
    const { flags, args } = await this.parse(RoleDelete)

    async function deleteFrom(dir: string): Promise<boolean> {
      const path = `${dir}/${args.name}.json`
      if (await access(path).then(() => true).catch(() => false)) {
        await rm(path)
        return true
      }
      return false
    }

    let deleted = false

    if (flags.local || (!flags.local && !flags.global)) {
      deleted = await deleteFrom(getLocalRolesDir())
    }

    if (flags.global) {
      deleted = await deleteFrom(getGlobalRolesDir()) || deleted
    }

    if (deleted) {
      this.log(`✅ 已删除: ${args.name}`)
    }
    else {
      this.error(`未找到角色: ${args.name}`)
    }
  }
}
