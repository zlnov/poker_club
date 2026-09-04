/**
 * Profile page component.
 *
 * Displays the current user's profile.
 * Phase 1 — placeholder page for the Application Shell.
 *
 * See 05_FE_UX.md section 4 (Application Navigation) and
 * 04_FE_SPEC.md section 11 (Routing).
 */

import { PageContainer, PageHeader, EmptyState } from '../components/ui'

/**
 * Profile page.
 *
 * Placeholder for the user profile screen.
 * Full implementation will come in a future phase.
 */
export function ProfilePage() {
  return (
    <PageContainer>
      <PageHeader title="Profile" description="Your account information" />

      <EmptyState
        title="Profile not configured"
        description="User profile management will be available in a future phase."
      />
    </PageContainer>
  )
}
