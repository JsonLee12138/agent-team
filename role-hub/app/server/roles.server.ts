import { createHash } from 'node:crypto'
import { listRoleRows, listRoleRowsByRepo, type RoleRow } from './db.server'
import type { PaginatedResponse, Repository, RoleRecord, RolesQueryParams } from '~/types'

function parseTags(raw: RoleRow['tags']): string[] {
  if (!raw) return []
  if (Array.isArray(raw)) return raw
  if (typeof raw === 'string') {
    try {
      const parsed = JSON.parse(raw)
      return Array.isArray(parsed) ? parsed.map(String) : []
    } catch {
      return []
    }
  }
  return []
}

function toIsoString(value: string | Date): string {
  if (value instanceof Date) return value.toISOString()
  return new Date(value).toISOString()
}

function toDisplayName(name: string): string {
  return name
    .replace(/[-_]+/g, ' ')
    .split(' ')
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ')
}

function buildRoleRecord(row: RoleRow): RoleRecord {
  const roleName = row.name || row.role_path.split('/').pop() || 'unknown'
  const id = `${row.repo_owner}/${row.repo_name}/${row.role_path}`
  const tags = parseTags(row.tags)
  const score = row.score ?? 0
  const nowIso = toIsoString(row.updated_at || row.created_at)

  return {
    id,
    role_name: roleName,
    display_name: toDisplayName(roleName),
    description: row.description ?? '',
    source_owner: row.repo_owner,
    source_repo: row.repo_name,
    role_path: row.role_path,
    source_ref: 'main',
    status: 'verified',
    folder_hash: createHash('sha1').update(id).digest('hex').slice(0, 8),
    install_count: row.install_count ?? 0,
    tags,
    last_verified_at: nowIso,
    created_at: toIsoString(row.created_at),
    updated_at: toIsoString(row.updated_at),
    readme: undefined,
  }
}

function matchesSearch(role: RoleRecord, term: string): boolean {
  const haystack = [
    role.role_name,
    role.display_name,
    role.description,
    role.source_owner,
    role.source_repo,
    role.role_path,
    role.tags.join(' '),
  ]
    .join(' ')
    .toLowerCase()
  return haystack.includes(term)
}

function filterRoles(roles: RoleRecord[], params: RolesQueryParams): RoleRecord[] {
  let filtered = roles
  if (params.search) {
    const term = params.search.toLowerCase().trim()
    if (term) {
      filtered = filtered.filter((role) => matchesSearch(role, term))
    }
  }

  if (params.framework) {
    const frameworks = Array.isArray(params.framework) ? params.framework : [params.framework]
    const normalized = frameworks.map((fw) => fw.toLowerCase())
    if (normalized.length > 0) {
      filtered = filtered.filter((role) => role.tags.some((tag) => normalized.includes(tag.toLowerCase())))
    }
  }

  return filtered
}

function sortRoles(roles: RoleRecord[], sort?: RolesQueryParams['sort']): RoleRecord[] {
  const sorted = [...roles]
  switch (sort) {
    case 'newest':
      sorted.sort((a, b) => b.updated_at.localeCompare(a.updated_at))
      break
    case 'installs':
      sorted.sort((a, b) => b.install_count - a.install_count)
      break
    default:
      sorted.sort((a, b) => b.install_count - a.install_count)
      break
  }
  return sorted
}

export async function getRoles(params: RolesQueryParams = {}): Promise<PaginatedResponse<RoleRecord>> {
  const rows = await listRoleRows()
  const roles = rows.map(buildRoleRecord)
  const filtered = filterRoles(roles, params)
  const sorted = sortRoles(filtered, params.sort)

  const page = Math.max(1, Number(params.page ?? 1))
  const limit = Math.max(1, Math.min(50, Number(params.limit ?? 12)))
  const start = (page - 1) * limit
  const paged = sorted.slice(start, start + limit)

  return {
    data: paged,
    total: sorted.length,
    page,
    limit,
    hasMore: start + limit < sorted.length,
  }
}

export async function getRoleByName(name: string): Promise<RoleRecord | null> {
  const rows = await listRoleRows()
  const target = name.toLowerCase()
  const match = rows.map(buildRoleRecord).find((role) => role.role_name.toLowerCase() === target)
  return match ?? null
}

export async function getRepo(owner: string, repo: string): Promise<Repository | null> {
  const rows = await listRoleRowsByRepo(owner, repo)
  if (rows.length === 0) return null
  const roles = rows.map(buildRoleRecord)
  const lastSyncedAt = roles
    .map((role) => role.updated_at)
    .sort()
    .at(-1) ?? new Date().toISOString()

  return {
    owner,
    repo,
    description: `Indexed repository for ${owner}/${repo}.`,
    stars: 0,
    license: 'Unknown',
    last_synced_at: lastSyncedAt,
    sync_status: 'healthy',
    roles,
  }
}

export async function getRepos(): Promise<Repository[]> {
  const rows = await listRoleRows()
  const grouped = new Map<string, RoleRecord[]>()
  rows.map(buildRoleRecord).forEach((role) => {
    const key = `${role.source_owner}/${role.source_repo}`
    const existing = grouped.get(key) ?? []
    existing.push(role)
    grouped.set(key, existing)
  })

  return Array.from(grouped.entries()).map(([key, roles]) => {
    const [owner, repo] = key.split('/')
    const lastSyncedAt = roles
      .map((role) => role.updated_at)
      .sort()
      .at(-1) ?? new Date().toISOString()

    return {
      owner,
      repo,
      description: `Indexed repository for ${owner}/${repo}.`,
      stars: 0,
      license: 'Unknown',
      last_synced_at: lastSyncedAt,
      sync_status: 'healthy',
      roles,
    }
  })
}
