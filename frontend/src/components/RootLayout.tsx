/**
 * Root layout component for the application.
 *
 * Provides the application shell with:
 * - Mantine AppShell (header + sidebar navigation)
 * - Telegram environment detection
 * - Theme application (Telegram theme or default)
 * - Safe area handling for Telegram Mini App
 * - Background video layer (persistent across navigation)
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
  Burger,
  Drawer,
} from '@mantine/core'
import { useDisclosure } from '@mantine/hooks'
import { useTelegramEnvironment } from '../hooks'
import { applyTelegramThemeVariables, backgroundSurfaces } from '../styles'
import { Navigation } from './Navigation'
// ColorScheme switcher - temporarily disabled, Dark mode is default
// import { ThemeToggle } from './ThemeToggle'
import { BackgroundVideo } from './BackgroundVideo'
import { useTelegramAuth } from '../auth/useTelegramAuth'

/**
 * Root layout component.
 *
 * Wraps all child routes with the application shell.
 * Handles Telegram environment initialization and theme application.
 * Telegram authentication is handled by useTelegramAuth hook.
 * BackgroundVideo is always rendered and persists across navigation.
 */
export function RootLayout() {
  const telegramEnv = useTelegramEnvironment()
  const [drawerOpened, { open: openDrawer, close: closeDrawer }] =
    useDisclosure(false)

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
    <>
      {/* Background video - persistent across all navigation, never unmounts */}
      <BackgroundVideo />

      {/* App shell wrapper - ensures all UI is above background video */}
      <div style={{ position: 'relative', zIndex: 1, height: '100vh' }}>
        <AppShell
          header={{ height: 64 }}
          navbar={{
            width: 260,
            breakpoint: 'sm',
            collapsed: { mobile: true },
          }}
          padding={0}
        >
          <AppShellHeader bg={backgroundSurfaces.header}>
            <Group h="100%" px="md" justify="space-between" align="center">
              <Group gap="sm">
                <Burger
                  opened={drawerOpened}
                  onClick={openDrawer}
                  size="sm"
                  hiddenFrom="sm"
                  aria-label="Open navigation"
                />
                <Title order={3} fw={600}>
                  Poker Club
                </Title>
              </Group>
              {/* ColorScheme switcher - temporarily disabled, Dark mode is default */}
              {/* <ThemeToggle /> */}
            </Group>
          </AppShellHeader>

          <AppShellNavbar bg={backgroundSurfaces.navbar}>
            <Navigation />
          </AppShellNavbar>

          <AppShellMain bg={backgroundSurfaces.main}>
            <Outlet />
          </AppShellMain>

          <Drawer
            opened={drawerOpened}
            onClose={closeDrawer}
            title="Navigation"
            styles={{
              content: { backgroundColor: backgroundSurfaces.drawer },
            }}
            size={260}
            padding="md"
            zIndex={1000}
            hiddenFrom="sm"
          >
            <Navigation onNavigate={closeDrawer} />
          </Drawer>
        </AppShell>
      </div>
    </>
  )
}
