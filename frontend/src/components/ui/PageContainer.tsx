/**
 * PageContainer — a responsive content wrapper with max-width and horizontal padding.
 *
 * Provides a consistent content width across all pages, with responsive
 * behavior for mobile, tablet, and desktop.
 *
 * See 05_FE_UX.md section 19 (Responsive Design) and
 * 04_FE_SPEC.md section 21 (Adaptive design).
 */

import { Container, type ContainerProps } from '@mantine/core'
import { forwardRef } from 'react'

export interface PageContainerProps extends ContainerProps {}

/**
 * Responsive page container with a maximum content width.
 *
 * Usage:
 *   <PageContainer>
 *     <PageHeader title="Clubs" />
 *     <Card>...</Card>
 *   </PageContainer>
 */
export const PageContainer = forwardRef<HTMLDivElement, PageContainerProps>(
  ({ children, size = 'lg', py = 'lg', ...rest }, ref) => {
    return (
      <Container ref={ref} size={size} py={py} {...rest}>
        {children}
      </Container>
    )
  },
)

PageContainer.displayName = 'PageContainer'
