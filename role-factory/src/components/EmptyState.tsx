import { Search } from 'lucide-react'
import { Link } from '@tanstack/react-router'
import type { ReactNode } from 'react'

interface EmptyStateProps {
  title?: string
  description?: string
  showClearButton?: boolean
  onClear?: () => void
  actionLabel?: string
  onAction?: () => void
  icon?: ReactNode
}

export function EmptyState({
  title = 'No roles found',
  description = "Try adjusting your search or filters to find what you're looking for.",
  showClearButton = true,
  onClear,
  actionLabel,
  onAction,
  icon,
}: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-6 py-24">
      {icon ?? <Search size={64} className="text-text-sub" />}
      <h2 className="text-2xl font-bold text-text-main font-ui m-0">{title}</h2>
      <p className="text-base text-text-sub font-ui text-center m-0">{description}</p>
      {actionLabel && onAction ? (
        <button onClick={onAction} className="btn-primary text-sm">
          {actionLabel}
        </button>
      ) : null}
      {showClearButton && onClear && !actionLabel && (
        <button onClick={onClear} className="btn-primary text-sm">
          Clear all filters
        </button>
      )}
      {showClearButton && !onClear && !actionLabel && (
        <Link to="/" className="btn-primary text-sm no-underline">
          Back to Home
        </Link>
      )}
    </div>
  )
}
