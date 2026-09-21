/**
 * PlayerHistoryTable component.
 *
 * Displays player game history in a web table format with sorting support.
 * All values come from backend API.
 *
 * Per agent_task_7_tables.md:
 * Columns: Дата | Игра | Place | Invested | Payout | Profit | ROI %
 * - Clicking a row opens a modal with a link to Game Details
 * - Default sort: date descending
 * - Sortable by each column
 * - Integer values for non-percentage fields (no currency symbols)
 * - 2 decimal places for percentage fields (ROI)
 * - Positive values without +, negative with -
 * - Winner (Place = 1) highlighted with gold text color
 */

import { Table, Text, Modal, Button, Stack } from '@mantine/core'
import { IconTrophy } from '@tabler/icons-react'
import { useState } from 'react'
import type { PlayerGameHistory } from '../../types'
import { formatPercentage, formatDateShort } from '../../utils'
import { useNavigate } from 'react-router-dom'

interface PlayerHistoryTableProps {
  history: PlayerGameHistory[]
}

type SortKey =
  'date' | 'game' | 'place' | 'invested' | 'payout' | 'profit' | 'roi'
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

export function PlayerHistoryTable({ history }: PlayerHistoryTableProps) {
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
      case 'invested':
        aValue = a.totalInvested
        bValue = b.totalInvested
        break
      case 'payout':
        aValue = b.payoutAmount ?? 0
        bValue = a.payoutAmount ?? 0
        break
      case 'profit':
        aValue = a.profit
        bValue = b.profit
        break
      case 'roi':
        aValue = a.roi
        bValue = b.roi
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

  return (
    <>
      <Table variant="striped">
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
            <Table.Th
              ta="center"
              style={{ cursor: 'pointer' }}
              onClick={() => handleSort('place')}
            >
              Place{getSortIndicator('place')}
            </Table.Th>
            <Table.Th
              ta="center"
              style={{ cursor: 'pointer' }}
              onClick={() => handleSort('invested')}
            >
              Invested{getSortIndicator('invested')}
            </Table.Th>
            <Table.Th
              ta="center"
              style={{ cursor: 'pointer' }}
              onClick={() => handleSort('payout')}
            >
              Payout{getSortIndicator('payout')}
            </Table.Th>
            <Table.Th
              ta="center"
              style={{ cursor: 'pointer' }}
              onClick={() => handleSort('profit')}
            >
              Profit{getSortIndicator('profit')}
            </Table.Th>
            <Table.Th
              ta="center"
              style={{ cursor: 'pointer' }}
              onClick={() => handleSort('roi')}
            >
              ROI %{getSortIndicator('roi')}
            </Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {sortedHistory.map((h) => (
            <Table.Tr
              key={h.gameId}
              style={{ cursor: 'pointer' }}
              onClick={() => setSelectedGame(h)}
            >
              <Table.Td ta="left">
                {h.startTime ? formatDateShort(h.startTime) : '—'}
              </Table.Td>
              <Table.Td ta="center">{h.gameName}</Table.Td>
              <Table.Td ta="center">
                {h.place === 1 ? (
                  <Text c="yellow" size="sm" fw={500}>
                    <IconTrophy size={14} /> {h.place}
                  </Text>
                ) : (
                  h.place
                )}
              </Table.Td>
              <Table.Td ta="center">{formatInt(h.totalInvested)}</Table.Td>
              <Table.Td ta="center">
                {h.payoutAmount !== undefined ? formatInt(h.payoutAmount) : '—'}
              </Table.Td>
              <Table.Td ta="center">
                {h.place === 1 ? (
                  <Text c="yellow" size="sm" fw={500}>
                    {formatProfit(h.profit)}
                  </Text>
                ) : (
                  formatProfit(h.profit)
                )}
              </Table.Td>
              <Table.Td ta="center">{formatPercentage(h.roi)}</Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>

      {/* Game Details Modal */}
      <Modal
        opened={!!selectedGame}
        onClose={() => setSelectedGame(null)}
        title={selectedGame ? selectedGame.gameName : ''}
        centered
        size="md"
      >
        {selectedGame && (
          <Stack gap="sm">
            <Text>Invested: {formatInt(selectedGame.totalInvested)}</Text>
            <Text>
              Payout:{' '}
              {selectedGame.payoutAmount !== undefined
                ? formatInt(selectedGame.payoutAmount)
                : '—'}
            </Text>
            <Text>ROI: {formatPercentage(selectedGame.roi)}</Text>
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
