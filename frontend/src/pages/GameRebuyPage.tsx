/**
 * Game Rebuy Management page.
 *
 * Phase 6 scope (RM_FE_6.md):
 * - Rebuy management for active games
 * - List participants with rebuy counts
 * - Register rebuy for a participant
 * - Adjust rebuy count for a participant
 *
 * This page is accessible only to Banker/Owner/Admin.
 * It provides the rebuy management interface with:
 * - Participant list with rebuy counts and amounts
 * - Add Rebuy and Change Rebuy buttons for each participant
 * - TOTAL row
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
  NumberInput,
} from '@mantine/core'
import { IconArrowLeft } from '@tabler/icons-react'
import { useParams, useNavigate } from 'react-router-dom'
import { PageContainer, PageHeader } from '../components/ui'
import {
  useGame,
  useGameMonitor,
  useGameParticipants,
  useClub,
  useRegisterRebuy,
  useFixRebuy,
} from '../features'
import { ApiClientError } from '../api'
import { notifications } from '@mantine/notifications'
import { useMediaQuery } from '@mantine/hooks'

/**
 * Game Rebuy Management page.
 *
 * Displays:
 * - Participant list with rebuy counts and amounts
 * - Add Rebuy and Change Rebuy buttons for each participant
 * - TOTAL row
 */
export function GameRebuyPage() {
  const { gameId } = useParams<{ gameId: string }>()
  const navigate = useNavigate()
  const gameIdNum = gameId ? parseInt(gameId, 10) : 0

  const { data: game } = useGame(gameIdNum)
  const { data: monitorData } = useGameMonitor(gameIdNum)
  const { data: participants, refetch: refetchParticipants } =
    useGameParticipants(gameIdNum)
  const { data: club } = useClub(game?.clubId ?? 0)

  const registerRebuy = useRegisterRebuy()
  const fixRebuy = useFixRebuy()

  const isMobile = useMediaQuery('(max-width: 768px)')

  // Modal states
  const [addRebuyPlayerId, setAddRebuyPlayerId] = useState<number | null>(null)
  const [changeRebuyPlayerId, setChangeRebuyPlayerId] = useState<number | null>(
    null,
  )
  const [changeRebuyValue, setChangeRebuyValue] = useState<number>(0)

  // Use monitor data if available (has full participant info), otherwise use participants
  const displayParticipants = monitorData?.participants ?? participants ?? []
  // Only confirmed participants are shown in game data tables
  const confirmedParticipants = displayParticipants.filter(
    (p) => p.status === 'confirmed',
  )

  const handleAddRebuy = async () => {
    if (!addRebuyPlayerId) return
    try {
      await registerRebuy.mutateAsync({
        gameId: gameIdNum,
        playerId: addRebuyPlayerId,
      })
      notifications.show({
        title: 'Rebuy Registered',
        message: 'Rebuy has been registered.',
        color: 'green',
      })
      setAddRebuyPlayerId(null)
      refetchParticipants()
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

  const handleSaveChangeRebuy = async () => {
    if (!changeRebuyPlayerId) return
    try {
      await fixRebuy.mutateAsync({
        gameId: gameIdNum,
        playerId: changeRebuyPlayerId,
        rebuyCount: changeRebuyValue,
      })
      notifications.show({
        title: 'Rebuy Updated',
        message: `Rebuy count set to ${changeRebuyValue}.`,
        color: 'green',
      })
      setChangeRebuyPlayerId(null)
      refetchParticipants()
    } catch (err) {
      if (err instanceof ApiClientError) {
        notifications.show({
          title: 'Failed to update rebuy',
          message: err.message,
          color: 'red',
        })
      }
    }
  }

  if (!game) {
    return (
      <PageContainer>
        <PageHeader title="Rebuy Management" description="Loading..." />
      </PageContainer>
    )
  }

  const rebuyPrice = game.rebuyPrice ?? 0

  return (
    <PageContainer>
      <PageHeader
        title="Rebuy Management"
        description={game.gameType}
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

      {/* Game Info */}
      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Title order={4} mb="md">
          Game Info
        </Title>
        <Stack gap="sm">
          <Group gap="sm" justify="space-between">
            <Text size="sm" c="dimmed">
              Club
            </Text>
            <Text size="sm" fw={500}>
              {club?.name ?? '—'}
            </Text>
          </Group>
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
              Type
            </Text>
            <Text size="sm" fw={500}>
              {game.gameType}
            </Text>
          </Group>
          <Group gap="sm" justify="space-between">
            <Text size="sm" c="dimmed">
              Status
            </Text>
            <Badge color="green" variant="light">
              {game.status}
            </Badge>
          </Group>
        </Stack>
      </Card>

      {/* Rebuy Configuration */}
      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Title order={4} mb="md">
          Rebuy Configuration
        </Title>
        <Group gap="sm" justify="space-between">
          <Text size="sm" c="dimmed">
            Rebuy Price
          </Text>
          <Text size="sm" fw={500}>
            {game.rebuyAllowed ? Math.round(rebuyPrice) : 'Not allowed'}
          </Text>
        </Group>
        {game.maxRebuys !== undefined && (
          <Group gap="sm" justify="space-between">
            <Text size="sm" c="dimmed">
              Max Rebuys
            </Text>
            <Text size="sm" fw={500}>
              {game.maxRebuys}
            </Text>
          </Group>
        )}
      </Card>

      {/* Participants with Rebuy */}
      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Title order={4} mb="md">
          Participants
        </Title>
        {confirmedParticipants.length === 0 ? (
          <Text size="sm" c="dimmed">
            No participants in this game.
          </Text>
        ) : (
          <Table variant={isMobile ? 'compact' : 'striped'}>
            <Table.Thead>
              <Table.Tr>
                <Table.Th>
                  <Text size="xs" c="dimmed">
                    Игрок
                  </Text>
                </Table.Th>
                <Table.Th ta="center">
                  <Text size="xs" c="dimmed">
                    Rebuy
                  </Text>
                </Table.Th>
                <Table.Th ta="center">
                  <Text size="xs" c="dimmed">
                    Actions
                  </Text>
                </Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {confirmedParticipants.map((p) => {
                const rebuyAmount = p.rebuyCount * rebuyPrice
                return (
                  <Table.Tr key={p.player.id}>
                    <Table.Td>
                      {[p.player.firstName, p.player.lastName]
                        .filter(Boolean)
                        .join(' ') || p.player.nickname}
                    </Table.Td>
                    <Table.Td ta="center">
                      {isMobile ? (
                        <Text
                          size="sm"
                          c="violet"
                          style={{
                            cursor: game.rebuyAllowed ? 'pointer' : 'default',
                          }}
                          onClick={() => {
                            if (game.rebuyAllowed) {
                              setChangeRebuyPlayerId(p.player.id)
                              setChangeRebuyValue(p.rebuyCount)
                            }
                          }}
                        >
                          {p.rebuyCount} / {Math.round(rebuyAmount)}
                        </Text>
                      ) : (
                        <Text size="sm">
                          {p.rebuyCount} / {Math.round(rebuyAmount)}
                        </Text>
                      )}
                    </Table.Td>
                    {!isMobile && (
                      <Table.Td ta="center">
                        <Group gap="xs" justify="center">
                          <Button
                            color="violet"
                            size="xs"
                            onClick={() => setAddRebuyPlayerId(p.player.id)}
                            disabled={!game.rebuyAllowed}
                          >
                            Добавить Rebuy
                          </Button>
                          <Button
                            color="violet"
                            size="xs"
                            variant="outline"
                            onClick={() => {
                              setChangeRebuyPlayerId(p.player.id)
                              setChangeRebuyValue(p.rebuyCount)
                            }}
                            disabled={!game.rebuyAllowed}
                          >
                            Изменить Rebuy
                          </Button>
                        </Group>
                      </Table.Td>
                    )}
                    {isMobile && (
                      <Table.Td ta="center">
                        <Button
                          color="violet"
                          size="xs"
                          onClick={() => setAddRebuyPlayerId(p.player.id)}
                          disabled={!game.rebuyAllowed}
                        >
                          Add Rebuy
                        </Button>
                      </Table.Td>
                    )}
                  </Table.Tr>
                )
              })}
              {/* TOTAL row */}
              <Table.Tr>
                <Table.Td fw={700}>TOTAL</Table.Td>
                <Table.Td fw={700} ta="center">
                  {confirmedParticipants.reduce(
                    (sum, p) => sum + p.rebuyCount,
                    0,
                  )}
                  {' / '}
                  {Math.round(
                    confirmedParticipants.reduce(
                      (sum, p) => sum + p.rebuyCount * rebuyPrice,
                      0,
                    ),
                  )}
                </Table.Td>
                <Table.Td />
              </Table.Tr>
            </Table.Tbody>
          </Table>
        )}
      </Card>

      {/* Add Rebuy Confirmation Modal */}
      <Modal
        opened={addRebuyPlayerId !== null}
        onClose={() => setAddRebuyPlayerId(null)}
        title="Add Rebuy"
        centered
        size="sm"
      >
        <Stack gap="md">
          {addRebuyPlayerId !== null &&
            (() => {
              const participant = displayParticipants.find(
                (p) => p.player.id === addRebuyPlayerId,
              )
              const currentCount = participant ? participant.rebuyCount : 0
              const newCount = currentCount + 1
              return (
                <>
                  <Alert color="blue" variant="light" title="Add Rebuy">
                    Add a rebuy for this player?
                  </Alert>
                  <Text size="sm">Current: {currentCount}</Text>
                  <Text size="sm">New: {newCount}</Text>
                </>
              )
            })()}
          <Group justify="flex-end" gap="sm">
            <Button variant="subtle" onClick={() => setAddRebuyPlayerId(null)}>
              Cancel
            </Button>
            <Button
              color="green"
              onClick={handleAddRebuy}
              loading={registerRebuy.isPending}
            >
              Confirm
            </Button>
          </Group>
        </Stack>
      </Modal>

      {/* Change Rebuy Modal */}
      <Modal
        opened={changeRebuyPlayerId !== null}
        onClose={() => setChangeRebuyPlayerId(null)}
        title="Change Rebuy"
        centered
        size="sm"
      >
        <Stack gap="md">
          {changeRebuyPlayerId !== null &&
            (() => {
              const participant = displayParticipants.find(
                (p) => p.player.id === changeRebuyPlayerId,
              )
              const oldCount = participant ? participant.rebuyCount : 0
              return (
                <>
                  <Alert color="blue" variant="light" title="Change Rebuy">
                    Rebuy: {oldCount}
                  </Alert>
                  <NumberInput
                    label="New Rebuy Count"
                    placeholder="0"
                    value={changeRebuyValue}
                    onChange={(value) =>
                      setChangeRebuyValue(Number(value) || 0)
                    }
                    allowNegative={false}
                    min={0}
                  />
                </>
              )
            })()}
          <Group justify="flex-end" gap="sm">
            <Button
              variant="subtle"
              onClick={() => setChangeRebuyPlayerId(null)}
            >
              Cancel
            </Button>
            <Button color="green" onClick={handleSaveChangeRebuy}>
              Save
            </Button>
          </Group>
        </Stack>
      </Modal>
    </PageContainer>
  )
}
