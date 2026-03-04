import { Copy, Check } from 'lucide-react'
import { useCopyToClipboard } from '@/hooks/useCopyToClipboard'

interface CopyButtonProps {
  text: string
  label?: string
  className?: string
}

export function CopyButton({ text, label = 'Copy Install', className = '' }: CopyButtonProps) {
  const { copied, copy } = useCopyToClipboard()

  return (
    <button
      onClick={(e) => {
        e.preventDefault()
        e.stopPropagation()
        copy(text)
      }}
      className={`flex items-center gap-1.5 text-primary text-xs font-medium cursor-pointer bg-transparent border-none hover:opacity-80 transition-opacity ${className}`}
    >
      {copied ? <Check size={14} /> : <Copy size={14} />}
      <span>{copied ? 'Copied!' : label}</span>
    </button>
  )
}
