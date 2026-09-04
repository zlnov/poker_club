/**
 * Card — a standardized content panel with consistent styling.
 *
 * Wraps Mantine Card with default props for the Poker Club Design System.
 *
 * See 05_FE_UX.md section 2 (Design System — Cards) and
 * 04_FE_SPEC.md section 8 (Page layout).
 */

import { Card as MantineCard, type CardProps } from '@mantine/core'
import { forwardRef } from 'react'

export type PokerCardProps = CardProps

/**
 * Standardized card component with consistent padding, radius, and border.
 *
 * Usage:
 *   <Card>
 *     <Card.Section>...</Card.Section>
 *   </Card>
 */
export const Card = forwardRef<HTMLDivElement, PokerCardProps>(
  ({ children, ...rest }, ref) => {
    return (
      <MantineCard
        ref={ref}
        shadow="sm"
        padding="lg"
        radius="md"
        withBorder
        {...rest}
      >
        {children}
      </MantineCard>
    )
  },
)

Card.displayName = 'Card'

// Re-export Card.Section for convenience
export const CardSection = MantineCard.Section
