import {
  createRouter,
  createRoute,
  createRootRoute,
  Outlet,
} from '@tanstack/react-router'
import { Navbar } from '@/components/Navbar'
import { HomePage } from '@/pages/HomePage'
import { RoleDetailPage } from '@/pages/RoleDetailPage'
import { RepoDetailPage } from '@/pages/RepoDetailPage'

const rootRoute = createRootRoute({
  component: () => (
    <div className="min-h-screen bg-bg font-ui">
      <Navbar />
      <main>
        <Outlet />
      </main>
    </div>
  ),
})

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: HomePage,
})

const roleDetailRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/roles/$name',
  component: () => {
    const { name } = roleDetailRoute.useParams()
    return <RoleDetailPage name={name} />
  },
})

const repoDetailRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/repos/$owner/$repo',
  component: () => {
    const { owner, repo } = repoDetailRoute.useParams()
    return <RepoDetailPage owner={owner} repo={repo} />
  },
})

const routeTree = rootRoute.addChildren([
  indexRoute,
  roleDetailRoute,
  repoDetailRoute,
])

export const router = createRouter({ routeTree })

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}
