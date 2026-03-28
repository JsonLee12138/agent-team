import inquirer from 'inquirer'
import { kebabCase, parseSkillInput, type Skill } from '../core/role/schema'

export async function promptRoleName(): Promise<string> {
  const { name } = await inquirer.prompt<{ name: string }>({
    type: 'input',
    name: 'name',
    message: '角色名称 (例如: 前端架构师):',
    validate: (input: string) => input.trim().length > 0 || '请输入角色名称',
  })
  return name
}

export async function promptFileName(defaultValue: string): Promise<string> {
  const { fileName } = await inquirer.prompt<{ fileName: string }>({
    type: 'input',
    name: 'fileName',
    message: `文件名称 (直接回车使用: ${defaultValue}):`,
    default: defaultValue,
  })
  return fileName || defaultValue
}

export async function promptDescription(defaultValue: string): Promise<string> {
  const { description } = await inquirer.prompt<{ description: string }>({
    type: 'input',
    name: 'description',
    message: '角色描述:',
    default: defaultValue,
  })
  return description || defaultValue
}

export async function promptVersion(): Promise<string> {
  const { version } = await inquirer.prompt<{ version: string }>({
    type: 'input',
    name: 'version',
    message: '版本 (直接回车使用 1.0.0):',
    default: '1.0.0',
  })
  return version || '1.0.0'
}

export async function promptPrompt(defaultValue: string): Promise<string> {
  const { prompt } = await inquirer.prompt<{ prompt: string }>({
    type: 'input',
    name: 'prompt',
    message: '系统提示词 (prompt):',
    default: defaultValue,
  })
  return prompt || defaultValue
}

export async function promptList(
  message: string,
  initial: string[] = [],
): Promise<string[]> {
  const { lines } = await inquirer.prompt<{ lines: string }>({
    type: 'input',
    name: 'lines',
    message: `${message}\n(输完一行后回车，空行结束)`,
    default: initial.join('\n'),
  })

  return lines
    .split('\n')
    .map(l => l.trim())
    .filter(l => l.length > 0)
}

export async function promptSkills(): Promise<Skill[]> {
  console.log('\n关联的 skills (格式: owner/repo@suffix:name:description，空行结束):')
  console.log('示例: pexoai/pexo-skills@pexoai-agent:前端架构助手:擅长React生态')

  const { skillInput } = await inquirer.prompt<{ skillInput: string }>({
    type: 'input',
    name: 'skillInput',
    message: 'skill 输入:',
  })

  if (!skillInput.trim())
    return []

  return skillInput
    .split('\n')
    .map(l => l.trim())
    .filter(l => l.length > 0)
    .map(parseSkillInput)
}

export interface RoleAnswers {
  name: string
  fileName: string
  version: string
  description: string
  prompt: string
  inScope: string[]
  outOfScope: string[]
  skills: Skill[]
}

export async function promptRoleInteractive(): Promise<RoleAnswers> {
  console.log('=== 角色创建向导 ===\n')

  const name = await promptRoleName()
  const autoFileName = kebabCase(name)
  const fileName = await promptFileName(autoFileName)
  const description = await promptDescription(`${name}角色`)
  const version = await promptVersion()
  const defaultPrompt = `你是一位${name}，具有丰富的专业知识和经验。`
  const prompt = await promptPrompt(defaultPrompt)
  const inScope = await promptList('职责范围内的工作 (inScope):')
  const outOfScope = await promptList('职责范围外的工作 (outOfScope):')
  const skills = await promptSkills()

  return {
    name,
    fileName,
    version,
    description,
    prompt,
    inScope,
    outOfScope,
    skills,
  }
}
