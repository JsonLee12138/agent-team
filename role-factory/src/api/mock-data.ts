import type { RoleRecord, Repository } from '@/types'

export const MOCK_ROLES: RoleRecord[] = [
  {
    id: '1',
    role_name: 'frontend-architect',
    display_name: 'Frontend Architect',
    description: 'Expert in Vite, Turborepo, and modern frontend architecture. Specializes in monorepo setups, build optimization, and developer tooling.',
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
    readme: `# Frontend Architect

This role provides a comprehensive set of tools and best practices for building modern web applications. It includes configuration for Vite, Turborepo, and Vitest, ensuring a smooth and efficient development experience.

## Features
- Vite project scaffolding and plugin configuration
- Turborepo monorepo setup with optimized task pipelines
- Vitest unit and integration test configuration
- UnoCSS atomic CSS integration
- TypeScript strict mode with path aliases

## Usage
\`\`\`bash
agent-team role install frontend-architect
\`\`\`

## Configuration
Set the following environment variables:
- \`VITE_API_URL\` - Backend API endpoint
- \`VITE_ENV\` - Environment (development/staging/production)`,
  },
  {
    id: '2',
    role_name: 'pencil-designer',
    display_name: 'Pencil Designer',
    description: 'Senior UI/UX designer specializing in design systems. Creates high-fidelity mockups and design tokens using Pencil MCP.',
    source_owner: 'agent-team',
    source_repo: 'roles',
    role_path: 'skills/pencil-designer',
    source_ref: 'main',
    status: 'verified',
    folder_hash: 'def456',
    install_count: 8900,
    tags: ['design', 'ui-ux', 'pencil', 'design-system'],
    last_verified_at: '2026-03-03T14:00:00Z',
    created_at: '2026-02-01T09:00:00Z',
    updated_at: '2026-03-03T14:00:00Z',
    readme: `# Pencil Designer

A senior UI/UX designer role that specializes in design systems and visual design using Pencil MCP tools.

## Capabilities
- Design system creation and maintenance
- High-fidelity mockup generation
- Component library design
- Responsive layout planning`,
  },
  {
    id: '3',
    role_name: 'solo-ops',
    display_name: 'Solo Ops',
    description: 'Automate your solo developer workflows with ease. Handles CI/CD, deployment, and infrastructure tasks.',
    source_owner: 'agent-team',
    source_repo: 'roles',
    role_path: 'skills/solo-ops',
    source_ref: 'main',
    status: 'verified',
    folder_hash: 'ghi789',
    install_count: 15600,
    tags: ['devops', 'ci-cd', 'deployment', 'automation'],
    last_verified_at: '2026-03-04T08:00:00Z',
    created_at: '2026-01-20T10:00:00Z',
    updated_at: '2026-03-04T08:00:00Z',
    readme: `# Solo Ops

Automate your solo developer workflows. Handles CI/CD pipelines, deployments, and infrastructure management.`,
  },
  {
    id: '4',
    role_name: 'api-guardian',
    display_name: 'API Guardian',
    description: 'Backend API specialist focused on REST/GraphQL design, validation, and security best practices.',
    source_owner: 'community',
    source_repo: 'agent-roles',
    role_path: 'roles/api-guardian',
    source_ref: 'main',
    status: 'verified',
    folder_hash: 'jkl012',
    install_count: 6700,
    tags: ['backend', 'api', 'security', 'rest', 'graphql'],
    last_verified_at: '2026-03-02T16:00:00Z',
    created_at: '2026-02-10T11:00:00Z',
    updated_at: '2026-03-02T16:00:00Z',
  },
  {
    id: '5',
    role_name: 'test-sentinel',
    display_name: 'Test Sentinel',
    description: 'Quality assurance role that ensures comprehensive test coverage with Vitest, Playwright, and testing best practices.',
    source_owner: 'community',
    source_repo: 'agent-roles',
    role_path: 'roles/test-sentinel',
    source_ref: 'main',
    status: 'verified',
    folder_hash: 'mno345',
    install_count: 4200,
    tags: ['testing', 'vitest', 'playwright', 'qa'],
    last_verified_at: '2026-03-01T12:00:00Z',
    created_at: '2026-02-15T08:00:00Z',
    updated_at: '2026-03-01T12:00:00Z',
  },
  {
    id: '6',
    role_name: 'data-weaver',
    display_name: 'Data Weaver',
    description: 'Database design and data modeling specialist. Expertise in PostgreSQL, Prisma ORM, and migration strategies.',
    source_owner: 'community',
    source_repo: 'agent-roles',
    role_path: 'roles/data-weaver',
    source_ref: 'main',
    status: 'verified',
    folder_hash: 'pqr678',
    install_count: 3800,
    tags: ['database', 'postgresql', 'prisma', 'data-modeling'],
    last_verified_at: '2026-02-28T10:00:00Z',
    created_at: '2026-02-20T14:00:00Z',
    updated_at: '2026-02-28T10:00:00Z',
  },
]

export const MOCK_REPOS: Repository[] = [
  {
    owner: 'agent-team',
    repo: 'roles',
    description: 'The official role repository for the agent-team project.',
    stars: 1200,
    license: 'MIT',
    last_synced_at: '2026-03-04T13:00:00Z',
    sync_status: 'healthy',
    roles: MOCK_ROLES.filter((r) => r.source_owner === 'agent-team'),
  },
  {
    owner: 'community',
    repo: 'agent-roles',
    description: 'Community-contributed roles for agent-team workflows.',
    stars: 480,
    license: 'Apache-2.0',
    last_synced_at: '2026-03-04T11:30:00Z',
    sync_status: 'healthy',
    roles: MOCK_ROLES.filter((r) => r.source_owner === 'community'),
  },
]

export const CATEGORIES = ['All Roles', 'Trending', 'Recently Added'] as const
export const FRAMEWORKS = ['Gemini', 'OpenAI', 'Claude', 'Vite', 'React'] as const
