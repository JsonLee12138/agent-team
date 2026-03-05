import { test, expect } from '@playwright/test'
import type { Page } from '@playwright/test'

const roles = [
  {
    id: '1',
    role_name: 'frontend-architect',
    display_name: 'Frontend Architect',
    description: 'Expert in Vite, Turborepo, and modern frontend architecture.',
    source_owner: 'agent-team',
    source_repo: 'roles',
    role_path: 'skills/frontend-architect',
    source_ref: 'main',
    status: 'verified',
    folder_hash: 'abc123',
    install_count: 12400,
    tags: ['frontend', 'vite', 'turborepo', 'architecture'],
    last_verified_at: '2026-03-04T10:00:00Z',
    created_at: '2026-01-15T08:00:00Z',
    updated_at: '2026-03-04T10:00:00Z',
    readme: '# Frontend Architect\n\nUse this role to bootstrap modern frontend systems.',
  },
  {
    id: '2',
    role_name: 'pencil-designer',
    display_name: 'Pencil Designer',
    description: 'Senior UI/UX designer specializing in design systems.',
    source_owner: 'agent-team',
    source_repo: 'roles',
    role_path: 'skills/pencil-designer',
    source_ref: 'main',
    status: 'verified',
    folder_hash: 'def456',
    install_count: 8900,
    tags: ['design', 'ui-ux', 'pencil'],
    last_verified_at: '2026-03-03T14:00:00Z',
    created_at: '2026-02-01T09:00:00Z',
    updated_at: '2026-03-03T14:00:00Z',
  },
  {
    id: '3',
    role_name: 'solo-ops',
    display_name: 'Solo Ops',
    description: 'Automate your solo developer workflows with ease.',
    source_owner: 'agent-team',
    source_repo: 'roles',
    role_path: 'skills/solo-ops',
    source_ref: 'main',
    status: 'verified',
    folder_hash: 'ghi789',
    install_count: 15600,
    tags: ['devops', 'ci-cd', 'deployment'],
    last_verified_at: '2026-03-04T08:00:00Z',
    created_at: '2026-01-20T10:00:00Z',
    updated_at: '2026-03-04T08:00:00Z',
  },
  {
    id: '4',
    role_name: 'api-guardian',
    display_name: 'API Guardian',
    description: 'Backend API specialist focused on REST/GraphQL design.',
    source_owner: 'community',
    source_repo: 'agent-roles',
    role_path: 'roles/api-guardian',
    source_ref: 'main',
    status: 'verified',
    folder_hash: 'jkl012',
    install_count: 6700,
    tags: ['backend', 'api', 'security'],
    last_verified_at: '2026-03-02T16:00:00Z',
    created_at: '2026-02-10T11:00:00Z',
    updated_at: '2026-03-02T16:00:00Z',
  },
  {
    id: '5',
    role_name: 'test-sentinel',
    display_name: 'Test Sentinel',
    description: 'Quality assurance role with Vitest and Playwright expertise.',
    source_owner: 'community',
    source_repo: 'agent-roles',
    role_path: 'roles/test-sentinel',
    source_ref: 'main',
    status: 'verified',
    folder_hash: 'mno345',
    install_count: 4200,
    tags: ['testing', 'vitest', 'playwright'],
    last_verified_at: '2026-03-01T12:00:00Z',
    created_at: '2026-02-15T08:00:00Z',
    updated_at: '2026-03-01T12:00:00Z',
  },
  {
    id: '6',
    role_name: 'support-pilot',
    display_name: 'Support Pilot',
    description: 'Customer success and support workflows for AI teams.',
    source_owner: 'community',
    source_repo: 'agent-roles',
    role_path: 'roles/support-pilot',
    source_ref: 'main',
    status: 'verified',
    folder_hash: 'pqr678',
    install_count: 2100,
    tags: ['support', 'ops'],
    last_verified_at: '2026-03-01T09:00:00Z',
    created_at: '2026-02-18T08:00:00Z',
    updated_at: '2026-03-01T09:00:00Z',
  },
]

const repoPayload = {
  owner: 'agent-team',
  repo: 'roles',
  description: 'The official role repository for the agent-team project.',
  stars: 4200,
  license: 'MIT',
  last_synced_at: new Date().toISOString(),
  sync_status: 'healthy',
  roles: roles.slice(0, 3),
}

async function setupApiRoutes(page: Page) {
  await page.route('**/api/v1/roles**', async (route) => {
    const url = new URL(route.request().url())
    if (url.pathname === '/api/v1/roles') {
      const search = url.searchParams.get('search')?.toLowerCase()
      const filtered = search
        ? roles.filter((role) =>
          role.role_name.includes(search) ||
          role.display_name.toLowerCase().includes(search) ||
          role.description.toLowerCase().includes(search),
        )
        : roles
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: filtered,
          total: filtered.length,
          page: 1,
          limit: 12,
          hasMore: false,
        }),
      })
    }

    const name = decodeURIComponent(url.pathname.split('/').pop() ?? '')
    const role = roles.find((entry) => entry.role_name === name)
    if (!role) {
      return route.fulfill({
        status: 404,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'Role not found' }),
      })
    }
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(role),
    })
  })

  await page.route('**/api/v1/repos**', async (route) => {
    const url = new URL(route.request().url())
    if (url.pathname === '/api/v1/repos') {
      return route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify([repoPayload]),
      })
    }

    const segments = url.pathname.split('/')
    const repo = segments.pop()
    const owner = segments.pop()
    if (owner !== repoPayload.owner || repo !== repoPayload.repo) {
      return route.fulfill({
        status: 404,
        contentType: 'application/json',
        body: JSON.stringify({ message: 'Repository not found' }),
      })
    }
    return route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(repoPayload),
    })
  })
}

test.describe('role-factory core routes', () => {
  test.beforeEach(async ({ page }) => {
    await setupApiRoutes(page)
  })

  test('Home and Roles Directory render with role list', async ({ page }) => {
    await page.goto('/')
    await expect(page.getByRole('heading', { name: 'Find the perfect role for your Agent' })).toBeVisible()
    await expect(page.getByText('6 roles')).toBeVisible()
    await expect(page.getByRole('link', { name: 'frontend-architect' })).toBeVisible()
  })

  test('Role Detail renders content and install command', async ({ page }) => {
    await page.goto('/roles/frontend-architect')
    await expect(page.getByRole('heading', { name: 'frontend-architect' })).toBeVisible()
    await expect(page.getByText('README.md')).toBeVisible()
    await expect(page.locator('code', { hasText: 'agent-team role install frontend-architect' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'agent-team/roles' })).toBeVisible()
  })

  test('Repo Detail renders repository info', async ({ page }) => {
    await page.goto('/repos/agent-team/roles')
    await expect(page.getByRole('heading', { name: 'agent-team/roles' })).toBeVisible()
    await expect(page.getByText('The official role repository for the agent-team project.')).toBeVisible()
    await expect(page.getByRole('link', { name: 'frontend-architect' })).toBeVisible()
  })

  test('Empty states render for missing data', async ({ page }) => {
    await page.goto('/')
    await page.getByPlaceholder('Search 2,400+ roles...').fill('no-such-role')
    await expect(page.getByRole('heading', { name: 'No roles found' })).toBeVisible()

    await page.goto('/roles/does-not-exist')
    await expect(page.getByRole('heading', { name: 'Role not found' })).toBeVisible()

    await page.goto('/repos/missing/repo')
    await expect(page.getByRole('heading', { name: 'Repository not found' })).toBeVisible()
  })
})
