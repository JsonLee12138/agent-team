import { CheckCircle2, Download } from 'lucide-react'
import { Link } from '@remix-run/react'
import { CopyButton } from './CopyButton'
import { Tag } from './Tag'
import type { RoleRecord } from '~/types'

interface RoleCardProps {
  role: RoleRecord
}

export function RoleCard({ role }: RoleCardProps) {
  const rolePath = `/roles/${encodeURIComponent(role.role_name)}`

  return (
    <Link
      to={rolePath}
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
        <span className="text-xs text-text-sub font-ui flex items-center gap-1">
          <Download size={12} />
          {role.install_count >= 1000 ? `${(role.install_count / 1000).toFixed(1)}k` : role.install_count}
        </span>
        <CopyButton text={`agent-team role-repo add ${role.source_owner}/${role.source_repo} --role ${role.role_name}`} />
      </div>
    </Link>
  )
}
