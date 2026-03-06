import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import { useLoaderData, useNavigation, useRevalidator } from '@remix-run/react'
import { AlertTriangle } from 'lucide-react'
import { RoleCard } from '~/components/RoleCard'
import { PageSkeleton, Skeleton, RoleCardSkeleton } from '~/components/Skeleton'
import { EmptyState } from '~/components/EmptyState'
import { getRepo } from '~/server/roles.server'
import { getErrorMessage } from '~/utils/errors'
import { buildCanonical, buildMeta } from '~/utils/seo'
import type { Repository } from '~/types'

interface LoaderData {
  repository: Repository | null
  error?: string
}

export async function loader({ params }: LoaderFunctionArgs) {
  const owner = params.owner
  const repo = params.repo
  if (!owner || !repo) {
    return json<LoaderData>({ repository: null, error: 'Repository owner and name are required.' }, { status: 400 })
  }

  try {
    const repository = await getRepo(owner, repo)
    return json<LoaderData>({ repository })
  } catch (error) {
    return json<LoaderData>({ repository: null, error: getErrorMessage(error) }, { status: 500 })
  }
}

export const meta: MetaFunction<typeof loader> = ({ data, params }) => {
  const owner = params.owner ?? ''
  const repo = params.repo ?? ''
  if (!data?.repository) {
    return [{ title: `${owner}/${repo} | Role Hub` }]
  }
  const { repository } = data
  const canonical = buildCanonical(`/repos/${repository.owner}/${repository.repo}`)
  const description =
    repository.description || `Browse AI roles published by ${repository.owner}/${repository.repo} on Role Hub.`
  return buildMeta({ title: `${repository.owner}/${repository.repo}`, description, canonical })
}

export default function RepoDetailPage() {
  const { repository, error } = useLoaderData<typeof loader>()
  const navigation = useNavigation()
  const revalidator = useRevalidator()

  if (navigation.state !== 'idle') {
    return (
      <PageSkeleton>
        <div className="flex flex-col gap-10 px-6 lg:px-20 py-15">
          <div className="flex flex-col gap-3">
            <Skeleton className="h-8 w-64" />
            <Skeleton className="h-4 w-96" />
          </div>
          <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
            {Array.from({ length: 3 }).map((_, i) => (
              <RoleCardSkeleton key={i} />
            ))}
          </div>
        </div>
      </PageSkeleton>
    )
  }

  if (error) {
    return (
      <EmptyState
        title="Unable to load repository"
        description={error}
        showClearButton={false}
        actionLabel={revalidator.state !== 'idle' ? 'Retrying...' : 'Retry'}
        onAction={() => revalidator.revalidate()}
        icon={<AlertTriangle size={64} className="text-warning" />}
      />
    )
  }

  if (!repository) {
    return (
      <EmptyState
        title="Repository not found"
        description="The repository you're looking for doesn't exist or hasn't been indexed yet."
        showClearButton
      />
    )
  }

  return (
    <div className="flex flex-col gap-10 px-6 lg:px-20 py-15">
      <div className="flex flex-col lg:flex-row items-start lg:items-end justify-between gap-6">
        <div className="flex flex-col gap-3">
          <h1 className="text-3xl lg:text-[32px] font-extrabold text-text-main font-ui m-0">
            {repository.owner}/{repository.repo}
          </h1>
          <p className="text-base text-text-sub font-ui m-0">{repository.description}</p>
        </div>

        <div className="flex items-center gap-4">
          <a
            href={`https://github.com/${repository.owner}/${repository.repo}`}
            target="_blank"
            rel="noopener noreferrer"
            className="btn-secondary text-sm no-underline"
          >
            View on GitHub
          </a>
          <div className="flex flex-col items-end gap-1">
            <span className="text-xs text-text-sub font-ui">
              Last synced: {formatRelativeTime(repository.last_synced_at)}
            </span>
            <div className="flex items-center gap-1.5">
              <div
                className={`w-2 h-2 rounded-full ${
                  repository.sync_status === 'healthy'
                    ? 'bg-verified'
                    : repository.sync_status === 'syncing'
                      ? 'bg-warning'
                      : 'bg-invalid'
                }`}
              />
              <span
                className={`text-xs font-semibold font-ui ${
                  repository.sync_status === 'healthy'
                    ? 'text-verified'
                    : repository.sync_status === 'syncing'
                      ? 'text-warning'
                      : 'text-invalid'
                }`}
              >
                {repository.sync_status.charAt(0).toUpperCase() + repository.sync_status.slice(1)}
              </span>
            </div>
          </div>
        </div>
      </div>

      {repository.roles.length === 0 ? (
        <EmptyState title="No roles yet" description="This repository hasn't published any roles yet." showClearButton={false} />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
          {repository.roles.map((role) => (
            <RoleCard key={role.id} role={role} />
          ))}
        </div>
      )}
    </div>
  )
}

function formatRelativeTime(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime()
  const hours = Math.floor(diff / (1000 * 60 * 60))
  if (hours < 1) return 'just now'
  if (hours < 24) return `${hours} hours ago`
  const days = Math.floor(hours / 24)
  return `${days} days ago`
}
