/**
 * ClubMemberStatsMobile component.
 *
 * Displays club member statistics in a mobile-friendly compact table format.
 * Tapping a row opens a Modal with a table showing additional details.
 *
 * Per agent_task_7_tables.md:
 * Main table columns: Игрок | Игр | Среднее место | Побед
 * Extended table (modal): Total Invested | Profit | ROI % | Winrate %
 * - Winner (gamesWon > 0) highlighted with gold text color
 * - 0.1rem gap between parameter and value
 * - Column headers: mantine-font-size-xs, mantine-color-dimmed
 * - Integer values for non-percentage fields
 * - 2 decimal places for percentage fields (ROI, Winrate, Average Place)
 * - Positive values without +, negative with -
 * - Default sort: Побед (wins) descending
 */

import { Table, Text, Modal, Stack } from '@mantine/core'
import { IconTrophy } from '@tabler/icons-react'
import { useState } from 'react'
import type { ClubMemberStatistics } from '../../types'
import { formatPercentage } from '../../utils'

interface ClubMemberStatsMobileProps {
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

export function ClubMemberStatsMobile({ members }: ClubMemberStatsMobileProps) {
  const [selectedMember, setSelectedMember] =
    useState<ClubMemberStatistics | null>(null)

  // Sort by gamesWon descending by default
  const sortedMembers = [...members].sort((a, b) => b.gamesWon - a.gamesWon)

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
            <Table.Th>Игрок</Table.Th>
            <Table.Th ta="center">Игр</Table.Th>
            <Table.Th ta="center">Среднее место</Table.Th>
            <Table.Th ta="center">Побед</Table.Th>
          </Table.Tr>
        </Table.Thead>
        <Table.Tbody>
          {sortedMembers.map((m) => (
            <Table.Tr
              key={m.playerId}
              onClick={() => setSelectedMember(m)}
              style={{ cursor: 'pointer' }}
            >
              <Table.Td>
                <Text size="xs" fw={500}>
                  {m.playerName}
                </Text>
              </Table.Td>
              <Table.Td ta="center">
                <Text size="xs">{formatInt(m.games)}</Text>
              </Table.Td>
              <Table.Td ta="center">
                <Text size="xs">{m.avgPlace.toFixed(1)}</Text>
              </Table.Td>
              <Table.Td ta="center">
                {m.gamesWon > 0 ? (
                  <Text c="yellow" size="xs" fw={500}>
                    <IconTrophy size={12} /> {formatInt(m.gamesWon)}
                  </Text>
                ) : (
                  <Text size="xs">{formatInt(m.gamesWon)}</Text>
                )}
              </Table.Td>
            </Table.Tr>
          ))}
        </Table.Tbody>
      </Table>

      {/* Member Details Modal with table format */}
      <Modal
        opened={!!selectedMember}
        onClose={() => setSelectedMember(null)}
        title={selectedMember ? selectedMember.playerName : ''}
        centered
        size="md"
      >
        {selectedMember && (
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
                  <Table.Th>Total Invested</Table.Th>
                  <Table.Th ta="center">Profit</Table.Th>
                  <Table.Th ta="center">ROI %</Table.Th>
                  <Table.Th ta="center">Winrate %</Table.Th>
                </Table.Tr>
              </Table.Thead>
              <Table.Tbody>
                <Table.Tr>
                  <Table.Td>{formatInt(selectedMember.totalInvested)}</Table.Td>
                  <Table.Td ta="center">
                    {selectedMember.profit >= 0 ? (
                      <Text c="green" size="sm" fw={500}>
                        {formatProfit(selectedMember.profit)}
                      </Text>
                    ) : (
                      <Text c="red" size="sm" fw={500}>
                        {formatProfit(selectedMember.profit)}
                      </Text>
                    )}
                  </Table.Td>
                  <Table.Td ta="center">
                    {formatPercentage(selectedMember.roi)}
                  </Table.Td>
                  <Table.Td ta="center">
                    {formatPercentage(selectedMember.winrate)}
                  </Table.Td>
                </Table.Tr>
              </Table.Tbody>
            </Table>
          </Stack>
        )}
      </Modal>
    </>
  )
}
