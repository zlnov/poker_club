/**
 * ErrorState — displays an error message with an optional retry action.
 *
 * Used when data fetching fails or an operation errors.
 *
 * See 04_FE_SPEC.md section 24 (Error Handling) and
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
import { IconAlertCircle } from '@tabler/icons-react'
import { forwardRef } from 'react'

export interface ErrorStateProps {
  /** Error message to display. */
  message?: string
  /** Optional retry handler. */
  onRetry?: () => void
  /** Whether a retry is in progress. */
  isRetrying?: boolean
  /** Optional error code for more specific messaging. */
  errorCode?: string
  /** Optional icon. Defaults to alert icon. */
  icon?: React.ReactNode
  /** Button props for the retry button. */
  retryButtonProps?: ButtonProps
}

/**
 * Full-page error state.
 * Centers the error message vertically and horizontally.
 */
export const ErrorState = forwardRef<HTMLDivElement, ErrorStateProps>(
  (
    {
      message = 'Something went wrong',
      onRetry,
      isRetrying = false,
      errorCode,
      icon,
      retryButtonProps,
    },
    ref,
  ) => {
    return (
      <Center ref={ref} style={{ minHeight: '200px' }} bg="transparent">
        <Stack align="center" gap="md" style={{ maxWidth: '400px' }}>
          {icon ?? <IconAlertCircle size={48} stroke={1.5} />}
          <Title order={4}>Error</Title>
          <Text size="sm" c="dimmed" ta="center">
            {message}
          </Text>
          {errorCode && (
            <Text size="xs" c="dimmed">
              Error code: {errorCode}
            </Text>
          )}
          {onRetry && (
            <Button
              onClick={onRetry}
              loading={isRetrying}
              leftSection={<IconAlertCircle size={16} />}
              {...retryButtonProps}
            >
              Try again
            </Button>
          )}
        </Stack>
      </Center>
    )
  },
)

ErrorState.displayName = 'ErrorState'
