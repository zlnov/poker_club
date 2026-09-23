/**
 * Event Journal section component for the Club Dashboard.
 *
 * Displays the ЖУРНАЛ СОБЫТИЙ block with a table of correction events.
 *
 * Layout:
 * - Full width (no columns)
 * - Web: shows 3 latest records in main table, click to expand to 10 in modal
 * - Mobile: shows 7 records in compact table, click on row to see details in modal
 *
 * Visible only for Admin/Owner roles.
 *
 * See 05_FE_UX.md section 4 (Application Navigation) and
 * 04_FE_SPEC.md section 11 (Routing).
 */

import {
  Card,
  Group,
  Stack,
  Text,
  Title,
  Table,
  Modal,
  ScrollArea,
  Button,
} from '@mantine/core'
import { IconAlertTriangle } from '@tabler/icons-react'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useMediaQuery } from '@mantine/hooks'
import {
  useClubCorrectionEvents,
  useClubMembers,
  useClubGames,
  participantKeys,
} from '../features'
import type { GameParticipantSummary } from '../types'
import { useApiClient } from '../hooks'
import { useQueries } from '@tanstack/react-query'
import { backgroundSurfaces } from '../styles'

interface EventJournalSectionProps {
  /** Club ID. */
  clubId: number
}

/**
 * Formats a date to dd.mm.yyyy format.
 */
function formatDateDDMMYYYY(isoDate: string): string {
  if (!isoDate) return '—'
  try {
    const date = new Date(isoDate)
    const day = String(date.getDate()).padStart(2, '0')
    const month = String(date.getMonth() + 1).padStart(2, '0')
    const year = date.getFullYear()
    return `${day}.${month}.${year}`
  } catch {
    return '—'
  }
}

/**
 * Extracts the field name from metadata.
 * Prioritizes *_amount fields over *_count fields.
 */
function extractFieldName(metadata?: Record<string, unknown>): string | null {
  if (!metadata) return null

  const keys = Object.keys(metadata)

  // Find money fields (*_amount) first
  const amountField = keys.find(
    (k) => k.startsWith('new_') && k.endsWith('_amount'),
  )
  if (amountField) {
    return amountField.replace(/^new_/, '')
  }

  // Then look for count fields (*_count)
  const countField = keys.find(
    (k) => k.startsWith('new_') && k.endsWith('_count'),
  )
  if (countField) {
    return countField.replace(/^new_/, '')
  }

  // Then look for old_ fields
  const oldAmountField = keys.find(
    (k) => k.startsWith('old_') && k.endsWith('_amount'),
  )
  if (oldAmountField) {
    return oldAmountField.replace(/^old_/, '')
  }

  const oldCountField = keys.find(
    (k) => k.startsWith('old_') && k.endsWith('_count'),
  )
  if (oldCountField) {
    return oldCountField.replace(/^old_/, '')
  }

  // Look for field key directly
  const field = metadata['field'] as string | undefined
  if (field) return field

  return null
}

/**
 * Gets the old value for the extracted field from metadata.
 */
function getOldValue(metadata?: Record<string, unknown>): number | undefined {
  if (!metadata) return undefined
  const fieldName = extractFieldName(metadata)
  if (!fieldName) return undefined
  return metadata[`old_${fieldName}`] as number | undefined
}

/**
 * Gets the new value for the extracted field from metadata.
 */
function getNewValue(metadata?: Record<string, unknown>): number | undefined {
  if (!metadata) return undefined
  const fieldName = extractFieldName(metadata)
  if (!fieldName) return undefined
  return metadata[`new_${fieldName}`] as number | undefined
}

/**
 * Formats a value based on field type.
 */
function formatValue(
  value: number | undefined,
  fieldName: string | null,
): string {
  if (value === undefined || value === null) return '—'

  // Money fields
  if (fieldName && fieldName.endsWith('_amount')) {
    const rounded = Math.round(value)
    const formatted = Math.abs(rounded)
      .toString()
      .replace(/\B(?=(\d{3})+(?!\d))/g, ' ')
    return `${rounded >= 0 ? '+' : '-'}${formatted} ₽`
  }

  // Default: integer
  return Math.round(value).toString()
}

/**
 * Formats the metadata field name for display.
 */
function formatMetadata(metadata?: Record<string, unknown>): string {
  const field = extractFieldName(metadata)
  if (!field) return '—'

  // Truncate long field names
  if (field.length > 16) {
    return `${field.substring(0, 13)}…`
  }
  return field
}

/**
 * Event Journal section component.
 *
 * Displays the ЖУРНАЛ СОБЫТИЙ block with correction events table.
 * Layout adapts to mobile (compact table with 7 rows) and desktop (full table with 3 rows).
 */
export function EventJournalSection({ clubId }: EventJournalSectionProps) {
  const isMobile = useMediaQuery('(max-width: 768px)')
  const [isModalOpen, setIsModalOpen] = useState(false)
  const [selectedEvent, setSelectedEvent] = useState<any>(null)
  const navigate = useNavigate()

  const { data: events } = useClubCorrectionEvents(clubId)
  const { data: members } = useClubMembers(clubId)
  const { data: games } = useClubGames(clubId)
  const apiClient = useApiClient()

  // Build a map of player_id -> nickname from game participants
  const playerNicknames = new Map<number, string>()
  const clubMemberNicknames = new Map<number, string>()

  // Populate club member nicknames (created_by is club_members.id)
  if (members) {
    for (const member of members) {
      clubMemberNicknames.set(
        member.clubMemberId,
        member.nickname ||
          `${member.firstName} ${member.lastName}`.trim() ||
          '—',
      )
    }
  }

  // Fetch participants for all games to get player nicknames
  const gameIds = games?.map((g) => g.id) ?? []
  const participantQueries = useQueries({
    queries: gameIds.map((gameId) => ({
      queryKey: participantKeys.list(gameId),
      queryFn: async () => {
        const response = await apiClient.get<{
          participants: GameParticipantSummary[]
        }>(`/games/${gameId}/participants`)
        return response.data.participants
      },
      enabled: !!gameId,
    })),
  })

  // Populate player nicknames from game participants
  for (let i = 0; i < participantQueries.length; i++) {
    const participants = participantQueries[i]?.data
    if (participants) {
      for (const p of participants) {
        playerNicknames.set(
          p.player.id,
          p.player.nickname ||
            `${p.player.firstName} ${p.player.lastName}`.trim() ||
            '—',
        )
      }
    }
  }

  // Web: show 3 latest records in main table
  // Mobile: show 7 records in compact table
  const displayLimit = isMobile ? 7 : 3
  const displayEvents = events?.slice(0, displayLimit) ?? []
  const hasMore = (events?.length ?? 0) > displayLimit

  // For modal: show 10 records
  const modalEvents = events?.slice(0, 10) ?? []

  const handleRowClick = (event: any) => {
    setSelectedEvent(event)
    // In mobile view, the selectedEvent modal (Детали записи) opens
    // In web view, the isModalOpen modal (Журнал событий — 10 записей) opens
    if (!isMobile) {
      setIsModalOpen(true)
    }
  }

  // Helper to get player nickname
  const getPlayerNickname = (playerId: number): string => {
    return playerNicknames.get(playerId) || `ID: ${playerId}`
  }

  // Helper to get club member nickname (for created_by)
  const getClubMemberNickname = (clubMemberId: number): string => {
    return clubMemberNicknames.get(clubMemberId) || `ID: ${clubMemberId}`
  }

  return (
    <Card
      mt="lg"
      padding="lg"
      radius="md"
      withBorder
      bg={backgroundSurfaces.card}
    >
      <Group gap="xs" mb="md">
        <IconAlertTriangle size={20} />
        <Title order={4} mb={0}>
          ЖУРНАЛ СОБЫТИЙ
        </Title>
        {hasMore && !isMobile && (
          <Text
            size="sm"
            c="blue"
            style={{ cursor: 'pointer', textDecoration: 'underline' }}
            onClick={() => setIsModalOpen(true)}
          >
            Развернуть →
          </Text>
        )}
      </Group>

      {displayEvents.length === 0 ? (
        <Stack
          style={{
            justifyContent: 'center',
            alignItems: 'center',
            minHeight: '80px',
          }}
        >
          <Text size="sm" c="dimmed" ta="center">
            Пока нет корректировок
          </Text>
        </Stack>
      ) : isMobile ? (
        // Mobile: compact table with fewer columns (matching ClubMemberStatsMobile style)
        <Table
          variant="compact"
          styles={{
            td: { padding: '0.1rem' },
            th: {
              fontSize: 'var(--mantine-font-size-xs)',
              color: 'var(--mantine-color-dimmed)',
            },
          }}
        >
          <Table.Thead>
            <Table.Tr>
              <Table.Th ta="center">Game</Table.Th>
              <Table.Th>Player</Table.Th>
              <Table.Th ta="center">Metadata</Table.Th>
              <Table.Th ta="center">Created_at</Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {displayEvents.map((event) => (
              <Table.Tr
                key={event.id}
                onClick={() => handleRowClick(event)}
                style={{ cursor: 'pointer' }}
              >
                <Table.Td ta="center">
                  <Text
                    size="xs"
                    component="a"
                    href={`/games/${event.game_id}`}
                    style={{ color: 'var(--mantine-color-blue-6)' }}
                    onClick={(e) => {
                      e.stopPropagation()
                      navigate(`/games/${event.game_id}`)
                    }}
                  >
                    #{event.game_id}
                  </Text>
                </Table.Td>
                <Table.Td>
                  <Text size="xs" fw={500}>
                    @{getPlayerNickname(event.player_id)}
                  </Text>
                </Table.Td>
                <Table.Td ta="center">
                  <Text size="xs" lineClamp={1}>
                    {formatMetadata(event.metadata)}
                  </Text>
                </Table.Td>
                <Table.Td ta="center">
                  <Text size="xs">{formatDateDDMMYYYY(event.created_at)}</Text>
                </Table.Td>
              </Table.Tr>
            ))}
          </Table.Tbody>
        </Table>
      ) : (
        // Web: full table with all columns (matching ClubMemberStatsTable style)
        <Table variant="striped">
          <Table.Thead>
            <Table.Tr>
              <Table.Th ta="center">Game</Table.Th>
              <Table.Th>Player</Table.Th>
              <Table.Th ta="center">Old_value</Table.Th>
              <Table.Th ta="center">New_value</Table.Th>
              <Table.Th>Metadata</Table.Th>
              <Table.Th ta="center">Created_at</Table.Th>
              <Table.Th ta="center">Created_by</Table.Th>
            </Table.Tr>
          </Table.Thead>
          <Table.Tbody>
            {displayEvents.map((event) => {
              const fieldName = extractFieldName(event.metadata)
              return (
                <Table.Tr key={event.id}>
                  <Table.Td ta="center">
                    <Text
                      size="sm"
                      component="a"
                      href={`/games/${event.game_id}`}
                      style={{
                        color: 'var(--mantine-color-blue-6)',
                        cursor: 'pointer',
                      }}
                      onClick={(e) => {
                        e.preventDefault()
                        navigate(`/games/${event.game_id}`)
                      }}
                    >
                      #{event.game_id}
                    </Text>
                  </Table.Td>
                  <Table.Td>
                    <Text size="sm">@{getPlayerNickname(event.player_id)}</Text>
                  </Table.Td>
                  <Table.Td ta="center">
                    <Text size="sm">
                      {formatValue(getOldValue(event.metadata), fieldName)}
                    </Text>
                  </Table.Td>
                  <Table.Td ta="center">
                    <Text size="sm">
                      {formatValue(getNewValue(event.metadata), fieldName)}
                    </Text>
                  </Table.Td>
                  <Table.Td>
                    <Text size="sm" lineClamp={1}>
                      {formatMetadata(event.metadata)}
                    </Text>
                  </Table.Td>
                  <Table.Td ta="center">
                    <Text size="sm">
                      {formatDateDDMMYYYY(event.created_at)}
                    </Text>
                  </Table.Td>
                  <Table.Td ta="center">
                    <Text size="sm">
                      @{getClubMemberNickname(event.created_by)}
                    </Text>
                  </Table.Td>
                </Table.Tr>
              )
            })}
          </Table.Tbody>
        </Table>
      )}

      {/* Modal for expanded view (web) */}
      <Modal
        opened={isModalOpen}
        onClose={() => setIsModalOpen(false)}
        title="ЖУРНАЛ СОБЫТИЙ — ПОСЛЕДНИЕ 10 ЗАПИСЕЙ"
        size="70%"
        centered
        padding="md"
        withCloseButton={false}
      >
        <ScrollArea style={{ maxHeight: '500px' }}>
          <Table variant="striped">
            <Table.Thead>
              <Table.Tr>
                <Table.Th ta="center">Game</Table.Th>
                <Table.Th>Player</Table.Th>
                <Table.Th ta="center">Old_value</Table.Th>
                <Table.Th ta="center">New_value</Table.Th>
                <Table.Th>Metadata</Table.Th>
                <Table.Th ta="center">Created_at</Table.Th>
                <Table.Th ta="center">Created_by</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {modalEvents.map((event) => {
                const fieldName = extractFieldName(event.metadata)
                return (
                  <Table.Tr key={event.id}>
                    <Table.Td ta="center">
                      <Text
                        size="sm"
                        component="a"
                        href={`/games/${event.game_id}`}
                        style={{
                          color: 'var(--mantine-color-blue-6)',
                          cursor: 'pointer',
                        }}
                        onClick={(e) => {
                          e.preventDefault()
                          navigate(`/games/${event.game_id}`)
                        }}
                      >
                        #{event.game_id}
                      </Text>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm">
                        @{getPlayerNickname(event.player_id)}
                      </Text>
                    </Table.Td>
                    <Table.Td ta="center">
                      <Text size="sm">
                        {formatValue(getOldValue(event.metadata), fieldName)}
                      </Text>
                    </Table.Td>
                    <Table.Td ta="center">
                      <Text size="sm">
                        {formatValue(getNewValue(event.metadata), fieldName)}
                      </Text>
                    </Table.Td>
                    <Table.Td>
                      <Text size="sm" lineClamp={1}>
                        {formatMetadata(event.metadata)}
                      </Text>
                    </Table.Td>
                    <Table.Td ta="center">
                      <Text size="sm">
                        {formatDateDDMMYYYY(event.created_at)}
                      </Text>
                    </Table.Td>
                    <Table.Td ta="center">
                      <Text size="sm">
                        @{getClubMemberNickname(event.created_by)}
                      </Text>
                    </Table.Td>
                  </Table.Tr>
                )
              })}
            </Table.Tbody>
          </Table>
        </ScrollArea>

        <Group justify="flex-end" mt="md">
          <Button
            size="sm"
            variant="subtle"
            onClick={() => setIsModalOpen(false)}
          >
            Закрыть
          </Button>
        </Group>
      </Modal>

      {/* Mobile modal for row details */}
      {isMobile && (
        <Modal
          opened={!!selectedEvent}
          onClose={() => setSelectedEvent(null)}
          title="Детали записи"
          size="md"
          centered
          padding="md"
          withCloseButton={false}
        >
          {selectedEvent && (
            <Table
              variant="compact"
              styles={{
                td: { padding: '0.1rem' },
                th: {
                  fontSize: 'var(--mantine-font-size-xs)',
                  color: 'var(--mantine-color-dimmed)',
                },
              }}
            >
              <Table.Thead>
                <Table.Tr>
                  <Table.Th ta="center">Old_value</Table.Th>
                  <Table.Th ta="center">New_value</Table.Th>
                  <Table.Th ta="center">Created_by</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                <Table.Tr>
                  <Table.Td ta="center">
                    <Text size="xs">
                      {formatValue(
                        getOldValue(selectedEvent.metadata),
                        extractFieldName(selectedEvent.metadata),
                      )}
                    </Text>
                  </Table.Td>
                  <Table.Td ta="center">
                    <Text size="xs">
                      {formatValue(
                        getNewValue(selectedEvent.metadata),
                        extractFieldName(selectedEvent.metadata),
                      )}
                    </Text>
                  </Table.Td>
                  <Table.Td ta="center">
                    <Text size="xs">
                      @{getClubMemberNickname(selectedEvent.created_by)}
                    </Text>
                  </Table.Td>
                </Table.Tr>
              </Table.Tbody>
            </Table>
          )}

          <Group justify="flex-end" mt="md">
            <Button
              size="sm"
              variant="subtle"
              onClick={() => setSelectedEvent(null)}
            >
              Закрыть
            </Button>
          </Group>
        </Modal>
      )}
    </Card>
  )
}
