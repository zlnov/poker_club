/**
 * CashTournamentToggle component.
 *
 * Provides a toggle between Cash and Tournament statistics views.
 * Tournament is enabled but shows a placeholder when selected.
 */

import { SegmentedControl, Text, Tooltip, Group } from '@mantine/core'
import { IconTrophy } from '@tabler/icons-react'

interface CashTournamentToggleProps {
  value: 'cash' | 'tournament'
  onChange: (value: 'cash' | 'tournament') => void
  tournamentPlaceholder?: boolean
}

export function CashTournamentToggle({
  value,
  onChange,
  tournamentPlaceholder = false,
}: CashTournamentToggleProps) {
  return (
    <Group gap="sm" mb="md">
      <SegmentedControl
        value={value}
        onChange={(v) => onChange(v as 'cash' | 'tournament')}
        data={[
          { label: 'Cash', value: 'cash' },
          { label: 'Tournament', value: 'tournament' },
        ]}
      />
      {value === 'tournament' && tournamentPlaceholder && (
        <Tooltip
          label="Tournament statistics are not available yet"
          position="top"
          withArrow
        >
          <Text size="xs" c="dimmed" style={{ cursor: 'help' }}>
            <IconTrophy size={14} />
          </Text>
        </Tooltip>
      )}
    </Group>
  )
}
