/**
 * EmptyState — displays a message when there is no data to show.
 *
 * Provides a consistent visual treatment for empty content areas,
 * optionally with a call-to-action button.
 *
 * See 04_FE_SPEC.md section 23 (Loading States) and
 * 05_FE_UX.md section 11 (Loading / Empty / Error States).
 */

import {
  Button,
  Center,
  Stack,
  Text,
  Title,
  type ButtonProps,
} from '@mantine/core'
import { IconInfoCircle } from '@tabler/icons-react'
import { forwardRef } from 'react'

export interface EmptyStateProps {
  /** Title for the empty state. */
  title?: string
  /** Description explaining why the content is empty. */
  description?: string
  /** Optional action button label. */
  actionLabel?: string
  /** Optional action handler. */
  onAction?: () => void
  /** Optional icon. Defaults to info icon. */
  icon?: React.ReactNode
  /** Button props for the action button. */
  actionButtonProps?: ButtonProps
  /** Whether the action is loading. */
  actionLoading?: boolean
}

/**
 * Full-page empty state.
 * Centers the message vertically and horizontally.
 */
export const EmptyState = forwardRef<HTMLDivElement, EmptyStateProps>(
  (
    {
      title = 'No data',
      description = 'There is nothing to display here yet.',
      actionLabel,
      onAction,
      icon,
      actionButtonProps,
      actionLoading = false,
    },
    ref,
  ) => {
    return (
      <Center ref={ref} style={{ minHeight: '200px' }} bg="transparent">
        <Stack align="center" gap="md" style={{ maxWidth: '400px' }}>
          {icon ?? <IconInfoCircle size={48} stroke={1.5} />}
          <Title order={4}>{title}</Title>
          <Text size="sm" c="dimmed" ta="center">
            {description}
          </Text>
          {actionLabel && onAction && (
            <Button
              onClick={onAction}
              loading={actionLoading}
              {...actionButtonProps}
            >
              {actionLabel}
            </Button>
          )}
        </Stack>
      </Center>
    )
  },
)

EmptyState.displayName = 'EmptyState'
