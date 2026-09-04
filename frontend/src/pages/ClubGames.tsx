/**
 * Club games page component.
 *
 * Displays the list of games for a club.
 *
 * See 05_FE_UX.md section 4 (Application Navigation) and
 * 04_FE_SPEC.md section 11 (Routing).
 */

import { Card, Group, Stack, Text, Title, Badge, Button } from '@mantine/core'
import {
  IconChessKing,
  IconCash,
  IconTrophy,
  IconClock,
} from '@tabler/icons-react'
import { useParams, useNavigate } from 'react-router-dom'
import type { ReactNode } from 'react'
import {
  PageContainer,
  PageHeader,
  EmptyState,
  LoadingState,
  ErrorState,
} from '../components/ui'
import { useClubGames } from '../features'
import { formatCurrency, formatDate } from '../utils'
import { ApiClientError } from '../api'
import type { GameStatus, GameType } from '../types'

// --- Status helpers ---

function getStatusColor(status: GameStatus): string {
  switch (status) {
    case 'planned':
      return 'blue'
    case 'active':
      return 'green'
    case 'finished':
      return 'gray'
    case 'cancelled':
      return 'red'
    default:
      return 'gray'
  }
}

function getGameTypeIcon(type: GameType): ReactNode {
  switch (type) {
    case 'cash':
      return <IconCash size={16} />
    case 'tournament':
      return <IconTrophy size={16} />
    default:
      return <IconChessKing size={16} />
  }
}

/**
 * Club games page.
 *
 * Displays:
 * - List of games for the club
 * - Loading, empty, and error states
 */
export function ClubGames() {
  const { clubId } = useParams<{ clubId: string }>()
  const navigate = useNavigate()
  const clubIdNum = clubId ? parseInt(clubId, 10) : 0

  const {
    data: games,
    isLoading,
    isError,
    error,
    refetch,
  } = useClubGames(clubIdNum)

  if (isLoading) {
    return (
      <PageContainer>
        <PageHeader title="Games" description="Club games" />
        <LoadingState message="Loading games..." />
      </PageContainer>
    )
  }

  if (isError) {
    return (
      <PageContainer>
        <PageHeader title="Games" description="Club games" />
        <ErrorState
          message={
            error instanceof ApiClientError
              ? error.message
              : 'Failed to load games'
          }
          onRetry={() => refetch()}
        />
      </PageContainer>
    )
  }

  if (!games || games.length === 0) {
    return (
      <PageContainer>
        <PageHeader
          title="Games"
          description="Club games"
          action={
            <Button
              leftSection={<IconChessKing size={16} />}
              onClick={() => navigate(`/clubs/${clubId}/games`)}
            >
              Create Game
            </Button>
          }
        />
        <EmptyState
          title="No games yet"
          description="No games have been created for this club yet."
        />
      </PageContainer>
    )
  }

  return (
    <PageContainer>
      <PageHeader
        title="Games"
        description="Club games"
        action={
          <Button
            leftSection={<IconChessKing size={16} />}
            onClick={() => navigate(`/clubs/${clubId}/games`)}
          >
            Create Game
          </Button>
        }
      />

      <Stack gap="sm" mt="md">
        {games.map((game) => (
          <Card
            key={game.id}
            padding="lg"
            radius="md"
            withBorder
            style={{ cursor: 'pointer' }}
            onClick={() => navigate(`/games/${game.id}`)}
          >
            <Group justify="space-between">
              <Group gap="sm">
                {getGameTypeIcon(game.type)}
                <Stack gap={4}>
                  <Title order={4} mb={0}>
                    {game.type === 'cash' ? 'Cash Game' : 'Tournament'}
                  </Title>
                  <Text size="sm" c="dimmed">
                    {formatDate(game.startTime)}
                  </Text>
                </Stack>
              </Group>
              <Group gap="xs">
                <Badge color={getStatusColor(game.status)} variant="light">
                  {game.status}
                </Badge>
                <Badge
                  color={game.type === 'cash' ? 'violet' : 'orange'}
                  variant="light"
                >
                  {game.type}
                </Badge>
              </Group>
            </Group>

            <Group gap="md" mt="sm">
              <Group gap="xs">
                <IconCash size={16} />
                <Text size="sm">
                  Buy-in: {formatCurrency(game.buyInAmount, game.currency)}
                </Text>
              </Group>
              {game.rebuyAllowed && (
                <Group gap="xs">
                  <IconCash size={16} />
                  <Text size="sm">
                    Rebuy: {formatCurrency(game.rebuyPrice ?? 0, game.currency)}
                  </Text>
                </Group>
              )}
              <Group gap="xs">
                <IconClock size={16} />
                <Text size="sm">
                  {game.minPlayers}–{game.maxPlayers} players
                </Text>
              </Group>
            </Group>
          </Card>
        ))}
      </Stack>
    </PageContainer>
  )
}
