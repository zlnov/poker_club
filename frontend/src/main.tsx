/**
 * Application entry point.
 *
 * Initializes the React application with:
 * - QueryClientProvider (TanStack Query for server state)
 * - RouterProvider (React Router for client-side routing)
 *
 * See 04_FE_SPEC.md section 9 (Frontend Architecture).
 */

import { createRoot } from 'react-dom/client'
import { RouterProvider } from 'react-router-dom'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { router } from './app/router'
import './styles/index.css'

/**
 * Singleton QueryClient instance.
 *
 * Configured with sensible defaults for the Poker Club Frontend.
 * See 04_FE_SPEC.md section 15 (Server State).
 */
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      // Data is considered fresh for 30 seconds
      staleTime: 30_000,
      // Retry failed queries once
      retry: 1,
      // Refetch on window focus for fresh data
      refetchOnWindowFocus: true,
    },
    mutations: {
      // Don't retry mutations by default — they have side effects
      retry: false,
    },
  },
})

createRoot(document.getElementById('root')!).render(
  <QueryClientProvider client={queryClient}>
    <RouterProvider router={router} />
  </QueryClientProvider>,
)
