import { Command, Args } from '@oclif/core'
import { getGlobalRolesDir, getLocalRolesDir, readRoleFile } from '../../core/role/schema'

export default class RoleShow extends Command {
  static description = '查看角色详情'

  static examples = [
    '$ otcc role show 前端架构师',
  ]

  static args = {
    name: Args.string({ required: true, description: '角色文件名' }),
  }

  async run(): Promise<void> {
    const { args } = await this.parse(RoleShow)

    const localPath = `${getLocalRolesDir()}/${args.name}.json`
    const globalPath = `${getGlobalRolesDir()}/${args.name}.json`

    const localRole = await readRoleFile(localPath)
    if (localRole) {
      this.log('=== 角色详情 ===\n')
      this.log(JSON.stringify(localRole, null, 2))
      return
    }

    const globalRole = await readRoleFile(globalPath)
    if (globalRole) {
      this.log('=== 角色详情 ===\n')
      this.log(JSON.stringify(globalRole, null, 2))
      return
    }

    this.error(`未找到角色: ${args.name}`)
  }
}
