/**
 * GameConfigurationCards component.
 *
 * Displays game configuration in a card-based layout for Planned/Active/Canceled games.
 * Four cards with icons, each containing a table with 2-column key-value layout.
 *
 * Per task requirements:
 * - Web view: 2x2 grid of cards
 * - Mobile view: 4 cards stacked vertically
 * - Each card has an icon and title in the header
 * - Table inside each card with left-aligned labels and center-aligned values
 * - Adjust Game button remains unchanged
 */

import {
  Card,
  Group,
  Text,
  Title,
  Button,
  Table,
  Grid,
  Stack,
} from '@mantine/core'
import {
  IconEdit,
  IconInfoCircle,
  IconCash,
  IconUsers,
  IconCalendar,
} from '@tabler/icons-react'
import { useMediaQuery } from '@mantine/hooks'
import type { GameDetails } from '../types'

interface GameConfigurationCardsProps {
  game: GameDetails
  clubName?: string
  isPlanned: boolean
  canManageGame: boolean
  onAdjustGame: () => void
}

/**
 * Formats a date string for display in tables.
 * Shows date and time in dd.mm.yyyy hh:mm format.
 */
function formatDateTime(isoDate: string): string {
  if (!isoDate) return '—'
  try {
    const date = new Date(isoDate)
    const day = String(date.getDate()).padStart(2, '0')
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const year = date.getFullYear()
    const hours = String(date.getHours()).padStart(2, '0')
    const minutes = String(date.getMinutes()).padStart(2, '0')
    return `${day}.${month}.${year} ${hours}:${minutes}`
  } catch {
    return '—'
  }
}

/**
 * Formats a date string for display (date only).
 */
function formatDateOnly(isoDate: string): string {
  if (!isoDate) return '—'
  try {
    const date = new Date(isoDate)
    const day = String(date.getDate()).padStart(2, '0')
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const year = date.getFullYear()
    return `${day}.${month}.${year}`
  } catch {
    return '—'
  }
}

/**
 * Formats a time string from a date.
 */
function formatTimeOnly(isoDate: string): string {
  if (!isoDate) return '—'
  try {
    const date = new Date(isoDate)
    const hours = String(date.getHours()).padStart(2, '0')
    const minutes = String(date.getMinutes()).padStart(2, '0')
    return `${hours}:${minutes}`
  } catch {
    return '—'
  }
}

/**
 * Renders a key-value table row with left-aligned label and center-aligned value.
 */
function TableRow({ label, value }: { label: string; value: React.ReactNode }) {
  return (
    <Table.Tr>
      <Table.Td>
        <Text size="sm" c="dimmed">
          {label}
        </Text>
      </Table.Td>
      <Table.Td ta="center">
        <Text size="sm" fw={500}>
          {value}
        </Text>
      </Table.Td>
    </Table.Tr>
  )
}

/**
 * Renders a configuration card with icon, title, and table content.
 */
function ConfigCard({
  icon,
  title,
  children,
}: {
  icon: React.ReactNode
  title: string
  children: React.ReactNode
}) {
  return (
    <Card shadow="sm" padding="md" radius="md" withBorder>
      <Group gap="xs" mb="sm">
        {icon}
        <Title order={5} mb={0}>
          {title}
        </Title>
      </Group>
      <Table variant="striped">
        <Table.Tbody>{children}</Table.Tbody>
      </Table>
    </Card>
  )
}

export function GameConfigurationCards({
  game,
  clubName,
  isPlanned,
  canManageGame,
  onAdjustGame,
}: GameConfigurationCardsProps) {
  const isMobile = useMediaQuery('(max-width: 768px)')

  return (
    <Card mt="lg" padding="lg" radius="md" withBorder>
      <Group justify="space-between" mb="md">
        <Title order={4} mb={0}>
          Game Configuration
        </Title>
        {isPlanned && canManageGame && (
          <Button
            leftSection={<IconEdit size={16} />}
            variant="subtle"
            size="sm"
            onClick={onAdjustGame}
          >
            Adjust Game
          </Button>
        )}
      </Group>

      {isMobile ? (
        // Mobile: 4 cards stacked vertically
        <Stack gap="md">
          {/* Game Info Card */}
          <ConfigCard icon={<IconInfoCircle size={20} />} title="Game info">
            <TableRow label="Game ID" value={game.id} />
            <TableRow label="Club Name" value={clubName || '—'} />
          </ConfigCard>

          {/* Money Card */}
          <ConfigCard icon={<IconCash size={20} />} title="Money">
            <TableRow label="Currency" value={game.currency} />
            <TableRow label="Money Model" value={game.moneyModel} />
            <TableRow label="Chip Value" value={game.chipValue} />
            <TableRow label="Buy-in Amount" value={game.buyInAmount} />
            <TableRow
              label="Rebuy Allowed"
              value={game.rebuyAllowed ? 'Yes' : 'No'}
            />
            {game.rebuyAllowed && (
              <>
                <TableRow
                  label="Rebuy Price"
                  value={game.rebuyPrice !== undefined ? game.rebuyPrice : '—'}
                />
                <TableRow
                  label="Max Rebuys"
                  value={
                    game.maxRebuys !== undefined
                      ? game.maxRebuys === 0
                        ? 'Unlimited'
                        : game.maxRebuys
                      : '—'
                  }
                />
              </>
            )}
          </ConfigCard>

          {/* Players & Ranking Card */}
          <ConfigCard icon={<IconUsers size={20} />} title="Players & Ranking">
            <TableRow
              label="Min / Max Players"
              value={`${game.minPlayers} / ${game.maxPlayers}`}
            />
            <TableRow label="Ranking Primary" value={game.rankingPrimary} />
            <TableRow
              label="Ranking Secondary"
              value={game.rankingSecondary || '—'}
            />
          </ConfigCard>

          {/* Schedule Card */}
          <ConfigCard icon={<IconCalendar size={20} />} title="Schedule">
            <TableRow
              label="Scheduled Date"
              value={game.startTime ? formatDateOnly(game.startTime) : '—'}
            />
            <TableRow
              label="Scheduled Time"
              value={game.startTime ? formatTimeOnly(game.startTime) : '—'}
            />
            <TableRow
              label="Created"
              value={game.createdAt ? formatDateTime(game.createdAt) : '—'}
            />
            <TableRow
              label="Last Update"
              value={game.updatedAt ? formatDateTime(game.updatedAt) : '—'}
            />
          </ConfigCard>
        </Stack>
      ) : (
        // Web: 2x2 grid of cards
        <Grid>
          <Grid.Col span={6}>
            <ConfigCard icon={<IconInfoCircle size={20} />} title="Game info">
              <TableRow label="Game ID" value={game.id} />
              <TableRow label="Club Name" value={clubName || '—'} />
            </ConfigCard>
          </Grid.Col>
          <Grid.Col span={6}>
            <ConfigCard icon={<IconCash size={20} />} title="Money">
              <TableRow label="Currency" value={game.currency} />
              <TableRow label="Money Model" value={game.moneyModel} />
              <TableRow label="Chip Value" value={game.chipValue} />
              <TableRow label="Buy-in Amount" value={game.buyInAmount} />
              <TableRow
                label="Rebuy Allowed"
                value={game.rebuyAllowed ? 'Yes' : 'No'}
              />
              {game.rebuyAllowed && (
                <>
                  <TableRow
                    label="Rebuy Price"
                    value={
                      game.rebuyPrice !== undefined ? game.rebuyPrice : '—'
                    }
                  />
                  <TableRow
                    label="Max Rebuys"
                    value={
                      game.maxRebuys !== undefined
                        ? game.maxRebuys === 0
                          ? 'Unlimited'
                          : game.maxRebuys
                        : '—'
                    }
                  />
                </>
              )}
            </ConfigCard>
          </Grid.Col>
          <Grid.Col span={6}>
            <ConfigCard
              icon={<IconUsers size={20} />}
              title="Players & Ranking"
            >
              <TableRow
                label="Min / Max Players"
                value={`${game.minPlayers} / ${game.maxPlayers}`}
              />
              <TableRow label="Ranking Primary" value={game.rankingPrimary} />
              <TableRow
                label="Ranking Secondary"
                value={game.rankingSecondary || '—'}
              />
            </ConfigCard>
          </Grid.Col>
          <Grid.Col span={6}>
            <ConfigCard icon={<IconCalendar size={20} />} title="Schedule">
              <TableRow
                label="Scheduled Date"
                value={game.startTime ? formatDateOnly(game.startTime) : '—'}
              />
              <TableRow
                label="Scheduled Time"
                value={game.startTime ? formatTimeOnly(game.startTime) : '—'}
              />
              <TableRow
                label="Created"
                value={game.createdAt ? formatDateTime(game.createdAt) : '—'}
              />
              <TableRow
                label="Last Update"
                value={game.updatedAt ? formatDateTime(game.updatedAt) : '—'}
              />
            </ConfigCard>
          </Grid.Col>
        </Grid>
      )}
    </Card>
  )
}
