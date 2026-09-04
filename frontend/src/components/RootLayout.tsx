/**
 * Root layout component for the application.
 *
 * Provides the application shell with:
 * - Mantine AppShell (header + sidebar navigation)
 * - Telegram environment detection
 * - Theme application (Telegram theme or default)
 * - Safe area handling for Telegram Mini App
 *
 * See 04_FE_SPEC.md section 6 (Telegram Integration Layer) and
 * section 7 (Telegram and Standard Web compatibility).
 */

import { useEffect } from 'react'
import { Outlet } from 'react-router-dom'
import {
  AppShell,
  AppShellHeader,
  AppShellNavbar,
  AppShellMain,
  Group,
  Title,
} from '@mantine/core'
import { useTelegramEnvironment } from '../hooks'
import { applyTelegramThemeVariables } from '../styles'
import { Navigation } from './Navigation'
import { ThemeToggle } from './ThemeToggle'
import { useTelegramAuth } from '../auth/useTelegramAuth'

/**
 * Root layout component.
 *
 * Wraps all child routes with the application shell.
 * Handles Telegram environment initialization and theme application.
 * Telegram authentication is handled by useTelegramAuth hook.
 */
export function RootLayout() {
  const telegramEnv = useTelegramEnvironment()
  // Initialize Telegram authentication (hook handles deduplication)
  useTelegramAuth()

  useEffect(() => {
    // Apply Telegram theme if available, otherwise use default theme
    if (telegramEnv.isTelegram && telegramEnv.webApp?.theme) {
      applyTelegramThemeVariables({
        bgColor: telegramEnv.webApp.theme.bgColor,
        textColor: telegramEnv.webApp.theme.textColor,
      })
    } else {
      // Standard Web mode — use default theme
      applyTelegramThemeVariables(undefined)
    }
  }, [telegramEnv])

  return (
    <AppShell
      header={{ height: 64 }}
      navbar={{
        width: 260,
        breakpoint: 'sm',
        collapsed: { mobile: true },
      }}
      padding={0}
    >
      <AppShellHeader>
        <Group h="100%" px="md" justify="space-between" align="center">
          <Title order={3} fw={600}>
            Poker Club
          </Title>
          <ThemeToggle />
        </Group>
      </AppShellHeader>

      <AppShellNavbar>
        <Navigation />
      </AppShellNavbar>

      <AppShellMain>
        <Outlet />
      </AppShellMain>
    </AppShell>
  )
}
