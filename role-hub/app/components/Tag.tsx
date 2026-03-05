interface TagProps {
  label: string
  onClick?: () => void
}

export function Tag({ label, onClick }: TagProps) {
  return (
    <span
      onClick={onClick}
      className={`border border-border rounded-full px-3 py-1 text-xs text-text-main font-ui ${onClick ? 'cursor-pointer hover:border-primary/50 transition-colors' : ''}`}
    >
      {label}
    </span>
  )
}
