/**
 * Navigation component for the Poker Club Frontend.
 *
 * Provides sidebar navigation between the main sections of the application.
 *
 * Navigation structure from 05_FE_UX.md section 4:
 * - Clubs (list)
 * - Club (dashboard, games, members, statistics, settings)
 * - Profile
 *
 * See 05_FE_UX.md section 6 (Club Navigation) and
 * 04_FE_SPEC.md section 11 (Routing).
 */

import {
  NavLink as MantineNavLink,
  ScrollArea,
  Stack,
  UnstyledButton,
  useMantineColorScheme,
} from '@mantine/core'
import {
  IconHome,
  IconUsers,
  IconChartBar,
  IconSettings,
  IconSun,
  IconMoonStars,
  IconChessKing,
} from '@tabler/icons-react'
import { Link, useLocation, useParams } from 'react-router-dom'

/** Navigation item definition. */
interface NavItem {
  label: string
  href: string
  icon: React.ComponentType<{ size?: number; stroke?: number }>
  exact?: boolean
}

/** Club-level navigation items (shown when inside a club context). */
const clubNavItems: NavItem[] = [
  { label: 'Dashboard', href: '', icon: IconHome, exact: true },
  { label: 'Games', href: 'games', icon: IconChessKing },
  { label: 'Members', href: 'members', icon: IconUsers },
  { label: 'Statistics', href: 'statistics', icon: IconChartBar },
  { label: 'Settings', href: 'settings', icon: IconSettings },
]

/** Top-level navigation items (shown outside club context). */
const topNavItems: NavItem[] = [
  { label: 'Clubs', href: '/clubs', icon: IconChessKing },
  { label: 'Profile', href: '/profile', icon: IconUsers },
]

/**
 * Navigation component.
 *
 * Renders top-level navigation (Clubs, Profile) when not in a club context,
 * or club-level navigation (Dashboard, Games, Members, Statistics, Settings)
 * when inside a club context.
 */
export function Navigation() {
  const location = useLocation()
  const params = useParams()
  const { colorScheme, toggleColorScheme } = useMantineColorScheme()

  const clubId = params.clubId
  const showClubNav = !!clubId

  const isActive = (href: string, exact = false): boolean => {
    if (exact) {
      return location.pathname === href
    }
    return location.pathname.startsWith(href)
  }

  const renderNavItem = (item: NavItem, prefix = '') => {
    const fullHref = `${prefix}${item.href}`
    const active = isActive(fullHref, item.exact)

    return (
      <MantineNavLink
        key={fullHref}
        component={Link}
        to={fullHref}
        label={item.label}
        leftSection={<item.icon size={18} stroke={active ? 1.5 : 1} />}
        active={active}
        variant={active ? 'light' : 'subtle'}
      />
    )
  }

  return (
    <ScrollArea style={{ height: 'calc(100vh - 60px)' }}>
      <Stack gap={4} p="md">
        {showClubNav ? (
          <>
            {clubNavItems.map((item) =>
              renderNavItem(item, `/clubs/${clubId}/`),
            )}
          </>
        ) : (
          <>{topNavItems.map((item) => renderNavItem(item))}</>
        )}

        {/* Theme toggle at the bottom */}
        <UnstyledButton
          onClick={() => toggleColorScheme()}
          style={{
            display: 'flex',
            alignItems: 'center',
            width: '100%',
            padding: '8px 12px',
            borderRadius: 'var(--mantine-radius-sm)',
            color: 'var(--mantine-color-text)',
          }}
        >
          {colorScheme === 'dark' ? (
            <IconSun size={18} />
          ) : (
            <IconMoonStars size={18} />
          )}
          <span style={{ marginLeft: '0.5rem', fontSize: '0.875rem' }}>
            {colorScheme === 'dark' ? 'Light mode' : 'Dark mode'}
          </span>
        </UnstyledButton>
      </Stack>
    </ScrollArea>
  )
}
