/**
 * PlayerHistoryMobile component.
 *
 * Displays player game history in a mobile-friendly compact table format with sorting.
 * Tapping a row opens a Modal with a table showing additional details.
 *
 * Per agent_task_7_tables.md:
 * Main table columns: Дата | Игра | Place | Profit
 * Extended table (modal): Invested | Payout | ROI %
 * - Default sort: date descending
 * - Sortable by each column in main table
 * - 0.1rem gap between parameter and value
 * - Column headers: mantine-font-size-xs, mantine-color-dimmed
 * - Integer values for non-percentage fields
 * - 2 decimal places for percentage fields (ROI)
 * - Positive values without +, negative with -
 * - Winner (Place = 1) row has gold text color
 */

import { Table, Text, Modal, Button, Stack } from '@mantine/core'
import { IconTrophy } from '@tabler/icons-react'
import { useState } from 'react'
import type { PlayerGameHistory } from '../../types'
import { formatPercentage, formatDateShort } from '../../utils'
import { useNavigate } from 'react-router-dom'

interface PlayerHistoryMobileProps {
  history: PlayerGameHistory[]
}

type SortKey = 'date' | 'game' | 'place' | 'profit'
type SortDirection = 'asc' | 'desc'

/**
 * Formats a profit value: positive without +, negative with -.
 * No currency symbol per spec.
 */
function formatProfit(value: number): string {
  if (isNaN(value)) {
    return '—'
  }
  const formatted = Math.round(Math.abs(value)).toString()
  return value < 0 ? `-${formatted}` : formatted
}

/**
 * Formats a plain integer value.
 */
function formatInt(value: number): string {
  if (isNaN(value)) {
    return '—'
  }
  return Math.round(value).toString()
}

export function PlayerHistoryMobile({ history }: PlayerHistoryMobileProps) {
  const navigate = useNavigate()
  const [sortKey, setSortKey] = useState<SortKey>('date')
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc')
  const [selectedGame, setSelectedGame] = useState<PlayerGameHistory | null>(
    null,
  )

  const handleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortKey(key)
      setSortDirection('desc')
    }
  }

  const sortedHistory = [...history].sort((a, b) => {
    let aValue: number | string
    let bValue: number | string

    switch (sortKey) {
      case 'date':
        aValue = a.startTime ? new Date(a.startTime).getTime() : 0
        bValue = b.startTime ? new Date(b.startTime).getTime() : 0
        break
      case 'game':
        aValue = a.gameName
        bValue = b.gameName
        break
      case 'place':
        aValue = a.place ?? 0
        bValue = b.place ?? 0
        break
      case 'profit':
        aValue = a.profit
        bValue = b.profit
        break
      default:
        aValue = 0
        bValue = 0
    }

    if (aValue < bValue) return sortDirection === 'asc' ? -1 : 1
    if (aValue > bValue) return sortDirection === 'asc' ? 1 : -1
    return 0
  })

  const getSortIndicator = (key: SortKey) => {
    if (sortKey !== key) return ''
    return sortDirection === 'asc' ? ' ↑' : ' ↓'
  }

  const isWinner = (place?: number) => place === 1

  return (
    <>
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
            <Table.Th
              style={{ cursor: 'pointer' }}
              onClick={() => handleSort('date')}
            >
              Дата{getSortIndicator('date')}
            </Table.Th>
            <Table.Th
              ta="center"
              style={{ cursor: 'pointer' }}
              onClick={() => handleSort('game')}
            >
              Игра{getSortIndicator('game')}
            </Table.Th>
            <Table.Th ta="center">Place</Table.Th>
            <Table.Th
              ta="right"
              style={{ cursor: 'pointer' }}
              onClick={() => handleSort('profit')}
            >
              Profit{getSortIndicator('profit')}
            </Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {sortedHistory.map((h) => {
            const winner = isWinner(h.place)
            return (
              <Table.Tr
                key={h.gameId}
                onClick={() => setSelectedGame(h)}
                style={{ cursor: 'pointer' }}
              >
                <Table.Td ta="left">
                  <Text
                    size="xs"
                    c={winner ? 'yellow' : undefined}
                    fw={winner ? 500 : 400}
                  >
                    {h.startTime ? formatDateShort(h.startTime) : '—'}
                  </Text>
                </Table.Td>
                <Table.Td ta="center">
                  <Text
                    size="xs"
                    c={winner ? 'yellow' : undefined}
                    fw={winner ? 500 : 400}
                  >
                    {h.gameName}
                  </Text>
                </Table.Td>
                <Table.Td ta="center">
                  {h.place === 1 ? (
                    <Text c="yellow" size="xs" fw={500}>
                      <IconTrophy size={12} /> {h.place}
                    </Text>
                  ) : (
                    <Text size="xs" c={winner ? 'yellow' : undefined}>
                      {h.place}
                    </Text>
                  )}
                </Table.Td>
                <Table.Td ta="right">
                  <Text
                    size="xs"
                    c={winner ? 'yellow' : undefined}
                    fw={winner ? 500 : 400}
                  >
                    {formatProfit(h.profit)}
                  </Text>
                </Table.Td>
              </Table.Tr>
            )
          })}
        </Table.Tbody>
      </Table>

      {/* Game Details Modal with table format */}
      <Modal
        opened={!!selectedGame}
        onClose={() => setSelectedGame(null)}
        title={selectedGame ? selectedGame.gameName : ''}
        centered
        size="md"
      >
        {selectedGame && (
          <Stack gap="sm">
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
                  <Table.Th>Invested</Table.Th>
                  <Table.Th ta="center">Payout</Table.Th>
                  <Table.Th ta="center">ROI %</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                <Table.Tr>
                  <Table.Td>{formatInt(selectedGame.totalInvested)}</Table.Td>
                  <Table.Td ta="center">
                    {selectedGame.payoutAmount !== undefined
                      ? formatInt(selectedGame.payoutAmount)
                      : '—'}
                  </Table.Td>
                  <Table.Td ta="center">
                    {formatPercentage(selectedGame.roi)}
                  </Table.Td>
                </Table.Tr>
              </Table.Tbody>
            </Table>
            <Button
              color="blue"
              onClick={() => {
                navigate(`/games/${selectedGame.gameId}`)
                setSelectedGame(null)
              }}
            >
              Go to Game Details
            </Button>
          </Stack>
        )}
      </Modal>
    </>
  )
}
