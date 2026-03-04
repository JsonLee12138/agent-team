export type RoleStatus = 'discovered' | 'verified' | 'invalid' | 'unreachable'

export interface RoleRecord {
  id: string
  role_name: string
  display_name: string
  description: string
  source_owner: string
  source_repo: string
  role_path: string
  source_ref: string
  status: RoleStatus
  folder_hash: string
  install_count: number
  tags: string[]
  last_verified_at: string
  created_at: string
  updated_at: string
  readme?: string
}

export interface Repository {
  owner: string
  repo: string
  description: string
  stars: number
  license: string
  last_synced_at: string
  sync_status: 'healthy' | 'syncing' | 'error'
  roles: RoleRecord[]
}

export interface RolesQueryParams {
  search?: string
  category?: string
  framework?: string
  sort?: 'relevance' | 'installs' | 'newest'
  page?: number
  limit?: number
}

export interface PaginatedResponse<T> {
  data: T[]
  total: number
  page: number
  limit: number
  hasMore: boolean
}
