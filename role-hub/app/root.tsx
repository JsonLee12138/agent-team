import type { LinksFunction, MetaFunction } from '@remix-run/node'
import {
  Links,
  Meta,
  Outlet,
  Scripts,
  ScrollRestoration,
  LiveReload,
  useRouteError,
  isRouteErrorResponse,
} from '@remix-run/react'
import { Navbar } from '~/components/Navbar'
import { EmptyState } from '~/components/EmptyState'
import globalStyles from '~/styles/global.css'
import unoStyles from '~/styles/uno.css'

export const meta: MetaFunction = () => [
  { title: 'Role Hub' },
  { name: 'description', content: 'Discover, install, and share AI roles for agent-team.' },
]

export const links: LinksFunction = () => [
  { rel: 'stylesheet', href: globalStyles },
  { rel: 'stylesheet', href: unoStyles },
]

export function Layout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <Meta />
        <Links />
      </head>
      <body>
        {children}
        <ScrollRestoration />
        <Scripts />
        <LiveReload />
      </body>
    </html>
  )
}

export default function App() {
  return (
    <div className="min-h-screen bg-bg font-ui">
      <Navbar />
      <main>
        <Outlet />
      </main>
    </div>
  )
}

export function ErrorBoundary() {
  const error = useRouteError()
  let title = 'Something went wrong'
  let description = 'An unexpected error occurred.'

  if (isRouteErrorResponse(error)) {
    title = `${error.status} ${error.statusText}`
    description = error.data?.message || 'An unexpected error occurred.'
  } else if (error instanceof Error) {
    description = error.message
  }

  return (
    <Layout>
      <div className="min-h-screen bg-bg font-ui">
        <Navbar />
        <EmptyState title={title} description={description} showClearButton={false} />
      </div>
    </Layout>
  )
}
