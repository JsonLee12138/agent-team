import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { Link, useLoaderData, useNavigation, useRevalidator } from '@remix-run/react'
import { AlertTriangle, CheckCircle2, ChevronRight } from 'lucide-react'
import { CopyButton } from '~/components/CopyButton'
import { Tag } from '~/components/Tag'
import { PageSkeleton, Skeleton } from '~/components/Skeleton'
import { EmptyState } from '~/components/EmptyState'
import { getRoleByName } from '~/server/roles.server'
import { getErrorMessage } from '~/utils/errors'
import { buildCanonical, buildJsonLd, buildMeta } from '~/utils/seo'
import type { RoleRecord } from '~/types'

interface LoaderData {
  role: RoleRecord | null
  error?: string
}

export async function loader({ params }: LoaderFunctionArgs) {
  const name = params.name
  if (!name) {
    return json<LoaderData>({ role: null, error: 'Role name is required.' }, { status: 400 })
  }

  try {
    const role = await getRoleByName(name)
    return json<LoaderData>({ role })
  } catch (error) {
    return json<LoaderData>({ role: null, error: getErrorMessage(error) }, { status: 500 })
  }
}

export const meta: MetaFunction<typeof loader> = ({ data, params }) => {
  const name = params.name ?? 'Role'
  if (!data?.role) {
    return [{ title: `${name} | Role Hub` }]
  }
  const { role } = data
  const canonical = buildCanonical(`/roles/${role.role_name}`)
  const description = role.description || `Install and use the ${role.display_name} AI role with agent-team.`
  return [
    ...buildMeta({ title: role.display_name, description, canonical }),
    {
      'script:ld+json': buildJsonLd({
        '@context': 'https://schema.org',
        '@type': 'BreadcrumbList',
        itemListElement: [
          { '@type': 'ListItem', position: 1, name: 'Home', item: buildCanonical('/') },
          { '@type': 'ListItem', position: 2, name: 'Roles', item: buildCanonical('/') },
          { '@type': 'ListItem', position: 3, name: role.display_name, item: canonical },
        ],
      }),
    },
    {
      'script:ld+json': buildJsonLd({
        '@context': 'https://schema.org',
        '@type': 'SoftwareApplication',
        name: role.display_name,
        description,
        url: canonical,
        applicationCategory: 'DeveloperApplication',
        keywords: role.tags.join(', '),
      }),
    },
  ]
}

export default function RoleDetailPage() {
  const { role, error } = useLoaderData<typeof loader>()
  const navigation = useNavigation()
  const revalidator = useRevalidator()

  if (navigation.state !== 'idle') {
    return (
      <PageSkeleton>
        <div className="flex flex-col gap-8 px-6 lg:px-20 py-10">
          <Skeleton className="h-4 w-48" />
          <Skeleton className="h-10 w-64" />
          <Skeleton className="h-14 w-full max-w-150" />
          <div className="flex gap-20">
            <div className="flex-1 flex flex-col gap-4">
              <Skeleton className="h-6 w-32" />
              <Skeleton className="h-48 w-full" />
            </div>
            <Skeleton className="h-64 w-80" />
          </div>
        </div>
      </PageSkeleton>
    )
  }

  if (error) {
    return (
      <EmptyState
        title="Unable to load role"
        description={error}
        showClearButton={false}
        actionLabel={revalidator.state !== 'idle' ? 'Retrying...' : 'Retry'}
        onAction={() => revalidator.revalidate()}
        icon={<AlertTriangle size={64} className="text-warning" />}
      />
    )
  }

  if (!role) {
    return (
      <EmptyState
        title="Role not found"
        description="The role you're looking for doesn't exist or has been removed."
        showClearButton
      />
    )
  }

  const installCmd = `agent-team role-repo add ${role.source_owner}/${role.source_repo} --role ${role.role_name}`

  return (
    <div className="flex flex-col gap-10 px-6 lg:px-20 py-10">
      <nav className="flex items-center gap-3 text-sm font-ui">
        <Link to="/" className="text-text-sub no-underline hover:text-text-main transition-colors">
          Home
        </Link>
        <ChevronRight size={14} className="text-text-sub" />
        <Link to="/" className="text-text-sub no-underline hover:text-text-main transition-colors">
          Roles
        </Link>
        <ChevronRight size={14} className="text-text-sub" />
        <span className="text-primary font-semibold">{role.role_name}</span>
      </nav>

      <div className="flex flex-col gap-6">
        <div className="flex items-center gap-3">
          <h1 className="text-3xl lg:text-[40px] font-extrabold text-text-main font-ui m-0">{role.role_name}</h1>
          {role.status === 'verified' && <CheckCircle2 size={24} className="text-verified" />}
        </div>

        <div className="flex items-center justify-between bg-surface border border-border rounded-lg px-5 py-4 max-w-150">
          <code className="text-primary font-code text-sm lg:text-base">{installCmd}</code>
          <CopyButton text={installCmd} label="" className="ml-4" />
        </div>
      </div>

      <div className="flex flex-col lg:flex-row gap-12 lg:gap-20">
        <div className="flex-1 min-w-0 flex flex-col gap-5">
          <h2 className="text-2xl font-bold text-text-main font-ui m-0">README.md</h2>
          <div className="text-base text-text-sub font-ui leading-relaxed whitespace-pre-wrap">
            {role.readme ?? role.description}
          </div>
        </div>

        <div className="w-full lg:w-80 flex-shrink-0">
          <div className="border border-border rounded-xl p-6 flex flex-col gap-4">
            <h4 className="text-xs font-bold text-text-sub tracking-widest uppercase font-ui m-0">Metadata</h4>

            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded-full bg-surface border border-border" />
              <span className="text-sm font-semibold text-text-main font-ui">{role.source_owner}</span>
            </div>

            <div className="flex flex-col gap-2">
              <span className="text-xs text-text-sub font-ui">Repository</span>
              <Link
                to={`/repos/${encodeURIComponent(role.source_owner)}/${encodeURIComponent(role.source_repo)}`}
                className="text-sm font-medium text-primary no-underline hover:underline font-ui"
              >
                {role.source_owner}/{role.source_repo}
              </Link>
            </div>

            <div className="flex gap-6">
              <div className="flex flex-col gap-1">
                <span className="text-xs text-text-sub font-ui">Installs</span>
                <span className="text-sm font-semibold text-text-main font-ui">{formatCount(role.install_count)}</span>
              </div>
            </div>

            {role.tags.length > 0 && (
              <div className="flex flex-wrap gap-2 pt-2">
                {role.tags.map((tag) => (
                  <Tag key={tag} label={tag} />
                ))}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

function formatCount(n: number): string {
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`
  return String(n)
}
