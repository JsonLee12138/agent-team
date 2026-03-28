import { z } from 'zod'

export const SkillSchema = z.object({
  path: z.string(),
  name: z.string(),
  description: z.string(),
})

export const RoleSchema = z.object({
  name: z.string(),
  fileName: z.string().regex(/^[a-z0-9]+(-[a-z0-9]+)*$/, 'fileName 必须是 kebab-case'),
  version: z.string(),
  description: z.string(),
  prompt: z.string(),
  inScope: z.array(z.string()),
  outOfScope: z.array(z.string()),
  skills: z.array(SkillSchema),
})

export type Role = z.infer<typeof RoleSchema>
export type Skill = z.infer<typeof SkillSchema>

export function getLocalRolesDir(): string {
  return '.otcc/roles'
}

export function getGlobalRolesDir(): string {
  return `${process.env.HOME}/.claude/plugins/marketplaces/.otcc/roles`
}

export function kebabCase(str: string): string {
  return str
    .toLowerCase()
    .replace(/[\s_]+/g, '-')
    .replace(/[^a-z0-9-]/g, '')
    .replace(/^-+|-+$/g, '')
}

export function parseSkillInput(input: string): Skill {
  const parts = input.split(':')
  if (parts.length >= 3) {
    return {
      path: parts[0],
      name: parts[1],
      description: parts[2],
    }
  }
  else if (parts.length === 2) {
    return {
      path: parts[0],
      name: parts[1],
      description: '',
    }
  }
  const repoPart = parts[0].split('/').pop()?.split('@')[0] || parts[0]
  return {
    path: parts[0],
    name: repoPart,
    description: '',
  }
}

export function validateRole(data: unknown): Role {
  return RoleSchema.parse(data)
}

export async function readRoleFile(filePath: string): Promise<Role | null> {
  try {
    const { readFile } = await import('node:fs/promises')
    const content = await readFile(filePath, 'utf-8')
    return validateRole(JSON.parse(content))
  }
  catch {
    return null
  }
}
