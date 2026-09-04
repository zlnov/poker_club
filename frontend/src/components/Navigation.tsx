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
  useMantineColorScheme,
  Group,
  Text,
  useMantineTheme,
} from '@mantine/core'
import {
  IconHome,
  IconUsers,
  IconChartBar,
  IconSettings,
  IconChessKing,
  IconClubs,
} from '@tabler/icons-react'
import { Link, useLocation, useParams } from 'react-router-dom'
import { ThemeToggle } from './ThemeToggle'

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
  { label: 'Clubs', href: '/clubs', icon: IconClubs },
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
  const { colorScheme } = useMantineColorScheme()
  const theme = useMantineTheme()

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
    <ScrollArea style={{ height: 'calc(100vh - 64px)' }}>
      <Stack gap={4} p="md" pt="xl">
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
        <Stack
          gap="xs"
          pt="lg"
          style={{ borderTop: `1px solid ${theme.colors.gray[2]}` }}
        >
          <Text size="xs" c="dimmed" fw={500}>
            Appearance
          </Text>
          <Group gap="xs">
            <ThemeToggle size="sm" />
            <Text size="sm" c="dimmed">
              {colorScheme === 'dark' ? 'Dark mode' : 'Light mode'}
            </Text>
          </Group>
        </Stack>
      </Stack>
    </ScrollArea>
  )
}
