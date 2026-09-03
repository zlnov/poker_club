/**
 * Query client factory.
 *
 * Creates a singleton QueryClient instance with sensible defaults.
 *
 * See 04_FE_SPEC.md section 15 (Server State).
 */

import { QueryClient } from '@tanstack/react-query'

/**
 * Creates a QueryClient instance with sensible defaults for the Poker Club Frontend.
 * See 04_FE_SPEC.md section 15 (Server State).
 */
export function createQueryClient(): QueryClient {
  return new QueryClient({
    defaultOptions: {
      queries: {
        staleTime: 30_000,
        retry: 1,
        refetchOnWindowFocus: true,
      },
      mutations: {
        retry: false,
      },
    },
  })
}
