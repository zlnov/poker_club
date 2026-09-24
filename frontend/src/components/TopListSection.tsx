/**
 * Top List section component for the Club Dashboard.
 *
 * Displays the TOP LIST block with two cards:
 * - Top 3 players by total profit
 * - Top 3 players by ROI (min 5 games)
 *
 * Layout:
 * - Web: two cards in a row (grid 2 columns)
 * - Mobile: two cards in a column
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
  SimpleGrid,
} from '@mantine/core'
import { IconChartBar } from '@tabler/icons-react'
import { useMediaQuery } from '@mantine/hooks'
import { useClubMemberStatistics } from '../features'
import { backgroundSurfaces } from '../styles'

interface TopListSectionProps {
  /** Club ID. */
  clubId: number
}

/**
 * Medal colors for top 3 positions.
 */
const MEDAL_COLORS = ['gold', 'silver', 'bronze'] as const

/**
 * Formats a profit value with sign and space as thousands separator.
 */
function formatProfit(value: number): string {
  const rounded = Math.round(value)
  const sign = rounded >= 0 ? '+' : ''
  const absValue = Math.abs(rounded)
  const formatted = absValue.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ' ')
  return `${sign}${formatted}`
}

/**
 * Formats ROI percentage with one decimal place and sign.
 */
function formatRoi(value: number): string {
  const sign = value >= 0 ? '+' : ''
  return `${sign}${value.toFixed(1)} %`
}

/**
 * Returns the color for a medal position.
 */
function getMedalColor(position: number): string {
  return MEDAL_COLORS[position] ?? 'gray'
}

/**
 * Formats player name for display.
 */
function formatPlayerName(name: string): string {
  return name || '—'
}

/**
 * Top 3 players by total profit card.
 */
function TopProfitCard({
  members,
}: {
  members: ReturnType<typeof useClubMemberStatistics>['data']
}) {
  const topPlayers = members
    ? [...members].sort((a, b) => b.profit - a.profit).slice(0, 3)
    : []

  return (
    <Card
      padding="md"
      radius="md"
      withBorder
      bg={backgroundSurfaces.card}
      style={{ height: '100%', display: 'flex', flexDirection: 'column' }}
    >
      <Title order={5} mb="sm">
        <Text size="s">ТОП-3 ПО ПРИБЫЛИ</Text>
      </Title>

      {topPlayers.length === 0 ? (
        <Stack
          style={{ flex: 1, justifyContent: 'center', alignItems: 'center' }}
        >
          <Text size="sm" c="dimmed" ta="center">
            Пока нет завершённых игр
          </Text>
        </Stack>
      ) : (
        <Table
          variant="unstyled"
          style={{ flex: 1, tableLayout: 'fixed' }}
          mt="auto"
        >
          <thead>
            <tr>
              <th style={{ width: '40px', textAlign: 'center' }}>#</th>
              <th style={{ textAlign: 'left' }}>Игрок</th>
              <th style={{ width: '80px', textAlign: 'center' }}>Игр</th>
              <th style={{ width: '140px', textAlign: 'right' }}>
                Total Profit
              </th>
            </tr>
          </thead>
          <tbody>
            {topPlayers.map((member, index) => (
              <tr key={member.playerId}>
                <td style={{ textAlign: 'center' }}>
                  <Text size="sm" fw={700} c={getMedalColor(index)}>
                    {index + 1}
                  </Text>
                </td>
                <td>
                  <Text size="sm">@{formatPlayerName(member.playerName)}</Text>
                </td>
                <td style={{ textAlign: 'center' }}>
                  <Text size="sm">{member.games}</Text>
                </td>
                <td style={{ textAlign: 'right' }}>
                  <Text
                    size="sm"
                    fw={600}
                    c={member.profit >= 0 ? 'green' : 'red'}
                  >
                    {formatProfit(member.profit)}
                  </Text>
                </td>
              </tr>
            ))}
          </tbody>
        </Table>
      )}
    </Card>
  )
}

/**
 * Top 3 players by ROI card (min 5 games).
 */
function TopRoiCard({
  members,
}: {
  members: ReturnType<typeof useClubMemberStatistics>['data']
}) {
  const topPlayers = members
    ? [...members]
        .filter((m) => m.games >= 5)
        .sort((a, b) => b.roi - a.roi)
        .slice(0, 3)
    : []

  return (
    <Card
      padding="md"
      radius="md"
      withBorder
      bg={backgroundSurfaces.card}
      style={{ height: '100%', display: 'flex', flexDirection: 'column' }}
    >
      <Title order={5} mb="xs">
        <Text size="s">ТОП-3 ПО ROI</Text>
      </Title>
      <Text size="xs" c="dimmed" mb="sm">
        (минимум 5 игр)
      </Text>

      {topPlayers.length === 0 ? (
        <Stack
          style={{ flex: 1, justifyContent: 'center', alignItems: 'center' }}
        >
          <Text size="sm" c="dimmed" ta="center">
            Пока нет завершённых игр
          </Text>
        </Stack>
      ) : (
        <Table
          variant="unstyled"
          style={{ flex: 1, tableLayout: 'fixed' }}
          mt="auto"
        >
          <thead>
            <tr>
              <th style={{ width: '40px', textAlign: 'center' }}>#</th>
              <th style={{ textAlign: 'left' }}>Игрок</th>
              <th style={{ width: '80px', textAlign: 'center' }}>Игр</th>
              <th style={{ width: '140px', textAlign: 'right' }}>ROI %</th>
            </tr>
          </thead>
          <tbody>
            {topPlayers.map((member, index) => (
              <tr key={member.playerId}>
                <td style={{ textAlign: 'center' }}>
                  <Text size="sm" fw={700} c={getMedalColor(index)}>
                    {index + 1}
                  </Text>
                </td>
                <td>
                  <Text size="sm">@{formatPlayerName(member.playerName)}</Text>
                </td>
                <td style={{ textAlign: 'center' }}>
                  <Text size="sm">{member.games}</Text>
                </td>
                <td style={{ textAlign: 'right' }}>
                  <Text
                    size="sm"
                    fw={600}
                    c={member.roi >= 0 ? 'green' : 'red'}
                  >
                    {formatRoi(member.roi)}
                  </Text>
                </td>
              </tr>
            ))}
          </tbody>
        </Table>
      )}
    </Card>
  )
}

/**
 * Top List section component.
 *
 * Displays the TOP LIST block with two cards:
 * - Top 3 players by total profit
 * - Top 3 players by ROI (min 5 games)
 *
 * Layout adapts to mobile (column) and desktop (row).
 */
export function TopListSection({ clubId }: TopListSectionProps) {
  const isMobile = useMediaQuery('(max-width: 768px)')

  const { data: members } = useClubMemberStatistics(clubId, 'cash')

  return (
    <Card
      mt="lg"
      padding="lg"
      radius="md"
      withBorder
      bg={backgroundSurfaces.card}
    >
      <Group gap="xs" mb="md">
        <IconChartBar size={20} />
        <Title order={4} mb={0}>
          TOP LIST
        </Title>
      </Group>

      <SimpleGrid
        cols={isMobile ? 1 : 2}
        spacing="md"
        style={{ alignItems: 'stretch' }}
      >
        <TopProfitCard members={members} />
        <TopRoiCard members={members} />
      </SimpleGrid>
    </Card>
  )
}
