import { Command, Flags, Args } from '@oclif/core'
import { readdir } from 'node:fs/promises'
import { getGlobalRolesDir, getLocalRolesDir, readRoleFile } from '../../core/role/schema'

export default class RoleList extends Command {
  static description = '列出所有角色'

  static examples = [
    '$ otcc role list',
    '$ otcc role list --local',
    '$ otcc role list --global',
  ]

  static flags = {
    global: Flags.boolean({ char: 'g', description: '仅列出全局角色' }),
    local: Flags.boolean({ char: 'l', description: '仅列出本地角色' }),
  }

  async run(): Promise<void> {
    const { flags } = await this.parse(RoleList)

    const roles: Array<{ path: string; role: Awaited<ReturnType<typeof readRoleFile>> }> = []

    if (flags.local || (!flags.local && !flags.global)) {
      const localDir = getLocalRolesDir()
      try {
        const files = await readdir(localDir)
        for (const file of files) {
          if (file.endsWith('.json')) {
            const role = await readRoleFile(`${localDir}/${file}`)
            roles.push({ path: `${localDir}/${file}`, role })
          }
        }
      }
      catch {
        // 目录不存在，跳过
      }
    }

    if (flags.global || (!flags.local && !flags.global)) {
      const globalDir = getGlobalRolesDir()
      try {
        const files = await readdir(globalDir)
        for (const file of files) {
          if (file.endsWith('.json')) {
            const role = await readRoleFile(`${globalDir}/${file}`)
            roles.push({ path: `${globalDir}/${file}`, role })
          }
        }
      }
      catch {
        // 目录不存在，跳过
      }
    }

    if (roles.length === 0) {
      this.log('没有找到任何角色')
      return
    }

    this.log('=== 可用角色 ===\n')
    for (const { path, role } of roles) {
      if (role) {
        this.log(`📄 ${role.name} (${role.fileName})`)
        this.log(`   描述: ${role.description}`)
        this.log(`   路径: ${path}`)
        this.log(`   技能: ${role.skills.length} 个`)
        this.log()
      }
    }
  }
}
