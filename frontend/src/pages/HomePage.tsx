/**
 * Home page component.
 *
 * Displays a landing page that confirms the application is running.
 * Shows the current environment (Standard Web or Telegram Mini App).
 *
 * See 05_FE_UX.md section 4 (Application Navigation) and
 * 04_FE_SPEC.md section 11 (Routing).
 */

import { Card, Text, Title, Stack, Group } from '@mantine/core'
import { IconChessKing, IconWorld } from '@tabler/icons-react'
import { useTelegramEnvironment } from '../hooks'
import { PageContainer, PageHeader } from '../components/ui'

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
    <PageContainer>
      <PageHeader
        title="Poker Club"
        description="Manage your poker clubs, games, and statistics"
      />

      <Card mt="lg">
        <Stack gap="md">
          <Group gap="sm">
            <IconChessKing size={24} />
            <Title order={4} mb={0}>
              Application
            </Title>
          </Group>

          <Stack gap="xs">
            <Group gap="sm">
              <IconWorld size={16} />
              <Text size="sm">
                <strong>Environment:</strong>{' '}
                {telegramEnv.isTelegram ? 'Telegram Mini App' : 'Standard Web'}
              </Text>
            </Group>
            <Text size="sm">
              <strong>API Base URL:</strong>{' '}
              {import.meta.env.VITE_API_BASE_URL || 'not configured'}
            </Text>
          </Stack>
        </Stack>
      </Card>
    </PageContainer>
  )
}
