import type { ReactNode } from 'react'

interface SkeletonProps {
  className?: string
}

export function Skeleton({ className = '' }: SkeletonProps) {
  return <div className={`bg-surface rounded-lg animate-pulse ${className}`} />
}

export function RoleCardSkeleton() {
  return (
    <div className="card flex flex-col gap-4">
      <div className="flex items-center justify-between">
        <Skeleton className="h-5 w-36" />
        <Skeleton className="h-5 w-5 rounded-full" />
      </div>
      <Skeleton className="h-10 w-full" />
      <div className="flex gap-2">
        <Skeleton className="h-6 w-16 rounded-full" />
        <Skeleton className="h-6 w-20 rounded-full" />
      </div>
      <div className="flex items-center justify-between pt-3 border-t border-border">
        <Skeleton className="h-4 w-12" />
        <Skeleton className="h-4 w-20" />
      </div>
    </div>
  )
}

export function PageSkeleton({ children }: { children?: ReactNode }) {
  return (
    <div className="animate-pulse">
      {children ?? (
        <div className="flex flex-col gap-6 p-8">
          <Skeleton className="h-8 w-64" />
          <Skeleton className="h-4 w-96" />
          <Skeleton className="h-48 w-full" />
        </div>
      )}
    </div>
  )
}
