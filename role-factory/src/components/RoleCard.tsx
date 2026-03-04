import { CheckCircle2 } from 'lucide-react'
import { Link } from '@tanstack/react-router'
import { CopyButton } from './CopyButton'
import { Tag } from './Tag'
import type { RoleRecord } from '@/types'

interface RoleCardProps {
  role: RoleRecord
}

export function RoleCard({ role }: RoleCardProps) {
  return (
    <Link
      to="/roles/$name"
      params={{ name: role.role_name }}
      className="card flex flex-col gap-4 no-underline hover:-translate-y-0.5 transition-all duration-200"
    >
      <div className="flex items-center justify-between w-full">
        <h3 className="text-lg font-semibold text-text-main font-ui m-0">{role.role_name}</h3>
        {role.status === 'verified' && <CheckCircle2 size={18} className="text-verified flex-shrink-0" />}
      </div>

      <p className="text-sm text-text-sub font-ui leading-relaxed m-0 line-clamp-2">{role.description}</p>

      {role.tags.length > 0 && (
        <div className="flex flex-wrap gap-2">
          {role.tags.slice(0, 3).map((tag) => (
            <Tag key={tag} label={tag} />
          ))}
        </div>
      )}

      <div className="flex items-center justify-between w-full pt-3 border-t border-border">
        <span className="text-xs text-text-sub font-ui">v1.0.2</span>
        <CopyButton text={`agent-team role install ${role.role_name}`} />
      </div>
    </Link>
  )
}
