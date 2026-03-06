import type { LoaderFunctionArgs, MetaFunction } from '@remix-run/node'
import { json } from '@remix-run/node'
import {
  useLoaderData,
  useNavigation,
  useSearchParams,
  useRevalidator,
} from '@remix-run/react'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { AlertTriangle } from 'lucide-react'
import { SearchBar } from '~/components/SearchBar'
import { RoleCard } from '~/components/RoleCard'
import { Sidebar } from '~/components/Sidebar'
import { EmptyState } from '~/components/EmptyState'
import { RoleCardSkeleton } from '~/components/Skeleton'
import { getRoles } from '~/server/roles.server'
import { getErrorMessage } from '~/utils/errors'
import { buildCanonical, buildJsonLd, buildMeta } from '~/utils/seo'
import type { PaginatedResponse, RoleRecord, RolesQueryParams } from '~/types'

interface LoaderData {
  data: PaginatedResponse<RoleRecord>
  error?: string
  params: RolesQueryParams
}

const emptyData: PaginatedResponse<RoleRecord> = {
  data: [],
  total: 0,
  page: 1,
  limit: 12,
  hasMore: false,
}

export const meta: MetaFunction = () => {
  const canonical = buildCanonical('/')
  const description = 'Discover, install, and share AI roles for agent-team. Browse 2,400+ community-built roles.'
  return [
    ...buildMeta({ title: 'Role Hub', description, canonical }),
    {
      'script:ld+json': buildJsonLd({
        '@context': 'https://schema.org',
        '@type': 'WebSite',
        name: 'Role Hub',
        url: canonical,
        description,
        potentialAction: {
          '@type': 'SearchAction',
          target: {
            '@type': 'EntryPoint',
            urlTemplate: `${canonical}?search={search_term_string}`,
          },
          'query-input': 'required name=search_term_string',
        },
      }),
    },
  ]
}

export async function loader({ request }: LoaderFunctionArgs) {
  const url = new URL(request.url)
  const params: RolesQueryParams = {
    search: url.searchParams.get('search') ?? undefined,
    category: url.searchParams.get('category') ?? undefined,
    framework: url.searchParams.getAll('framework'),
    sort: (url.searchParams.get('sort') as RolesQueryParams['sort']) ?? undefined,
    page: url.searchParams.get('page') ? Number(url.searchParams.get('page')) : undefined,
    limit: url.searchParams.get('limit') ? Number(url.searchParams.get('limit')) : undefined,
  }

  try {
    const data = await getRoles(params)
    return json<LoaderData>({ data, params })
  } catch (error) {
    return json<LoaderData>(
      {
        data: emptyData,
        error: getErrorMessage(error),
        params,
      },
      { status: 500 },
    )
  }
}

export default function HomePage() {
  const { data, error, params } = useLoaderData<typeof loader>()
  const [searchParams, setSearchParams] = useSearchParams()
  const navigation = useNavigation()
  const revalidator = useRevalidator()

  const [search, setSearch] = useState(params.search ?? '')
  const [category, setCategory] = useState(params.category ?? 'All Roles')
  const [frameworks, setFrameworks] = useState<string[]>(
    Array.isArray(params.framework) ? params.framework : params.framework ? [params.framework] : [],
  )
  const [sort, setSort] = useState<RolesQueryParams['sort']>(params.sort ?? 'relevance')

  const debouncedSearch = useDebounce(search, 300)

  useEffect(() => {
    setSearch(params.search ?? '')
    setCategory(params.category ?? 'All Roles')
    setFrameworks(Array.isArray(params.framework) ? params.framework : params.framework ? [params.framework] : [])
    setSort(params.sort ?? 'relevance')
  }, [params.search, params.category, params.framework, params.sort])

  useEffect(() => {
    const nextParams = new URLSearchParams()
    if (debouncedSearch) nextParams.set('search', debouncedSearch)
    if (category && category !== 'All Roles') nextParams.set('category', category)
    frameworks.forEach((fw) => nextParams.append('framework', fw))
    if (sort && sort !== 'relevance') nextParams.set('sort', sort)
    if (nextParams.toString() !== searchParams.toString()) {
      setSearchParams(nextParams, { replace: true })
    }
  }, [debouncedSearch, category, frameworks, sort, searchParams, setSearchParams])

  const handleFrameworkToggle = useCallback((fw: string) => {
    setFrameworks((prev) => (prev.includes(fw) ? prev.filter((f) => f !== fw) : [...prev, fw]))
  }, [])

  const handleClearFilters = useCallback(() => {
    setSearch('')
    setCategory('All Roles')
    setFrameworks([])
    setSort('relevance')
    setSearchParams(new URLSearchParams(), { replace: true })
  }, [setSearchParams])

  const isLoading = navigation.state !== 'idle'
  const isFetching = revalidator.state !== 'idle'
  const roles = useMemo(() => data?.data ?? [], [data])

  return (
    <div className="flex flex-col w-full">
      <section className="flex flex-col items-center gap-6 py-16 lg:py-20 px-6">
        <h1 className="text-4xl lg:text-[56px] font-extrabold text-text-main font-ui text-center leading-tight m-0">
          Find the perfect role for your Agent
        </h1>
        <p className="text-base lg:text-lg text-text-sub font-ui text-center max-w-150 m-0">
          Discover, install, and share AI roles for agent-team. Powered by open-source community.
        </p>
        <SearchBar
          value={search}
          onChange={setSearch}
          placeholder="Search 2,400+ roles..."
          size="large"
          className="w-full max-w-150"
        />
      </section>

      <section className="flex gap-15 px-6 lg:px-20 pb-20">
        <div className="hidden lg:block">
          <Sidebar
            selectedCategory={category}
            onCategoryChange={setCategory}
            selectedFrameworks={frameworks}
            onFrameworkToggle={handleFrameworkToggle}
          />
        </div>

        <div className="flex flex-col gap-6 flex-1 min-w-0">
          <div className="flex items-center justify-between">
            <span className="text-sm text-text-sub font-ui">
              {error ? 'Failed to load roles' : data ? `${data.total} roles` : 'Loading...'}
            </span>
            <select
              value={sort}
              onChange={(e) => setSort(e.target.value as RolesQueryParams['sort'])}
              className="bg-surface border border-border rounded-lg px-3 py-2 text-sm text-text-main font-ui outline-none cursor-pointer"
            >
              <option value="relevance">Relevance</option>
              <option value="installs">Most Installs</option>
              <option value="newest">Newest</option>
            </select>
          </div>

          {isLoading ? (
            <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
              {Array.from({ length: 6 }).map((_, i) => (
                <RoleCardSkeleton key={i} />
              ))}
            </div>
          ) : error ? (
            <EmptyState
              title="Unable to load roles"
              description={error}
              showClearButton={false}
              actionLabel={isFetching ? 'Retrying...' : 'Retry'}
              onAction={() => revalidator.revalidate()}
              icon={<AlertTriangle size={64} className="text-warning" />}
            />
          ) : roles.length === 0 ? (
            <EmptyState onClear={handleClearFilters} />
          ) : (
            <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
              {roles.map((role) => (
                <RoleCard key={role.id} role={role} />
              ))}
            </div>
          )}
        </div>
      </section>
    </div>
  )
}

function useDebounce(value: string, delay: number) {
  const [debounced, setDebounced] = useState(value)
  useEffect(() => {
    const timer = setTimeout(() => setDebounced(value), delay)
    return () => clearTimeout(timer)
  }, [value, delay])
  return debounced
}
