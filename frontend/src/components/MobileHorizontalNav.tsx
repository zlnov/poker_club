/**
 * Primary horizontal navigation for mobile layouts.
 */

import { Box, Group, Text, UnstyledButton } from '@mantine/core'
import {
  IconChartBar,
  IconChessKing,
  IconHome,
  IconUser,
  IconUsers,
} from '@tabler/icons-react'
import { Link, useLocation, useParams } from 'react-router-dom'
import { useLastClubId } from '../hooks/useLastClubId'
import { navAccent } from './navTheme'

interface MobileNavItem {
  label: string
  path: string
  icon: React.ComponentType<{ size?: number; stroke?: number; color?: string }>
  exact?: boolean
  kind: 'club' | 'profile'
}

const items: MobileNavItem[] = [
  { label: 'Dashboard', path: '', icon: IconHome, exact: true, kind: 'club' },
  { label: 'Games', path: 'games', icon: IconChessKing, kind: 'club' },
  { label: 'Members', path: 'members', icon: IconUsers, kind: 'club' },
  { label: 'Stats', path: 'statistics', icon: IconChartBar, kind: 'club' },
  { label: 'Profile', path: '/profile', icon: IconUser, kind: 'profile' },
]

export function MobileHorizontalNav() {
  const location = useLocation()
  const params = useParams()
  const { lastClubId } = useLastClubId()

  const routeClubId = params.clubId
    ? Number.parseInt(params.clubId, 10)
    : null
  const effectiveClubId =
    routeClubId && Number.isFinite(routeClubId) ? routeClubId : lastClubId

  const hrefFor = (item: MobileNavItem): string => {
    if (item.kind === 'profile') return '/profile'
    if (!effectiveClubId) return '/clubs'
    const base = `/clubs/${effectiveClubId}`
    return item.path ? `${base}/${item.path}` : base
  }

  const isActive = (item: MobileNavItem): boolean => {
    if (item.kind === 'profile') {
      return location.pathname.startsWith('/profile')
    }
    if (!effectiveClubId) return false
    const base = `/clubs/${effectiveClubId}`
    const full = item.path ? `${base}/${item.path}` : base
    if (item.exact) {
      return location.pathname === full || location.pathname === `${full}/`
    }
    return location.pathname.startsWith(full)
  }

  return (
    <Box
      hiddenFrom="sm"
      style={{
        borderBottom: `1px solid ${navAccent.border}`,
        background: 'rgba(28, 28, 28, 0.75)',
        padding: '4px 8px 0',
      }}
    >
      <Group gap={0} justify="space-between" wrap="nowrap">
        {items.map((item) => {
          const active = isActive(item)
          const Icon = item.icon
          return (
            <UnstyledButton
              key={item.label}
              component={Link}
              to={hrefFor(item)}
              style={{
                flex: 1,
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'center',
                gap: 2,
                padding: '8px 4px 10px',
                borderBottom: active
                  ? `2px solid ${navAccent.color}`
                  : '2px solid transparent',
                color: active ? navAccent.color : 'rgba(255,255,255,0.65)',
              }}
            >
              <Icon size={18} stroke={active ? 1.75 : 1.4} />
              <Text size="10px" fw={active ? 600 : 500} lh={1.1}>
                {item.label}
              </Text>
            </UnstyledButton>
          )
        })}
      </Group>
    </Box>
  )
}
