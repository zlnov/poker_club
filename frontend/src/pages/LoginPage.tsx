/**
 * Login page component.
 *
 * Handles Standard Web authentication with login + password.
 * Uses Mantine Form for validation.
 *
 * See 05_FE_UX.md section 10 (Forms) and
 * 07_AUTH.md section 4 (Standard Web Authentication).
 */

import { useState } from 'react'
import { useNavigate, useLocation } from 'react-router-dom'
import {
  Button,
  Stack,
  TextInput,
  Title,
  Text,
  Alert,
  Group,
  Center,
  Paper,
} from '@mantine/core'
import { useForm } from '@mantine/form'
import { IconInfoCircle, IconLock, IconUser } from '@tabler/icons-react'
import { useAuth } from '../auth'
import { useTelegramEnvironment } from '../hooks'
import { ApiClientError } from '../api'

interface LoginFormValues {
  login: string
  password: string
}

/**
 * Login page.
 *
 * Displays login form for Standard Web authentication.
 * In Telegram Mini App mode, shows a message to use Telegram authentication.
 */
export function LoginPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const { login, isLoading: authLoading } = useAuth()
  const telegramEnv = useTelegramEnvironment()
  const [error, setError] = useState<string | null>(null)

  // Get redirect target from location state or default to /clubs
  const from = (location.state as { from?: string })?.from || '/clubs'

  const form = useForm<LoginFormValues>({
    initialValues: {
      login: '',
      password: '',
    },
    validate: {
      login: (value) =>
        value.length < 3 ? 'Login must be at least 3 characters' : null,
      password: (value) =>
        value.length < 12 ? 'Password must be at least 12 characters' : null,
    },
  })

  const handleSubmit = form.onSubmit(async (values) => {
    setError(null)
    try {
      await login(values)
      navigate(from, { replace: true })
    } catch (err) {
      if (err instanceof ApiClientError) {
        switch (err.code) {
          case 'INVALID_CREDENTIALS':
            setError('Invalid login or password')
            break
          case 'RATE_LIMIT_EXCEEDED':
            setError('Too many login attempts. Please try again later.')
            break
          default:
            setError(err.message || 'Login failed. Please try again.')
        }
      } else if (err instanceof Error) {
        setError(err.message || 'Login failed. Please try again.')
      } else {
        setError('An unexpected error occurred. Please try again.')
      }
    }
  })

  // In Telegram Mini App mode, show a different UI
  if (telegramEnv.isTelegram) {
    return (
      <Center style={{ minHeight: '100vh' }}>
        <Paper
          shadow="md"
          radius="md"
          p="xl"
          withBorder
          style={{ maxWidth: '400px', width: '100%' }}
        >
          <Stack gap="md" align="center">
            <Title order={1} ta="center">
              Poker Club
            </Title>
            <Text c="dimmed" ta="center">
              Telegram Mini App Mode
            </Text>

            <Stack gap="md" align="center">
              <Text ta="center" size="lg" fw={500}>
                Authentication via Telegram
              </Text>
              <Text ta="center" c="dimmed" size="sm">
                You are running in Telegram Mini App mode. Authentication is
                handled automatically via Telegram.
              </Text>
              <Alert
                color="blue"
                variant="filled"
                title="Auto-authentication"
                icon={<IconInfoCircle size={16} />}
              >
                The app will automatically authenticate you using your Telegram
                account. If authentication doesn't happen automatically, please
                restart the Mini App.
              </Alert>
            </Stack>
          </Stack>
        </Paper>
      </Center>
    )
  }

  return (
    <Center style={{ minHeight: '100vh' }}>
      <Paper
        shadow="md"
        radius="md"
        p="xl"
        withBorder
        style={{ maxWidth: '400px', width: '100%' }}
      >
        <Stack gap="lg" align="center">
          <Title order={1} ta="center">
            Poker Club
          </Title>
          <Text c="dimmed" ta="center">
            Sign in to your account
          </Text>

          <form onSubmit={handleSubmit} style={{ width: '100%' }}>
            <Stack gap="md">
              {error && (
                <Alert
                  color="red"
                  variant="filled"
                  title="Login Failed"
                  icon={<IconInfoCircle size={16} />}
                >
                  {error}
                </Alert>
              )}

              <TextInput
                {...form.getInputProps('login')}
                label="Login"
                placeholder="Enter your login"
                autoComplete="username"
                required
                error={form.errors.login}
                leftSection={<IconUser size={16} />}
              />

              <TextInput
                {...form.getInputProps('password')}
                type="password"
                label="Password"
                placeholder="Enter your password"
                autoComplete="current-password"
                required
                error={form.errors.password}
                leftSection={<IconLock size={16} />}
              />

              <Button
                type="submit"
                fullWidth
                size="lg"
                loading={authLoading || form.submitting}
              >
                Sign In
              </Button>
            </Stack>
          </form>

          <Group justify="center" gap="xs">
            <Text c="dimmed" size="sm">
              Running in Telegram?
            </Text>
            <Text size="sm" c="blue" fw={500}>
              Open the Mini App for automatic authentication
            </Text>
          </Group>
        </Stack>
      </Paper>
    </Center>
  )
}
