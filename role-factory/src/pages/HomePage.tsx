import { useState, useMemo, useCallback } from 'react'
import { SearchBar } from '@/components/SearchBar'
import { RoleCard } from '@/components/RoleCard'
import { Sidebar } from '@/components/Sidebar'
import { EmptyState } from '@/components/EmptyState'
import { RoleCardSkeleton } from '@/components/Skeleton'
import { useRoles } from '@/hooks/useQueries'

export function HomePage() {
  const [search, setSearch] = useState('')
  const [category, setCategory] = useState('All Roles')
  const [frameworks, setFrameworks] = useState<string[]>([])
  const [sort, setSort] = useState<'relevance' | 'installs' | 'newest'>('relevance')

  const debouncedSearch = useDebounce(search, 300)

  const { data, isLoading } = useRoles({
    search: debouncedSearch,
    category,
    sort,
  })

  const handleFrameworkToggle = useCallback((fw: string) => {
    setFrameworks((prev) => (prev.includes(fw) ? prev.filter((f) => f !== fw) : [...prev, fw]))
  }, [])

  const handleClearFilters = useCallback(() => {
    setSearch('')
    setCategory('All Roles')
    setFrameworks([])
    setSort('relevance')
  }, [])

  const roles = useMemo(() => data?.data ?? [], [data])

  return (
    <div className="flex flex-col w-full">
      {/* Hero Section */}
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

      {/* Grid Area */}
      <section className="flex gap-15 px-6 lg:px-20 pb-20">
        {/* Sidebar - hidden on mobile */}
        <div className="hidden lg:block">
          <Sidebar
            selectedCategory={category}
            onCategoryChange={setCategory}
            selectedFrameworks={frameworks}
            onFrameworkToggle={handleFrameworkToggle}
          />
        </div>

        {/* Main Grid */}
        <div className="flex flex-col gap-6 flex-1 min-w-0">
          {/* Sort Controls */}
          <div className="flex items-center justify-between">
            <span className="text-sm text-text-sub font-ui">
              {data ? `${data.total} roles` : 'Loading...'}
            </span>
            <select
              value={sort}
              onChange={(e) => setSort(e.target.value as typeof sort)}
              className="bg-surface border border-border rounded-lg px-3 py-2 text-sm text-text-main font-ui outline-none cursor-pointer"
            >
              <option value="relevance">Relevance</option>
              <option value="installs">Most Installs</option>
              <option value="newest">Newest</option>
            </select>
          </div>

          {/* Role Cards Grid */}
          {isLoading ? (
            <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-6">
              {Array.from({ length: 6 }).map((_, i) => (
                <RoleCardSkeleton key={i} />
              ))}
            </div>
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
  useMemo(() => {
    const timer = setTimeout(() => setDebounced(value), delay)
    return () => clearTimeout(timer)
  }, [value, delay])
  return debounced
}
