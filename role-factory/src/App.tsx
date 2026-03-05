import { RouterProvider } from '@tanstack/react-router'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { router } from './router'
import { ApiError } from '@/api/client'

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5,
      retry: (failureCount: number, error: unknown) => {
        if (error instanceof ApiError) {
          if (error.status === 0 || error.status === 408 || error.status === 429 || error.status >= 500) {
            return failureCount < 2
          }
          return false
        }
        return failureCount < 2
      },
      retryDelay: (attempt: number) => Math.min(1000 * 2 ** attempt, 8000),
    },
  },
})

export default function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  )
}
