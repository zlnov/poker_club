/**
 * Club settings page component.
 *
 * Displays the settings for a club.
 * Phase 1 — placeholder page for the Application Shell.
 *
 * See 05_FE_UX.md section 4 (Application Navigation) and
 * 04_FE_SPEC.md section 11 (Routing).
 */

import { Center, Text, Title } from '@mantine/core'

/**
 * Club settings page.
 *
 * Placeholder for the club settings screen.
 * Full implementation will come in RM_FE_03 (Clubs & Membership).
 */
export function ClubSettings() {
  return (
    <Center style={{ minHeight: '400px' }}>
      <div style={{ textAlign: 'center' }}>
        <Title order={2} mb="sm">
          Settings
        </Title>
        <Text c="dimmed">
          Club settings will be implemented in a future phase.
        </Text>
      </div>
    </Center>
  )
}
