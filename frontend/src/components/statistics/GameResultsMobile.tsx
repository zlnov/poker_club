/**
 * GameResultsMobile component.
 *
 * Displays game results in a mobile-friendly compact table format.
 * Winner (Place = 1) is highlighted with gold text color.
 * Tapping a row opens a Modal with additional details.
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
 */

import { Table, Text, Modal, Stack } from '@mantine/core'
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
        <Table.Tbody>
          {results.map((r) => (
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
              <Table.Td>
                <Text size="xs" fw={500}>
                  {r.playerName}
                </Text>
              </Table.Td>
              <Table.Td ta="right">
                {r.place === 1 ? (
                  <Text c="yellow" size="xs" fw={500}>
                    {formatProfit(r.profit)}
                  </Text>
                ) : (
                  <Text size="xs">{formatProfit(r.profit)}</Text>
                )}
              </Table.Td>
              <Table.Td ta="right">
                <Text size="xs">{formatPercentage(r.roi)}</Text>
              </Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>

      {/* Extended Details Modal */}
      <Modal
        opened={!!selectedResult}
        onClose={() => setSelectedResult(null)}
        title={selectedResult ? selectedResult.playerName : ''}
        centered
        size="md"
      >
        {selectedResult && (
          <Stack gap="sm">
            <Text>
              <Text span c="dimmed" size="xs">
                Invested:{' '}
              </Text>
              {formatInt(selectedResult.totalInvested)}
            </Text>
            <Text>
              <Text span c="dimmed" size="xs">
                Chips End:{' '}
              </Text>
              {selectedResult.chipsEnd !== undefined
                ? formatInt(selectedResult.chipsEnd)
                : '—'}
            </Text>
            <Text>
              <Text span c="dimmed" size="xs">
                Payout:{' '}
              </Text>
              {selectedResult.payoutAmount !== undefined
                ? formatInt(selectedResult.payoutAmount)
                : '—'}
            </Text>
          </Stack>
        )}
      </Modal>
    </>
  )
}
