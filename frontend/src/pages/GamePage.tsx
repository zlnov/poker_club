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

import { useState, useEffect } from 'react'
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
  NumberInput,
  Checkbox,
  Grid,
  Table,
} from '@mantine/core'
import { DateTimePicker } from '@mantine/dates'
import { useForm } from '@mantine/form'
import { useMediaQuery } from '@mantine/hooks'
import {
  IconChessKing,
  IconCash,
  IconTrophy,
  IconPlayerPlay,
  IconX,
  IconCheck,
  IconEdit,
  IconInfoCircle,
  IconUser,
  IconCalendar,
  IconCurrencyDollar,
  IconCoin,
  IconClock,
  IconChevronLeft,
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
import {
  useGame,
  useStartGame,
  useCancelGame,
  useUpdateBanker,
  useUpdateGame,
  useGameParticipants,
  useAcceptGameParticipation,
  useDeclineGameParticipation,
  useConfirmGameParticipation,
  useGameMonitor,
  useFinishGame,
  useRegisterRebuy,
  useSetCurrentStack,
  useGameResults,
} from '../features'
import { useClub, useClubMembers } from '../features'
import { useAuth } from '../auth'
import { ApiClientError } from '../api'
import type { GameBankCheck } from '../types'
import { notifications } from '@mantine/notifications'
import type { GameStatus, GameType, ClubMemberRole, GameConfig } from '../types'
import { GameResultsTable } from '../components/statistics/GameResultsTable'
import { GameResultsMobile } from '../components/statistics/GameResultsMobile'
import { GameConfigurationCards } from '../components/GameConfigurationCards'
import { GameConfigurationFinished } from '../components/GameConfigurationFinished'

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
function mapGameType(backendType: string, durationSeconds?: number): GameType {
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

// --- Form options ---

const gameTypeOptions: { value: GameType; label: string }[] = [
  { value: 'cash_time', label: 'Cash (Timed)' },
  { value: 'cash_open', label: 'Cash (Open)' },
  { value: 'tournament', label: 'Tournament' },
]

const currencyOptions = [
  { value: 'RUB', label: 'RUB' },
  { value: 'USD', label: 'USD' },
  { value: 'EUR', label: 'EUR' },
]

const moneyModelOptions = [
  { value: 'real', label: 'Real' },
  { value: 'points', label: 'Points' },
  { value: 'virtual', label: 'Virtual' },
  { value: 'practice', label: 'Practice' },
]

const rankingPrimaryOptions = [
  { value: 'profit', label: 'Profit' },
  { value: 'chips', label: 'Chips' },
  { value: 'place', label: 'Place' },
]

const rankingSecondaryOptions = [
  { value: 'roi', label: 'ROI' },
  { value: 'chips', label: 'Chips' },
  { value: 'place', label: 'Place' },
]

// --- Form values ---

interface GameFormValues {
  gameType: GameType
  currency: string
  moneyModel: string
  chipValue: number
  buyInAmount: number
  rebuyAllowed: boolean
  rebuyPrice: number
  maxRebuys: number
  durationSeconds: number
  startTime: Date | null
  minPlayers: number
  maxPlayers: number
  rankingPrimary: string
  rankingSecondary: string
  bankerId: number
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
  const [adjustGameOpen, setAdjustGameOpen] = useState(false)
  const [selectedBankerId, setSelectedBankerId] = useState<number | null>(null)

  const { data: game, isLoading, isError, error, refetch } = useGame(gameIdNum)

  const { data: club } = useClub(game?.clubId ?? 0)
  const { data: members } = useClubMembers(game?.clubId ?? 0)

  const isOwner = club?.isOwner ?? false
  const isAdmin = club?.isAdmin ?? false
  const canManageGame = isOwner || isAdmin

  // Check if current user is the banker
  const currentMember = members?.find((m) => m.playerId === currentPlayerId)
  const isBanker = currentMember?.clubMemberId === game?.bankerId

  const canStartGame = canManageGame || isBanker

  const startGame = useStartGame()
  const cancelGame = useCancelGame()
  const updateBanker = useUpdateBanker()
  const updateGame = useUpdateGame()
  const { data: participants } = useGameParticipants(gameIdNum)
  const acceptParticipation = useAcceptGameParticipation()
  const declineParticipation = useDeclineGameParticipation()
  const confirmParticipation = useConfirmGameParticipation()

  // Phase 6: Active Game hooks
  const { data: monitorData } = useGameMonitor(gameIdNum)
  const finishGame = useFinishGame()
  const registerRebuy = useRegisterRebuy()
  const setCurrentStack = useSetCurrentStack()

  // Phase 7: Game Results hook (for finished games)
  const { data: gameResultsData } = useGameResults(gameIdNum)

  const isMobile = useMediaQuery('(max-width: 768px)')

  // Active game state
  const isActive = game?.status === 'active'
  const canManageGameActions = canManageGame || isBanker

  // Modal states for active game
  const [currentStackModalOpen, setCurrentStackModalOpen] = useState(false)
  const [currentStackValue, setCurrentStackValue] = useState<number>(0)
  const [finishGameModalOpen, setFinishGameModalOpen] = useState(false)
  const [rebuyPlayerId, setRebuyPlayerId] = useState<number | null>(null)

  // Build banker select options (active members only, excluding current banker)
  const bankerOptions = (members ?? [])
    .filter((m) => m.status === 'active')
    .map((m) => ({
      value: String(m.playerId),
      label: formatMemberName(m),
      disabled: m.clubMemberId === game?.bankerId,
    }))

  // Resolve banker name from club members
  const bankerMember = members?.find((m) => m.clubMemberId === game?.bankerId)
  const bankerName = bankerMember
    ? formatMemberName(bankerMember)
    : `Banker #${game?.bankerId ?? '—'}`

  // Build form for adjusting game configuration
  const gameType = game
    ? mapGameType(game.gameType, game.durationSeconds)
    : undefined

  const adjustForm = useForm<GameFormValues>({
    initialValues: {
      gameType: gameType ?? 'cash_open',
      currency: game?.currency ?? 'RUB',
      moneyModel: game?.moneyModel ?? 'real',
      chipValue: game?.chipValue ?? 1,
      buyInAmount: game?.buyInAmount ?? 0,
      rebuyAllowed: game?.rebuyAllowed ?? false,
      rebuyPrice: game?.rebuyPrice ?? 0,
      maxRebuys: game?.maxRebuys ?? 0,
      durationSeconds: game?.durationSeconds ?? 0,
      startTime: game?.startTime ? new Date(game.startTime) : null,
      minPlayers: game?.minPlayers ?? 2,
      maxPlayers: game?.maxPlayers ?? 10,
      rankingPrimary: game?.rankingPrimary ?? 'profit',
      rankingSecondary: game?.rankingSecondary ?? '',
      bankerId: game?.bankerId ?? 0,
    },
    validate: {
      gameType: (value) => (!value ? 'Game type is required' : null),
      currency: (value) => (!value ? 'Currency is required' : null),
      moneyModel: (value) => (!value ? 'Money model is required' : null),
      chipValue: (value) =>
        value <= 0 ? 'Chip value must be greater than 0' : null,
      buyInAmount: (value) =>
        value < 0 ? 'Buy-in amount cannot be negative' : null,
      minPlayers: (value) => (value < 1 ? 'Minimum 1 player required' : null),
      maxPlayers: (value) => (value < 1 ? 'Maximum 1 player required' : null),
      rankingPrimary: (value) =>
        !value ? 'Ranking primary is required' : null,
      bankerId: (value) =>
        !value || value <= 0 ? 'Banker must be selected' : null,
    },
  })

  // Update form values when game data changes
  useEffect(() => {
    if (game) {
      adjustForm.setValues({
        gameType: mapGameType(game.gameType, game.durationSeconds),
        currency: game.currency,
        moneyModel: game.moneyModel,
        chipValue: game.chipValue,
        buyInAmount: game.buyInAmount,
        rebuyAllowed: game.rebuyAllowed,
        rebuyPrice: game.rebuyPrice ?? 0,
        maxRebuys: game.maxRebuys ?? 0,
        durationSeconds: game.durationSeconds ?? 0,
        startTime: game.startTime ? new Date(game.startTime) : null,
        minPlayers: game.minPlayers,
        maxPlayers: game.maxPlayers,
        rankingPrimary: game.rankingPrimary,
        rankingSecondary: game.rankingSecondary ?? '',
        bankerId: game.bankerId,
      })
    }
  }, [game])

  const handleAdjustGame = adjustForm.onSubmit(async (values) => {
    const config: Partial<GameConfig> = {
      gameType: values.gameType,
      currency: values.currency,
      moneyModel: values.moneyModel,
      chipValue: values.chipValue,
      buyInAmount: values.buyInAmount,
      rebuyAllowed: values.rebuyAllowed,
      rebuyPrice: values.rebuyAllowed ? values.rebuyPrice : undefined,
      maxRebuys: values.rebuyAllowed ? values.maxRebuys : undefined,
      durationSeconds:
        values.gameType === 'cash_time' && values.durationSeconds > 0
          ? values.durationSeconds
          : undefined,
      startTime: values.startTime ? values.startTime.toISOString() : undefined,
      minPlayers: values.minPlayers,
      maxPlayers: values.maxPlayers,
      rankingPrimary: values.rankingPrimary,
      rankingSecondary: values.rankingSecondary || undefined,
      bankerId: values.bankerId,
    }

    try {
      await updateGame.mutateAsync({ gameId: gameIdNum, config })
      notifications.show({
        title: 'Game updated',
        message: 'The game configuration has been updated successfully.',
        color: 'green',
      })
      setAdjustGameOpen(false)
      refetch()
    } catch (err) {
      if (err instanceof ApiClientError) {
        notifications.show({
          title: 'Failed to update game',
          message: err.message,
          color: 'red',
        })
      } else {
        notifications.show({
          title: 'Failed to update game',
          message: 'An unexpected error occurred.',
          color: 'red',
        })
      }
    }
  })

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

  // Compute bank check for finish game validation
  function computeBankCheck(): GameBankCheck {
    if (!game) {
      return { totalBank: 0, totalPayout: 0, difference: 0, mismatch: false }
    }
    const buyInAmount = game.buyInAmount
    const rebuyPrice = game.rebuyPrice ?? 0
    const chipValue = game.chipValue

    let totalBank = 0
    let totalPayout = 0

    const data = monitorData ?? { game, participants: participants ?? [] }
    const participantList = data.participants.filter(
      (p) => p.status === 'confirmed',
    )

    for (const p of participantList) {
      const invested = p.buyInCount * buyInAmount + p.rebuyCount * rebuyPrice
      totalBank += invested
      if (p.chipsEnd !== undefined) {
        totalPayout += p.chipsEnd * chipValue
      }
    }

    const difference = totalPayout - totalBank
    return {
      totalBank,
      totalPayout,
      difference,
      mismatch: Math.abs(difference) > 0.01,
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
          <Button variant="subtle" size="sm" onClick={() => navigate(-1)}>
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
            <Badge
              color={getStatusColor(game.status)}
              variant="light"
              size="lg"
            >
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

      {/* Game Configuration (Cards for Planned/Active/Canceled) */}
      {game.status !== 'finished' && (
        <GameConfigurationCards
          game={game}
          clubName={club?.name}
          isPlanned={isPlanned}
          canManageGame={canManageGame}
          onAdjustGame={() => setAdjustGameOpen(true)}
        />
      )}

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

      {/* Game Configuration (Finished games) */}
      {game.status === 'finished' && gameResultsData && (
        <GameConfigurationFinished
          game={game}
          bankerName={bankerName}
          gameResults={gameResultsData.results || []}
          isPlanned={isPlanned}
          canManageGame={canManageGame}
          onAdjustGame={() => setAdjustGameOpen(true)}
        />
      )}

      {/* Game Results (finished games only) */}
      {game.status === 'finished' && gameResultsData && (
        <Card mt="lg" padding="lg" radius="md" withBorder>
          <Title order={4} mb="md">
            Results
          </Title>

          {/* Results Table */}
          {!isMobile ? (
            <GameResultsTable results={gameResultsData.results || []} />
          ) : (
            <GameResultsMobile results={gameResultsData.results || []} />
          )}
        </Card>
      )}

      {/* Participants */}
      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Title order={4} mb="md">
          Participants
        </Title>
        {participants && participants.length > 0 ? (
          <Stack gap="sm">
            {participants.map((p) => {
              const isAccepted = p.status === 'accepted'
              const canConfirm = isAccepted && canManageGame

              return (
                <Group key={p.player.id} justify="space-between">
                  <Group gap="sm">
                    <IconUser size={20} />
                    <Stack gap={2}>
                      <Text fw={500}>
                        {[p.player.firstName, p.player.lastName]
                          .filter(Boolean)
                          .join(' ') || p.player.nickname}
                      </Text>
                      <Text size="sm" c="dimmed">
                        {p.player.nickname}
                      </Text>
                    </Stack>
                  </Group>
                  <Group gap="xs">
                    <Badge
                      color={
                        p.status === 'confirmed'
                          ? 'green'
                          : p.status === 'accepted'
                            ? 'yellow'
                            : p.status === 'declined'
                              ? 'red'
                              : 'blue'
                      }
                      variant="light"
                    >
                      {p.status}
                    </Badge>
                    {canConfirm && (
                      <Button
                        size="xs"
                        color="green"
                        onClick={() =>
                          confirmParticipation.mutateAsync({
                            gameId: gameIdNum,
                            playerId: p.player.id,
                          })
                        }
                        loading={confirmParticipation.isPending}
                      >
                        Confirm
                      </Button>
                    )}
                  </Group>
                </Group>
              )
            })}
          </Stack>
        ) : (
          <Text size="sm" c="dimmed">
            No participants yet.
          </Text>
        )}
      </Card>

      {/* Accept/Decline Game Participation (current user only, invited status) */}
      {participants &&
        (() => {
          const currentUserParticipant = participants.find(
            (p) => p.player.id === currentPlayerId,
          )
          const canAcceptDecline =
            currentUserParticipant &&
            currentUserParticipant.status === 'invited'
          return canAcceptDecline ? (
            <Card mt="lg" padding="lg" radius="md" withBorder>
              <Stack gap="sm" align={isMobile ? 'stretch' : 'center'}>
                <Button
                  color="green"
                  leftSection={<IconCheck size={16} />}
                  onClick={() => acceptParticipation.mutateAsync(gameIdNum)}
                  loading={acceptParticipation.isPending}
                  size={isMobile ? 'md' : 'sm'}
                  w={isMobile ? '100%' : 'auto'}
                >
                  Принять участие
                </Button>
                <Button
                  color="red"
                  variant="outline"
                  leftSection={<IconX size={16} />}
                  onClick={() => declineParticipation.mutateAsync(gameIdNum)}
                  loading={declineParticipation.isPending}
                  size={isMobile ? 'md' : 'sm'}
                  w={isMobile ? '100%' : 'auto'}
                >
                  Отказаться
                </Button>
              </Stack>
            </Card>
          ) : null
        })()}

      {/* Active Game Section */}
      {isActive && (
        <>
          {/* Game Timer */}
          <Card mt="lg" padding="lg" radius="md" withBorder>
            <Group justify="space-between" mb="md">
              <Title order={4} mb={0}>
                Game Timer
              </Title>
              <Badge
                color={game.timerNotified ? 'yellow' : 'green'}
                variant="light"
              >
                {game.timerNotified ? 'Timer Expired' : 'Running'}
              </Badge>
            </Group>
            {game.durationSeconds && game.durationSeconds > 0 ? (
              <GameTimer
                startTime={game.startTime}
                durationSeconds={game.durationSeconds}
                timerPausedAt={game.timerPausedAt}
                timerPausedDuration={game.timerPausedDuration}
              />
            ) : (
              <Text size="sm" c="dimmed">
                No time limit set for this game.
              </Text>
            )}
          </Card>

          {/* Game Management Actions (Banker/Owner/Admin only) */}
          {canManageGameActions && (
            <Card mt="lg" padding="lg" radius="md" withBorder>
              <Title order={4} mb="md">
                Game Management
              </Title>
              <Stack gap="sm">
                <Button
                  leftSection={<IconPlayerPlay size={16} />}
                  onClick={() => navigate(`/games/${gameIdNum}/manage`)}
                  fullWidth={isMobile}
                >
                  Войти в игру
                </Button>
              </Stack>
            </Card>
          )}

          {/* Player Game View (Member) */}
          {!canManageGameActions && (
            <Card mt="lg" padding="lg" radius="md" withBorder>
              <Title order={4} mb="md">
                Мои данные
              </Title>
              {participants &&
                (() => {
                  const myParticipant = participants.find(
                    (p) => p.player.id === currentPlayerId,
                  )
                  if (!myParticipant) {
                    return (
                      <Text size="sm" c="dimmed">
                        You are not a participant in this game.
                      </Text>
                    )
                  }
                  return (
                    <Stack gap="sm">
                      <Table
                        variant={isMobile ? 'compact' : 'striped'}
                        style={isMobile ? { fontSize: '0.85rem' } : undefined}
                      >
                        <Table.Thead>
                          <Table.Tr>
                            <Table.Th>Buy-in</Table.Th>
                            <Table.Th>Rebuy</Table.Th>
                            <Table.Th>Invested</Table.Th>
                          </Table.Tr>
                        </Table.Thead>
                        <Table.Tbody>
                          <Table.Tr>
                            <Table.Td>{myParticipant.buyInCount}</Table.Td>
                            <Table.Td>{myParticipant.rebuyCount}</Table.Td>
                            <Table.Td>
                              {Math.round(
                                myParticipant.buyInCount * game.buyInAmount +
                                  myParticipant.rebuyCount *
                                    (game.rebuyPrice ?? 0),
                              )}
                            </Table.Td>
                          </Table.Tr>
                        </Table.Tbody>
                      </Table>

                      <Button
                        leftSection={<IconCoin size={16} />}
                        onClick={() => {
                          setCurrentStackValue(0)
                          setCurrentStackModalOpen(true)
                        }}
                        fullWidth={isMobile}
                      >
                        Ввести текущий стек
                      </Button>

                      <Title order={5} mb="xs">
                        Данные соперников
                      </Title>
                      <Table
                        variant={isMobile ? 'compact' : 'striped'}
                        style={isMobile ? { fontSize: '0.85rem' } : undefined}
                      >
                        <Table.Thead>
                          <Table.Tr>
                            <Table.Th>Игрок</Table.Th>
                            <Table.Th>Rebuy</Table.Th>
                            <Table.Th>Invested</Table.Th>
                            <Table.Th>Chips End</Table.Th>
                          </Table.Tr>
                        </Table.Thead>
                        <Table.Tbody>
                          {participants
                            .filter(
                              (p) =>
                                p.player.id !== currentPlayerId &&
                                p.status === 'confirmed',
                            )
                            .map((p) => (
                              <Table.Tr key={p.player.id}>
                                <Table.Td>
                                  {p.player.tgUserID || p.player.id}
                                </Table.Td>
                                <Table.Td>
                                  {p.rebuyCount} /{' '}
                                  {Math.round(
                                    p.rebuyCount * (game.rebuyPrice ?? 0),
                                  )}
                                </Table.Td>
                                <Table.Td>
                                  {Math.round(
                                    p.buyInCount * game.buyInAmount +
                                      p.rebuyCount * (game.rebuyPrice ?? 0),
                                  )}
                                </Table.Td>
                                <Table.Td>
                                  {p.chipsEnd !== undefined
                                    ? Math.round(p.chipsEnd)
                                    : '—'}
                                </Table.Td>
                              </Table.Tr>
                            ))}
                        </Table.Tbody>
                      </Table>
                    </Stack>
                  )
                })()}
            </Card>
          )}
        </>
      )}

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

      {/* Adjust Game Modal */}
      <Modal
        opened={adjustGameOpen}
        onClose={() => setAdjustGameOpen(false)}
        title="Adjust Game"
        size="lg"
        centered
      >
        <form onSubmit={handleAdjustGame}>
          <Stack gap="md">
            {/* Game Type */}
            <Select
              label="Game Type"
              placeholder="Select game type"
              data={gameTypeOptions}
              value={adjustForm.values.gameType}
              onChange={(value) =>
                adjustForm.setFieldValue(
                  'gameType',
                  (value ?? 'cash_open') as GameType,
                )
              }
              required
              error={adjustForm.errors.gameType}
              leftSection={<IconChessKing size={16} />}
            />

            {/* Banker */}
            <Select
              label="Banker"
              placeholder="Select banker"
              data={bankerOptions}
              value={
                adjustForm.values.bankerId
                  ? String(adjustForm.values.bankerId)
                  : null
              }
              onChange={(value) =>
                adjustForm.setFieldValue(
                  'bankerId',
                  value ? parseInt(value, 10) : 0,
                )
              }
              required
              error={adjustForm.errors.bankerId}
              leftSection={<IconUser size={16} />}
              disabled={bankerOptions.length === 0}
            />

            {/* Currency */}
            <Select
              label="Currency"
              placeholder="Select currency"
              data={currencyOptions}
              value={adjustForm.values.currency}
              onChange={(value) =>
                adjustForm.setFieldValue('currency', value ?? 'RUB')
              }
              required
              error={adjustForm.errors.currency}
              leftSection={<IconCurrencyDollar size={16} />}
            />

            {/* Money Model */}
            <Select
              label="Money Model"
              placeholder="Select money model"
              data={moneyModelOptions}
              value={adjustForm.values.moneyModel}
              onChange={(value) =>
                adjustForm.setFieldValue('moneyModel', value ?? 'real')
              }
              required
              error={adjustForm.errors.moneyModel}
            />

            {/* Chip Value */}
            <NumberInput
              label="Chip Value"
              placeholder="1.00"
              value={adjustForm.values.chipValue}
              onChange={(value) =>
                adjustForm.setFieldValue('chipValue', Number(value) || 0)
              }
              required
              error={adjustForm.errors.chipValue}
              leftSection={<IconCoin size={16} />}
              decimalScale={2}
              allowNegative={false}
            />

            {/* Buy-in Amount */}
            <NumberInput
              label="Buy-in Amount"
              placeholder="0.00"
              value={adjustForm.values.buyInAmount}
              onChange={(value) =>
                adjustForm.setFieldValue('buyInAmount', Number(value) || 0)
              }
              required
              error={adjustForm.errors.buyInAmount}
              leftSection={<IconCash size={16} />}
              decimalScale={2}
              allowNegative={false}
            />

            {/* Rebuy Allowed */}
            <Checkbox
              label="Allow Rebuy"
              description="Enable rebuy for this game"
              checked={adjustForm.values.rebuyAllowed}
              onChange={(e) =>
                adjustForm.setFieldValue(
                  'rebuyAllowed',
                  e.currentTarget.checked,
                )
              }
            />

            {/* Rebuy Price (conditional) */}
            {adjustForm.values.rebuyAllowed && (
              <>
                <NumberInput
                  label="Rebuy Price"
                  placeholder="0.00"
                  value={adjustForm.values.rebuyPrice}
                  onChange={(value) =>
                    adjustForm.setFieldValue('rebuyPrice', Number(value) || 0)
                  }
                  decimalScale={2}
                  allowNegative={false}
                  leftSection={<IconCash size={16} />}
                />

                <NumberInput
                  label="Max Rebuys"
                  placeholder="0"
                  value={adjustForm.values.maxRebuys}
                  onChange={(value) =>
                    adjustForm.setFieldValue('maxRebuys', Number(value) || 0)
                  }
                  description="0 = unlimited"
                  allowNegative={false}
                />
              </>
            )}

            {/* Duration (only for cash_time) */}
            {adjustForm.values.gameType === 'cash_time' && (
              <NumberInput
                label="Duration (seconds)"
                placeholder="1800"
                value={adjustForm.values.durationSeconds}
                onChange={(value) =>
                  adjustForm.setFieldValue(
                    'durationSeconds',
                    Number(value) || 0,
                  )
                }
                description="Game duration in seconds. 0 = no time limit."
                leftSection={<IconClock size={16} />}
                allowNegative={false}
              />
            )}

            {/* Start Time */}
            <DateTimePicker
              label="Start Time"
              placeholder="Select date and time"
              value={adjustForm.values.startTime}
              onChange={(value) => {
                if (value === null) {
                  adjustForm.setFieldValue('startTime', null)
                } else if (typeof value === 'string') {
                  const parsed = new Date(value)
                  adjustForm.setFieldValue(
                    'startTime',
                    isNaN(parsed.getTime()) ? null : parsed,
                  )
                } else {
                  adjustForm.setFieldValue('startTime', value)
                }
              }}
              valueFormat="DD.MM.YYYY HH:mm"
              defaultTimeValue="19:00"
              leftSection={<IconCalendar size={16} />}
              clearable
              nextIcon={<IconChevronRight size={14} />}
              previousIcon={<IconChevronLeft size={14} />}
              timePickerProps={{
                withDropdown: true,
                format: '24h',
                hoursStep: 1,
                minutesStep: 15,
              }}
              styles={{
                day: {
                  '&[data-weekend]': {
                    color: 'var(--mantine-color-red-6)',
                  },
                  '&[data-outside]': {
                    opacity: 'var(--mantine-color-gray-3)',
                  },
                },
              }}
            />

            {/* Min/Max Players */}
            <Grid>
              <Grid.Col span={6}>
                <NumberInput
                  label="Min Players"
                  placeholder="2"
                  value={adjustForm.values.minPlayers}
                  onChange={(value) =>
                    adjustForm.setFieldValue('minPlayers', Number(value) || 2)
                  }
                  required
                  error={adjustForm.errors.minPlayers}
                  allowNegative={false}
                />
              </Grid.Col>
              <Grid.Col span={6}>
                <NumberInput
                  label="Max Players"
                  placeholder="10"
                  value={adjustForm.values.maxPlayers}
                  onChange={(value) =>
                    adjustForm.setFieldValue('maxPlayers', Number(value) || 10)
                  }
                  required
                  error={adjustForm.errors.maxPlayers}
                  allowNegative={false}
                />
              </Grid.Col>
            </Grid>

            {/* Ranking */}
            <Select
              label="Ranking Primary"
              placeholder="Select primary ranking"
              data={rankingPrimaryOptions}
              value={adjustForm.values.rankingPrimary}
              onChange={(value) =>
                adjustForm.setFieldValue('rankingPrimary', value ?? 'profit')
              }
              required
              error={adjustForm.errors.rankingPrimary}
            />

            <Select
              label="Ranking Secondary (optional)"
              placeholder="Select secondary ranking"
              data={rankingSecondaryOptions}
              value={adjustForm.values.rankingSecondary || null}
              onChange={(value) =>
                adjustForm.setFieldValue('rankingSecondary', value ?? '')
              }
              clearable
            />

            {/* Form actions */}
            <Group justify="flex-end" gap="sm">
              <Button
                variant="subtle"
                onClick={() => setAdjustGameOpen(false)}
                disabled={updateGame.isPending}
              >
                Cancel
              </Button>
              <Button type="submit" loading={updateGame.isPending}>
                Save
              </Button>
            </Group>
          </Stack>
        </form>
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

      {/* Set Current Stack Modal (Member) */}
      <Modal
        opened={currentStackModalOpen}
        onClose={() => setCurrentStackModalOpen(false)}
        title="Ввести текущий стек"
        centered
        size="sm"
      >
        <Stack gap="md">
          <NumberInput
            label="Current Stack"
            placeholder="0"
            value={currentStackValue}
            onChange={(value) => setCurrentStackValue(Number(value) || 0)}
            allowNegative={false}
            min={0}
            leftSection={<IconCoin size={16} />}
          />
          <Group justify="flex-end" gap="sm">
            <Button
              variant="subtle"
              onClick={() => setCurrentStackModalOpen(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={async () => {
                try {
                  await setCurrentStack.mutateAsync({
                    gameId: gameIdNum,
                    stack: currentStackValue,
                  })
                  notifications.show({
                    title: 'Current Stack Set',
                    message: 'Current stack has been recorded.',
                    color: 'green',
                  })
                  setCurrentStackModalOpen(false)
                  refetch()
                } catch (err) {
                  if (err instanceof ApiClientError) {
                    notifications.show({
                      title: 'Failed to set current stack',
                      message: err.message,
                      color: 'red',
                    })
                  }
                }
              }}
              loading={setCurrentStack.isPending}
            >
              Confirm
            </Button>
          </Group>
        </Stack>
      </Modal>

      {/* Finish Game Confirmation Modal */}
      <Modal
        opened={finishGameModalOpen}
        onClose={() => setFinishGameModalOpen(false)}
        title="Finish Game"
        centered
        size="md"
      >
        <Stack gap="md">
          {(() => {
            const bankCheck = computeBankCheck()
            if (bankCheck.mismatch) {
              return (
                <>
                  <Alert color="red" variant="light" title="Balance Mismatch">
                    <Stack gap="xs">
                      <Text>Invested: {Math.round(bankCheck.totalBank)}</Text>
                      <Text>
                        Chips End: {Math.round(bankCheck.totalPayout)}
                      </Text>
                      <Text>
                        Difference: {Math.round(bankCheck.difference)}
                      </Text>
                    </Stack>
                  </Alert>
                  <Text size="sm">
                    The game can still be finished, but the results may be
                    incorrect.
                  </Text>
                  <Group justify="flex-end" gap="sm">
                    <Button
                      variant="subtle"
                      onClick={() => setFinishGameModalOpen(false)}
                    >
                      Return to game
                    </Button>
                    <Button
                      color="red"
                      onClick={async () => {
                        try {
                          await finishGame.mutateAsync(gameIdNum)
                          notifications.show({
                            title: 'Game Finished',
                            message: 'The game has been finished successfully.',
                            color: 'green',
                          })
                          setFinishGameModalOpen(false)
                          //navigate(`/games/${gameIdNum}`)
                        } catch (err) {
                          if (err instanceof ApiClientError) {
                            notifications.show({
                              title: 'Failed to finish game',
                              message: err.message,
                              color: 'red',
                            })
                          }
                        }
                      }}
                      loading={finishGame.isPending}
                    >
                      Finish anyway
                    </Button>
                  </Group>
                </>
              )
            }
            return (
              <>
                <Alert
                  color="green"
                  variant="light"
                  title="Game Balance Correct"
                >
                  <Stack gap="xs">
                    <Text>
                      Total Invested: {Math.round(bankCheck.totalBank)}
                    </Text>
                    <Text>
                      Total Chips End: {Math.round(bankCheck.totalPayout)}
                    </Text>
                    <Text>
                      Total Payout: {Math.round(bankCheck.totalPayout)}
                    </Text>
                  </Stack>
                </Alert>
                <Group justify="flex-end" gap="sm">
                  <Button
                    variant="subtle"
                    onClick={() => setFinishGameModalOpen(false)}
                  >
                    Cancel
                  </Button>
                  <Button
                    color="green"
                    onClick={async () => {
                      try {
                        await finishGame.mutateAsync(gameIdNum)
                        notifications.show({
                          title: 'Game Finished',
                          message: 'The game has been finished successfully.',
                          color: 'green',
                        })
                        setFinishGameModalOpen(false)
                        //navigate(`/games/${gameIdNum}`)
                      } catch (err) {
                        if (err instanceof ApiClientError) {
                          notifications.show({
                            title: 'Failed to finish game',
                            message: err.message,
                            color: 'red',
                          })
                        }
                      }
                    }}
                    loading={finishGame.isPending}
                  >
                    Finish Game
                  </Button>
                </Group>
              </>
            )
          })()}
        </Stack>
      </Modal>

      {/* Add Rebuy Modal */}
      <Modal
        opened={rebuyPlayerId !== null}
        onClose={() => setRebuyPlayerId(null)}
        title="Add Rebuy"
        centered
        size="sm"
      >
        <Stack gap="md">
          <Text>Add a rebuy for this player?</Text>
          <Group justify="flex-end" gap="sm">
            <Button variant="subtle" onClick={() => setRebuyPlayerId(null)}>
              Cancel
            </Button>
            <Button
              color="green"
              onClick={async () => {
                if (rebuyPlayerId) {
                  try {
                    await registerRebuy.mutateAsync({
                      gameId: gameIdNum,
                      playerId: rebuyPlayerId,
                    })
                    notifications.show({
                      title: 'Rebuy Registered',
                      message: 'Rebuy has been registered.',
                      color: 'green',
                    })
                    setRebuyPlayerId(null)
                    refetch()
                  } catch (err) {
                    if (err instanceof ApiClientError) {
                      notifications.show({
                        title: 'Failed to register rebuy',
                        message: err.message,
                        color: 'red',
                      })
                    }
                  }
                }
              }}
              loading={registerRebuy.isPending}
            >
              Confirm
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

/**
 * Game timer component for Active Game.
 * Displays a countdown based on the game's start time and duration.
 * Backend is the source of truth; this is a local visual countdown.
 */
function GameTimer({
  startTime,
  durationSeconds,
  timerPausedAt,
  timerPausedDuration,
}: {
  startTime: string
  durationSeconds: number
  timerPausedAt?: string
  timerPausedDuration?: number
}) {
  const [timeLeft, setTimeLeft] = useState<string>('')

  useEffect(() => {
    const updateTimer = () => {
      const start = new Date(startTime).getTime()
      const now = Date.now()
      const elapsed = (now - start) / 1000

      // Account for paused duration
      let pausedDuration = 0
      if (timerPausedDuration) {
        pausedDuration = timerPausedDuration
      }
      // If timer is currently paused, add time since pause
      if (timerPausedAt) {
        const pauseStart = new Date(timerPausedAt).getTime()
        pausedDuration += (now - pauseStart) / 1000
      }

      const remaining = durationSeconds - elapsed + pausedDuration
      if (remaining <= 0) {
        setTimeLeft('00:00')
      } else {
        const mins = Math.floor(remaining / 60)
        const secs = Math.floor(remaining % 60)
        setTimeLeft(
          `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`,
        )
      }
    }

    updateTimer()
    const interval = setInterval(updateTimer, 1000)
    return () => clearInterval(interval)
  }, [startTime, durationSeconds, timerPausedAt, timerPausedDuration])

  return (
    <Group gap="sm" align="center">
      <IconClock size={24} />
      <Text size="xl" fw={700} ff="monospace">
        {timeLeft}
      </Text>
      <Text size="sm" c="dimmed">
        remaining
      </Text>
    </Group>
  )
}
