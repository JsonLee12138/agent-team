import { parse as parseYaml } from './yaml-lite.server'

export interface GitHubRoleContent {
  system_md?: string
  skills?: string[]
  in_scope?: string[]
  out_of_scope?: string[]
}

const cache = new Map<string, { data: GitHubRoleContent; expires: number }>()
const CACHE_TTL_MS = 5 * 60 * 1000

function rawUrl(owner: string, repo: string, branch: string, path: string): string {
  return `https://raw.githubusercontent.com/${owner}/${repo}/${branch}/${path}`
}

async function fetchText(url: string): Promise<string | null> {
  try {
    const res = await fetch(url, { signal: AbortSignal.timeout(5000) })
    if (!res.ok) return null
    return await res.text()
  } catch {
    return null
  }
}

export async function fetchRoleContent(
  owner: string,
  repo: string,
  rolePath: string,
  branch = 'main',
): Promise<GitHubRoleContent> {
  const cacheKey = `${owner}/${repo}/${rolePath}@${branch}`
  const cached = cache.get(cacheKey)
  if (cached && cached.expires > Date.now()) {
    return cached.data
  }

  const prefix = rolePath.includes('/') ? rolePath : rolePath
  const systemMdUrl = rawUrl(owner, repo, branch, `${prefix}/system.md`)
  const roleYamlUrl = rawUrl(owner, repo, branch, `${prefix}/references/role.yaml`)

  const [systemMd, roleYamlText] = await Promise.all([
    fetchText(systemMdUrl),
    fetchText(roleYamlUrl),
  ])

  const result: GitHubRoleContent = {}

  if (systemMd) {
    result.system_md = systemMd
  }

  if (roleYamlText) {
    const parsed = parseYaml(roleYamlText)
    if (parsed.skills && Array.isArray(parsed.skills)) {
      result.skills = parsed.skills.map(String)
    }
    const scope = parsed.scope as Record<string, unknown> | undefined
    if (scope) {
      if (Array.isArray(scope.in_scope)) {
        result.in_scope = scope.in_scope.map(String)
      }
      if (Array.isArray(scope.out_of_scope)) {
        result.out_of_scope = scope.out_of_scope.map(String)
      }
    }
  }

  cache.set(cacheKey, { data: result, expires: Date.now() + CACHE_TTL_MS })
  return result
}
