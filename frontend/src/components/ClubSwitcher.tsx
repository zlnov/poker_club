/**
 * Global club switcher for the app shell.
 *
 * Sidebar (`inline`): expands in place and pushes NAVIGATION down.
 * Compact header: uses a portal dropdown so it fits the top bar.
 */

import { useState } from 'react'
import {
  Box,
  Collapse,
  Divider,
  Group,
  Menu,
  Stack,
  Text,
  UnstyledButton,
} from '@mantine/core'
import {
  IconChevronDown,
  IconSelector,
  IconSettings,
  IconUsersGroup,
} from '@tabler/icons-react'
import { useNavigate, useLocation, useParams } from 'react-router-dom'
import { useClubs } from '../features'
import { useLastClubId } from '../hooks/useLastClubId'
import { navAccent } from './navTheme'

interface ClubSwitcherProps {
  /** Compact trigger for the mobile header (dropdown instead of inline). */
  compact?: boolean
  /** Called after a navigation action (e.g. close mobile drawer). */
  onNavigate?: () => void
}

function useClubSwitcherState() {
  const navigate = useNavigate()
  const location = useLocation()
  const params = useParams()
  const { data: clubs } = useClubs()
  const { lastClubId, setLastClubId } = useLastClubId()

  const routeClubId = params.clubId
    ? Number.parseInt(params.clubId, 10)
    : null
  const isClubsListPage = location.pathname === '/clubs'
  const hasRouteClub = !!routeClubId && Number.isFinite(routeClubId)

  const effectiveClubId = isClubsListPage
    ? null
    : hasRouteClub
      ? routeClubId
      : lastClubId

  const currentClub =
    effectiveClubId != null
      ? (clubs?.find((c) => c.id === effectiveClubId) ?? null)
      : null

  const switchClubs =
    clubs?.filter((c) =>
      effectiveClubId != null ? c.id !== effectiveClubId : true,
    ) ?? []

  const settingsEnabled = effectiveClubId != null

  const handleSelectClub = (clubId: number, onNavigate?: () => void) => {
    setLastClubId(clubId)
    navigate(`/clubs/${clubId}`)
    onNavigate?.()
  }

  const handleSettings = (onNavigate?: () => void) => {
    if (!effectiveClubId) return
    navigate(`/clubs/${effectiveClubId}/settings`)
    onNavigate?.()
  }

  return {
    currentClub,
    switchClubs,
    settingsEnabled,
    handleSelectClub,
    handleSettings,
  }
}

function ClubPanelContent({
  switchClubs,
  settingsEnabled,
  onSelectClub,
  onSettings,
  framed = true,
}: {
  switchClubs: { id: number; name: string }[]
  settingsEnabled: boolean
  onSelectClub: (clubId: number) => void
  onSettings: () => void
  /** Draw outer frame (sidebar expand). Off inside Menu dropdown. */
  framed?: boolean
}) {
  return (
    <Box
      style={
        framed
          ? {
              marginTop: 8,
              borderRadius: 10,
              border: `1px solid ${navAccent.border}`,
              background: navAccent.surface,
              overflow: 'hidden',
            }
          : undefined
      }
    >
      <UnstyledButton
        disabled={!settingsEnabled}
        onClick={onSettings}
        style={{
          width: '100%',
          display: 'flex',
          alignItems: 'center',
          gap: 10,
          padding: '10px 12px',
          color: settingsEnabled ? undefined : navAccent.muted,
          opacity: settingsEnabled ? 1 : 0.55,
          cursor: settingsEnabled ? 'pointer' : 'not-allowed',
        }}
      >
        <IconSettings size={16} />
        <Text size="sm">Settings</Text>
      </UnstyledButton>

      <Divider color={navAccent.border} />

      <Text size="sm" px="12px" pt="10px" pb={4}>
        Switch Club
      </Text>

      {switchClubs.length === 0 ? (
        <Text
          size="xs"
          px="12px"
          pb="sm"
          style={{ color: navAccent.label }}
        >
          No other clubs
        </Text>
      ) : (
        <Stack gap={0} pb={4}>
          {switchClubs.map((club) => (
            <UnstyledButton
              key={club.id}
              onClick={() => onSelectClub(club.id)}
              style={{
                width: '100%',
                display: 'flex',
                alignItems: 'center',
                gap: 10,
                padding: '8px 12px',
              }}
            >
              <Box
                style={{
                  width: 22,
                  height: 22,
                  borderRadius: '50%',
                  background: navAccent.hoverBg,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  fontSize: 11,
                  fontWeight: 600,
                  flexShrink: 0,
                }}
              >
                {club.name.charAt(0).toUpperCase()}
              </Box>
              <Text size="sm" truncate>
                {club.name}
              </Text>
            </UnstyledButton>
          ))}
        </Stack>
      )}
    </Box>
  )
}

export function ClubSwitcher({
  compact = false,
  onNavigate,
}: ClubSwitcherProps) {
  const [opened, setOpened] = useState(false)
  const {
    currentClub,
    switchClubs,
    settingsEnabled,
    handleSelectClub,
    handleSettings,
  } = useClubSwitcherState()

  const trigger = (
    <UnstyledButton
      aria-label="Club switcher"
      aria-expanded={opened}
      onClick={compact ? undefined : () => setOpened((v) => !v)}
      style={{
        width: '100%',
        borderRadius: 10,
        border: `1px solid ${navAccent.border}`,
        background: navAccent.surface,
        padding: compact ? '6px 10px' : '10px 12px',
      }}
    >
      <Group gap="xs" wrap="nowrap" justify="space-between">
        <Group gap="xs" wrap="nowrap" style={{ minWidth: 0, flex: 1 }}>
          <IconUsersGroup
            size={compact ? 16 : 18}
            color={navAccent.color}
            stroke={1.5}
          />
          <Stack gap={0} style={{ minWidth: 0, flex: 1 }}>
            <Text
              size={compact ? 'xs' : 'sm'}
              fw={500}
              truncate
              style={{ color: currentClub ? undefined : navAccent.muted }}
            >
              {currentClub?.name ?? 'Select a club'}
            </Text>
            {!compact && currentClub && (
              <Text size="xs" style={{ color: navAccent.label }}>
                Current club
              </Text>
            )}
          </Stack>
        </Group>
        {compact ? (
          <IconChevronDown
            size={14}
            style={{ opacity: 0.6, color: 'inherit', flexShrink: 0 }}
          />
        ) : (
          <IconSelector
            size={16}
            style={{
              opacity: 0.6,
              color: 'inherit',
              flexShrink: 0,
              transform: opened ? 'rotate(180deg)' : undefined,
              transition: 'transform 0.15s ease',
            }}
          />
        )}
      </Group>
    </UnstyledButton>
  )

  // Mobile header: keep dropdown so it doesn't stretch the top bar.
  if (compact) {
    return (
      <Menu
        shadow="md"
        width={260}
        position="bottom-end"
        withinPortal
        opened={opened}
        onChange={setOpened}
      >
        <Menu.Target>{trigger}</Menu.Target>
        <Menu.Dropdown
          style={{
            background: navAccent.panel,
            border: `1px solid ${navAccent.border}`,
            padding: 0,
          }}
        >
          <ClubPanelContent
            framed={false}
            switchClubs={switchClubs}
            settingsEnabled={settingsEnabled}
            onSelectClub={(id) => {
              setOpened(false)
              handleSelectClub(id, onNavigate)
            }}
            onSettings={() => {
              setOpened(false)
              handleSettings(onNavigate)
            }}
          />
        </Menu.Dropdown>
      </Menu>
    )
  }

  // Sidebar / drawer: expand inline and push NAVIGATION down.
  return (
    <Box>
      {trigger}
      <Collapse expanded={opened}>
        <ClubPanelContent
          switchClubs={switchClubs}
          settingsEnabled={settingsEnabled}
          onSelectClub={(id) => {
            setOpened(false)
            handleSelectClub(id, onNavigate)
          }}
          onSettings={() => {
            setOpened(false)
            handleSettings(onNavigate)
          }}
        />
      </Collapse>
    </Box>
  )
}
