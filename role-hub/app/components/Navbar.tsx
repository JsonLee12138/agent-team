import { Zap } from 'lucide-react'
import { Link } from '@remix-run/react'

export function Navbar() {
  return (
    <header className="flex items-center justify-between h-20 px-8 lg:px-20 border-b border-border w-full">
      <Link to="/" className="flex items-center gap-2 no-underline">
        <Zap size={24} className="text-primary" />
        <span className="text-xl font-extrabold text-text-main font-ui">Role Hub</span>
      </Link>
      <nav className="hidden md:flex items-center gap-8">
        <Link to="/" className="text-sm font-medium text-text-main no-underline hover:text-primary transition-colors font-ui">
          Directory
        </Link>
        <a
          href="https://github.com/agent-team"
          target="_blank"
          rel="noopener noreferrer"
          className="text-sm font-medium text-text-sub no-underline hover:text-primary transition-colors font-ui"
        >
          GitHub
        </a>
        <a
          href="#"
          className="text-sm font-medium text-text-sub no-underline hover:text-primary transition-colors font-ui"
        >
          Docs
        </a>
      </nav>
    </header>
  )
}
