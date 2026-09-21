/**
 * Club statistics page component.
 *
 * Displays the statistics for a club.
 * Path: Club → Statistics → Club
 *
 * Per agent_task_7_tables.md:
 * - Cash/Tournament toggle at top (default: Cash)
 * - Player Statistics block (above Club Statistics):
 *   - Sub-block 1: Summary table with player metrics
 *   - Sub-block 2: Personal game history table
 * - Club Statistics block with two sub-blocks:
 *   - Sub-block 1: General club statistics (icon cards for web, table for mobile)
 *   - Sub-block 2: Club member statistics table (tournament table)
 *
 * Club statistics is the same for all game types.
 * Member statistics depends on selected game type.
 */

import {
  Card,
  Group,
  Text,
  Title,
  SimpleGrid,
  Table,
  Button,
} from '@mantine/core'
import {
  IconUsers,
  IconChessKing,
  IconCash,
  IconTrophy,
  IconClock,
  IconChartBar,
  IconArrowLeft,
} from '@tabler/icons-react'
import { useParams, useNavigate } from 'react-router-dom'
import { useState } from 'react'
import {
  PageContainer,
  PageHeader,
  EmptyState,
  LoadingState,
  ErrorState,
} from '../components/ui'
import { useClubStatistics, useClubMemberStatistics } from '../features'
import { usePlayerStatistics, usePlayerGameHistory } from '../features'
import { CashTournamentToggle } from '../components/statistics'
import { PlayerStatsSummary } from '../components/statistics/PlayerStatsSummary'
import { PlayerHistoryTable } from '../components/statistics/PlayerHistoryTable'
import { PlayerHistoryMobile } from '../components/statistics/PlayerHistoryMobile'
import { ClubMemberStatsTable } from '../components/statistics/ClubMemberStatsTable'
import { ClubMemberStatsMobile } from '../components/statistics/ClubMemberStatsMobile'
import { useMediaQuery } from '@mantine/hooks'
import { ApiClientError } from '../api'
import { formatDuration } from '../utils'
import { useAuth } from '../auth'

export function ClubStatistics() {
  const { clubId } = useParams<{ clubId: string }>()
  const clubIdNum = clubId ? parseInt(clubId, 10) : 0
  const navigate = useNavigate()
  const isMobile = useMediaQuery('(max-width: 768px)')
  const { user } = useAuth()
  const currentPlayerId = user?.id ?? 0

  const [gameType, setGameType] = useState<'cash' | 'tournament'>('cash')

  const {
    data: stats,
    isLoading,
    isError,
    error,
    refetch,
  } = useClubStatistics(clubIdNum)

  const {
    data: memberStats,
    isLoading: memberStatsLoading,
    isError: memberStatsError,
    error: memberStatsErrorObj,
    refetch: refetchMemberStats,
  } = useClubMemberStatistics(clubIdNum, gameType)

  // Player statistics for the current user
  const {
    data: playerStats,
    isLoading: playerStatsLoading,
    isError: playerStatsError,
    error: playerStatsErrorObj,
    refetch: refetchPlayerStats,
  } = usePlayerStatistics(currentPlayerId, clubIdNum, gameType)

  const {
    data: playerHistory,
    isLoading: playerHistoryLoading,
    isError: playerHistoryError,
    error: playerHistoryErrorObj,
  } = usePlayerGameHistory(currentPlayerId, clubIdNum, gameType)

  const isLoadingAll =
    isLoading ||
    memberStatsLoading ||
    playerStatsLoading ||
    playerHistoryLoading
  const isErrorAll =
    isError || memberStatsError || playerStatsError || playerHistoryError
  const errorAll =
    error || memberStatsErrorObj || playerStatsErrorObj || playerHistoryErrorObj

  if (isLoadingAll) {
    return (
      <PageContainer>
        <PageHeader title="Statistics" description="Club statistics" />
        <LoadingState message="Loading statistics..." />
      </PageContainer>
    )
  }

  if (isErrorAll) {
    return (
      <PageContainer>
        <PageHeader title="Statistics" description="Club statistics" />
        <ErrorState
          message={
            errorAll instanceof ApiClientError
              ? errorAll.message
              : 'Failed to load statistics'
          }
          onRetry={() => {
            refetch()
            refetchMemberStats()
            refetchPlayerStats()
          }}
        />
      </PageContainer>
    )
  }

  if (!stats) {
    return (
      <PageContainer>
        <PageHeader title="Statistics" description="Club statistics" />
        <EmptyState
          title="No statistics available"
          description="Statistics will be available after games are played."
        />
      </PageContainer>
    )
  }

  return (
    <PageContainer>
      <PageHeader
        title="Statistics"
        description="Club statistics"
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

      {/* Game Type Toggle */}
      <Group mb="md">
        <CashTournamentToggle value={gameType} onChange={setGameType} />
      </Group>

      {/* Player Statistics Block */}
      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Title order={4} mb="md">
          Player Statistics
        </Title>

        {/* Sub-block 1: Summary Statistics */}
        {playerStats && playerStats.totalGames > 0 ? (
          <PlayerStatsSummary stats={playerStats} />
        ) : (
          <EmptyState
            title="Игрок еще не сыграл ни одной игры"
            description="Statistics will be available after the player completes their first game."
          />
        )}

        {/* Sub-block 2: Personal Game History */}
        <Title order={4} mb="md">
          Personal Game History
        </Title>
        {playerHistory && playerHistory.length === 0 ? (
          <EmptyState title="История игр пуста" />
        ) : (
          <>
            {!isMobile ? (
              <PlayerHistoryTable history={playerHistory || []} />
            ) : (
              <PlayerHistoryMobile history={playerHistory || []} />
            )}
          </>
        )}
      </Card>

      {/* Club Statistics Block */}
      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Title order={4} mb="md">
          Club Statistics
        </Title>

        {/* Sub-block 1: General Club Statistics */}
        {!isMobile ? (
          // Web: Icon cards with center alignment
          <SimpleGrid cols={4} spacing="md" mb="lg">
            <Card shadow="sm" padding="md" radius="md" withBorder ta="center">
              <Group gap="xs" mb="xs" justify="center">
                <IconUsers size={20} />
                <Text size="sm" c="dimmed">
                  Members
                </Text>
              </Group>
              <Text size="2xl" fw={700} ta="center">
                {stats.totalMembers}
              </Text>
            </Card>
            <Card shadow="sm" padding="md" radius="md" withBorder ta="center">
              <Group gap="xs" mb="xs" justify="center">
                <IconChessKing size={20} />
                <Text size="sm" c="dimmed">
                  Games played
                </Text>
              </Group>
              <Text size="2xl" fw={700} ta="center">
                {stats.totalGames}
              </Text>
            </Card>
            <Card shadow="sm" padding="md" radius="md" withBorder ta="center">
              <Group gap="xs" mb="xs" justify="center">
                <IconCash size={20} />
                <Text size="sm" c="dimmed">
                  Cash Games
                </Text>
              </Group>
              <Text size="2xl" fw={700} ta="center">
                {stats.cashGames}
              </Text>
            </Card>
            <Card shadow="sm" padding="md" radius="md" withBorder ta="center">
              <Group gap="xs" mb="xs" justify="center">
                <IconTrophy size={20} />
                <Text size="sm" c="dimmed">
                  Tournaments
                </Text>
              </Group>
              <Text size="2xl" fw={700} ta="center">
                {stats.tournamentGames}
              </Text>
            </Card>
          </SimpleGrid>
        ) : (
          // Mobile: Table format
          <Table
            variant="compact"
            styles={{
              td: { padding: '0.1rem' },
              th: {
                fontSize: 'var(--mantine-font-size-xs)',
                color: 'var(--mantine-color-dimmed)',
              },
            }}
            mb="lg"
          >
            <Table.Tbody>
              <Table.Tr>
                <Table.Td>
                  <Text size="xs" c="dimmed">
                    Members
                  </Text>
                </Table.Td>
                <Table.Td ta="center">
                  <Text size="xs" fw={500}>
                    {stats.totalMembers}
                  </Text>
                </Table.Td>
              </Table.Tr>
              <Table.Tr>
                <Table.Td>
                  <Text size="xs" c="dimmed">
                    Games played
                  </Text>
                </Table.Td>
                <Table.Td ta="center">
                  <Text size="xs" fw={500}>
                    {stats.totalGames}
                  </Text>
                </Table.Td>
              </Table.Tr>
              <Table.Tr>
                <Table.Td>
                  <Text size="xs" c="dimmed">
                    Cash Games
                  </Text>
                </Table.Td>
                <Table.Td ta="center">
                  <Text size="xs" fw={500}>
                    {stats.cashGames}
                  </Text>
                </Table.Td>
              </Table.Tr>
              <Table.Tr>
                <Table.Td>
                  <Text size="xs" c="dimmed">
                    Tournaments
                  </Text>
                </Table.Td>
                <Table.Td ta="center">
                  <Text size="xs" fw={500}>
                    {stats.tournamentGames}
                  </Text>
                </Table.Td>
              </Table.Tr>
              <Table.Tr>
                <Table.Td>
                  <Text size="xs" c="dimmed">
                    Total Buy-in
                  </Text>
                </Table.Td>
                <Table.Td ta="center">
                  <Text size="xs" fw={500}>
                    {Math.round(stats.totalBuyInAmount)}
                  </Text>
                </Table.Td>
              </Table.Tr>
              <Table.Tr>
                <Table.Td>
                  <Text size="xs" c="dimmed">
                    Total Rebuy
                  </Text>
                </Table.Td>
                <Table.Td ta="center">
                  <Text size="xs" fw={500}>
                    {Math.round(stats.totalRebuyAmount)}
                  </Text>
                </Table.Td>
              </Table.Tr>
              <Table.Tr>
                <Table.Td>
                  <Text size="xs" c="dimmed">
                    Total Bank
                  </Text>
                </Table.Td>
                <Table.Td ta="center">
                  <Text size="xs" fw={500}>
                    {Math.round(stats.totalBank)}
                  </Text>
                </Table.Td>
              </Table.Tr>
              <Table.Tr>
                <Table.Td>
                  <Text size="xs" c="dimmed">
                    Средняя продолжительность игры
                  </Text>
                </Table.Td>
                <Table.Td ta="center">
                  <Text size="xs" fw={500}>
                    {stats.averageGameDuration}
                  </Text>
                </Table.Td>
              </Table.Tr>
              <Table.Tr>
                <Table.Td>
                  <Text size="xs" c="dimmed">
                    Общая продолжительность игр
                  </Text>
                </Table.Td>
                <Table.Td ta="center">
                  <Text size="xs" fw={500}>
                    {formatDuration(
                      stats.totalGames *
                        parseDurationToSeconds(stats.averageGameDuration),
                    )}
                  </Text>
                </Table.Td>
              </Table.Tr>
            </Table.Tbody>
          </Table>
        )}

        {/* Financial Summary (web only - shown as cards with center alignment) */}
        {!isMobile && (
          <SimpleGrid cols={3} spacing="md" mb="lg">
            <Card shadow="sm" padding="md" radius="md" withBorder ta="center">
              <Group gap="xs" mb="xs" justify="center">
                <IconCash size={20} />
                <Text size="sm" c="dimmed">
                  Total Buy-in
                </Text>
              </Group>
              <Text size="lg" fw={600} ta="center">
                {Math.round(stats.totalBuyInAmount)}
              </Text>
            </Card>
            <Card shadow="sm" padding="md" radius="md" withBorder ta="center">
              <Group gap="xs" mb="xs" justify="center">
                <IconCash size={20} />
                <Text size="sm" c="dimmed">
                  Total Rebuy
                </Text>
              </Group>
              <Text size="lg" fw={600} ta="center">
                {Math.round(stats.totalRebuyAmount)}
              </Text>
            </Card>
            <Card shadow="sm" padding="md" radius="md" withBorder ta="center">
              <Group gap="xs" mb="xs" justify="center">
                <IconChartBar size={20} />
                <Text size="sm" c="dimmed">
                  Total Bank
                </Text>
              </Group>
              <Text size="lg" fw={600} ta="center">
                {Math.round(stats.totalBank)}
              </Text>
            </Card>
          </SimpleGrid>
        )}

        {/* Game Duration (web only - shown as card with center alignment) */}
        {!isMobile && (
          <SimpleGrid cols={2} spacing="md" mb="lg">
            <Card shadow="sm" padding="md" radius="md" withBorder ta="center">
              <Group gap="xs" mb="xs" justify="center">
                <IconClock size={20} />
                <Text size="sm" c="dimmed">
                  Средняя продолжительность игры
                </Text>
              </Group>
              <Text size="lg" fw={600} ta="center">
                {stats.averageGameDuration}
              </Text>
            </Card>
            <Card shadow="sm" padding="md" radius="md" withBorder ta="center">
              <Group gap="xs" mb="xs" justify="center">
                <IconClock size={20} />
                <Text size="sm" c="dimmed">
                  Общая продолжительность игр
                </Text>
              </Group>
              <Text size="lg" fw={600} ta="center">
                {formatDuration(
                  stats.totalGames *
                    parseDurationToSeconds(stats.averageGameDuration),
                )}
              </Text>
            </Card>
          </SimpleGrid>
        )}

        {/* Sub-block 2: Club Member Statistics */}
        <Title order={4} mb="md">
          Club Members Statistics
        </Title>
        {memberStats && memberStats.length === 0 ? (
          <EmptyState title="No member statistics available" />
        ) : (
          <>
            {!isMobile ? (
              <ClubMemberStatsTable members={memberStats || []} />
            ) : (
              <ClubMemberStatsMobile members={memberStats || []} />
            )}
          </>
        )}
      </Card>
    </PageContainer>
  )
}

/**
 * Parses a duration string (HH:MM:SS or MM:SS) back to seconds.
 */
function parseDurationToSeconds(duration: string): number {
  if (!duration || duration === '—') return 0
  const parts = duration.split(':')
  if (parts.length === 3) {
    const h = parseInt(parts[0], 10)
    const m = parseInt(parts[1], 10)
    const s = parseInt(parts[2], 10)
    return h * 3600 + m * 60 + s
  }
  if (parts.length === 2) {
    const m = parseInt(parts[0], 10)
    const s = parseInt(parts[1], 10)
    return m * 60 + s
  }
  return 0
}
