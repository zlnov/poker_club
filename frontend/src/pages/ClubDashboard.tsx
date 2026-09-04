/**
 * Club dashboard page component.
 *
 * Displays the club dashboard — an overview of the club with statistics
 * and quick actions.
 *
 * See 05_FE_UX.md section 4 (Application Navigation) and
 * 04_FE_SPEC.md section 11 (Routing).
 */

import { Button, Card, Group, Text, Title, SimpleGrid } from '@mantine/core'
import {
  IconUsers,
  IconChessKing,
  IconCash,
  IconTrophy,
  IconClock,
  IconEdit,
  IconChartBar,
} from '@tabler/icons-react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  PageContainer,
  PageHeader,
  EmptyState,
  LoadingState,
  ErrorState,
} from '../components/ui'
import { useClub, useClubStatistics } from '../features'
import { formatCurrency } from '../utils'
import { ApiClientError } from '../api'

/**
 * Club dashboard page.
 *
 * Displays:
 * - Club information (name, role)
 * - Club statistics (members, games, bank, etc.)
 * - Quick action buttons
 */
export function ClubDashboard() {
  const { clubId } = useParams<{ clubId: string }>()
  const navigate = useNavigate()
  const clubIdNum = clubId ? parseInt(clubId, 10) : 0

  const {
    data: club,
    isLoading: clubLoading,
    isError: clubError,
    error: clubFetchError,
    refetch: refetchClub,
  } = useClub(clubIdNum)

  const {
    data: stats,
    isLoading: statsLoading,
    isError: statsError,
    refetch: refetchStats,
  } = useClubStatistics(clubIdNum)

  const isLoading = clubLoading || statsLoading
  const isError = clubError || statsError

  if (isLoading) {
    return (
      <PageContainer>
        <PageHeader title="Club Dashboard" />
        <LoadingState message="Loading club information..." />
      </PageContainer>
    )
  }

  if (isError) {
    return (
      <PageContainer>
        <PageHeader title="Club Dashboard" />
        <ErrorState
          message={
            clubFetchError instanceof ApiClientError
              ? clubFetchError.message
              : 'Failed to load club information'
          }
          onRetry={() => {
            refetchClub()
            refetchStats()
          }}
        />
      </PageContainer>
    )
  }

  if (!club) {
    return (
      <PageContainer>
        <PageHeader title="Club Dashboard" />
        <EmptyState
          title="Club not found"
          description="The club you're looking for doesn't exist or you don't have access."
        />
      </PageContainer>
    )
  }

  const isOwner = club.isOwner
  const isAdmin = club.isAdmin

  return (
    <PageContainer>
      <PageHeader
        title={club.name}
        description={
          isOwner
            ? 'You are the owner of this club'
            : isAdmin
              ? 'You are an admin of this club'
              : 'Club member'
        }
        action={
          (isOwner || isAdmin) && (
            <Button
              leftSection={<IconEdit size={16} />}
              variant="subtle"
              onClick={() => navigate(`/clubs/${club.id}/settings`)}
            >
              Edit Settings
            </Button>
          )
        }
      />

      {/* Club Statistics */}
      {stats && (
        <Card mt="lg" padding="lg" radius="md" withBorder>
          <Title order={4} mb="md">
            Club Statistics
          </Title>
          <SimpleGrid cols={4} spacing="md">
            <Card shadow="sm" padding="md" radius="md" withBorder>
              <Group gap="xs" mb="xs">
                <IconUsers size={20} />
                <Text size="sm" c="dimmed">
                  Members
                </Text>
              </Group>
              <Text size="2xl" fw={700}>
                {stats.totalMembers}
              </Text>
            </Card>
            <Card shadow="sm" padding="md" radius="md" withBorder>
              <Group gap="xs" mb="xs">
                <IconChessKing size={20} />
                <Text size="sm" c="dimmed">
                  Total Games
                </Text>
              </Group>
              <Text size="2xl" fw={700}>
                {stats.totalGames}
              </Text>
            </Card>
            <Card shadow="sm" padding="md" radius="md" withBorder>
              <Group gap="xs" mb="xs">
                <IconCash size={20} />
                <Text size="sm" c="dimmed">
                  Cash Games
                </Text>
              </Group>
              <Text size="2xl" fw={700}>
                {stats.cashGames}
              </Text>
            </Card>
            <Card shadow="sm" padding="md" radius="md" withBorder>
              <Group gap="xs" mb="xs">
                <IconTrophy size={20} />
                <Text size="sm" c="dimmed">
                  Tournaments
                </Text>
              </Group>
              <Text size="2xl" fw={700}>
                {stats.tournamentGames}
              </Text>
            </Card>
          </SimpleGrid>

          <SimpleGrid cols={3} spacing="md" mt="md">
            <Card shadow="sm" padding="md" radius="md" withBorder>
              <Group gap="xs" mb="xs">
                <IconCash size={20} />
                <Text size="sm" c="dimmed">
                  Total Buy-in
                </Text>
              </Group>
              <Text size="lg" fw={600}>
                {formatCurrency(stats.totalBuyInAmount, 'USD')}
              </Text>
            </Card>
            <Card shadow="sm" padding="md" radius="md" withBorder>
              <Group gap="xs" mb="xs">
                <IconCash size={20} />
                <Text size="sm" c="dimmed">
                  Total Rebuy
                </Text>
              </Group>
              <Text size="lg" fw={600}>
                {formatCurrency(stats.totalRebuyAmount, 'USD')}
              </Text>
            </Card>
            <Card shadow="sm" padding="md" radius="md" withBorder>
              <Group gap="xs" mb="xs">
                <IconChartBar size={20} />
                <Text size="sm" c="dimmed">
                  Total Bank
                </Text>
              </Group>
              <Text size="lg" fw={600}>
                {formatCurrency(stats.totalBank, 'USD')}
              </Text>
            </Card>
          </SimpleGrid>

          <SimpleGrid cols={1} spacing="md" mt="md">
            <Card shadow="sm" padding="md" radius="md" withBorder>
              <Group gap="xs" mb="xs">
                <IconClock size={20} />
                <Text size="sm" c="dimmed">
                  Avg Game Duration
                </Text>
              </Group>
              <Text size="lg" fw={600}>
                {stats.averageGameDuration}
              </Text>
            </Card>
          </SimpleGrid>
        </Card>
      )}

      {/* Quick Actions */}
      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Title order={4} mb="md">
          Quick Actions
        </Title>
        <Group gap="sm">
          <Button
            leftSection={<IconChessKing size={16} />}
            onClick={() => navigate(`/clubs/${club.id}/games`)}
          >
            View Games
          </Button>
          <Button
            leftSection={<IconUsers size={16} />}
            onClick={() => navigate(`/clubs/${club.id}/members`)}
          >
            Manage Members
          </Button>
          <Button
            leftSection={<IconChartBar size={16} />}
            onClick={() => navigate(`/clubs/${club.id}/statistics`)}
          >
            View Statistics
          </Button>
        </Group>
      </Card>
    </PageContainer>
  )
}
