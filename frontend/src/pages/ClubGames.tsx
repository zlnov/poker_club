/**
 * Club games page component.
 *
 * Displays the list of games for a club.
 * Phase 1 — placeholder page for the Application Shell.
 *
 * See 05_FE_UX.md section 4 (Application Navigation) and
 * 04_FE_SPEC.md section 11 (Routing).
 */

import { Center, Text, Title } from '@mantine/core'

/**
 * Club games page.
 *
 * Placeholder for the club games screen.
 * Full implementation will come in RM_FE_05 (Games).
 */
export function ClubGames() {
  return (
    <Center style={{ minHeight: '400px' }}>
      <div style={{ textAlign: 'center' }}>
        <Title order={2} mb="sm">
          Games
        </Title>
        <Text c="dimmed">
          Club games list will be implemented in a future phase.
        </Text>
      </div>
    </Center>
  )
}
