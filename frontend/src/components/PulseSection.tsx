/**
 * Pulse section component for the Club Dashboard.
 *
 * Displays the PULSE block with two cards:
 * - Active Game card (shows active game details or "no active games")
 * - Upcoming Game card (shows next planned game or "no upcoming games")
 *
 * Layout:
 * - Web: two cards in a row (grid 2 columns)
 * - Mobile: two cards in a column
 *
 * See 05_FE_UX.md section 4 (Application Navigation) and
 * 04_FE_SPEC.md section 11 (Routing).
 */

import {
  Card,
  Group,
  Stack,
  Text,
  Title,
  Badge,
  Button,
  SimpleGrid,
} from '@mantine/core'
import {
  IconChessKing,
  IconCash,
  IconTrophy,
  IconClock,
  IconChevronRight,
} from '@tabler/icons-react'
import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMediaQuery } from '@mantine/hooks'
import type { GameSummary, GameParticipantSummary, GameType } from '../types'
import { useGameParticipants } from '../features'
import { formatDateShort } from '../utils'
import { backgroundSurfaces } from '../styles'

interface PulseSectionProps {
  /** Active game (if any). */
  activeGame: GameSummary | null
  /** Upcoming planned game (if any). */
  upcomingGame: GameSummary | null
  /** Current player ID for status lookup. */
  currentPlayerId: number
}

/**
 * Formats a time string from an ISO date.
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
 * Returns the game type label for display.
 */
function getGameTypeLabel(type: GameType): string {
  switch (type) {
    case 'cash_time':
      return 'Cash'
    case 'cash_open':
      return 'Cash'
    case 'tournament':
      return 'Tournament'
    default:
      return type
  }
}

/**
 * Returns the game type badge color.
 */
function getGameTypeBadgeColor(type: GameType): string {
  switch (type) {
    case 'cash_time':
    case 'cash_open':
      return 'blue'
    case 'tournament':
      return 'violet'
    default:
      return 'gray'
  }
}

/**
 * Returns the game type icon.
 */
function getGameTypeIcon(type: GameType): React.ReactNode {
  switch (type) {
    case 'cash_time':
    case 'cash_open':
      return <IconCash size={16} />
    case 'tournament':
      return <IconTrophy size={16} />
    default:
      return <IconChessKing size={16} />
  }
}

/**
 * Gets the participant status text for the current player.
 */
function getPlayerStatusText(
  participants: GameParticipantSummary[] | undefined,
  currentPlayerId: number,
): { text: string; color: string } {
  if (!participants) {
    return { text: 'Вы приглашены', color: 'yellow' }
  }

  const participant = participants.find((p) => p.player.id === currentPlayerId)

  if (!participant) {
    return { text: 'Вы приглашены', color: 'yellow' }
  }

  switch (participant.status) {
    case 'invited':
      return { text: 'Вы приглашены', color: 'yellow' }
    case 'accepted':
    case 'confirmed':
      return { text: 'Вы участвуете', color: 'green' }
    case 'declined':
      return { text: 'Вы отказались', color: 'gray' }
    default:
      return { text: 'Вы приглашены', color: 'yellow' }
  }
}

/**
 * Computes the bank for an active game: Σ buy-in + rebuy.
 */
function computeBank(
  game: GameSummary,
  participants: GameParticipantSummary[] | undefined,
): number {
  if (!participants) return 0

  const confirmedParticipants = participants.filter(
    (p) => p.status === 'confirmed',
  )
  let totalBank = 0

  for (const p of confirmedParticipants) {
    const invested =
      p.buyInCount * game.buyInAmount + p.rebuyCount * (game.rebuyPrice ?? 0)
    totalBank += invested
  }

  return totalBank
}

/**
 * Formats remaining time for a cash_time game.
 */
function formatRemainingTime(
  startTime: string,
  durationSeconds: number | undefined,
): string | null {
  if (!durationSeconds || durationSeconds <= 0) return null

  const start = new Date(startTime).getTime()
  const now = Date.now()
  const elapsed = (now - start) / 1000
  const remaining = durationSeconds - elapsed

  if (remaining <= 0) return '00:00:00'

  const h = Math.floor(remaining / 3600)
  const m = Math.floor((remaining % 3600) / 60)
  const s = Math.floor(remaining % 60)

  return `${h.toString().padStart(2, '0')}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
}

/**
 * Active Game card component.
 */
function ActiveGameCard({
  game,
  participants,
  tick,
}: {
  game: GameSummary | null
  participants: GameParticipantSummary[] | undefined
  tick: number
}) {
  const navigate = useNavigate()
  // tick is used to trigger re-renders for the timer countdown
  void tick

  if (!game) {
    return (
      <Card
        padding="md"
        radius="md"
        withBorder
        bg={backgroundSurfaces.card}
        style={{ minHeight: 'auto' }}
      >
        <Group gap="xs">
          <Text c="dimmed" size="sm">
            ○
          </Text>
          <Text size="sm" c="dimmed">
            Нет активных игр
          </Text>
        </Group>
      </Card>
    )
  }

  const confirmedParticipants =
    participants?.filter((p) => p.status === 'confirmed') ?? []
  const participantCount = confirmedParticipants.length
  const bank = computeBank(game, participants)
  const remainingTime = formatRemainingTime(
    game.startTime,
    game.durationSeconds,
  )

  return (
    <Card
      padding="md"
      radius="md"
      withBorder
      bg={backgroundSurfaces.card}
      style={{ height: '100%', display: 'flex', flexDirection: 'column' }}
    >
      <Group justify="space-between" mb="sm">
        <Group gap="xs">
          <Text c="green" size="sm" fw={500}>
            ● ИГРА ИДЁТ
          </Text>
          {getGameTypeIcon(game.type)}
        </Group>
        <Badge color={getGameTypeBadgeColor(game.type)} variant="light">
          {getGameTypeLabel(game.type)}
        </Badge>
      </Group>

      <Stack gap="xs" mt="sm" style={{ flex: 1 }}>
        <Group gap="xs">
          <Text size="sm" c="dimmed">
            Старт:
          </Text>
          <Text size="sm">{formatTimeOnly(game.startTime)}</Text>
        </Group>

        {remainingTime && (
          <Group gap="xs">
            <Text size="sm" c="dimmed">
              Длительность:
            </Text>
            <Text size="sm" style={{ fontFamily: 'monospace' }}>
              {remainingTime}
            </Text>
            {game.durationSeconds && (
              <Text size="sm" c="dimmed">
                (до{' '}
                {formatTimeOnly(
                  new Date(
                    new Date(game.startTime).getTime() +
                      game.durationSeconds * 1000,
                  ).toISOString(),
                )}
                )
              </Text>
            )}
          </Group>
        )}

        <Group gap="xs">
          <Text size="sm" c="dimmed">
            Участников:
          </Text>
          <Text size="sm">{participantCount}</Text>
        </Group>

        <Group gap="xs">
          <Text size="sm" c="dimmed">
            Банк:
          </Text>
          <Text size="lg" fw={700}>
            {Math.round(bank)}
          </Text>
        </Group>
      </Stack>

      <Group justify="flex-end" mt="sm">
        <Button
          size="sm"
          rightSection={<IconChevronRight size={16} />}
          onClick={() => navigate(`/games/${game.id}`)}
        >
          Перейти в игру
        </Button>
      </Group>
    </Card>
  )
}

/**
 * Upcoming Game card component.
 */
function UpcomingGameCard({
  game,
  participants,
  currentPlayerId,
}: {
  game: GameSummary | null
  participants: GameParticipantSummary[] | undefined
  currentPlayerId: number
}) {
  const navigate = useNavigate()

  if (!game) {
    return (
      <Card
        padding="md"
        radius="md"
        withBorder
        bg={backgroundSurfaces.card}
        style={{ minHeight: 'auto' }}
      >
        <Group gap="xs">
          <Text c="dimmed" size="sm">
            ○
          </Text>
          <Text size="sm" c="dimmed">
            Нет запланированных игр
          </Text>
        </Group>
      </Card>
    )
  }

  const statusInfo = getPlayerStatusText(participants, currentPlayerId)

  return (
    <Card
      padding="md"
      radius="md"
      withBorder
      bg={backgroundSurfaces.card}
      style={{ height: '100%', display: 'flex', flexDirection: 'column' }}
    >
      <Group justify="space-between" mb="sm">
        <Text size="sm" fw={500}>
          БЛИЖАЙШАЯ ИГРА
        </Text>
        <Badge color={getGameTypeBadgeColor(game.type)} variant="light">
          {getGameTypeLabel(game.type)}
        </Badge>
      </Group>

      <Text size="xs" c="dimmed" mb="sm">
        ID: #{game.id}
      </Text>

      <Stack gap="xs" mt="sm" style={{ flex: 1 }}>
        <Group gap="xs">
          <Text size="sm" c="dimmed">
            Дата:
          </Text>
          <Text size="sm">{formatDateShort(game.startTime)}</Text>
        </Group>

        <Group gap="xs">
          <Text size="sm" c="dimmed">
            Время:
          </Text>
          <Text size="sm">{formatTimeOnly(game.startTime)}</Text>
        </Group>

        <Group gap="xs">
          <Text size="sm" c="dimmed">
            Buy-in:
          </Text>
          <Text size="sm">{Math.round(game.buyInAmount)}</Text>
        </Group>

        {game.rebuyAllowed && game.rebuyPrice !== undefined && (
          <Group gap="xs">
            <Text size="sm" c="dimmed">
              Rebuy:
            </Text>
            <Text size="sm">{Math.round(game.rebuyPrice)}</Text>
          </Group>
        )}

        <Group gap="xs">
          <Text size="sm" c="dimmed">
            Банкир:
          </Text>
          <Text size="sm">{game.bankerName || '—'}</Text>
        </Group>

        <Group gap="xs">
          <Text size="sm" c="dimmed">
            Участников:
          </Text>
          <Text size="sm">{game.currentPlayers || 0}</Text>
        </Group>
      </Stack>

      {/* My Status block */}
      <Card
        mt="sm"
        padding="sm"
        radius="md"
        withBorder
        bg={backgroundSurfaces.card}
      >
        <Group gap="xs">
          <Text c={statusInfo.color} size="sm" fw={500}>
            ●
          </Text>
          <Text size="sm" c={statusInfo.color} fw={500}>
            {statusInfo.text}
          </Text>
        </Group>
      </Card>

      <Group justify="flex-end" mt="sm">
        <Button
          size="sm"
          rightSection={<IconChevronRight size={16} />}
          onClick={() => navigate(`/games/${game.id}`)}
        >
          Перейти в карточку
        </Button>
      </Group>
    </Card>
  )
}

/**
 * Pulse section component.
 *
 * Displays the PULSE block with Active Game and Upcoming Game cards.
 * Layout adapts to mobile (column) and desktop (row).
 */
export function PulseSection({
  activeGame,
  upcomingGame,
  currentPlayerId,
}: PulseSectionProps) {
  const isMobile = useMediaQuery('(max-width: 768px)')
  const [tick, setTick] = useState(0)

  // Poll timer for active game countdown
  useEffect(() => {
    if (activeGame) {
      const interval = setInterval(() => {
        setTick((t) => t + 1)
      }, 1000)
      return () => clearInterval(interval)
    }
  }, [activeGame])

  // Fetch participants for active game (for bank calculation)
  const activeParticipants = useGameParticipants(activeGame?.id ?? 0)
  // Fetch participants for upcoming game (for status lookup)
  const upcomingParticipants = useGameParticipants(upcomingGame?.id ?? 0)

  return (
    <Card
      mt="lg"
      padding="lg"
      radius="md"
      withBorder
      bg={backgroundSurfaces.card}
    >
      <Group gap="xs" mb="md">
        <IconClock size={20} />
        <Title order={4} mb={0}>
          PULSE
        </Title>
      </Group>

      <SimpleGrid
        cols={isMobile ? 1 : 2}
        spacing="md"
        style={{ alignItems: 'stretch' }}
      >
        <ActiveGameCard
          game={activeGame}
          participants={activeParticipants.data}
          tick={tick}
        />
        <UpcomingGameCard
          game={upcomingGame}
          participants={upcomingParticipants.data}
          currentPlayerId={currentPlayerId}
        />
      </SimpleGrid>
    </Card>
  )
}
