import { test, expect } from '@playwright/test'

test.describe('role-factory core routes', () => {
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
