/**
 * Club statistics page component.
 *
 * Displays the statistics for a club.
 * Phase 1 — placeholder page for the Application Shell.
 *
 * See 05_FE_UX.md section 4 (Application Navigation) and
 * 04_FE_SPEC.md section 11 (Routing).
 */

import { Center, Text, Title } from '@mantine/core'

/**
 * Club statistics page.
 *
 * Placeholder for the club statistics screen.
 * Full implementation will come in RM_FE_07 (Results & Statistics).
 */
export function ClubStatistics() {
  return (
    <Center style={{ minHeight: '400px' }}>
      <div style={{ textAlign: 'center' }}>
        <Title order={2} mb="sm">
          Statistics
        </Title>
        <Text c="dimmed">
          Club statistics will be implemented in a future phase.
        </Text>
      </div>
    </Center>
  )
}
