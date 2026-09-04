/**
 * Section — a vertical rhythm container for grouping related content.
 *
 * Provides consistent vertical spacing between groups of elements.
 *
 * See 05_FE_UX.md section 2 (Design System — Spacing) and
 * 04_FE_SPEC.md section 8 (Page layout).
 */

import { Stack, Title, type StackProps } from '@mantine/core'
import { forwardRef } from 'react'

export interface SectionProps extends StackProps {
  /** Section title (optional). */
  title?: string
}

/**
 * Vertical rhythm container for grouping related content.
 *
 * Usage:
 *   <Section title="Game Settings">
 *     <TextInput ... />
 *     <Button>Save</Button>
 *   </Section>
 */
export const Section = forwardRef<HTMLDivElement, SectionProps>(
  ({ children, title, ...rest }, ref) => {
    return (
      <Stack ref={ref} gap="lg" {...rest}>
        {title && (
          <Title order={4} mb={8}>
            {title}
          </Title>
        )}
        {children}
      </Stack>
    )
  },
)

Section.displayName = 'Section'
