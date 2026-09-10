/**
 * Application providers.
 *
 * Wraps the application with:
 * - MantineProvider (Design System + theme)
 * - QueryClientProvider (TanStack Query for server state)
 * - Notifications (Mantine notifications for toasts)
 * - AuthProvider (authentication state)
 *
 * See 04_FE_SPEC.md section 9 (Frontend Architecture) and
 * section 15 (Server State).
 */

import { Notifications } from '@mantine/notifications'
import { QueryClientProvider } from '@tanstack/react-query'
import { type ReactNode, useMemo } from 'react'
import { MantineProvider } from '@mantine/core'
import { pokerClubTheme } from '../styles/theme'
import { AuthProvider } from '../auth'
import { TelegramEnvironmentProvider } from '../telegram'
import { createQueryClient } from './queryClient'

export interface AppProvidersProps {
  children: ReactNode
}

/**
 * Root providers component.
 *
 * Wraps children with all necessary context providers.
 * The color scheme (light/dark) is determined by:
 * 1. Telegram theme (if in Telegram Mini App mode)
 * 2. System preference (prefers-color-scheme)
 *
 * See 05_FE_UX.md section 18 (Telegram Theme) and
 * 04_FE_SPEC.md section 22 (Telegram Theme).
 */
export function AppProviders({ children }: AppProvidersProps) {
  const queryClient = useMemo(() => createQueryClient(), [])

  return (
    <QueryClientProvider client={queryClient}>
      <MantineProvider
        theme={pokerClubTheme}
        defaultColorScheme="light"
        withCssVariables
      >
        <Notifications
          position="top-right"
          zIndex={1000}
          limit={5}
          autoClose={5000}
        />
        <TelegramEnvironmentProvider>
          <AuthProvider>{children}</AuthProvider>
        </TelegramEnvironmentProvider>
      </MantineProvider>
    </QueryClientProvider>
  )
}
