import { Search } from 'lucide-react'
import { Link } from '@tanstack/react-router'

interface EmptyStateProps {
  title?: string
  description?: string
  showClearButton?: boolean
  onClear?: () => void
}

export function EmptyState({
  title = 'No roles found',
  description = "Try adjusting your search or filters to find what you're looking for.",
  showClearButton = true,
  onClear,
}: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-6 py-24">
      <Search size={64} className="text-text-sub" />
      <h2 className="text-2xl font-bold text-text-main font-ui m-0">{title}</h2>
      <p className="text-base text-text-sub font-ui text-center m-0">{description}</p>
      {showClearButton && onClear && (
        <button onClick={onClear} className="btn-primary text-sm">
          Clear all filters
        </button>
      )}
      {showClearButton && !onClear && (
        <Link to="/" className="btn-primary text-sm no-underline">
          Back to Home
        </Link>
      )}
    </div>
  )
}
