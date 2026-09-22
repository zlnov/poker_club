/**
 * GameResultsMobile component.
 *
 * Displays game results in a mobile-friendly compact table format.
 * Winner (Place = 1) is highlighted with gold text color.
 * Tapping a row opens a Modal with a table showing additional details.
 *
 * Per agent_task_7_tables.md:
 * Main table columns: Place | Player | Profit | ROI
 * Extended table (modal): Invested | Chips End | Payout
 * - Winner (Place = 1) highlighted with gold text color
 * - 0.1rem gap between parameter and value
 * - Column headers: mantine-font-size-xs, mantine-color-dimmed
 * - Integer values for non-percentage fields
 * - 2 decimal places for percentage fields (ROI)
 * - Positive values without +, negative with -
 * - Default sort: Place ascending
 * - Negative profit values shown in red
 */

import { Table, Text, Modal, Button, Stack } from '@mantine/core'
import { IconTrophy } from '@tabler/icons-react'
import { useState } from 'react'
import type { GameResult } from '../../types'
import { formatPercentage } from '../../utils'

interface GameResultsMobileProps {
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

export function GameResultsMobile({ results }: GameResultsMobileProps) {
  const [selectedResult, setSelectedResult] = useState<GameResult | null>(null)

  // Sort by Place ascending by default
  const sortedResults = [...results].sort((a, b) => {
    const aPlace = a.place ?? 0
    const bPlace = b.place ?? 0
    return aPlace - bPlace
  })

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
            <Table.Th ta="center">Place</Table.Th>
            <Table.Th ta="left">Player</Table.Th>
            <Table.Th ta="center">Profit</Table.Th>
            <Table.Th ta="center">ROI</Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {sortedResults.map((r) => (
            <Table.Tr
              key={r.playerId}
              onClick={() => setSelectedResult(r)}
              style={{ cursor: 'pointer' }}
            >
              <Table.Td ta="center">
                {r.place === 1 ? (
                  <Text c="yellow" size="xs" fw={500}>
                    <IconTrophy size={12} /> {r.place}
                  </Text>
                ) : (
                  <Text size="xs">{r.place}</Text>
                )}
              </Table.Td>
              <Table.Td ta="left">
                <Text size="xs" fw={500}>
                  {r.playerName}
                </Text>
              </Table.Td>
              <Table.Td ta="center">
                {r.profit < 0 ? (
                  <Text c="red" size="xs" fw={500}>
                    {formatProfit(r.profit)}
                  </Text>
                ) : r.place === 1 ? (
                  <Text c="yellow" size="xs" fw={500}>
                    {formatProfit(r.profit)}
                  </Text>
                ) : (
                  <Text size="xs">{formatProfit(r.profit)}</Text>
                )}
              </Table.Td>
              <Table.Td ta="center">
                <Text size="xs">{formatPercentage(r.roi)}</Text>
              </Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>

      {/* Extended Details Modal with table format */}
      <Modal
        opened={!!selectedResult}
        onClose={() => setSelectedResult(null)}
        title={selectedResult ? selectedResult.playerName : ''}
        centered
        size="md"
      >
        {selectedResult && (
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
                  <Table.Th ta="center">Invested</Table.Th>
                  <Table.Th ta="center">Chips End</Table.Th>
                  <Table.Th ta="center">Payout</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                <Table.Tr>
                  <Table.Td ta="center">
                    {formatInt(selectedResult.totalInvested)}
                  </Table.Td>
                  <Table.Td ta="center">
                    {selectedResult.chipsEnd !== undefined
                      ? formatInt(selectedResult.chipsEnd)
                      : '—'}
                  </Table.Td>
                  <Table.Td ta="center">
                    {selectedResult.payoutAmount !== undefined
                      ? formatInt(selectedResult.payoutAmount)
                      : '—'}
                  </Table.Td>
                </Table.Tr>
              </Table.Tbody>
            </Table>
            <Button
              color="blue"
              onClick={() => {
                // Navigate to game details page
                window.location.href = `/games/${selectedResult.gameId || 0}`
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
