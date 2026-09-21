/**
 * Player Statistics page component.
 *
 * Displays statistics for a specific player in a club.
 * Path: Club → Statistics → Player
 *
 * Shows:
 * - Cash/Tournament toggle (default: Cash)
 * - Player summary statistics (ПодБлок 1)
 * - Personal game history (ПодБлок 2)
 *
 * Tournament toggle is active but shows placeholder when selected.
 *
 * Per agent_task_7_tables.md:
 * - Player Statistics block with two sub-blocks
 * - Sub-block 1: Summary table with all metrics
 * - Sub-block 2: Game history table with sorting
 */

import { useState } from 'react'
import { Card, Title, Button } from '@mantine/core'
import { IconArrowLeft } from '@tabler/icons-react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  PageContainer,
  PageHeader,
  EmptyState,
  LoadingState,
  ErrorState,
} from '../components/ui'
import { usePlayerStatistics, usePlayerGameHistory } from '../features'
import { CashTournamentToggle } from '../components/statistics'
import { PlayerStatsSummary } from '../components/statistics/PlayerStatsSummary'
import { PlayerHistoryTable } from '../components/statistics/PlayerHistoryTable'
import { PlayerHistoryMobile } from '../components/statistics/PlayerHistoryMobile'
import { useMediaQuery } from '@mantine/hooks'
import { ApiClientError } from '../api'

export function PlayerStatisticsPage() {
  const { clubId, playerId } = useParams<{ clubId: string; playerId: string }>()
  const clubIdNum = clubId ? parseInt(clubId, 10) : 0
  const playerIdNum = playerId ? parseInt(playerId, 10) : 0
  const navigate = useNavigate()
  const isMobile = useMediaQuery('(max-width: 768px')

  const [gameType, setGameType] = useState<'cash' | 'tournament'>('cash')

  const {
    data: stats,
    isLoading: statsLoading,
    isError: statsError,
    error: statsErrorObj,
    refetch: refetchStats,
  } = usePlayerStatistics(playerIdNum, clubIdNum, gameType)

  const {
    data: history,
    isLoading: historyLoading,
    isError: historyError,
    error: historyErrorObj,
  } = usePlayerGameHistory(playerIdNum, clubIdNum, gameType)

  const isLoading = statsLoading || historyLoading
  const isError = statsError || historyError
  const error = statsErrorObj || historyErrorObj

  if (isLoading) {
    return (
      <PageContainer>
        <PageHeader title="Statistics" description="Player statistics" />
        <LoadingState message="Loading statistics..." />
      </PageContainer>
    )
  }

  if (isError) {
    return (
      <PageContainer>
        <PageHeader title="Statistics" description="Player statistics" />
        <ErrorState
          message={
            error instanceof ApiClientError
              ? error.message
              : 'Failed to load statistics'
          }
          onRetry={() => refetchStats()}
        />
      </PageContainer>
    )
  }

  // Tournament placeholder
  if (gameType === 'tournament') {
    return (
      <PageContainer>
        <PageHeader
          title="Statistics"
          description="Player statistics"
          action={
            <Button
              variant="subtle"
              size="sm"
              leftSection={<IconArrowLeft size={16} />}
              onClick={() => navigate(-1)}
            >
              Back
            </Button>
          }
        />

        <Card mt="lg" padding="lg" radius="md" withBorder>
          <Title order={4} mb="md">
            Player Statistics
          </Title>
          <CashTournamentToggle value={gameType} onChange={setGameType} />
          <EmptyState
            title="Tournament statistics are not available yet"
            description="Tournament statistics will be available in a future update."
          />
        </Card>
      </PageContainer>
    )
  }

  // Empty state for new players
  if (!stats || stats.totalGames === 0) {
    return (
      <PageContainer>
        <PageHeader
          title="Statistics"
          description="Player statistics"
          action={
            <Button
              variant="subtle"
              size="sm"
              leftSection={<IconArrowLeft size={16} />}
              onClick={() => navigate(-1)}
            >
              Back
            </Button>
          }
        />

        <Card mt="lg" padding="lg" radius="md" withBorder>
          <Title order={4} mb="md">
            Player Statistics
          </Title>
          <CashTournamentToggle value={gameType} onChange={setGameType} />
          <EmptyState
            title="Игрок еще не сыграл ни одной игры"
            description="Statistics will be available after the player completes their first game."
          />
        </Card>
      </PageContainer>
    )
  }

  return (
    <PageContainer>
      <PageHeader
        title="Statistics"
        description="Player statistics"
        action={
          <Button
            variant="subtle"
            size="sm"
            leftSection={<IconArrowLeft size={16} />}
            onClick={() => navigate(-1)}
          >
            Back
          </Button>
        }
      />

      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Title order={4} mb="md">
          Player Statistics
        </Title>
        <CashTournamentToggle value={gameType} onChange={setGameType} />

        {/* Sub-block 1: Summary Statistics */}
        <PlayerStatsSummary stats={stats} />

        {/* Sub-block 2: Game History */}
        <Title order={4} mb="md">
          Personal Game History
        </Title>
        {history && history.length === 0 ? (
          <EmptyState title="История игр пуста" />
        ) : (
          <>
            {!isMobile ? (
              <PlayerHistoryTable history={history || []} />
            ) : (
              <PlayerHistoryMobile history={history || []} />
            )}
          </>
        )}
      </Card>
    </PageContainer>
  )
}
