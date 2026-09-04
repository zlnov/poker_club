/**
 * Club statistics page component.
 *
 * Displays the statistics for a club.
 *
 * See 05_FE_UX.md section 4 (Application Navigation) and
 * 04_FE_SPEC.md section 11 (Routing).
 */

import { Card, Group, Text, Title, SimpleGrid } from '@mantine/core'
import {
  IconUsers,
  IconChessKing,
  IconCash,
  IconTrophy,
  IconClock,
  IconChartBar,
} from '@tabler/icons-react'
import { useParams } from 'react-router-dom'
import {
  PageContainer,
  PageHeader,
  EmptyState,
  LoadingState,
  ErrorState,
} from '../components/ui'
import { useClubStatistics } from '../features'
import { formatCurrency } from '../utils'
import { ApiClientError } from '../api'

/**
 * Club statistics page.
 *
 * Displays:
 * - Club aggregate statistics (members, games, bank, etc.)
 * - Loading, empty, and error states
 */
export function ClubStatistics() {
  const { clubId } = useParams<{ clubId: string }>()
  const clubIdNum = clubId ? parseInt(clubId, 10) : 0

  const {
    data: stats,
    isLoading,
    isError,
    error,
    refetch,
  } = useClubStatistics(clubIdNum)

  if (isLoading) {
    return (
      <PageContainer>
        <PageHeader title="Statistics" description="Club statistics" />
        <LoadingState message="Loading statistics..." />
      </PageContainer>
    )
  }

  if (isError) {
    return (
      <PageContainer>
        <PageHeader title="Statistics" description="Club statistics" />
        <ErrorState
          message={
            error instanceof ApiClientError
              ? error.message
              : 'Failed to load statistics'
          }
          onRetry={() => refetch()}
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
      <PageHeader title="Statistics" description="Club aggregate statistics" />

      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Title order={4} mb="md">
          Overview
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
      </Card>

      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Title order={4} mb="md">
          Financial Summary
        </Title>
        <SimpleGrid cols={3} spacing="md">
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
      </Card>

      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Title order={4} mb="md">
          Game Duration
        </Title>
        <SimpleGrid cols={1} spacing="md">
          <Card shadow="sm" padding="md" radius="md" withBorder>
            <Group gap="xs" mb="xs">
              <IconClock size={20} />
              <Text size="sm" c="dimmed">
                Average Game Duration
              </Text>
            </Group>
            <Text size="lg" fw={600}>
              {stats.averageGameDuration}
            </Text>
          </Card>
        </SimpleGrid>
      </Card>
    </PageContainer>
  )
}
