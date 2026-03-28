import { Command, Args } from '@oclif/core'
import { readFile } from 'node:fs/promises'
import { getGlobalRolesDir, getLocalRolesDir, validateRole } from '../../core/role/schema'

export default class RoleValidate extends Command {
  static description = '验证角色文件'

  static examples = [
    '$ otcc role validate 前端架构师',
  ]

  static args = {
    name: Args.string({ required: true, description: '角色文件名' }),
  }

  async run(): Promise<void> {
    const { args } = await this.parse(RoleValidate)

    const localPath = `${getLocalRolesDir()}/${args.name}.json`
    const globalPath = `${getGlobalRolesDir()}/${args.name}.json`

    let content: string | null = null
    let path: string | null = null

    try {
      content = await readFile(localPath, 'utf-8')
      path = localPath
    }
    catch {
      try {
        content = await readFile(globalPath, 'utf-8')
        path = globalPath
      }
      catch {
        this.error(`未找到角色文件: ${args.name}`)
      }
    }

    try {
      const parsed = JSON.parse(content!)
      const validated = validateRole(parsed)
      this.log(`✅ 角色验证通过: ${path}`)
      this.log(`   名称: ${validated.name}`)
      this.log(`   文件名: ${validated.fileName}`)
      this.log(`   版本: ${validated.version}`)
      this.log(`   描述: ${validated.description}`)
      this.log(`   inScope: ${validated.inScope.length} 项`)
      this.log(`   outOfScope: ${validated.outOfScope.length} 项`)
      this.log(`   skills: ${validated.skills.length} 个`)
    }
    catch (e) {
      this.error(`角色验证失败: ${e}`)
    }
  }
}
