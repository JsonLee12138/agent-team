import type { RoleRecord, Repository, RolesQueryParams, PaginatedResponse } from '@/types'

const DEFAULT_API_BASE = '/api/v1'
const API_BASE_URL = (import.meta.env.VITE_API_URL || DEFAULT_API_BASE).replace(/\/$/, '')
const API_TIMEOUT_MS = Number(import.meta.env.VITE_API_TIMEOUT ?? 12_000)

type QueryValue = string | number | boolean | Array<string | number | boolean> | null | undefined
type QueryParams = Record<string, QueryValue>

export class ApiError extends Error {
  status: number
  details?: unknown

  constructor(message: string, status: number, details?: unknown) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.details = details
  }
}

const isAbsoluteUrl = (value: string) => /^https?:\/\//i.test(value)

function buildUrl(path: string, query?: QueryParams): URL {
  const base = API_BASE_URL || DEFAULT_API_BASE
  const normalizedBase = base.replace(/\/$/, '')
  const normalizedPath = path.replace(/^\//, '')
  const combined = `${normalizedBase}/${normalizedPath}`
  const url = isAbsoluteUrl(normalizedBase)
    ? new URL(combined)
    : new URL(combined, window.location.origin)

  if (query) {
    Object.entries(query).forEach(([key, value]) => {
      if (value === undefined || value === null || value === '') return
      if (Array.isArray(value)) {
        value.forEach((entry) => url.searchParams.append(key, String(entry)))
      } else {
        url.searchParams.set(key, String(value))
      }
    })
  }

  return url
}

async function requestJson<T>(path: string, options: { method?: string; query?: QueryParams } = {}): Promise<T> {
  const { method = 'GET', query } = options
  const controller = new AbortController()
  const timeoutId = setTimeout(() => controller.abort(), API_TIMEOUT_MS)

  try {
    const response = await fetch(buildUrl(path, query), {
      method,
      headers: {
        Accept: 'application/json',
      },
      signal: controller.signal,
    })

    if (response.status === 204) {
      return null as T
    }

    const contentType = response.headers.get('content-type') ?? ''
    const body = contentType.includes('application/json') ? await response.json() : await response.text()

    if (!response.ok) {
      const message = typeof body === 'string'
        ? body || response.statusText
        : (body as { message?: string })?.message || response.statusText
      throw new ApiError(message, response.status, body)
    }

    return body as T
  } catch (error) {
    if (error instanceof ApiError) throw error
    if (error instanceof DOMException && error.name === 'AbortError') {
      throw new ApiError('Request timed out. Please try again.', 408)
    }
    if (error instanceof Error) {
      throw new ApiError(error.message, 0)
    }
    throw new ApiError('Unknown error', 0)
  } finally {
    clearTimeout(timeoutId)
  }
}

export function getErrorMessage(error: unknown): string {
  if (error instanceof ApiError) return error.message
  if (error instanceof Error) return error.message
  return 'Something went wrong. Please try again.'
}

export async function fetchRoles(params: RolesQueryParams = {}): Promise<PaginatedResponse<RoleRecord>> {
  const {
    search,
    category,
    framework,
    sort = 'relevance',
    page = 1,
    limit = 12,
  } = params

  return requestJson<PaginatedResponse<RoleRecord>>('/roles', {
    query: {
      search,
      category: category && category !== 'All Roles' ? category : undefined,
      framework,
      sort,
      page,
      limit,
    },
  })
}

export async function fetchRoleById(id: string): Promise<RoleRecord | null> {
  try {
    return await requestJson<RoleRecord>(`/roles/${encodeURIComponent(id)}`)
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) return null
    throw error
  }
}

export async function fetchRoleByName(name: string): Promise<RoleRecord | null> {
  try {
    return await requestJson<RoleRecord>(`/roles/${encodeURIComponent(name)}`)
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) return null
    throw error
  }
}

export async function fetchRepo(owner: string, repo: string): Promise<Repository | null> {
  try {
    return await requestJson<Repository>(`/repos/${encodeURIComponent(owner)}/${encodeURIComponent(repo)}`)
  } catch (error) {
    if (error instanceof ApiError && error.status === 404) return null
    throw error
  }
}

export async function fetchRepos(): Promise<Repository[]> {
  return requestJson<Repository[]>('/repos')
}
