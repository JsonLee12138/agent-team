import type { RoleRecord, Repository, RolesQueryParams, PaginatedResponse } from '@/types'
import { MOCK_ROLES, MOCK_REPOS } from './mock-data'

const delay = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

export async function fetchRoles(params: RolesQueryParams = {}): Promise<PaginatedResponse<RoleRecord>> {
  await delay(300)
  const { search, category, sort = 'relevance', page = 1, limit = 12 } = params

  let filtered = MOCK_ROLES.filter((r) => r.status === 'verified')

  if (search) {
    const q = search.toLowerCase()
    filtered = filtered.filter(
      (r) =>
        r.role_name.includes(q) ||
        r.display_name.toLowerCase().includes(q) ||
        r.description.toLowerCase().includes(q) ||
        r.tags.some((t) => t.includes(q)),
    )
  }

  if (category === 'Trending') {
    filtered = [...filtered].sort((a, b) => b.install_count - a.install_count)
  } else if (category === 'Recently Added') {
    filtered = [...filtered].sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())
  }

  if (sort === 'installs') {
    filtered = [...filtered].sort((a, b) => b.install_count - a.install_count)
  } else if (sort === 'newest') {
    filtered = [...filtered].sort((a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime())
  }

  const start = (page - 1) * limit
  const data = filtered.slice(start, start + limit)

  return {
    data,
    total: filtered.length,
    page,
    limit,
    hasMore: start + limit < filtered.length,
  }
}

export async function fetchRoleById(id: string): Promise<RoleRecord | null> {
  await delay(200)
  return MOCK_ROLES.find((r) => r.id === id) ?? null
}

export async function fetchRoleByName(name: string): Promise<RoleRecord | null> {
  await delay(200)
  return MOCK_ROLES.find((r) => r.role_name === name) ?? null
}

export async function fetchRepo(owner: string, repo: string): Promise<Repository | null> {
  await delay(200)
  return MOCK_REPOS.find((r) => r.owner === owner && r.repo === repo) ?? null
}

export async function fetchRepos(): Promise<Repository[]> {
  await delay(200)
  return MOCK_REPOS
}
