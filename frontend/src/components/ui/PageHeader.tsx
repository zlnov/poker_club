/**
 * PageHeader — a standardized page title with optional description and actions.
 *
 * Provides consistent typography and spacing for page headers across
 * all screens.
 *
 * See 05_FE_UX.md section 2 (Design System — Typography) and
 * 04_FE_SPEC.md section 11 (Routing).
 */

import {
  Group,
  Stack,
  Text,
  Title,
  type TitleProps,
  type TextProps,
  type GroupProps,
} from '@mantine/core'
import { forwardRef } from 'react'

export interface PageHeaderProps extends GroupProps {
  /** Page title. */
  title: string
  /** Optional description text below the title. */
  description?: string
  /** Optional action elements (buttons, etc.) aligned to the right. */
  action?: React.ReactNode
  /** Title order (HTML heading level). Defaults to 2. */
  titleOrder?: TitleProps['order']
  /** Description text props. */
  descriptionProps?: TextProps
}

/**
 * Standardized page header with title, optional description, and actions.
 *
 * Usage:
 *   <PageHeader
 *     title="Clubs"
 *     description="Manage your poker clubs"
 *     action={<Button>Create Club</Button>}
 *   />
 */
export const PageHeader = forwardRef<HTMLDivElement, PageHeaderProps>(
  (
    { title, description, action, titleOrder = 2, descriptionProps, ...rest },
    ref,
  ) => {
    return (
      <Group ref={ref} justify="space-between" {...rest}>
        <Stack gap={4}>
          <Title order={titleOrder}>{title}</Title>
          {description && (
            <Text size="sm" c="dimmed" {...descriptionProps}>
              {description}
            </Text>
          )}
        </Stack>
        {action}
      </Group>
    )
  },
)

PageHeader.displayName = 'PageHeader'
