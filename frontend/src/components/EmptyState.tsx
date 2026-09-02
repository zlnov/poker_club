/**
 * Empty state component.
 *
 * Displays a message when there is no data to show,
 * optionally with a call-to-action button.
 *
 * See 04_FE_SPEC.md section 23 (Loading States) and
 * 05_FE_UX.md section 11 (Loading / Empty / Error States).
 */

import { Button, Center, Stack, Text, Title } from '@mantine/core'
import { IconInfoCircle } from '@tabler/icons-react'

export interface EmptyStateProps {
  /** Title for the empty state. */
  title?: string
  /** Description explaining why the content is empty. */
  description?: string
  /** Optional action button label. */
  actionLabel?: string
  /** Optional action handler. */
  onAction?: () => void
}

/**
 * Full-page empty state.
 * Centers the message vertically and horizontally.
 */
export function EmptyState({
  title = 'No data',
  description = 'There is nothing to display here yet.',
  actionLabel,
  onAction,
}: EmptyStateProps) {
  return (
    <Center style={{ minHeight: '200px' }} bg="transparent">
      <Stack align="center" gap="md" style={{ maxWidth: '400px' }}>
        <IconInfoCircle size={48} stroke={1.5} />
        <Title order={4}>{title}</Title>
        <Text size="sm" c="dimmed" ta="center">
          {description}
        </Text>
        {actionLabel && onAction && (
          <Button onClick={onAction}>{actionLabel}</Button>
        )}
      </Stack>
    </Center>
  )
}
