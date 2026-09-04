/**
 * LoadingState — displays a loading indicator with an optional message.
 *
 * Used for initial loading, data fetching, and async operations.
 *
 * See 04_FE_SPEC.md section 23 (Loading States) and
 * 05_FE_UX.md section 11 (Loading / Empty / Error States).
 */

import { Center, Loader, Stack, Text, type LoaderProps } from '@mantine/core'
import { forwardRef } from 'react'

export interface LoadingStateProps {
  /** Optional message to display below the loader. */
  message?: string
  /** Loader size. Defaults to 'md'. */
  size?: LoaderProps['size']
}

/**
 * Full-page loading state.
 * Centers the loader vertically and horizontally.
 */
export const LoadingState = forwardRef<HTMLDivElement, LoadingStateProps>(
  ({ message, size = 'md' }, ref) => {
    return (
      <Center ref={ref} style={{ minHeight: '200px' }} bg="transparent">
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
  },
)

LoadingState.displayName = 'LoadingState'

/**
 * Inline loading state.
 * Displays a small loader next to a message.
 */
export const InlineLoading = forwardRef<HTMLDivElement, LoadingStateProps>(
  ({ message = 'Loading...' }, ref) => {
    return (
      <Center ref={ref}>
        <Loader size="xs" />
        <Text size="sm" c="dimmed" ml="xs">
          {message}
        </Text>
      </Center>
    )
  },
)

InlineLoading.displayName = 'InlineLoading'
