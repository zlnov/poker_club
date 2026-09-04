/**
 * Clubs page component.
 *
 * Displays the list of clubs the user belongs to.
 * Phase 1 — placeholder page for the Application Shell.
 *
 * See 05_FE_UX.md section 4 (Application Navigation) and
 * 04_FE_SPEC.md section 11 (Routing).
 */

import { PageContainer, PageHeader, EmptyState } from '../components/ui'

/**
 * Clubs page.
 *
 * Placeholder for the club list screen.
 * Full implementation will come in RM_FE_03 (Clubs & Membership).
 */
export function ClubsPage() {
  return (
    <PageContainer>
      <PageHeader title="Clubs" description="Your poker clubs" />

      <EmptyState
        title="No clubs yet"
        description="You haven't joined or created any clubs. Club management will be available in a future phase."
      />
    </PageContainer>
  )
}
