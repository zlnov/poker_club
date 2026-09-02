/**
 * Loading state component.
 *
 * Displays a loading indicator with an optional message.
 * Used for initial loading, data fetching, and async operations.
 *
 * See 04_FE_SPEC.md section 23 (Loading States) and
 * 05_FE_UX.md section 11 (Loading / Empty / Error States).
 */

import { Center, Loader, Stack, Text } from '@mantine/core'

export interface LoadingStateProps {
  /** Optional message to display below the loader. */
  message?: string
  /** Loader size. Defaults to 'md'. */
  size?: 'xs' | 'sm' | 'md' | 'lg' | 'xl'
}

/**
 * Full-page loading state.
 * Centers the loader vertically and horizontally.
 */
export function LoadingState({ message, size = 'md' }: LoadingStateProps) {
  return (
    <Center style={{ minHeight: '200px' }} bg="transparent">
      <Stack align="center" gap="sm">
        <Loader size={size} />
        {message && (
          <Text size="sm" c="dimmed">
            {message}
          </Text>
        )}
      </Stack>
    </Center>
  )
}

/**
 * Inline loading state.
 * Displays a small loader next to a message.
 */
export function InlineLoading({
  message = 'Loading...',
}: {
  message?: string
}) {
  return (
    <Center>
      <Loader size="xs" />
      <Text size="sm" c="dimmed" ml="xs">
        {message}
      </Text>
    </Center>
  )
}
