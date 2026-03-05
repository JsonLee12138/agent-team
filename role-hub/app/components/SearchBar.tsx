import { Search } from 'lucide-react'

interface SearchBarProps {
  value: string
  onChange: (value: string) => void
  placeholder?: string
  className?: string
  size?: 'default' | 'large'
}

export function SearchBar({
  value,
  onChange,
  placeholder = 'Search roles...',
  className = '',
  size = 'default',
}: SearchBarProps) {
  return (
    <div
      className={`flex items-center gap-3 bg-surface border border-border rounded-lg font-ui ${
        size === 'large' ? 'px-5 py-4 text-base' : 'px-4 py-3 text-sm'
      } ${className}`}
    >
      <Search size={size === 'large' ? 20 : 18} className="text-text-sub flex-shrink-0" />
      <input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="bg-transparent border-none outline-none text-text-main placeholder-text-sub w-full font-ui"
      />
    </div>
  )
}
