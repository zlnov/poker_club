/**
 * PlayerStatsSummary component.
 *
 * Displays player statistics summary as cards.
 * All values come from backend API.
 *
 * Per agent_task_7_tables.md:
 * - Web: 3 cards per row, 3 rows (7 cards in first 2 rows, 3 in last)
 * - Mobile: 3 cards per row, 4 rows (12 cards total)
 * - Integer values for non-percentage fields (no currency symbols)
 * - 2 decimal places for percentage fields (ROI, Winrate, Average Place)
 * - Positive values without +, negative with -
 * - Winner (Games Won > 0) highlighted with gold color and trophy icon
 *
 * Web layout:
 * Row 1: Всего игр | Games Won | Total Profit
 * Row 2: ROI % | Winrate % | Games in profit | Average Place
 * Row 3: Biggest Win | Biggest Loss | Total Invested
 * Row 4: Rebuy (count) | Buy-in (amount) | Rebuy (amount)
 *
 * Mobile layout:
 * Row 1: Всего игр | Games Won | Games in profit
 * Row 2: Biggest Win | Biggest Loss | Total Profit
 * Row 3: Average Place | ROI % | Winrate %
 * Row 4: Rebuy (count) | Buy-in (amount) | Total Invested
 */

import { Card, SimpleGrid, Group, Text, Stack } from '@mantine/core'
import {
  IconUsers,
  IconCash,
  IconTrendingUp,
  IconTrophy,
  IconPigMoney,
  IconChartBar,
} from '@tabler/icons-react'
import type { PlayerStatistics } from '../../types'
import { formatPercentage } from '../../utils'

interface PlayerStatsSummaryProps {
  stats: PlayerStatistics
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

/**
 * Renders a stat card with icon, label, and value.
 */
function StatCard({
  icon,
  label,
  value,
  valueColor,
  valueFw,
  valueSize,
}: {
  icon: React.ReactNode
  label: string
  value: React.ReactNode
  valueColor?: string
  valueFw?: number
  valueSize?: string
}) {
  return (
    <Card shadow="sm" padding="md" radius="md" withBorder ta="center">
      <Group gap="xs" mb="xs" justify="center">
        {icon}
        <Text size="sm" c="dimmed">
          {label}
        </Text>
      </Group>
      <Text
        size={valueSize ?? 'lg'}
        fw={valueFw ?? 600}
        c={valueColor}
        ta="center"
      >
        {value}
      </Text>
    </Card>
  )
}

export function PlayerStatsSummary({ stats }: PlayerStatsSummaryProps) {
  const isWinner = stats.gamesWon > 0

  return (
    <Stack gap="md" mb="lg">
      {/* Row 1: Всего игр | ROI % | Winrate % */}
      <SimpleGrid cols={3} spacing="md">
        <StatCard
          icon={<IconUsers size={20} />}
          label="Всего игр"
          value={formatInt(stats.totalGames)}
          valueSize="xl"
          valueFw={700}
        />

        <StatCard
          icon={<IconTrendingUp size={20} />}
          label="ROI %"
          value={formatPercentage(stats.roi)}
        />

        <StatCard
          icon={<IconTrophy size={20} />}
          label="Winrate %"
          value={formatPercentage(stats.winrate)}
        />
      </SimpleGrid>

      {/* Row 2: Games Won | Games in profit | Average Place */}
      <SimpleGrid cols={3} spacing="md">
        <StatCard
          icon={<IconTrophy size={20} />}
          label="Games Won"
          value={
            isWinner ? (
              <Group gap="xs" justify="center">
                <IconTrophy size={14} color="gold" />
                <span>{formatInt(stats.gamesWon)}</span>
              </Group>
            ) : (
              formatInt(stats.gamesWon)
            )
          }
          valueColor={isWinner ? 'yellow' : undefined}
          valueSize="xl"
          valueFw={700}
        />

        <StatCard
          icon={<IconPigMoney size={20} />}
          label="Games in profit"
          value={formatInt(stats.gamesInProfit)}
          valueSize="xl"
          valueFw={700}
        />

        <StatCard
          icon={<IconUsers size={20} />}
          label="Average Place"
          value={stats.avgPlace.toFixed(1)}
        />
      </SimpleGrid>

      {/* Row 3: Total Profit | Total Invested | Biggest Win | Biggest Loss */}
      <SimpleGrid cols={4} spacing="md">
        <StatCard
          icon={<IconChartBar size={20} />}
          label="Total Profit"
          value={formatProfit(stats.totalProfit)}
          valueColor={stats.totalProfit >= 0 ? 'green' : 'red'}
        />

        <StatCard
          icon={<IconPigMoney size={20} />}
          label="Total Invested"
          value={formatInt(stats.totalInvested)}
        />

        <StatCard
          icon={<IconTrendingUp size={20} />}
          label="Biggest Win"
          value={formatProfit(stats.biggestWin)}
          valueColor="green"
        />

        <StatCard
          icon={<IconTrophy size={20} />}
          label="Biggest Loss"
          value={formatProfit(stats.biggestLoss)}
          valueColor="red"
        />
      </SimpleGrid>

      {/* Row 4: Rebuy | Buy-in (stretched to full width) */}
      <SimpleGrid cols={3} spacing="md">
        <StatCard
          icon={<IconCash size={20} />}
          label="Rebuy Count"
          value={formatInt(stats.totalRebuyCount)}
          valueSize="xl"
          valueFw={700}
        />

        <StatCard
          icon={<IconCash size={20} />}
          label="Rebuy"
          value={formatInt(stats.totalRebuyAmount)}
          valueSize="xl"
          valueFw={700}
        />

        <StatCard
          icon={<IconCash size={20} />}
          label="Buy-in"
          value={formatInt(stats.totalBuyInAmount)}
        />
      </SimpleGrid>
    </Stack>
  )
}
