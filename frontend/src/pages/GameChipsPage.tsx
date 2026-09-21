/**
 * Game Chips End Management page.
 *
 * Phase 6 scope (RM_FE_6.md):
 * - Chips End management for active games
 * - List participants with invested amounts and chips end
 * - Set chips end for a participant
 *
 * This page is accessible only to Banker/Owner/Admin.
 * It provides the chips end management interface with:
 * - Web: table with Игрок, Invested, Chips end columns; clicking Chips end opens modal
 * - Mini App: accordion with same table; clicking Chips end opens modal
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
  NumberInput,
  Accordion,
} from '@mantine/core'
import { IconArrowLeft, IconChevronRight } from '@tabler/icons-react'
import { useParams, useNavigate } from 'react-router-dom'
import { PageContainer, PageHeader } from '../components/ui'
import {
  useGame,
  useGameMonitor,
  useGameParticipants,
  useClub,
  useSetChipsEnd,
} from '../features'
import { ApiClientError } from '../api'
import { notifications } from '@mantine/notifications'
import { useMediaQuery } from '@mantine/hooks'

/**
 * Game Chips End Management page.
 *
 * Displays:
 * - Participant list with invested amounts and chips end
 * - Click on Chips end field to open modal for input
 */
export function GameChipsPage() {
  const { gameId } = useParams<{ gameId: string }>()
  const navigate = useNavigate()
  const gameIdNum = gameId ? parseInt(gameId, 10) : 0

  const { data: game } = useGame(gameIdNum)
  const { data: monitorData } = useGameMonitor(gameIdNum)
  const { data: participants, refetch: refetchParticipants } =
    useGameParticipants(gameIdNum)
  const { data: club } = useClub(game?.clubId ?? 0)

  const setChipsEnd = useSetChipsEnd()

  const isMobile = useMediaQuery('(max-width: 768px)')

  // Modal state for chips end input
  const [chipsEndPlayerId, setChipsEndPlayerId] = useState<number | null>(null)
  const [chipsEndValue, setChipsEndValue] = useState<number>(0)

  // Use monitor data if available (has full participant info), otherwise use participants
  const displayParticipants = monitorData?.participants ?? participants ?? []
  // Only confirmed participants are shown in game data tables
  const confirmedParticipants = displayParticipants.filter(
    (p) => p.status === 'confirmed',
  )

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

  if (!game) {
    return (
      <PageContainer>
        <PageHeader title="Chips End" description="Loading..." />
      </PageContainer>
    )
  }

  const buyInAmount = game.buyInAmount
  const rebuyPrice = game.rebuyPrice ?? 0
  const chipValue = game.chipValue

  return (
    <PageContainer>
      <PageHeader
        title="Chips End"
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

      {/* Chips End Configuration */}
      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Title order={4} mb="md">
          Chips End Configuration
        </Title>
        <Group gap="sm" justify="space-between">
          <Text size="sm" c="dimmed">
            Chip Value
          </Text>
          <Text size="sm" fw={500}>
            {chipValue}
          </Text>
        </Group>
      </Card>

      {/* Participants with Chips End */}
      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Title order={4} mb="md">
          Participants
        </Title>
        {confirmedParticipants.length === 0 ? (
          <Text size="sm" c="dimmed">
            No participants in this game.
          </Text>
        ) : (
          <>
            {/* Web version: full table */}
            {!isMobile && (
              <Table variant="striped">
                <Table.Thead>
                  <Table.Tr>
                    <Table.Th>
                      <Text size="xs" c="dimmed">
                        Игрок
                      </Text>
                    </Table.Th>
                    <Table.Th ta="center">
                      <Text size="xs" c="dimmed">
                        Invested
                      </Text>
                    </Table.Th>
                    <Table.Th ta="center">
                      <Text size="xs" c="dimmed">
                        Chips End
                      </Text>
                    </Table.Th>
                  </Table.Tr>
                </Table.Thead>
                <Table.Tbody>
                  {confirmedParticipants.map((p) => {
                    const invested =
                      p.buyInCount * buyInAmount + p.rebuyCount * rebuyPrice
                    return (
                      <Table.Tr key={p.player.id}>
                        <Table.Td>
                          {[p.player.firstName, p.player.lastName]
                            .filter(Boolean)
                            .join(' ') || p.player.nickname}
                        </Table.Td>
                        <Table.Td ta="center">{Math.round(invested)}</Table.Td>
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
            )}

            {/* Mobile version: accordion with table */}
            {isMobile && (
              <Accordion
                variant="contained"
                radius="sm"
                styles={{
                  content: { padding: '0 0 0 0' },
                  item: { width: '100%' },
                }}
              >
                <Accordion.Item value="chips">
                  <Accordion.Control
                    chevron={<IconChevronRight size={16} stroke={2} />}
                  >
                    <Text fw={500}>Participants</Text>
                  </Accordion.Control>
                  <Accordion.Panel>
                    <Table
                      variant="compact"
                      styles={{ td: { padding: '0.1rem' } }}
                    >
                      <Table.Thead>
                        <Table.Tr>
                          <Table.Th ta="center">
                            <Text size="xs" c="dimmed">
                              Игрок
                            </Text>
                          </Table.Th>
                          <Table.Th ta="center">
                            <Text size="xs" c="dimmed">
                              Invested
                            </Text>
                          </Table.Th>
                          <Table.Th ta="center">
                            <Text size="xs" c="dimmed">
                              Chips End
                            </Text>
                          </Table.Th>
                        </Table.Tr>
                      </Table.Thead>
                      <Table.Tbody>
                        {confirmedParticipants.map((p) => {
                          const invested =
                            p.buyInCount * buyInAmount +
                            p.rebuyCount * rebuyPrice
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
                  </Accordion.Panel>
                </Accordion.Item>
              </Accordion>
            )}
          </>
        )}
      </Card>

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
            <Button variant="subtle" onClick={() => setChipsEndPlayerId(null)}>
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
    </PageContainer>
  )
}
