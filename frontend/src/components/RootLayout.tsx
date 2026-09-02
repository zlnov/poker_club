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
} from '@mantine/core'
import { useTelegramEnvironment } from '../hooks'
import { applyTelegramThemeVariables } from '../styles'
import { Navigation } from './Navigation'

/**
 * Root layout component.
 *
 * Wraps all child routes with the application shell.
 * Handles Telegram environment initialization and theme application.
 */
export function RootLayout() {
  const telegramEnv = useTelegramEnvironment()

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
      header={{ height: 60 }}
      navbar={{
        width: 250,
        breakpoint: 'sm',
        collapsed: { mobile: true },
      }}
    >
      <AppShellHeader>
        <div
          style={{
            height: '100%',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            padding: '0 var(--mantine-spacing-md)',
          }}
        >
          <span style={{ fontWeight: 500, fontSize: '1.125rem' }}>
            Poker Club
          </span>
        </div>
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
