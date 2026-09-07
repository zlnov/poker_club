/**
 * Game page component.
 *
 * Displays the details of a single game and provides actions
 * based on the game status and the current user's permissions.
 *
 * Phase 5 scope (RM_FE_5.md):
 * - Game Details
 * - Start Game (planned → active)
 * - Cancel Game (planned → cancelled)
 * - Banker assignment (change banker for planned games)
 *
 * Phase 6 (not in scope):
 * - Participants, Buy-in, Rebuy, Chips, Timer, Polling, Finish Game
 *
 * See 05_FE_UX.md section 4 (Application Navigation) and
 * 04_FE_SPEC.md section 11 (Routing).
 */

import { useState } from 'react'
import {
  Card,
  Group,
  Stack,
  Text,
  Title,
  Badge,
  Button,
  Select,
  Modal,
  Alert,
} from '@mantine/core'
import {
  IconChessKing,
  IconCash,
  IconTrophy,
  IconPlayerPlay,
  IconX,
  IconEdit,
  IconInfoCircle,
  IconUser,
} from '@tabler/icons-react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  PageContainer,
  PageHeader,
  EmptyState,
  LoadingState,
  ErrorState,
} from '../components/ui'
import { useGame, useStartGame, useCancelGame, useUpdateBanker } from '../features'
import { useClub, useClubMembers } from '../features'
import { useAuth } from '../auth'
import { formatCurrency, formatDate, formatDuration } from '../utils'
import { ApiClientError } from '../api'
import { notifications } from '@mantine/notifications'
import type { GameStatus, GameType, ClubMemberRole } from '../types'

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

function getGameTypeLabel(type: GameType): string {
  switch (type) {
    case 'cash_time':
      return 'Cash (Timed)'
    case 'cash_open':
      return 'Cash (Open)'
    case 'tournament':
      return 'Tournament'
    default:
      return type
  }
}

/**
 * Maps a backend game_type to a frontend GameType.
 */
function mapGameType(
  backendType: string,
  durationSeconds?: number,
): GameType {
  if (backendType === 'tournament') {
    return 'tournament'
  }
  if (durationSeconds !== undefined && durationSeconds > 0) {
    return 'cash_time'
  }
  return 'cash_open'
}

/**
 * Formats a player's display name from a ClubMemberRole.
 */
function formatMemberName(member: ClubMemberRole): string {
  const parts = [member.firstName, member.lastName].filter(Boolean)
  return parts.length > 0 ? parts.join(' ') : member.nickname || 'Unknown'
}

/**
 * Game page.
 *
 * Displays:
 * - Game details (type, status, configuration)
 * - Banker information
 * - Start Game button (planned games, banker/owner/admin)
 * - Cancel Game button (planned games, owner/admin)
 * - Change Banker option (planned games, owner/admin)
 * - Loading, empty, and error states
 */
export function GamePage() {
  const { gameId } = useParams<{ gameId: string }>()
  const navigate = useNavigate()
  const gameIdNum = gameId ? parseInt(gameId, 10) : 0

  const { user } = useAuth()
  const currentPlayerId = user?.id ?? 0

  const [cancelConfirmOpen, setCancelConfirmOpen] = useState(false)
  const [changeBankerOpen, setChangeBankerOpen] = useState(false)
  const [selectedBankerId, setSelectedBankerId] = useState<number | null>(null)

  const {
    data: game,
    isLoading,
    isError,
    error,
    refetch,
  } = useGame(gameIdNum)

  const { data: club } = useClub(game?.clubId ?? 0)
  const { data: members } = useClubMembers(game?.clubId ?? 0)

  const isOwner = club?.isOwner ?? false
  const isAdmin = club?.isAdmin ?? false
  const canManageGame = isOwner || isAdmin

  // Check if current user is the banker
  const currentMember = members?.find(
    (m) => m.playerId === currentPlayerId,
  )
  const isBanker = currentMember?.clubMemberId === game?.bankerId

  const canStartGame = canManageGame || isBanker

  const startGame = useStartGame()
  const cancelGame = useCancelGame()
  const updateBanker = useUpdateBanker()

  // Resolve banker name from club members
  const bankerMember = members?.find(
    (m) => m.clubMemberId === game?.bankerId,
  )
  const bankerName = bankerMember
    ? formatMemberName(bankerMember)
    : `Banker #${game?.bankerId ?? '—'}`

  // Build banker select options (active members only, excluding current banker)
  const bankerOptions = (members ?? [])
    .filter((m) => m.status === 'active')
    .map((m) => ({
      value: String(m.playerId),
      label: formatMemberName(m),
      disabled: m.clubMemberId === game?.bankerId,
    }))

  const gameType = game
    ? mapGameType(game.gameType, game.durationSeconds)
    : undefined

  const handleStartGame = async () => {
    try {
      await startGame.mutateAsync(gameIdNum)
      notifications.show({
        title: 'Game started',
        message: 'The game has been started successfully.',
        color: 'green',
      })
      refetch()
    } catch (err) {
      if (err instanceof ApiClientError) {
        notifications.show({
          title: 'Failed to start game',
          message: err.message,
          color: 'red',
        })
      } else {
        notifications.show({
          title: 'Failed to start game',
          message: 'An unexpected error occurred.',
          color: 'red',
        })
      }
    }
  }

  const handleCancelGame = async () => {
    try {
      await cancelGame.mutateAsync(gameIdNum)
      notifications.show({
        title: 'Game cancelled',
        message: 'The game has been cancelled successfully.',
        color: 'green',
      })
      setCancelConfirmOpen(false)
      refetch()
    } catch (err) {
      if (err instanceof ApiClientError) {
        notifications.show({
          title: 'Failed to cancel game',
          message: err.message,
          color: 'red',
        })
      } else {
        notifications.show({
          title: 'Failed to cancel game',
          message: 'An unexpected error occurred.',
          color: 'red',
        })
      }
    } finally {
      setCancelConfirmOpen(false)
    }
  }

  const handleBankerChange = async () => {
    if (!selectedBankerId) return
    try {
      await updateBanker.mutateAsync({
        gameId: gameIdNum,
        bankerId: selectedBankerId,
      })
      notifications.show({
        title: 'Banker updated',
        message: 'The banker has been changed successfully.',
        color: 'green',
      })
      setChangeBankerOpen(false)
      setSelectedBankerId(null)
      refetch()
    } catch (err) {
      if (err instanceof ApiClientError) {
        notifications.show({
          title: 'Failed to change banker',
          message: err.message,
          color: 'red',
        })
      } else {
        notifications.show({
          title: 'Failed to change banker',
          message: 'An unexpected error occurred.',
          color: 'red',
        })
      }
    }
  }

  if (isLoading) {
    return (
      <PageContainer>
        <PageHeader title="Game" description="Game details" />
        <LoadingState message="Loading game..." />
      </PageContainer>
    )
  }

  if (isError) {
    return (
      <PageContainer>
        <PageHeader title="Game" description="Game details" />
        <ErrorState
          message={
            error instanceof ApiClientError
              ? error.message
              : 'Failed to load game'
          }
          onRetry={() => refetch()}
        />
      </PageContainer>
    )
  }

  if (!game) {
    return (
      <PageContainer>
        <PageHeader title="Game" description="Game details" />
        <EmptyState
          title="Game not found"
          description="The game you're looking for doesn't exist or you don't have access."
        />
      </PageContainer>
    )
  }

  const isPlanned = game.status === 'planned'
  const isCancelled = game.status === 'cancelled'

  return (
    <PageContainer>
      <PageHeader
        title="Game Details"
        description={getGameTypeLabel(gameType!)}
        action={
          <Button
            variant="subtle"
            size="sm"
            onClick={() => navigate(-1)}
          >
            Back
          </Button>
        }
      />

      {/* Game Status & Type */}
      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Group justify="space-between" mb="md">
          <Group gap="sm">
            {getGameTypeIcon(gameType!)}
            <Title order={3} mb={0}>
              {getGameTypeLabel(gameType!)}
            </Title>
          </Group>
          <Group gap="xs">
            <Badge color={getStatusColor(game.status)} variant="light" size="lg">
              {game.status}
            </Badge>
            <Badge color={getGameTypeColor(gameType!)} variant="light">
              {game.gameType}
            </Badge>
          </Group>
        </Group>

        {/* Game Lifecycle Actions (planned games only) */}
        {isPlanned && canStartGame && (
          <Group gap="sm" mt="md">
            <Button
              leftSection={<IconPlayerPlay size={16} />}
              onClick={handleStartGame}
              loading={startGame.isPending}
              color="green"
            >
              Start Game
            </Button>
            <Button
              leftSection={<IconX size={16} />}
              onClick={() => setCancelConfirmOpen(true)}
              loading={cancelGame.isPending}
              color="red"
              variant="outline"
            >
              Cancel Game
            </Button>
          </Group>
        )}

        {isPlanned && !canStartGame && (
          <Alert
            color="blue"
            variant="light"
            title="Waiting to start"
            icon={<IconInfoCircle size={16} />}
            mt="md"
          >
            The game will start when the banker or an admin begins it.
          </Alert>
        )}

        {isCancelled && (
          <Alert
            color="red"
            variant="light"
            title="Game Cancelled"
            icon={<IconX size={16} />}
            mt="md"
          >
            This game has been cancelled and cannot be started.
          </Alert>
        )}
      </Card>

      {/* Game Configuration */}
      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Title order={4} mb="md">
          Game Configuration
        </Title>
        <Stack gap="sm">
          <Group gap="sm" justify="space-between">
            <Text size="sm" c="dimmed">
              Currency
            </Text>
            <Text size="sm" fw={500}>
              {game.currency}
            </Text>
          </Group>
          <Group gap="sm" justify="space-between">
            <Text size="sm" c="dimmed">
              Money Model
            </Text>
            <Text size="sm" fw={500}>
              {game.moneyModel}
            </Text>
          </Group>
          <Group gap="sm" justify="space-between">
            <Text size="sm" c="dimmed">
              Chip Value
            </Text>
            <Text size="sm" fw={500}>
              {game.chipValue}
            </Text>
          </Group>
          <Group gap="sm" justify="space-between">
            <Text size="sm" c="dimmed">
              Buy-in Amount
            </Text>
            <Text size="sm" fw={500}>
              {formatCurrency(game.buyInAmount, game.currency)}
            </Text>
          </Group>
          <Group gap="sm" justify="space-between">
            <Text size="sm" c="dimmed">
              Rebuy Allowed
            </Text>
            <Text size="sm" fw={500}>
              {game.rebuyAllowed ? 'Yes' : 'No'}
            </Text>
          </Group>
          {game.rebuyAllowed && (
            <>
              <Group gap="sm" justify="space-between">
                <Text size="sm" c="dimmed">
                  Rebuy Price
                </Text>
                <Text size="sm" fw={500}>
                  {game.rebuyPrice !== undefined
                    ? formatCurrency(game.rebuyPrice, game.currency)
                    : '—'}
                </Text>
              </Group>
              <Group gap="sm" justify="space-between">
                <Text size="sm" c="dimmed">
                  Max Rebuys
                </Text>
                <Text size="sm" fw={500}>
                  {game.maxRebuys !== undefined
                    ? game.maxRebuys === 0
                      ? 'Unlimited'
                      : game.maxRebuys
                    : '—'}
                </Text>
              </Group>
            </>
          )}
          {game.durationSeconds !== undefined && game.durationSeconds > 0 && (
            <Group gap="sm" justify="space-between">
              <Text size="sm" c="dimmed">
                Duration
              </Text>
              <Text size="sm" fw={500}>
                {formatDuration(game.durationSeconds)}
              </Text>
            </Group>
          )}
          <Group gap="sm" justify="space-between">
            <Text size="sm" c="dimmed">
              Scheduled Start
            </Text>
            <Text size="sm" fw={500}>
              {game.startTime ? formatDate(game.startTime) : '—'}
            </Text>
          </Group>
          <Group gap="sm" justify="space-between">
            <Text size="sm" c="dimmed">
              Min / Max Players
            </Text>
            <Text size="sm" fw={500}>
              {game.minPlayers} / {game.maxPlayers}
            </Text>
          </Group>
          <Group gap="sm" justify="space-between">
            <Text size="sm" c="dimmed">
              Ranking Primary
            </Text>
            <Text size="sm" fw={500}>
              {game.rankingPrimary}
            </Text>
          </Group>
          {game.rankingSecondary && (
            <Group gap="sm" justify="space-between">
              <Text size="sm" c="dimmed">
                Ranking Secondary
              </Text>
              <Text size="sm" fw={500}>
                {game.rankingSecondary}
              </Text>
            </Group>
          )}
        </Stack>
      </Card>

      {/* Banker */}
      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Group justify="space-between" mb="md">
          <Title order={4} mb={0}>
            Banker
          </Title>
          {isPlanned && canManageGame && (
            <Button
              leftSection={<IconEdit size={16} />}
              variant="subtle"
              size="sm"
              onClick={() => setChangeBankerOpen(true)}
            >
              Change Banker
            </Button>
          )}
        </Group>
        <Group gap="sm">
          <IconUser size={20} />
          <Stack gap={2}>
            <Text fw={500}>{bankerName}</Text>
            <Text size="sm" c="dimmed">
              Club Member ID: {game.bankerId}
            </Text>
          </Stack>
        </Group>
      </Card>

      {/* Game Metadata */}
      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Title order={4} mb="md">
          Game Information
        </Title>
        <Stack gap="sm">
          <Group gap="sm" justify="space-between">
            <Text size="sm" c="dimmed">
              Game ID
            </Text>
            <Text size="sm" fw={500}>
              {game.id}
            </Text>
          </Group>
          <Group gap="sm" justify="space-between">
            <Text size="sm" c="dimmed">
              Club ID
            </Text>
            <Text size="sm" fw={500}>
              {game.clubId}
            </Text>
          </Group>
          <Group gap="sm" justify="space-between">
            <Text size="sm" c="dimmed">
              Created
            </Text>
            <Text size="sm" fw={500}>
              {formatDate(game.createdAt)}
            </Text>
          </Group>
          <Group gap="sm" justify="space-between">
            <Text size="sm" c="dimmed">
              Last Updated
            </Text>
            <Text size="sm" fw={500}>
              {formatDate(game.updatedAt)}
            </Text>
          </Group>
        </Stack>
      </Card>

      {/* Cancel Game Confirmation Modal */}
      <Modal
        opened={cancelConfirmOpen}
        onClose={() => setCancelConfirmOpen(false)}
        title="Cancel Game"
        centered
      >
        <Stack gap="md">
          <Alert
            color="red"
            variant="light"
            title="Warning"
            icon={<IconInfoCircle size={16} />}
          >
            This action is irreversible. The game will be cancelled and all
            participants will be notified.
          </Alert>
          <Group justify="flex-end" gap="sm">
            <Button
              variant="subtle"
              onClick={() => setCancelConfirmOpen(false)}
              disabled={cancelGame.isPending}
            >
              Cancel
            </Button>
            <Button
              color="red"
              onClick={handleCancelGame}
              loading={cancelGame.isPending}
            >
              Confirm Cancel
            </Button>
          </Group>
        </Stack>
      </Modal>

      {/* Change Banker Modal */}
      <Modal
        opened={changeBankerOpen}
        onClose={() => setChangeBankerOpen(false)}
        title="Change Banker"
        centered
        size="sm"
      >
        <Stack gap="md">
          <Text>
            Select a new banker for this game. The banker must be an active
            member of the club.
          </Text>
          <Select
            label="New Banker"
            placeholder="Select banker"
            data={bankerOptions}
            value={selectedBankerId ? String(selectedBankerId) : null}
            onChange={(value) =>
              setSelectedBankerId(value ? parseInt(value, 10) : 0)
            }
            required
            leftSection={<IconUser size={16} />}
          />
          <Group justify="flex-end" gap="sm">
            <Button
              variant="subtle"
              onClick={() => setChangeBankerOpen(false)}
              disabled={updateBanker.isPending}
            >
              Cancel
            </Button>
            <Button
              onClick={handleBankerChange}
              loading={updateBanker.isPending}
              disabled={!selectedBankerId}
            >
              Save
            </Button>
          </Group>
        </Stack>
      </Modal>
    </PageContainer>
  )
}

// --- Helper ---

function getGameTypeColor(type: GameType): string {
  switch (type) {
    case 'cash_time':
    case 'cash_open':
      return 'violet'
    case 'tournament':
      return 'orange'
    default:
      return 'gray'
  }
}
