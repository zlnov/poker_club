/**
 * ClubMemberStatsTable component.
 *
 * Displays club member statistics in a web table format.
 * All values come from backend API.
 *
 * Per agent_task_7_tables.md:
 * Columns: Игрок | Игр | Total Invested | Profit | ROI % | Winrate % | Среднее место | Побед
 * - Winner (gamesWon > 0) highlighted with gold text color
 * - Integer values for non-percentage fields
 * - 2 decimal places for percentage fields (ROI, Winrate, Average Place)
 * - Positive values without +, negative with -
 * - Default sort: Побед (wins) descending
 */

import { Table, Text } from '@mantine/core'
import { IconTrophy } from '@tabler/icons-react'
import type { ClubMemberStatistics } from '../../types'
import { formatPercentage } from '../../utils'

interface ClubMemberStatsTableProps {
  members: ClubMemberStatistics[]
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

export function ClubMemberStatsTable({ members }: ClubMemberStatsTableProps) {
  // Sort by gamesWon descending by default
  const sortedMembers = [...members].sort((a, b) => b.gamesWon - a.gamesWon)

  return (
    <Table variant="striped">
      <Table.Thead>
        <Table.Tr>
          <Table.Th>Игрок</Table.Th>
          <Table.Th ta="center">Игр</Table.Th>
          <Table.Th ta="center">Total Invested</Table.Th>
          <Table.Th ta="center">Profit</Table.Th>
          <Table.Th ta="center">ROI %</Table.Th>
          <Table.Th ta="center">Winrate %</Table.Th>
          <Table.Th ta="center">Среднее место</Table.Th>
          <Table.Th ta="center">Побед</Table.Th>
        </Table.Tr>
      </Table.Thead>
      <Table.Tbody>
        {sortedMembers.map((m) => (
          <Table.Tr key={m.playerId}>
            <Table.Td>{m.playerName}</Table.Td>
            <Table.Td ta="center">{formatInt(m.games)}</Table.Td>
            <Table.Td ta="center">{formatInt(m.totalInvested)}</Table.Td>
            <Table.Td ta="center">
              {m.profit >= 0 ? (
                <Text c="green" size="sm" fw={500}>
                  {formatProfit(m.profit)}
                </Text>
              ) : (
                <Text c="red" size="sm" fw={500}>
                  {formatProfit(m.profit)}
                </Text>
              )}
            </Table.Td>
            <Table.Td ta="center">{formatPercentage(m.roi)}</Table.Td>
            <Table.Td ta="center">{formatPercentage(m.winrate)}</Table.Td>
            <Table.Td ta="center">{m.avgPlace.toFixed(1)}</Table.Td>
            <Table.Td ta="center">
              {m.gamesWon > 0 ? (
                <Text c="yellow" size="sm" fw={500}>
                  <IconTrophy size={14} /> {formatInt(m.gamesWon)}
                </Text>
              ) : (
                formatInt(m.gamesWon)
              )}
            </Table.Td>
          </Table.Tr>
        ))}
      </Table.Tbody>
    </Table>
  )
}
