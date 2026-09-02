/**
 * Profile page component.
 *
 * Displays the current user's profile.
 * Phase 1 — placeholder page for the Application Shell.
 *
 * See 05_FE_UX.md section 4 (Application Navigation) and
 * 04_FE_SPEC.md section 11 (Routing).
 */

import { Center, Text, Title } from '@mantine/core'

/**
 * Profile page.
 *
 * Placeholder for the user profile screen.
 * Full implementation will come in a future phase.
 */
export function ProfilePage() {
  return (
    <Center style={{ minHeight: '400px' }}>
      <div style={{ textAlign: 'center' }}>
        <Title order={2} mb="sm">
          Profile
        </Title>
        <Text c="dimmed">
          User profile will be implemented in a future phase.
        </Text>
      </div>
    </Center>
  )
}
