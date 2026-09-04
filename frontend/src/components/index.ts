/**
 * Components barrel exports.
 *
 * See 04_FE_SPEC.md section 9 (Frontend Architecture).
 */

export { RootLayout } from './RootLayout'
export { Navigation } from './Navigation'
export { ThemeToggle } from './ThemeToggle'
export { ClubLayout } from './ClubLayout'
export { LoadingState, InlineLoading } from './LoadingState'
export { ErrorState } from './ErrorState'
export { EmptyState } from './EmptyState'
export { ProtectedRoute, PublicRoute } from './ProtectedRoute'

// UI primitives
export {
  PageContainer,
  PageHeader,
  Section,
  Card,
  CardSection,
  EmptyState as UiEmptyState,
  LoadingState as UiLoadingState,
  InlineLoading as UiInlineLoading,
  ErrorState as UiErrorState,
} from './ui'
