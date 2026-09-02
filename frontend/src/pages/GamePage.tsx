/**
 * Game page component.
 *
 * Displays the details of a single game.
 * Phase 1 — placeholder page for the Application Shell.
 *
 * See 05_FE_UX.md section 4 (Application Navigation) and
 * 04_FE_SPEC.md section 11 (Routing).
 */

import { Center, Text, Title } from '@mantine/core'

/**
 * Game page.
 *
 * Placeholder for the game details screen.
 * Full implementation will come in RM_FE_05 (Games) and
 * RM_FE_06 (Active Game).
 */
export function GamePage() {
  return (
    <Center style={{ minHeight: '400px' }}>
      <div style={{ textAlign: 'center' }}>
        <Title order={2} mb="sm">
          Game
        </Title>
        <Text c="dimmed">
          Game details will be implemented in a future phase.
        </Text>
      </div>
    </Center>
  )
}
