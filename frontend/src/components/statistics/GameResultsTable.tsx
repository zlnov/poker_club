/**
 * GameResultsTable component.
 *
 * Displays game results in a web table format.
 * Winner (Place = 1) is highlighted with gold text color.
 * No sorting is applied - results are ordered by Place from backend.
 *
 * Per agent_task_7_tables.md:
 * Columns: Place | Player | Total Invested | Chips End | Payout | Profit | ROI %
 * - Winner (Place = 1) highlighted with gold text color
 * - Integer values for non-percentage fields
 * - 2 decimal places for percentage fields (ROI)
 * - Positive values without +, negative with -
 */

import { Table, Text } from '@mantine/core'
import { IconTrophy } from '@tabler/icons-react'
import type { GameResult } from '../../types'
import { formatPercentage } from '../../utils'

interface GameResultsTableProps {
  results: GameResult[]
}

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

export function GameResultsTable({ results }: GameResultsTableProps) {
  return (
    <Table variant="striped">
      <Table.Thead>
        <Table.Tr>
          <Table.Th ta="center">Place</Table.Th>
          <Table.Th>Player</Table.Th>
          <Table.Th ta="center">Total Invested</Table.Th>
          <Table.Th ta="center">Chips End</Table.Th>
          <Table.Th ta="center">Payout</Table.Th>
          <Table.Th ta="center">Profit</Table.Th>
          <Table.Th ta="center">ROI %</Table.Th>
        </Table.Tr>
      </Table.Thead>
      <Table.Tbody>
        {results.map((r) => (
          <Table.Tr key={r.playerId}>
            <Table.Td ta="center">
              {r.place === 1 ? (
                <Text c="yellow" size="sm" fw={500}>
                  <IconTrophy size={14} /> {r.place}
                </Text>
              ) : (
                r.place
              )}
            </Table.Td>
            <Table.Td>{r.playerName}</Table.Td>
            <Table.Td ta="center">{formatInt(r.totalInvested)}</Table.Td>
            <Table.Td ta="center">
              {r.chipsEnd !== undefined ? formatInt(r.chipsEnd) : '—'}
            </Table.Td>
            <Table.Td ta="center">
              {r.payoutAmount !== undefined ? formatInt(r.payoutAmount) : '—'}
            </Table.Td>
            <Table.Td ta="center">
              {r.place === 1 ? (
                <Text c="yellow" size="sm" fw={500}>
                  {formatProfit(r.profit)}
                </Text>
              ) : (
                formatProfit(r.profit)
              )}
            </Table.Td>
            <Table.Td ta="center">{formatPercentage(r.roi)}</Table.Td>
          </Table.Tr>
        ))}
      </Table.Tbody>
    </Table>
  )
}
