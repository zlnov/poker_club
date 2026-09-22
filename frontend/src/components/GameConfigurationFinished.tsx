/**
 * GameConfigurationFinished component.
 *
 * Displays game configuration in a card-based layout for Finished games.
 * Three cards with icons, each containing a table with 2-column key-value layout.
 *
 * Per task requirements:
 * - Web view: 2 cards in first row, 1 card (full width) in second row
 * - Mobile view: 3 cards stacked vertically
 * - Each card has an icon and title in the header
 * - Table inside each card with left-aligned labels and center-aligned values
 *
 * Card contents:
 * 1. Game info: Игра, Дата завершения, Время завершения, Продолжительность, Players, Банкир, Банк
 * 2. Money: Currency, Money Model, Chip Value, Buy-in, Rebuy Allowed, Rebuy Price, Max Rebuys
 * 3. Ranking: Ranking Primary, Ranking Secondary
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
  IconChartBar,
} from '@tabler/icons-react'
import { useMediaQuery } from '@mantine/hooks'
import type { GameDetails, GameResult } from '../types'
import { formatDateShort, formatDuration } from '../utils'

interface GameConfigurationFinishedProps {
  game: GameDetails
  bankerName: string
  gameResults: GameResult[]
  isPlanned: boolean
  canManageGame: boolean
  onAdjustGame: () => void
}

/**
 * Formats a plain integer value (no currency symbol, no decimals).
 */
function formatInt(value: number): string {
  if (isNaN(value)) {
    return '—'
  }
  return Math.round(value).toString()
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

export function GameConfigurationFinished({
  game,
  bankerName,
  gameResults,
  isPlanned,
  canManageGame,
  onAdjustGame,
}: GameConfigurationFinishedProps) {
  const isMobile = useMediaQuery('(max-width: 768px)')

  // Calculate total bank from game results
  const totalBank = (gameResults || []).reduce(
    (sum, r) => sum + r.totalInvested,
    0,
  )

  // Get confirmed participants count
  const confirmedPlayers = (gameResults || []).filter(
    (r) => r.status === 'confirmed',
  ).length

  // Format game type label
  const gameTypeLabel = game.gameType === 'cash' ? 'Cash' : 'Tournament'
  const gameLabel = `${gameTypeLabel} ${game.id}`

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
        // Mobile: 3 cards stacked vertically
        <Stack gap="md">
          {/* Game Info Card */}
          <ConfigCard icon={<IconInfoCircle size={20} />} title="Game info">
            <TableRow label="Игра" value={gameLabel} />
            <TableRow
              label="Дата завершения"
              value={game.endTime ? formatDateShort(game.endTime) : '—'}
            />
            <TableRow
              label="Время завершения"
              value={game.endTime ? formatTimeOnly(game.endTime) : '—'}
            />
            <TableRow
              label="Продолжительность"
              value={
                game.durationSeconds && game.durationSeconds > 0
                  ? formatDuration(game.durationSeconds)
                  : '—'
              }
            />
            <TableRow label="Players" value={confirmedPlayers} />
            <TableRow label="Банкир" value={bankerName} />
            <TableRow label="Банк" value={formatInt(totalBank)} />
          </ConfigCard>

          {/* Money Card */}
          <ConfigCard icon={<IconCash size={20} />} title="Money">
            <TableRow label="Currency" value={game.currency} />
            <TableRow label="Money Model" value={game.moneyModel} />
            <TableRow label="Chip Value" value={game.chipValue} />
            <TableRow
              label="Buy-in Amount"
              value={formatInt(game.buyInAmount)}
            />
            <TableRow
              label="Rebuy Allowed"
              value={game.rebuyAllowed ? 'Yes' : 'No'}
            />
            {game.rebuyAllowed && (
              <>
                <TableRow
                  label="Rebuy Price"
                  value={
                    game.rebuyPrice !== undefined
                      ? formatInt(game.rebuyPrice)
                      : '—'
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

          {/* Ranking Card */}
          <ConfigCard icon={<IconChartBar size={20} />} title="Ranking">
            <TableRow label="Ranking Primary" value={game.rankingPrimary} />
            <TableRow
              label="Ranking Secondary"
              value={game.rankingSecondary || '—'}
            />
          </ConfigCard>
        </Stack>
      ) : (
        // Web: 2 cards in first row, 1 card (full width) in second row
        <Grid>
          <Grid.Col span={6}>
            <ConfigCard icon={<IconInfoCircle size={20} />} title="Game info">
              <TableRow label="Игра" value={gameLabel} />
              <TableRow
                label="Дата завершения"
                value={game.endTime ? formatDateShort(game.endTime) : '—'}
              />
              <TableRow
                label="Время завершения"
                value={game.endTime ? formatTimeOnly(game.endTime) : '—'}
              />
              <TableRow
                label="Продолжительность"
                value={
                  game.durationSeconds && game.durationSeconds > 0
                    ? formatDuration(game.durationSeconds)
                    : '—'
                }
              />
              <TableRow label="Players" value={confirmedPlayers} />
              <TableRow label="Банкир" value={bankerName} />
              <TableRow label="Банк" value={formatInt(totalBank)} />
            </ConfigCard>
          </Grid.Col>
          <Grid.Col span={6}>
            <ConfigCard icon={<IconCash size={20} />} title="Money">
              <TableRow label="Currency" value={game.currency} />
              <TableRow label="Money Model" value={game.moneyModel} />
              <TableRow label="Chip Value" value={game.chipValue} />
              <TableRow
                label="Buy-in Amount"
                value={formatInt(game.buyInAmount)}
              />
              <TableRow
                label="Rebuy Allowed"
                value={game.rebuyAllowed ? 'Yes' : 'No'}
              />
              {game.rebuyAllowed && (
                <>
                  <TableRow
                    label="Rebuy Price"
                    value={
                      game.rebuyPrice !== undefined
                        ? formatInt(game.rebuyPrice)
                        : '—'
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
          <Grid.Col span={12}>
            <ConfigCard icon={<IconChartBar size={20} />} title="Ranking">
              <TableRow label="Ranking Primary" value={game.rankingPrimary} />
              <TableRow
                label="Ranking Secondary"
                value={game.rankingSecondary || '—'}
              />
            </ConfigCard>
          </Grid.Col>
        </Grid>
      )}
    </Card>
  )
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
