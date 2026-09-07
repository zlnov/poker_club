/**
 * Club games page component.
 *
 * Displays the list of games for a club and allows creating new games.
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
  Modal,
  Select,
  TextInput,
  NumberInput,
  Checkbox,
  Grid,
} from '@mantine/core'
import {
  IconChessKing,
  IconCash,
  IconTrophy,
  IconClock,
  IconPlus,
  IconCalendar,
  IconCurrencyDollar,
  IconCoin,
  IconUser,
} from '@tabler/icons-react'
import { useParams, useNavigate } from 'react-router-dom'
import { useForm } from '@mantine/form'
import {
  PageContainer,
  PageHeader,
  EmptyState,
  LoadingState,
  ErrorState,
} from '../components/ui'
import { useClubGames, useCreateGame } from '../features'
import { useClub, useClubMembers } from '../features'
import { formatCurrency, formatDate } from '../utils'
import { ApiClientError } from '../api'
import { notifications } from '@mantine/notifications'
import type { GameStatus, GameType, GameConfig, ClubMemberRole } from '../types'

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

// --- Game type options for the form ---

const gameTypeOptions: { value: GameType; label: string }[] = [
  { value: 'cash_time', label: 'Cash (Timed)' },
  { value: 'cash_open', label: 'Cash (Open)' },
  { value: 'tournament', label: 'Tournament' },
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
  startTime: string
  minPlayers: number
  maxPlayers: number
  rankingPrimary: string
  rankingSecondary: string
  bankerId: number
}

/**
 * Formats a player's display name from a ClubMemberRole.
 */
function formatMemberName(member: ClubMemberRole): string {
  const parts = [member.firstName, member.lastName].filter(Boolean)
  return parts.length > 0 ? parts.join(' ') : member.nickname || 'Unknown'
}

/**
 * Club games page.
 *
 * Displays:
 * - List of games for the club
 * - Create game modal with full configuration form
 * - Loading, empty, and error states
 */
export function ClubGames() {
  const { clubId } = useParams<{ clubId: string }>()
  const navigate = useNavigate()
  const clubIdNum = clubId ? parseInt(clubId, 10) : 0

  const [createModalOpen, setCreateModalOpen] = useState(false)

  const { data: club } = useClub(clubIdNum)
  const isOwner = club?.isOwner ?? false
  const isAdmin = club?.isAdmin ?? false
  const canCreateGame = isOwner || isAdmin

  const {
    data: games,
    isLoading,
    isError,
    error,
    refetch,
  } = useClubGames(clubIdNum)

  const {
    data: members,
    isLoading: membersLoading,
  } = useClubMembers(clubIdNum)

  const {
    mutate: createGame,
    isPending: isCreating,
  } = useCreateGame()

  // Build banker select options from active club members
  const bankerOptions = (members ?? [])
    .filter((m) => m.status === 'active')
    .map((m) => ({
      value: String(m.playerId),
      label: formatMemberName(m),
    }))

  const form = useForm<GameFormValues>({
    initialValues: {
      gameType: 'cash_open',
      currency: 'USD',
      moneyModel: 'real',
      chipValue: 1,
      buyInAmount: 0,
      rebuyAllowed: false,
      rebuyPrice: 0,
      maxRebuys: 0,
      durationSeconds: 0,
      startTime: '',
      minPlayers: 2,
      maxPlayers: 10,
      rankingPrimary: 'profit',
      rankingSecondary: 'roi',
      bankerId: 0,
    },
    validate: {
      gameType: (value) => (!value ? 'Game type is required' : null),
      currency: (value) =>
        !value || value.trim().length < 3
          ? 'Currency code is required (e.g. USD)'
          : null,
      moneyModel: (value) => (!value ? 'Money model is required' : null),
      chipValue: (value) =>
        value <= 0 ? 'Chip value must be greater than 0' : null,
      buyInAmount: (value) =>
        value < 0 ? 'Buy-in amount cannot be negative' : null,
      minPlayers: (value) =>
        value < 1 ? 'Minimum 1 player required' : null,
      maxPlayers: (value) =>
        value < 1 ? 'Maximum 1 player required' : null,
      rankingPrimary: (value) => (!value ? 'Ranking primary is required' : null),
      bankerId: (value) =>
        !value || value <= 0 ? 'Banker must be selected' : null,
    },
  })

  // Update ranking defaults when game type changes
  const handleGameTypeChange = (value: GameType | null) => {
    if (value) {
      form.setFieldValue('gameType', value)
      if (value === 'tournament') {
        form.setFieldValue('rankingPrimary', 'place')
      } else {
        form.setFieldValue('rankingPrimary', 'profit')
      }
    }
  }

  const handleCreateGame = form.onSubmit(async (values) => {
    const config: GameConfig = {
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
      startTime: values.startTime || undefined,
      minPlayers: values.minPlayers,
      maxPlayers: values.maxPlayers,
      rankingPrimary: values.rankingPrimary,
      rankingSecondary: values.rankingSecondary || undefined,
      bankerId: values.bankerId,
    }

    try {
      await createGame({ clubId: clubIdNum, config })
      notifications.show({
        title: 'Game created',
        message: 'The game has been created successfully.',
        color: 'green',
      })
      setCreateModalOpen(false)
      form.reset()
      refetch()
    } catch (err) {
      if (err instanceof ApiClientError) {
        notifications.show({
          title: 'Failed to create game',
          message: err.message,
          color: 'red',
        })
      } else {
        notifications.show({
          title: 'Failed to create game',
          message: 'An unexpected error occurred.',
          color: 'red',
        })
      }
    }
  })

  const selectedGameType = form.values.gameType

  if (isLoading || membersLoading) {
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

  return (
    <PageContainer>
      <PageHeader
        title="Games"
        description="Club games"
        action={
          canCreateGame && (
            <Button
              leftSection={<IconPlus size={16} />}
              onClick={() => setCreateModalOpen(true)}
            >
              Create Game
            </Button>
          )
        }
      />

      {!games || games.length === 0 ? (
        <EmptyState
          title="No games yet"
          description="No games have been created for this club yet."
          actionLabel={canCreateGame ? 'Create Game' : undefined}
          onAction={canCreateGame ? () => setCreateModalOpen(true) : undefined}
          icon={<IconChessKing size={48} stroke={1.5} />}
        />
      ) : (
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
                      {getGameTypeLabel(game.type)}
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
                  <Badge color={getGameTypeColor(game.type)} variant="light">
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
      )}

      {/* Create Game Modal */}
      <Modal
        opened={createModalOpen}
        onClose={() => setCreateModalOpen(false)}
        title="Create Game"
        size="lg"
        centered
      >
        <form onSubmit={handleCreateGame}>
          <Stack gap="md">
            {/* Game Type */}
            <Select
              label="Game Type"
              placeholder="Select game type"
              data={gameTypeOptions}
              value={form.values.gameType}
              onChange={(value) =>
                handleGameTypeChange(value as GameType | null)
              }
              required
              error={form.errors.gameType}
              leftSection={<IconChessKing size={16} />}
            />

            {/* Banker */}
            <Select
              label="Banker"
              placeholder="Select banker"
              data={bankerOptions}
              value={form.values.bankerId ? String(form.values.bankerId) : null}
              onChange={(value) =>
                form.setFieldValue('bankerId', value ? parseInt(value, 10) : 0)
              }
              required
              error={form.errors.bankerId}
              leftSection={<IconUser size={16} />}
              disabled={bankerOptions.length === 0}
            />

            {/* Currency */}
            <TextInput
              label="Currency"
              placeholder="USD"
              value={form.values.currency}
              onChange={(e) => form.setFieldValue('currency', e.target.value)}
              required
              error={form.errors.currency}
              leftSection={<IconCurrencyDollar size={16} />}
            />

            {/* Money Model */}
            <Select
              label="Money Model"
              placeholder="Select money model"
              data={moneyModelOptions}
              value={form.values.moneyModel}
              onChange={(value) =>
                form.setFieldValue('moneyModel', value ?? 'real')
              }
              required
              error={form.errors.moneyModel}
            />

            {/* Chip Value */}
            <NumberInput
              label="Chip Value"
              placeholder="1.00"
              value={form.values.chipValue}
              onChange={(value) =>
                form.setFieldValue('chipValue', Number(value) || 0)
              }
              required
              error={form.errors.chipValue}
              leftSection={<IconCoin size={16} />}
              decimalScale={2}
              allowNegative={false}
            />

            {/* Buy-in Amount */}
            <NumberInput
              label="Buy-in Amount"
              placeholder="0.00"
              value={form.values.buyInAmount}
              onChange={(value) =>
                form.setFieldValue('buyInAmount', Number(value) || 0)
              }
              required
              error={form.errors.buyInAmount}
              leftSection={<IconCash size={16} />}
              decimalScale={2}
              allowNegative={false}
            />

            {/* Rebuy Allowed */}
            <Checkbox
              label="Allow Rebuy"
              description="Enable rebuy for this game"
              checked={form.values.rebuyAllowed}
              onChange={(e) =>
                form.setFieldValue('rebuyAllowed', e.currentTarget.checked)
              }
            />

            {/* Rebuy Price (conditional) */}
            {form.values.rebuyAllowed && (
              <>
                <NumberInput
                  label="Rebuy Price"
                  placeholder="0.00"
                  value={form.values.rebuyPrice}
                  onChange={(value) =>
                    form.setFieldValue('rebuyPrice', Number(value) || 0)
                  }
                  decimalScale={2}
                  allowNegative={false}
                  leftSection={<IconCash size={16} />}
                />

                <NumberInput
                  label="Max Rebuys"
                  placeholder="0"
                  value={form.values.maxRebuys}
                  onChange={(value) =>
                    form.setFieldValue('maxRebuys', Number(value) || 0)
                  }
                  description="0 = unlimited"
                  allowNegative={false}
                />
              </>
            )}

            {/* Duration (only for cash_time) */}
            {selectedGameType === 'cash_time' && (
              <NumberInput
                label="Duration (seconds)"
                placeholder="1800"
                value={form.values.durationSeconds}
                onChange={(value) =>
                  form.setFieldValue('durationSeconds', Number(value) || 0)
                }
                description="Game duration in seconds. 0 = no time limit."
                leftSection={<IconClock size={16} />}
                allowNegative={false}
              />
            )}

            {/* Start Time */}
            <TextInput
              label="Start Time"
              placeholder="2024-01-15T19:30:00"
              value={form.values.startTime}
              onChange={(e) => form.setFieldValue('startTime', e.target.value)}
              description="ISO 8601 format (optional)"
              leftSection={<IconCalendar size={16} />}
            />

            {/* Min/Max Players */}
            <Grid>
              <Grid.Col span={6}>
                <NumberInput
                  label="Min Players"
                  placeholder="2"
                  value={form.values.minPlayers}
                  onChange={(value) =>
                    form.setFieldValue('minPlayers', Number(value) || 2)
                  }
                  required
                  error={form.errors.minPlayers}
                  allowNegative={false}
                />
              </Grid.Col>
              <Grid.Col span={6}>
                <NumberInput
                  label="Max Players"
                  placeholder="10"
                  value={form.values.maxPlayers}
                  onChange={(value) =>
                    form.setFieldValue('maxPlayers', Number(value) || 10)
                  }
                  required
                  error={form.errors.maxPlayers}
                  allowNegative={false}
                />
              </Grid.Col>
            </Grid>

            {/* Ranking */}
            <Select
              label="Ranking Primary"
              placeholder="Select primary ranking"
              data={rankingPrimaryOptions}
              value={form.values.rankingPrimary}
              onChange={(value) =>
                form.setFieldValue('rankingPrimary', value ?? 'profit')
              }
              required
              error={form.errors.rankingPrimary}
            />

            <Select
              label="Ranking Secondary (optional)"
              placeholder="Select secondary ranking"
              data={rankingSecondaryOptions}
              value={form.values.rankingSecondary || null}
              onChange={(value) =>
                form.setFieldValue('rankingSecondary', value ?? '')
              }
              clearable
            />

            {/* Form actions */}
            <Group justify="flex-end" gap="sm">
              <Button
                variant="subtle"
                onClick={() => setCreateModalOpen(false)}
                disabled={isCreating}
              >
                Cancel
              </Button>
              <Button type="submit" loading={isCreating}>
                Create Game
              </Button>
            </Group>
          </Stack>
        </form>
      </Modal>
    </PageContainer>
  )
}
