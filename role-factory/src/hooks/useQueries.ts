import { useQuery } from '@tanstack/react-query'
import { fetchRoles, fetchRoleByName, fetchRepo } from '@/api/client'
import type { RolesQueryParams } from '@/types'

export function useRoles(params: RolesQueryParams = {}) {
  return useQuery({
    queryKey: ['roles', params],
    queryFn: () => fetchRoles(params),
  })
}

export function useRole(name: string) {
  return useQuery({
    queryKey: ['role', name],
    queryFn: () => fetchRoleByName(name),
    enabled: !!name,
  })
}

export function useRepo(owner: string, repo: string) {
  return useQuery({
    queryKey: ['repo', owner, repo],
    queryFn: () => fetchRepo(owner, repo),
    enabled: !!owner && !!repo,
  })
}
