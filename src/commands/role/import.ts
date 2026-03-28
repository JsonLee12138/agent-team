import { Command, Args } from '@oclif/core'
import { access, copyFile, mkdir } from 'node:fs/promises'
import { getGlobalRolesDir, getLocalRolesDir } from '../../core/role/schema'
import inquirer from 'inquirer'

export default class RoleImport extends Command {
  static description = '从全局模板导入角色'

  static examples = [
    '$ otcc role import 前端架构师',
  ]

  static args = {
    name: Args.string({ required: true, description: '角色文件名' }),
  }

  async run(): Promise<void> {
    const { args } = await this.parse(RoleImport)

    const globalDir = getGlobalRolesDir()
    const globalPath = `${globalDir}/${args.name}.json`
    const localDir = getLocalRolesDir()
    const localPath = `${localDir}/${args.name}.json`

    if (!(await access(globalPath).then(() => true).catch(() => false))) {
      this.error(`全局模板中未找到角色: ${args.name}`)
    }

    await mkdir(localDir, { recursive: true })

    if (await access(localPath).then(() => true).catch(() => false)) {
      this.log(`⚠️  本地已存在: ${localPath}`)
      const { confirm } = await inquirer.prompt<{ confirm: boolean }>({
        type: 'confirm',
        name: 'confirm',
        message: '是否覆盖?',
        default: false,
      })
      if (!confirm) {
        this.log('已取消')
        return
      }
    }

    await copyFile(globalPath, localPath)
    this.log(`✅ 已导入: ${localPath}`)
  }
}
