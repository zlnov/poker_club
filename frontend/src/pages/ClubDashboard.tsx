/**
 * Club dashboard page component.
 *
 * Displays the club dashboard with the PULSE section showing:
 * - Active Game card
 * - Upcoming Game card
 *
 * See 05_FE_UX.md section 4 (Application Navigation) and
 * 04_FE_SPEC.md section 11 (Routing).
 */

import { Card, Group, Title, UnstyledButton } from '@mantine/core'
import {
  IconChessKing,
  IconUsers,
  IconChartBar,
  IconChevronRight,
} from '@tabler/icons-react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  PageContainer,
  PageHeader,
  EmptyState,
  LoadingState,
  ErrorState,
} from '../components/ui'
import { useClub, useClubGames } from '../features'
import { useAuth } from '../auth'
import { ApiClientError } from '../api'
import { PulseSection } from '../components/PulseSection'
import { TopListSection } from '../components/TopListSection'
import { ModerationSection } from '../components/ModerationSection'
import { EventJournalSection } from '../components/EventJournalSection'
import { backgroundSurfaces } from '../styles'
import { navAccent } from '../components/navTheme'

/**
 * Club dashboard page.
 *
 * Displays:
 * - PULSE section (Active Game + Upcoming Game cards)
 * - TOP LIST section (Top 3 by Profit + Top 3 by ROI)
 * - MODERATION section (Admin/Owner only)
 * - EVENT JOURNAL section (Admin/Owner only)
 */
export function ClubDashboard() {
  const { clubId } = useParams<{ clubId: string }>()
  const navigate = useNavigate()
  const clubIdNum = clubId ? parseInt(clubId, 10) : 0
  const { user } = useAuth()
  const currentPlayerId = user?.id ?? 0

  const {
    data: club,
    isLoading: clubLoading,
    isError: clubError,
    error: clubFetchError,
    refetch: refetchClub,
  } = useClub(clubIdNum)

  const {
    data: games,
    isLoading: gamesLoading,
    isError: gamesError,
    refetch: refetchGames,
  } = useClubGames(clubIdNum)

  const isLoading = clubLoading || gamesLoading
  const isError = clubError || gamesError

  // Find active game
  const activeGame = games?.find((g) => g.status === 'active') ?? null

  // Find upcoming planned game (earliest start time)
  const upcomingGame =
    games
      ?.filter((g) => g.status === 'planned')
      .sort(
        (a, b) =>
          new Date(a.startTime).getTime() - new Date(b.startTime).getTime(),
      )[0] ?? null

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
            refetchGames()
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
      />

      {/* PULSE Section */}
      <PulseSection
        activeGame={activeGame}
        upcomingGame={upcomingGame}
        currentPlayerId={currentPlayerId}
      />

      {/* TOP LIST Section */}
      <TopListSection clubId={club.id} />

      {/* MODERATION Section (Admin/Owner only) */}
      {(isOwner || isAdmin) && <ModerationSection clubId={club.id} />}

      {/* EVENT JOURNAL Section (Admin/Owner only) */}
      {(isOwner || isAdmin) && <EventJournalSection clubId={club.id} />}

      {/* Quick Actions */}
      <Card
        mt="lg"
        padding="lg"
        radius="md"
        withBorder
        bg={backgroundSurfaces.card}
        style={{ borderColor: navAccent.border }}
      >
        <Title order={4} mb="md">
          Quick Actions
        </Title>
        <Group gap="sm" grow preventGrowOverflow={false}>
          {(
            [
              {
                label: 'View Games',
                icon: IconChessKing,
                to: `/clubs/${club.id}/games`,
              },
              {
                label: 'Manage Members',
                icon: IconUsers,
                to: `/clubs/${club.id}/members`,
              },
              {
                label: 'View Statistics',
                icon: IconChartBar,
                to: `/clubs/${club.id}/statistics`,
              },
            ] as const
          ).map((action) => (
            <UnstyledButton
              key={action.label}
              onClick={() => navigate(action.to)}
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'space-between',
                gap: 10,
                padding: '12px 14px',
                borderRadius: 10,
                border: `1px solid ${navAccent.color}`,
                background: 'rgba(16, 185, 129, 0.06)',
                minWidth: 140,
              }}
            >
              <Group gap="sm" wrap="nowrap">
                <action.icon size={18} color={navAccent.color} />
                <span style={{ fontSize: 14, fontWeight: 500 }}>
                  {action.label}
                </span>
              </Group>
              <IconChevronRight size={16} color={navAccent.color} />
            </UnstyledButton>
          ))}
        </Group>
      </Card>
    </PageContainer>
  )
}
