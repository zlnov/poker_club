/**
 * Root layout component for the application.
 *
 * Desktop: persistent left sidebar (no top header).
 * Mobile: header with logo drawer trigger + club switcher, horizontal nav,
 *         and a drawer that reuses the sidebar navigation.
 */

import { useEffect } from 'react'
import { Outlet, useLocation } from 'react-router-dom'
import {
  AppShell,
  AppShellHeader,
  AppShellNavbar,
  AppShellMain,
  Group,
  Title,
  Drawer,
  Image,
  UnstyledButton,
  Box,
} from '@mantine/core'
import { useDisclosure } from '@mantine/hooks'
import { useTelegramEnvironment } from '../hooks'
import { applyTelegramThemeVariables, backgroundSurfaces } from '../styles'
import { Navigation } from './Navigation'
import { ClubSwitcher } from './ClubSwitcher'
import { MobileHorizontalNav } from './MobileHorizontalNav'
import { LastClubSync } from './LastClubSync'
import { BackgroundVideo } from './BackgroundVideo'
import { useTelegramAuth } from '../auth/useTelegramAuth'
import { navAccent } from './navTheme'

/**
 * Root layout component.
 */
export function RootLayout() {
  const location = useLocation()
  const telegramEnv = useTelegramEnvironment()
  const [drawerOpened, { open: openDrawer, close: closeDrawer }] =
    useDisclosure(false)
  const isLanding = location.pathname === '/'
  const hideAppNav =
    location.pathname === '/login' || location.pathname === '/register'

  useTelegramAuth()

  useEffect(() => {
    if (telegramEnv.isTelegram && telegramEnv.webApp?.theme) {
      applyTelegramThemeVariables({
        bgColor: telegramEnv.webApp.theme.bgColor,
        textColor: telegramEnv.webApp.theme.textColor,
      })
    } else {
      applyTelegramThemeVariables(undefined)
    }
  }, [telegramEnv])

  // Close drawer on route change
  useEffect(() => {
    closeDrawer()
  }, [location.pathname, closeDrawer])

  return (
    <>
      <BackgroundVideo />

      {isLanding ? (
        <div style={{ position: 'relative', zIndex: 1 }}>
          <Outlet />
        </div>
      ) : (
        <div style={{ position: 'relative', zIndex: 1, height: '100vh' }}>
          {!hideAppNav && <LastClubSync />}
          <AppShell
            header={{ height: { base: 56, sm: 0 } }}
            navbar={{
              width: 272,
              breakpoint: 'sm',
              collapsed: { mobile: true, desktop: hideAppNav },
            }}
            padding={0}
          >
            {!hideAppNav && (
              <AppShellHeader
                bg={backgroundSurfaces.header}
                hiddenFrom="sm"
                style={{ borderBottom: `1px solid ${navAccent.border}` }}
              >
                <Group h="100%" px="sm" justify="space-between" wrap="nowrap">
                  <Group gap="xs" wrap="nowrap" style={{ minWidth: 0 }}>
                    <UnstyledButton
                      onClick={openDrawer}
                      aria-label="Open navigation"
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        width: 36,
                        height: 36,
                        borderRadius: 8,
                      }}
                    >
                      <Image
                        src="/landing/logo/logo.png"
                        alt="Poker Club"
                        w={28}
                        h={28}
                        fit="contain"
                      />
                    </UnstyledButton>
                    <Title order={4} fw={600} style={{ whiteSpace: 'nowrap' }}>
                      Poker Club
                    </Title>
                  </Group>
                  <Box style={{ maxWidth: 160, minWidth: 0, flexShrink: 1 }}>
                    <ClubSwitcher compact />
                  </Box>
                </Group>
              </AppShellHeader>
            )}

            {!hideAppNav && (
              <AppShellNavbar
                bg={navAccent.panel}
                visibleFrom="sm"
                style={{ borderRight: `1px solid ${navAccent.border}` }}
              >
                <Navigation />
              </AppShellNavbar>
            )}

            <AppShellMain bg={backgroundSurfaces.main}>
              {!hideAppNav && <MobileHorizontalNav />}
              <Outlet />
            </AppShellMain>

            {!hideAppNav && (
              <Drawer
                opened={drawerOpened}
                onClose={closeDrawer}
                title={null}
                withCloseButton={false}
                styles={{
                  content: { backgroundColor: navAccent.panel },
                  body: { padding: 0, height: '100%' },
                  header: { display: 'none' },
                }}
                size={280}
                padding={0}
                zIndex={1000}
                hiddenFrom="sm"
              >
                <Navigation onNavigate={closeDrawer} showBrand />
              </Drawer>
            )}
          </AppShell>
        </div>
      )}
    </>
  )
}
