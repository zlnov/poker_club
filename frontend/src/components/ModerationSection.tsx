/**
 * Moderation section component for the Club Dashboard.
 *
 * Displays the MODERATION block with four cards:
 * - Pending confirmations (counters)
 * - Game requires completion (active game warning)
 * - Game participation requests (table)
 * - Club membership requests (table)
 *
 * Layout:
 * - Web: 2 columns × 2 rows grid
 * - Mobile: 1 column, stacked
 *
 * Visible only for Admin/Owner roles.
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
  Table,
  Grid,
} from '@mantine/core'
import { IconAlertTriangle, IconChevronRight } from '@tabler/icons-react'
import { useNavigate } from 'react-router-dom'
import { useMediaQuery } from '@mantine/hooks'
import { useQueries } from '@tanstack/react-query'
import { useApiClient } from '../hooks'
import {
  useMembershipRequests,
  useClubInvites,
  useClubGames,
  participantKeys,
} from '../features'
import type {
  GameSummary,
  GameParticipantSummary,
  ClubMemberRole,
  InvitationInfo,
} from '../types'
import { formatDateShort } from '../utils'
import { backgroundSurfaces } from '../styles'

interface ModerationSectionProps {
  /** Club ID. */
  clubId: number
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
 * Returns the badge color based on count value.
 */
function getCountBadgeColor(count: number): string {
  if (count === 0) return 'gray'
  if (count <= 2) return 'blue'
  return 'orange'
}

/**
 * Formats a date to dd.mm.yyyy format.
 */
function formatDateDDMMYYYY(isoDate: string): string {
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
 * Formats a player name for display.
 */
function formatPlayerName(member: ClubMemberRole | InvitationInfo): string {
  const parts = [member.firstName, member.lastName].filter(Boolean)
  return parts.length > 0 ? parts.join(' ') : member.nickname || '—'
}

/**
 * Pending confirmations card with counters.
 */
function PendingConfirmationsCard({
  memberRequestCount,
  gameParticipationCount,
  unacceptedGameInvites,
  unacceptedClubInvites,
}: {
  memberRequestCount: number
  gameParticipationCount: number
  unacceptedGameInvites: number
  unacceptedClubInvites: number
}) {
  const counters = [
    { label: 'Заявки на вступление в клуб', count: memberRequestCount },
    { label: 'Заявки на участие в игре', count: gameParticipationCount },
    { label: 'Непринятые приглашения в игру', count: unacceptedGameInvites },
    { label: 'Непринятые приглашения в клуб', count: unacceptedClubInvites },
  ].filter((c) => c.count > 0)

  return (
    <Card
      padding="md"
      radius="md"
      withBorder
      bg={backgroundSurfaces.card}
      style={{ height: '100%', display: 'flex', flexDirection: 'column' }}
    >
      <Title order={5} mb="sm">
        <Text size="x">
          ОЖИДАЮТ ПОДТВЕРЖДЕНИЯ
        </Text>
      </Title>

      {counters.length === 0 ? (
        <Stack
          style={{ flex: 1, justifyContent: 'center', alignItems: 'center' }}
        >
          <Text size="sm" c="dimmed" ta="center">
            Нет ожидающих действий
          </Text>
        </Stack>
      ) : (
        <Stack gap={0} style={{ flex: 1, justifyContent: 'center' }}>
          {counters.map((counter, index) => (
            <Group
              key={counter.label}
              justify="space-between"
              style={{
                padding: '8px 0',
                borderBottom:
                  index < counters.length - 1
                    ? '1px solid var(--mantine-color-gray-3)'
                    : 'none',
              }}
            >
              <Text size="sm">{counter.label}</Text>
              <Badge color={getCountBadgeColor(counter.count)} variant="light">
                {counter.count}
              </Badge>
            </Group>
          ))}
        </Stack>
      )}
    </Card>
  )
}

/**
 * Active game warning card.
 */
function ActiveGameCard({
  activeGame,
  participants,
}: {
  activeGame: GameSummary | null
  participants: GameParticipantSummary[] | undefined
}) {
  const navigate = useNavigate()

  if (!activeGame) {
    return (
      <Card
        padding="md"
        radius="md"
        withBorder
        bg={backgroundSurfaces.card}
        style={{ height: '100%', display: 'flex', flexDirection: 'column' }}
      >
        <Title order={5} mb="sm">
          <Text size="x">
            ИГРА ТРЕБУЕТ ЗАВЕРШЕНИЯ
          </Text>
        </Title>
        <Stack
          style={{ flex: 1, justifyContent: 'center', alignItems: 'center' }}
        >
          <Group gap="xs">
            <Text c="dimmed" size="sm">
              ○
            </Text>
            <Text size="sm" c="dimmed">
              Нет активных игр
            </Text>
          </Group>
        </Stack>
      </Card>
    )
  }

  const confirmedCount =
    participants?.filter((p) => p.status === 'confirmed').length ?? 0

  return (
    <Card
      padding="md"
      radius="md"
      withBorder
      bg={backgroundSurfaces.card}
      style={{ height: '100%', display: 'flex', flexDirection: 'column' }}
    >
      <Title order={5} mb="sm">
        ИГРА ТРЕБУЕТ ЗАВЕРШЕНИЯ
      </Title>

      <Stack gap="xs" style={{ flex: 1 }}>
        <Group gap="xs">
          <Text c="orange" size="lg">
            ⚠
          </Text>
          <Text size="sm">Есть незавершённая игра</Text>
        </Group>

        <Text size="xs" c="dimmed">
          ID: #{activeGame.id} ·{' '}
          {activeGame.type === 'tournament' ? 'Tournament' : 'Cash'} · старт{' '}
          {formatDateShort(activeGame.startTime)},{' '}
          {formatTimeOnly(activeGame.startTime)}
        </Text>

        <Text size="xs" c="dimmed">
          Банкир: {activeGame.bankerName || '—'} · участников: {confirmedCount}
        </Text>
      </Stack>

      <Group justify="flex-end" mt="sm">
        <Button
          size="sm"
          rightSection={<IconChevronRight size={16} />}
          onClick={() => navigate(`/games/${activeGame.id}`)}
        >
          Перейти в игру
        </Button>
      </Group>
    </Card>
  )
}

/**
 * Game participation requests table card.
 */
function GameParticipationRequestsCard({
  gameParticipants,
  clubId,
}: {
  gameParticipants: {
    gameId: number
    game: GameSummary
    participants: GameParticipantSummary[]
  }[]
  clubId: number
}) {
  const navigate = useNavigate()

  // Collect all pending (accepted) game participation requests
  const allRequests = gameParticipants.flatMap((gp) =>
    gp.participants
      .filter((p) => p.status === 'accepted')
      .map((p) => ({
        playerName:
          p.player.nickname ||
          `${p.player.firstName} ${p.player.lastName}`.trim() ||
          '—',
        gameId: gp.gameId,
        gameDate: gp.game.startTime,
      })),
  )

  // Sort: first "Ожидает" (accepted), then by game date (nearest first)
  const sortedRequests = [...allRequests].sort((a, b) => {
    return new Date(a.gameDate).getTime() - new Date(b.gameDate).getTime()
  })

  const displayRequests = sortedRequests.slice(0, 5)
  const hasMore = sortedRequests.length > 5

  return (
    <Card
      padding="md"
      radius="md"
      withBorder
      bg={backgroundSurfaces.card}
      style={{ height: '100%', display: 'flex', flexDirection: 'column' }}
    >
      <Title order={5} mb="sm">
        <Text size="x">
          ЗАЯВКИ НА УЧАСТИЕ В ИГРАХ
        </Text>  
      </Title>

      {displayRequests.length === 0 ? (
        <Stack
          style={{ flex: 1, justifyContent: 'center', alignItems: 'center' }}
        >
          <Text size="sm" c="dimmed" ta="center">
            Нет активных заявок
          </Text>
        </Stack>
      ) : (
        <>
          <Table
            variant="unstyled"
            style={{ flex: 1, tableLayout: 'fixed' }}
            mt="auto"
          >
            <thead>
              <tr>
                <th style={{ textAlign: 'left' }}>Игрок</th>
                <th style={{ width: '80px', textAlign: 'center' }}>Игра</th>
                <th style={{ width: '120px', textAlign: 'center' }}>
                  Дата игры
                </th>
                <th style={{ width: '140px', textAlign: 'left' }}>Статус</th>
              </tr>
            </thead>
            <tbody>
              {displayRequests.map((req, index) => (
                <tr key={`${req.gameId}-${index}`}>
                  <td>
                    <Text size="sm">@{req.playerName}</Text>
                  </td>
                  <td style={{ textAlign: 'center' }}>
                    <Text
                      size="sm"
                      component="a"
                      href={`/games/${req.gameId}`}
                      style={{
                        color: 'var(--mantine-color-blue-6)',
                        cursor: 'pointer',
                      }}
                      onClick={(e) => {
                        e.preventDefault()
                        navigate(`/games/${req.gameId}`)
                      }}
                    >
                      #{req.gameId}
                    </Text>
                  </td>
                  <td style={{ textAlign: 'center' }}>
                    <Text size="sm">{formatDateDDMMYYYY(req.gameDate)}</Text>
                  </td>
                  <td>
                    <Group gap="xs">
                      <Text c="orange" size="sm">
                        ●
                      </Text>
                      <Text size="sm" c="orange">
                        Ожидает
                      </Text>
                    </Group>
                  </td>
                </tr>
              ))}
            </tbody>
          </Table>

          {hasMore && (
            <Group justify="flex-end" mt="sm">
              <Button
                size="sm"
                variant="subtle"
                rightSection={<IconChevronRight size={16} />}
                onClick={() => navigate(`/clubs/${clubId}/games`)}
              >
                Все заявки →
              </Button>
            </Group>
          )}
        </>
      )}
    </Card>
  )
}

/**
 * Club membership requests table card.
 */
function ClubMembershipRequestsCard({
  memberRequests,
  clubId,
}: {
  memberRequests: ClubMemberRole[] | undefined
  clubId: number
}) {
  const navigate = useNavigate()

  const pendingRequests =
    memberRequests?.filter((m) => m.status === 'pending') ?? []
  const sortedRequests = [...pendingRequests].sort(
    (a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime(),
  )

  const displayRequests = sortedRequests.slice(0, 5)
  const hasMore = sortedRequests.length > 5

  return (
    <Card
      padding="md"
      radius="md"
      withBorder
      bg={backgroundSurfaces.card}
      style={{ height: '100%', display: 'flex', flexDirection: 'column' }}
    >
      <Title order={5} mb="sm">
        <Text size="x">
          ЗАЯВКИ НА ВСТУПЛЕНИЕ В КЛУБ
        </Text>
      </Title>

      {displayRequests.length === 0 ? (
        <Stack
          style={{ flex: 1, justifyContent: 'center', alignItems: 'center' }}
        >
          <Text size="sm" c="dimmed" ta="center">
            Нет активных заявок
          </Text>
        </Stack>
      ) : (
        <>
          <Table
            variant="unstyled"
            style={{ flex: 1, tableLayout: 'fixed' }}
            mt="auto"
          >
            <thead>
              <tr>
                <th style={{ textAlign: 'left' }}>Игрок</th>
                <th style={{ width: '120px', textAlign: 'center' }}>
                  Дата заявки
                </th>
                <th style={{ width: '160px', textAlign: 'left' }}>Статус</th>
              </tr>
            </thead>
            <tbody>
              {displayRequests.map((req) => (
                <tr key={req.playerId}>
                  <td>
                    <Text size="sm">@{formatPlayerName(req)}</Text>
                  </td>
                  <td style={{ textAlign: 'center' }}>
                    <Text size="sm">{formatDateDDMMYYYY(req.createdAt)}</Text>
                  </td>
                  <td>
                    <Group gap="xs">
                      <Text c="orange" size="sm">
                        ●
                      </Text>
                      <Text size="sm" c="orange">
                        Ожидает подтверждения
                      </Text>
                    </Group>
                  </td>
                </tr>
              ))}
            </tbody>
          </Table>

          {hasMore && (
            <Group justify="flex-end" mt="sm">
              <Button
                size="sm"
                variant="subtle"
                rightSection={<IconChevronRight size={16} />}
                onClick={() => navigate(`/clubs/${clubId}/members`)}
              >
                Все заявки →
              </Button>
            </Group>
          )}
        </>
      )}
    </Card>
  )
}

/**
 * Moderation section component.
 *
 * Displays the MODERATION block with four cards:
 * - Pending confirmations (counters)
 * - Game requires completion (active game warning)
 * - Game participation requests (table)
 * - Club membership requests (table)
 *
 * Layout adapts to mobile (column) and desktop (2x2 grid).
 */
export function ModerationSection({ clubId }: ModerationSectionProps) {
  const isMobile = useMediaQuery('(max-width: 768px)')
  const apiClient = useApiClient()

  const { data: memberRequests } = useMembershipRequests(clubId)
  const { data: invites } = useClubInvites(clubId)
  const { data: games } = useClubGames(clubId)

  // Find active game
  const activeGame = games?.find((g) => g.status === 'active') ?? null

  // Fetch participants for all planned/active games using useQueries
  const gamesToFetch =
    games?.filter((g) => g.status === 'planned' || g.status === 'active') ?? []

  const participantQueries = useQueries({
    queries: gamesToFetch.map((game) => ({
      queryKey: participantKeys.list(game.id),
      queryFn: async () => {
        const response = await apiClient.get<{
          participants: GameParticipantSummary[]
        }>(`/games/${game.id}/participants`)
        return response.data.participants
      },
      enabled: !!game.id,
    })),
  })

  // Combine game data with participants
  const gamesWithParticipants = gamesToFetch.map((game, index) => ({
    gameId: game.id,
    game,
    participants: participantQueries[index]?.data ?? [],
  }))

  // Fetch participants for active game
  const activeGameParticipants = gamesWithParticipants.find(
    (gp) => gp.gameId === activeGame?.id,
  )

  // Calculate counters
  const memberRequestCount =
    memberRequests?.filter((m) => m.status === 'pending').length ?? 0
  const gameParticipationCount = gamesWithParticipants.reduce((total, gp) => {
    const accepted = gp.participants.filter((p) => p.status === 'accepted')
    return total + accepted.length
  }, 0)
  const unacceptedGameInvites = gamesWithParticipants.reduce((total, gp) => {
    const invited = gp.participants.filter((p) => p.status === 'invited')
    return total + invited.length
  }, 0)
  const unacceptedClubInvites = invites?.filter((i) => !i.accepted).length ?? 0

  return (
    <Card
      mt="lg"
      padding="lg"
      radius="md"
      withBorder
      bg={backgroundSurfaces.card}
    >
      <Group gap="xs" mb="md">
        <IconAlertTriangle size={20} />
        <Title order={4} mb={0}>
          MODERATION
        </Title>
      </Group>

      {isMobile ? (
        <Stack gap="md">
          <PendingConfirmationsCard
            memberRequestCount={memberRequestCount}
            gameParticipationCount={gameParticipationCount}
            unacceptedGameInvites={unacceptedGameInvites}
            unacceptedClubInvites={unacceptedClubInvites}
          />
          <ActiveGameCard
            activeGame={activeGame}
            participants={activeGameParticipants?.participants}
          />
          <GameParticipationRequestsCard
            gameParticipants={gamesWithParticipants}
            clubId={clubId}
          />
          <ClubMembershipRequestsCard
            memberRequests={memberRequests}
            clubId={clubId}
          />
        </Stack>
      ) : (
        <Grid>
          <Grid.Col span={6}>
            <PendingConfirmationsCard
              memberRequestCount={memberRequestCount}
              gameParticipationCount={gameParticipationCount}
              unacceptedGameInvites={unacceptedGameInvites}
              unacceptedClubInvites={unacceptedClubInvites}
            />
          </Grid.Col>
          <Grid.Col span={6}>
            <ActiveGameCard
              activeGame={activeGame}
              participants={activeGameParticipants?.participants}
            />
          </Grid.Col>
          <Grid.Col span={6}>
            <GameParticipationRequestsCard
              gameParticipants={gamesWithParticipants}
              clubId={clubId}
            />
          </Grid.Col>
          <Grid.Col span={6}>
            <ClubMembershipRequestsCard
              memberRequests={memberRequests}
              clubId={clubId}
            />
          </Grid.Col>
        </Grid>
      )}
    </Card>
  )
}
