import { Command, Flags, Args } from '@oclif/core'
import { access, mkdir, writeFile } from 'node:fs/promises'
import { getGlobalRolesDir, getLocalRolesDir, kebabCase, parseSkillInput, validateRole } from '../../core/role/schema'
import { promptRoleInteractive } from '../../prompts/role'
import inquirer from 'inquirer'

export default class RoleCreate extends Command {
  static description = '创建新角色'

  static examples = [
    '$ otcc role create --name 前端架构师',
    '$ otcc role create --interactive',
  ]

  static flags = {
    name: Flags.string({ char: 'n', description: '角色名称' }),
    'file-name': Flags.string({ description: '文件名 (kebab-case)' }),
    description: Flags.string({ char: 'd', description: '角色描述' }),
    prompt: Flags.string({ char: 'p', description: '系统提示词' }),
    'in-scope': Flags.string({ char: 'i', description: '职责范围内 (逗号分隔)' }),
    'out-of-scope': Flags.string({ char: 'o', description: '职责范围外 (逗号分隔)' }),
    skills: Flags.string({ char: 's', description: '关联的 skills' }),
    interactive: Flags.boolean({ char: 'I', description: '交互式创建' }),
    global: Flags.boolean({ char: 'g', description: '创建到全局目录' }),
    local: Flags.boolean({ char: 'l', description: '创建到本地目录' }),
  }

  static args = {
    name: Args.string({ description: '角色名称' }),
  }

  async run(): Promise<void> {
    const { flags, args } = await this.parse(RoleCreate)

    let roleData

    if (flags.interactive || (!flags.name && !args.name)) {
      roleData = await promptRoleInteractive()
    }
    else {
      const name = flags.name || args.name!
      const skills = flags.skills
        ? flags.skills.split(',').map(parseSkillInput)
        : []

      roleData = {
        name,
        fileName: flags['file-name'] || kebabCase(name),
        version: '1.0.0',
        description: flags.description || `${name}角色`,
        prompt: flags.prompt || `你是一位${name}，具有丰富的专业知识和经验。`,
        inScope: flags['in-scope']?.split(',').filter(Boolean) || [],
        outOfScope: flags['out-of-scope']?.split(',').filter(Boolean) || [],
        skills,
      }
    }

    let validated: ReturnType<typeof validateRole>
    try {
      validated = validateRole(roleData)
    }
    catch (e) {
      this.error(`角色数据验证失败: ${e}`)
    }

    const rolesDir = flags.global
      ? getGlobalRolesDir()
      : getLocalRolesDir()

    await mkdir(rolesDir, { recursive: true })

    const filePath = `${rolesDir}/${validated.fileName}.json`

    if (await access(filePath).then(() => true).catch(() => false)) {
      this.log(`⚠️  角色文件已存在: ${filePath}`)
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

    await writeFile(filePath, JSON.stringify(validated, null, 2), 'utf-8')
    this.log(`✅ 角色已保存: ${filePath}`)
  }
}
