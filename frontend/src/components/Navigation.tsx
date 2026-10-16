/**
 * Sidebar / drawer navigation for the Poker Club app shell.
 *
 * Sections:
 * - Brand + CURRENT CLUB switcher
 * - NAVIGATION (Dashboard, Games, Members, Statistics)
 * - Profile footer (desktop / drawer bottom)
 */

import {
  Box,
  Group,
  Image,
  Stack,
  Text,
  UnstyledButton,
} from '@mantine/core'
import {
  IconChartBar,
  IconChessKing,
  IconChevronRight,
  IconHome,
  IconUser,
  IconUsers,
} from '@tabler/icons-react'
import { Link, useLocation, useParams } from 'react-router-dom'
import { useAuth } from '../auth'
import { useLastClubId } from '../hooks/useLastClubId'
import { ClubSwitcher } from './ClubSwitcher'
import { navAccent } from './navTheme'

interface NavItem {
  label: string
  path: string
  icon: React.ComponentType<{ size?: number; stroke?: number; color?: string }>
  exact?: boolean
}

const clubNavItems: NavItem[] = [
  { label: 'Dashboard', path: '', icon: IconHome, exact: true },
  { label: 'Games', path: 'games', icon: IconChessKing },
  { label: 'Members', path: 'members', icon: IconUsers },
  { label: 'Statistics', path: 'statistics', icon: IconChartBar },
]

function SectionLabel({ children }: { children: React.ReactNode }) {
  return (
    <Text
      size="xs"
      fw={600}
      mb={8}
      style={{ color: navAccent.label }}
    >
      {children}
    </Text>
  )
}

/**
 * @param onNavigate - Optional callback after a link click (closes mobile drawer).
 * @param showBrand - Show logo + Poker Club title (true in desktop sidebar / drawer).
 */
export function Navigation({
  onNavigate,
  showBrand = true,
}: {
  onNavigate?: () => void
  showBrand?: boolean
}) {
  const location = useLocation()
  const params = useParams()
  const { user } = useAuth()
  const { lastClubId } = useLastClubId()

  const routeClubId = params.clubId
    ? Number.parseInt(params.clubId, 10)
    : null
  const effectiveClubId =
    routeClubId && Number.isFinite(routeClubId) ? routeClubId : lastClubId

  const displayName =
    user?.nickname ||
    [user?.firstName, user?.lastName].filter(Boolean).join(' ') ||
    'Profile'
  const initial = (displayName.charAt(0) || 'P').toUpperCase()

  const isNavActive = (item: NavItem): boolean => {
    if (!effectiveClubId) return false
    const base = `/clubs/${effectiveClubId}`
    const full = item.path ? `${base}/${item.path}` : base
    if (item.exact) {
      return location.pathname === full || location.pathname === `${full}/`
    }
    return location.pathname.startsWith(full)
  }

  const hrefFor = (item: NavItem): string => {
    if (!effectiveClubId) return '/clubs'
    const base = `/clubs/${effectiveClubId}`
    return item.path ? `${base}/${item.path}` : base
  }

  return (
    <Box
      style={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        minHeight: 0,
      }}
    >
      <Box style={{ flex: 1, overflowY: 'auto', padding: '16px 14px' }}>
        {showBrand && (
          <Group gap="sm" mb="lg" wrap="nowrap">
            <Image
              src="/landing/logo/logo.png"
              alt="Poker Club"
              w={36}
              h={36}
              fit="contain"
            />
            <Stack gap={0}>
              <Text fw={600} size="md" lh={1.2}>
                Poker Club
              </Text>
              <Text size="xs" style={{ color: navAccent.label }}>
                Club management
              </Text>
            </Stack>
          </Group>
        )}

        <Box mb="lg">
          <SectionLabel>Current Club</SectionLabel>
          <ClubSwitcher onNavigate={onNavigate} />
        </Box>

        <Box>
          <SectionLabel>Navigation</SectionLabel>
          <Stack gap={4}>
            {clubNavItems.map((item) => {
              const active = isNavActive(item)
              const Icon = item.icon
              return (
                <UnstyledButton
                  key={item.label}
                  component={Link}
                  to={hrefFor(item)}
                  onClick={onNavigate}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 10,
                    padding: '9px 12px',
                    borderRadius: 10,
                    background: active ? navAccent.activeBg : 'transparent',
                    borderLeft: active
                      ? `3px solid ${navAccent.color}`
                      : '3px solid transparent',
                    color: active ? '#fff' : 'rgba(255,255,255,0.78)',
                    transition: 'background 0.15s ease',
                  }}
                >
                  <Icon
                    size={18}
                    stroke={active ? 1.75 : 1.4}
                    color={active ? navAccent.color : undefined}
                  />
                  <Text size="sm" fw={active ? 600 : 500}>
                    {item.label}
                  </Text>
                </UnstyledButton>
              )
            })}
          </Stack>
        </Box>
      </Box>

      <Box
        style={{
          borderTop: `1px solid ${navAccent.border}`,
          padding: '12px 14px',
        }}
      >
        <UnstyledButton
          component={Link}
          to="/profile"
          onClick={onNavigate}
          style={{
            width: '100%',
            display: 'flex',
            alignItems: 'center',
            gap: 10,
            padding: '8px 10px',
            borderRadius: 10,
          }}
        >
          <Box
            style={{
              width: 36,
              height: 36,
              borderRadius: '50%',
              background: navAccent.activeBg,
              color: navAccent.color,
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
              fontWeight: 600,
              flexShrink: 0,
            }}
          >
            {user ? initial : <IconUser size={18} />}
          </Box>
          <Stack gap={0} style={{ flex: 1, minWidth: 0 }}>
            <Text size="sm" fw={600} truncate>
              {displayName}
            </Text>
            <Text size="xs" style={{ color: navAccent.label }}>
              Profile
            </Text>
          </Stack>
          <IconChevronRight
            size={16}
            style={{ opacity: 0.5, color: 'inherit' }}
          />
        </UnstyledButton>
      </Box>
    </Box>
  )
}
