/**
 * Home page component.
 *
 * Displays a landing page that confirms the application is running.
 * Shows the current environment (Standard Web or Telegram Mini App).
 *
 * See 05_FE_UX.md section 4 (Application Navigation) and
 * 04_FE_SPEC.md section 11 (Routing).
 */

import { Card, Text, Title, Stack } from '@mantine/core'
import { useTelegramEnvironment } from '../hooks'

/**
 * Home page.
 *
 * Displays:
 * - Application name
 * - Current environment (Standard Web or Telegram Mini App)
 * - Confirmation that the app is running
 */
export function HomePage() {
  const telegramEnv = useTelegramEnvironment()

  return (
    <Stack
      align="center"
      style={{ minHeight: '400px', justifyContent: 'center' }}
    >
      <Title order={1} ta="center">
        Poker Club
      </Title>
      <Text c="dimmed" ta="center">
        Application Shell — Phase 1
      </Text>

      <Card
        shadow="sm"
        padding="lg"
        radius="md"
        withBorder
        style={{ maxWidth: '400px', width: '100%' }}
      >
        <Stack gap="xs">
          <Text size="sm">
            <strong>Environment:</strong>{' '}
            {telegramEnv.isTelegram ? 'Telegram Mini App' : 'Standard Web'}
          </Text>
          <Text size="sm">
            <strong>API Base URL:</strong>{' '}
            {import.meta.env.VITE_API_BASE_URL || 'not configured'}
          </Text>
        </Stack>
      </Card>
    </Stack>
  )
}
