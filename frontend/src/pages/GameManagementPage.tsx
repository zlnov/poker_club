/**
 * Game Management page for Banker/Owner/Admin.
 *
 * Phase 6 scope (RM_FE_6.md):
 * - Active Game management
 * - Participants, Buy-in, Rebuy, Chips, Timer
 * - Finish Game
 *
 * This page is accessible only to Banker/Owner/Admin.
 * It provides the full game management interface with:
 * - Actions block (Players, Rebuy, Final stacks)
 * - Summary table with TOTAL row
 * - Finish Game button
 *
 * See 05_FE_UX.md section 7 (Game Navigation) and
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
  Table,
  Modal,
  Alert,
  Grid,
  Accordion,
  NumberInput,
} from '@mantine/core'
import {
  IconChevronRight,
  IconTrophy,
} from '@tabler/icons-react'
import { useParams, useNavigate } from 'react-router-dom'
import {
  PageContainer,
  PageHeader,
} from '../components/ui'
import {
  useGame,
  useGameMonitor,
  useGameParticipants,
  useClub,
  useClubMembers,
  useFinishGame,
  useSetChipsEnd,
  useRemoveGameParticipant,
  useAddGameParticipant,
} from '../features'
import { formatDateTime } from '../utils'
import { ApiClientError } from '../api'
import { notifications } from '@mantine/notifications'
import { useMediaQuery } from '@mantine/hooks'
import type { GameBankCheck, ClubMemberRole } from '../types'

/**
 * Formats a player's display name from a ClubMemberRole.
 */
function formatMemberName(member: ClubMemberRole): string {
  const parts = [member.firstName, member.lastName].filter(Boolean)
  return parts.length > 0 ? parts.join(' ') : member.nickname || 'Unknown'
}

/**
 * Game Management page.
 *
 * Displays:
 * - Actions block with Players, Rebuy, Final stacks sub-sections
 * - Summary table with TOTAL row
 * - Finish Game button
 */
export function GameManagementPage() {
  const { gameId } = useParams<{ gameId: string }>()
  const navigate = useNavigate()
  const gameIdNum = gameId ? parseInt(gameId, 10) : 0

  const { data: game, refetch: refetchGame } = useGame(gameIdNum)
  const { data: monitorData } = useGameMonitor(gameIdNum)
  const { data: participants, refetch: refetchParticipants } =
    useGameParticipants(gameIdNum)
  const { data: club } = useClub(game?.clubId ?? 0)
  const { data: members } = useClubMembers(game?.clubId ?? 0)

  const finishGame = useFinishGame()
  const setChipsEnd = useSetChipsEnd()
  const removeParticipant = useRemoveGameParticipant()
  const addParticipant = useAddGameParticipant()

  const isMobile = useMediaQuery('(max-width: 768px)')

  // Resolve banker name from club members
  const bankerMember = members?.find(
    (m) => m.clubMemberId === game?.bankerId,
  )
  const bankerName = bankerMember
    ? formatMemberName(bankerMember)
    : `Banker #${game?.bankerId ?? '—'}`

  // Modal states
  const [finishGameModalOpen, setFinishGameModalOpen] = useState(false)
  const [removePlayerId, setRemovePlayerId] = useState<number | null>(null)
  const [addPlayerId, setAddPlayerId] = useState<number | null>(null)
  const [chipsModalOpen, setChipsModalOpen] = useState(false)
  const [chipsEndPlayerId, setChipsEndPlayerId] = useState<number | null>(null)
  const [chipsEndValue, setChipsEndValue] = useState<number>(0)

  // Use monitor data if available (has full participant info), otherwise use participants
  const displayParticipants = monitorData?.participants ?? participants ?? []
  // Only confirmed participants are shown in game data tables
  const confirmedParticipants = displayParticipants.filter((p) => p.status === 'confirmed')

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

    for (const p of confirmedParticipants) {
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

  const handleFinishGame = async () => {
    try {
      await finishGame.mutateAsync(gameIdNum)
      notifications.show({
        title: 'Game Finished',
        message: 'The game has been finished successfully.',
        color: 'green',
      })
      setFinishGameModalOpen(false)
      refetchGame()
      refetchParticipants()
    } catch (err) {
      if (err instanceof ApiClientError) {
        notifications.show({
          title: 'Failed to finish game',
          message: err.message,
          color: 'red',
        })
      } else {
        notifications.show({
          title: 'Failed to finish game',
          message: 'An unexpected error occurred.',
          color: 'red',
        })
      }
    }
  }

  const handleSetChipsEnd = async () => {
    if (!chipsEndPlayerId) return
    try {
      await setChipsEnd.mutateAsync({
        gameId: gameIdNum,
        playerId: chipsEndPlayerId,
        chipsEnd: chipsEndValue,
      })
      notifications.show({
        title: 'Chips End Set',
        message: 'Chips end has been set.',
        color: 'green',
      })
      setChipsEndPlayerId(null)
      setChipsModalOpen(false)
      refetchParticipants()
    } catch (err) {
      if (err instanceof ApiClientError) {
        notifications.show({
          title: 'Failed to set chips',
          message: err.message,
          color: 'red',
        })
      }
    }
  }

  const handleRemoveParticipant = async () => {
    if (!removePlayerId) return
    try {
      await removeParticipant.mutateAsync({
        gameId: gameIdNum,
        playerId: removePlayerId,
      })
      notifications.show({
        title: 'Player Removed',
        message: 'Player has been removed from the game.',
        color: 'green',
      })
      setRemovePlayerId(null)
      refetchParticipants()
    } catch (err) {
      if (err instanceof ApiClientError) {
        notifications.show({
          title: 'Failed to remove player',
          message: err.message,
          color: 'red',
        })
      }
    }
  }

  const handleAddParticipant = async () => {
    if (!addPlayerId) return
    try {
      await addParticipant.mutateAsync({
        gameId: gameIdNum,
        playerId: addPlayerId,
      })
      notifications.show({
        title: 'Player Added',
        message: 'Player has been added to the game.',
        color: 'green',
      })
      setAddPlayerId(null)
      refetchParticipants()
    } catch (err) {
      if (err instanceof ApiClientError) {
        notifications.show({
          title: 'Failed to add player',
          message: err.message,
          color: 'red',
        })
      }
    }
  }

  if (!game) {
    return (
      <PageContainer>
        <PageHeader title="Game Management" description="Loading..." />
      </PageContainer>
    )
  }

  const bankCheck = computeBankCheck()

  return (
    <PageContainer>
      <PageHeader
        title="Game Management"
        description={game.gameType}
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

      {/* Game Info */}
      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Title order={4} mb="md">
          Game Info
        </Title>
        <Stack gap="sm">
          <Group gap="sm" justify="space-between">
            <Text size="sm" c="dimmed">Club</Text>
            <Text size="sm" fw={500}>{club?.name ?? '—'}</Text>
          </Group>
          <Group gap="sm" justify="space-between">
            <Text size="sm" c="dimmed">Game ID</Text>
            <Text size="sm" fw={500}>{game.id}</Text>
          </Group>
          <Group gap="sm" justify="space-between">
            <Text size="sm" c="dimmed">Type</Text>
            <Text size="sm" fw={500}>{game.gameType}</Text>
          </Group>
          <Group gap="sm" justify="space-between">
            <Text size="sm" c="dimmed">Banker</Text>
            <Text size="sm" fw={500}>
              {bankerName}
            </Text>
          </Group>
          <Group gap="sm" justify="space-between">
            <Text size="sm" c="dimmed">Status</Text>
            <Badge color="green" variant="light">{game.status}</Badge>
          </Group>
        </Stack>
      </Card>

      {/* Game Configuration (preview) */}
      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Title order={4} mb="md">
          Game Configuration
        </Title>
        <Grid>
          <Grid.Col span={isMobile ? 6 : 3}>
            <Stack gap="xs" align="center">
              <Text size="xs" c="dimmed">Buy-in</Text>
              <Text size="sm" fw={500}>
                {Math.round(game.buyInAmount)}
              </Text>
            </Stack>
          </Grid.Col>
          <Grid.Col span={isMobile ? 6 : 3}>
            <Stack gap="xs" align="center">
              <Text size="xs" c="dimmed">Rebuy</Text>
              <Text size="sm" fw={500}>
                {game.rebuyAllowed
                  ? Math.round(game.rebuyPrice ?? 0)
                  : 'Not allowed'}
              </Text>
            </Stack>
          </Grid.Col>
          <Grid.Col span={isMobile ? 6 : 3}>
            <Stack gap="xs" align="center">
              <Text size="xs" c="dimmed">Participants</Text>
              <Text size="sm" fw={500}>{displayParticipants.length}</Text>
            </Stack>
          </Grid.Col>
          <Grid.Col span={isMobile ? 6 : 3}>
            <Stack gap="xs" align="center">
              <Text size="xs" c="dimmed">Start Time</Text>
              <Text size="sm" fw={500}>{formatDateTime(game.startTime)}</Text>
            </Stack>
          </Grid.Col>
        </Grid>
      </Card>

      {/* Actions Block */}
      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Title order={4} mb="md">
          Actions
        </Title>
        <Stack gap="sm">
          {/* Players accordion */}
          <Accordion
            variant="contained"
            radius="sm"
            styles={{
              content: { padding: '0 0 0 0' },
              item: { width: '100%' },
            }}
          >
            <Accordion.Item value="players">
              <Accordion.Control
                chevron={<IconChevronRight size={16} stroke={2} />}
              >
                <Text fw={500}>Players</Text>
              </Accordion.Control>
              <Accordion.Panel>
                <Stack gap="xs" w="100%">
                   {members
                     ?.filter((m) => m.status === 'active')
                     .map((m) => {
                       const participant = displayParticipants.find(
                         (p) => p.player.id === m.playerId,
                       )
                       const statusColor =
                         participant?.status === 'confirmed'
                           ? 'green'
                           : participant?.status === 'accepted'
                             ? 'yellow'
                             : participant?.status === 'declined'
                               ? 'red'
                               : 'blue'

                       // Determine action based on game status and participant status
                       const isConfirmed = participant?.status === 'confirmed'
                       const canRemove = game.status === 'planned' && isConfirmed
                       const canAdd = !isConfirmed || game.status === 'active'

                       return (
                         <Group
                           key={m.playerId}
                           justify="space-between"
                           w="100%"
                           style={{ padding: '0 12px' }}
                         >
                           <Text
                             size="sm"
                             style={{ cursor: (canRemove || canAdd) ? 'pointer' : 'default' }}
                             onClick={() => {
                               if (canRemove) {
                                 setRemovePlayerId(m.playerId)
                               } else if (canAdd) {
                                 setAddPlayerId(m.playerId)
                               }
                             }}
                           >
                             {formatMemberName(m)}
                           </Text>
                           <Badge
                             color={statusColor}
                             variant="light"
                             size="sm"
                           >
                             {participant?.status ?? 'not in game'}
                           </Badge>
                         </Group>
                       )
                     })}
                </Stack>
              </Accordion.Panel>
            </Accordion.Item>
          </Accordion>

  {/* Rebuy button */}
          <Button
            color="violet"
            size="sm"
            onClick={() => navigate(`/games/${gameIdNum}/rebuy`)}
          >
            Управление Rebuy
          </Button>

          {/* Final stacks button */}
          <Button
            color="violet"
            size="sm"
            onClick={isMobile
              ? () => setChipsModalOpen(true)
              : () => navigate(`/games/${gameIdNum}/chips`)
            }
          >
            Ввести Chips End
          </Button>
        </Stack>
      </Card>

      {/* Summary Table */}
      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Title order={4} mb="md">
          Summary
        </Title>

        {/* Web version: full table */}
        {!isMobile && (
          <Table variant="striped">
            <Table.Thead>
              <Table.Tr>
                <Table.Th>Игрок</Table.Th>
                <Table.Th ta="center">Buy-in</Table.Th>
                <Table.Th ta="center">Rebuy</Table.Th>
                <Table.Th ta="center">Invested</Table.Th>
                <Table.Th ta="center">Chips End</Table.Th>
                <Table.Th ta="center">Payout</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {confirmedParticipants.map((p) => {
                const invested =
                  p.buyInCount * game.buyInAmount +
                  p.rebuyCount * (game.rebuyPrice ?? 0)
                const payout = p.chipsEnd !== undefined
                  ? Math.round(p.chipsEnd * game.chipValue)
                  : '—'
                return (
                  <Table.Tr key={p.player.id}>
                    <Table.Td>
                      {[p.player.firstName, p.player.lastName]
                        .filter(Boolean)
                        .join(' ') || p.player.nickname}
                    </Table.Td>
                    <Table.Td ta="center">{Math.round(p.buyInCount * game.buyInAmount)}</Table.Td>
                    <Table.Td ta="center">
                      {p.rebuyCount} / {Math.round(p.rebuyCount * (game.rebuyPrice ?? 0))}
                    </Table.Td>
                    <Table.Td ta="center">{Math.round(invested)}</Table.Td>
                    <Table.Td ta="center">
                      {p.chipsEnd !== undefined
                        ? Math.round(p.chipsEnd)
                        : '—'}
                    </Table.Td>
                    <Table.Td ta="center">{payout}</Table.Td>
                  </Table.Tr>
                )
              })}
              {/* TOTAL row */}
              <Table.Tr>
                <Table.Td fw={700}>TOTAL</Table.Td>
                <Table.Td fw={700} ta="center">
                  {Math.round(confirmedParticipants.reduce((sum, p) => sum + p.buyInCount * game.buyInAmount, 0))}
                </Table.Td>
                <Table.Td fw={700} ta="center">
                  {confirmedParticipants.reduce((sum, p) => sum + p.rebuyCount, 0)}
                  {' / '}
                  {Math.round(confirmedParticipants.reduce((sum, p) => sum + p.rebuyCount * (game.rebuyPrice ?? 0), 0))}
                </Table.Td>
                <Table.Td fw={700} ta="center">{Math.round(bankCheck.totalBank)}</Table.Td>
                <Table.Td fw={700} ta="center">
                  {Math.round(bankCheck.totalPayout)}
                </Table.Td>
                <Table.Td fw={700} ta="center">{Math.round(bankCheck.totalPayout)}</Table.Td>
              </Table.Tr>
            </Table.Tbody>
          </Table>
        )}

        {/* Mobile version: mini tables per player */}
        {isMobile && (
          <Stack gap="md">
            {confirmedParticipants.map((p) => {
              const invested =
                p.buyInCount * game.buyInAmount +
                p.rebuyCount * (game.rebuyPrice ?? 0)
              const payout = p.chipsEnd !== undefined
                ? Math.round(p.chipsEnd * game.chipValue)
                : '—'
              const playerName =
                [p.player.firstName, p.player.lastName]
                  .filter(Boolean)
                  .join(' ') || p.player.nickname

              return (
                <Card key={p.player.id} padding="sm" radius="sm" withBorder>
                  <Group justify="space-between" mb="x">
                    <Text fw={500} size="sm" c="violet">
                      {playerName}
                    </Text>
                    <Text fw={500} size="sm" c="violet">
                      {payout}
                    </Text>
                  </Group>

                  {/* Row 1: Buy-in | Rebuy */}
                  <Table
                    variant="compact"
                    mt="x"
                    styles={{
                      td: { padding: '0.1rem' },
                    }}
                  >
                    <Table.Tbody>
                      <Table.Tr>
                        <Table.Td ta="center">
                          <Text size="xs" c="dimmed">Buy-in</Text>
                        </Table.Td>
                        <Table.Td ta="center">
                          <Text size="xs" c="dimmed">Rebuy</Text>
                        </Table.Td>
                      </Table.Tr>
                      <Table.Tr>
                      <Table.Td ta="center">{Math.round(p.buyInCount * game.buyInAmount)}</Table.Td>
                        <Table.Td ta="center">
                          {p.rebuyCount} / {Math.round(p.rebuyCount * (game.rebuyPrice ?? 0))}
                        </Table.Td>
                      </Table.Tr>
                    </Table.Tbody>
                  </Table>

                  {/* Row 2: Invested | Chips End */}
                  <Table
                    variant="compact"
                    mt="xs"
                    styles={{
                      td: { padding: '0.1rem' },
                    }}
                  >
                    <Table.Tbody>
                      <Table.Tr>
                        <Table.Td ta="center">
                          <Text size="xs" c="dimmed">Invested</Text>
                        </Table.Td>
                        <Table.Td ta="center">
                          <Text size="xs" c="dimmed">Chips End</Text>
                        </Table.Td>
                      </Table.Tr>
                      <Table.Tr>
                        <Table.Td ta="center">{Math.round(invested)}</Table.Td>
                        <Table.Td ta="center">
                          {p.chipsEnd !== undefined
                            ? Math.round(p.chipsEnd)
                            : '—'}
                        </Table.Td>
                      </Table.Tr>
                    </Table.Tbody>
                  </Table>
                </Card>
              )
            })}

            {/* TOTAL mini-table */}
            <Card padding="sm" radius="sm" withBorder>
              <Group justify="space-between" mb="x">
                <Text fw={700} size="sm">TOTAL</Text>
                <Text fw={700} size="sm" c="violet">
                  {Math.round(bankCheck.totalPayout)}
                </Text>
              </Group>
              <Table
                variant="compact"
                mt="x"
                styles={{
                  td: { padding: '0.1rem' },
                }}
              >
                <Table.Tbody>
                  <Table.Tr>
                    <Table.Td ta="center">
                      <Text size="xs" c="dimmed">Buy-in</Text>
                    </Table.Td>
                    <Table.Td ta="center">
                      <Text size="xs" c="dimmed">Rebuy</Text>
                    </Table.Td>
                  </Table.Tr>
                  <Table.Tr>
                    <Table.Td ta="center">
                      {Math.round(confirmedParticipants.reduce((sum, p) => sum + p.buyInCount * game.buyInAmount, 0))}
                    </Table.Td>
                    <Table.Td ta="center">
                      {confirmedParticipants.reduce((sum, p) => sum + p.rebuyCount, 0)}
                      {' / '}
                      {Math.round(confirmedParticipants.reduce((sum, p) => sum + p.rebuyCount * (game.rebuyPrice ?? 0), 0))}
                    </Table.Td>
                  </Table.Tr>
                </Table.Tbody>
              </Table>
              <Table
                variant="compact"
                mt="xs"
                styles={{
                  td: { padding: '0.1rem' },
                }}
              >
                <Table.Tbody>
                  <Table.Tr>
                    <Table.Td ta="center">
                      <Text size="xs" c="dimmed">Invested</Text>
                    </Table.Td>
                    <Table.Td ta="center">
                      <Text size="xs" c="dimmed">Chips End</Text>
                    </Table.Td>
                  </Table.Tr>
                  <Table.Tr>
                    <Table.Td ta="center">{Math.round(bankCheck.totalBank)}</Table.Td>
                    <Table.Td ta="center">{Math.round(bankCheck.totalPayout)}</Table.Td>
                  </Table.Tr>
                </Table.Tbody>
              </Table>
            </Card>
          </Stack>
        )}

        {/* Balance check */}
        <Stack gap="xs" mt="md">
          {bankCheck.mismatch ? (
            <>
              <Alert color="red" variant="light" title="⚠ Balance mismatch">
                Difference: {Math.round(bankCheck.difference)}
              </Alert>
              <Text size="sm" c="dimmed">
                Invested: {Math.round(bankCheck.totalBank)} | Chips End: {Math.round(bankCheck.totalPayout)} | Difference: {Math.round(bankCheck.difference)}
              </Text>
            </>
          ) : (
            <Alert color="green" variant="light" title="✓ Game balance is correct">
              Total Invested: {Math.round(bankCheck.totalBank)}
            </Alert>
          )}
        </Stack>
      </Card>

      {/* Finish Game Button (outside Summary) */}
      <Group justify={isMobile ? 'center' : 'flex-end'} mt="lg" w={isMobile ? '100%' : 'auto'}>
        <Button
          color={bankCheck.mismatch ? 'red' : 'green'}
          leftSection={<IconTrophy size={16} />}
          onClick={() => setFinishGameModalOpen(true)}
          loading={finishGame.isPending}
          w={isMobile ? '100%' : 'auto'}
        >
          Finish Game
        </Button>
      </Group>

      {/* Chips End Modal (mini app) */}
      <Modal
        opened={chipsModalOpen}
        onClose={() => setChipsModalOpen(false)}
        title="Ввести Chips End"
        centered
        size="lg"
        fullScreen={isMobile}
      >
        <Stack gap="md">
          <Table variant="compact" styles={{ td: { padding: '0.1rem' } }}>
            <Table.Thead>
              <Table.Tr>
                <Table.Th ta="center">
                  <Text size="xs" c="dimmed">Игрок</Text>
                </Table.Th>
                <Table.Th ta="center">
                  <Text size="xs" c="dimmed">Invested</Text>
                </Table.Th>
                <Table.Th ta="center">
                  <Text size="xs" c="dimmed">Chips End</Text>
                </Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {confirmedParticipants.map((p) => {
                const invested =
                  p.buyInCount * game.buyInAmount +
                  p.rebuyCount * (game.rebuyPrice ?? 0)
                return (
                  <Table.Tr key={p.player.id}>
                    <Table.Td ta="center">
                      {[p.player.firstName, p.player.lastName]
                        .filter(Boolean)
                        .join(' ') || p.player.nickname}
                    </Table.Td>
                    <Table.Td ta="center">
                      {Math.round(invested)}
                    </Table.Td>
                    <Table.Td ta="center">
                      {p.chipsEnd !== undefined ? (
                        <Text
                          size="sm"
                          c="violet"
                          style={{ cursor: 'pointer' }}
                          onClick={() => {
                            setChipsEndPlayerId(p.player.id)
                            setChipsEndValue(p.chipsEnd ?? 0)
                          }}
                        >
                          {Math.round(p.chipsEnd)}
                        </Text>
                      ) : (
                        <Text
                          size="sm"
                          c="dimmed"
                          style={{ cursor: 'pointer' }}
                          onClick={() => {
                            setChipsEndPlayerId(p.player.id)
                            setChipsEndValue(0)
                          }}
                        >
                          —
                        </Text>
                      )}
                    </Table.Td>
                  </Table.Tr>
                )
              })}
            </Table.Tbody>
          </Table>
        </Stack>
      </Modal>

      {/* Set Chips End Modal */}
      <Modal
        opened={chipsEndPlayerId !== null}
        onClose={() => setChipsEndPlayerId(null)}
        title="Ввести Chips End"
        centered
        size="sm"
      >
        <Stack gap="md">
          <NumberInput
            label="Chips End"
            placeholder="0"
            value={chipsEndValue}
            onChange={(value) => setChipsEndValue(Number(value) || 0)}
            allowNegative={false}
            min={0}
          />
          <Group justify="flex-end" gap="sm">
            <Button
              variant="subtle"
              onClick={() => setChipsEndPlayerId(null)}
            >
              Cancel
            </Button>
            <Button
              color="green"
              onClick={handleSetChipsEnd}
              loading={setChipsEnd.isPending}
            >
              Confirm
            </Button>
          </Group>
        </Stack>
      </Modal>

      {/* Remove Player Confirmation Modal */}
      <Modal
        opened={removePlayerId !== null}
        onClose={() => setRemovePlayerId(null)}
        title="Убрать из игры"
        centered
        size="sm"
      >
        <Stack gap="md">
          <Alert color="red" variant="light" title="Внимание">
            This will remove the player from the game. This action cannot be undone.
          </Alert>
          <Group justify="flex-end" gap="sm">
            <Button
              variant="subtle"
              onClick={() => setRemovePlayerId(null)}
            >
              Отмена
            </Button>
            <Button
              color="red"
              onClick={handleRemoveParticipant}
              loading={removeParticipant.isPending}
            >
              Убрать из игры
            </Button>
          </Group>
        </Stack>
      </Modal>

      {/* Add Player Confirmation Modal */}
      <Modal
        opened={addPlayerId !== null}
        onClose={() => setAddPlayerId(null)}
        title="Добавить в игру"
        centered
        size="sm"
      >
        <Stack gap="md">
          <Text>
            Add this player to the game?
          </Text>
          <Group justify="flex-end" gap="sm">
            <Button
              variant="subtle"
              onClick={() => setAddPlayerId(null)}
            >
              Отмена
            </Button>
            <Button
              color="green"
              onClick={handleAddParticipant}
              loading={addParticipant.isPending}
            >
              Добавить в игру
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
          {bankCheck.mismatch ? (
            <>
              <Alert color="red" variant="light" title="Balance Mismatch">
                <Stack gap="xs">
                  <Text>Invested: {Math.round(bankCheck.totalBank)}</Text>
                  <Text>Chips End: {Math.round(bankCheck.totalPayout)}</Text>
                  <Text>Difference: {Math.round(bankCheck.difference)}</Text>
                </Stack>
              </Alert>
              <Text size="sm">
                The game can still be finished, but the results may be incorrect.
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
                  onClick={handleFinishGame}
                  loading={finishGame.isPending}
                >
                  Finish anyway
                </Button>
              </Group>
            </>
          ) : (
            <>
              <Alert color="green" variant="light" title="Game Balance Correct">
                <Stack gap="xs">
                  <Text>Total Invested: {Math.round(bankCheck.totalBank)}</Text>
                  <Text>Total Chips End: {Math.round(bankCheck.totalPayout)}</Text>
                  <Text>Total Payout: {Math.round(bankCheck.totalPayout)}</Text>
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
                  onClick={handleFinishGame}
                  loading={finishGame.isPending}
                >
                  Finish Game
                </Button>
              </Group>
            </>
          )}
        </Stack>
      </Modal>
    </PageContainer>
  )
}
