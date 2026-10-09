/**
 * Login page — email/password + Continue with Telegram (Web OIDC).
 */

import { useState } from 'react'
import { useNavigate, useLocation, Link } from 'react-router-dom'
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
  Divider,
} from '@mantine/core'
import { useForm } from '@mantine/form'
import { IconBrandTelegram, IconInfoCircle, IconLock, IconMail } from '@tabler/icons-react'
import { useAuth } from '../auth'
import { requestTelegramIdToken } from '../auth/telegramWebLogin'
import { useTelegramEnvironment } from '../hooks'
import { ApiClientError } from '../api'
import { useApiClient } from '../hooks'
import { getEnvironment } from '../env'

interface LoginFormValues {
  email: string
  password: string
}

export function LoginPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const { login, loginWithTelegramWeb, isLoading: authLoading } = useAuth()
  const apiClient = useApiClient()
  const telegramEnv = useTelegramEnvironment()
  const [error, setError] = useState<string | null>(null)
  const [telegramLoading, setTelegramLoading] = useState(false)

  const from = (location.state as { from?: string })?.from || '/clubs'
  const env = getEnvironment()

  const form = useForm<LoginFormValues>({
    initialValues: {
      email: '',
      password: '',
    },
    validate: {
      email: (value) =>
        /^\S+@\S+\.\S+$/.test(value) ? null : 'Enter a valid email',
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
            setError('Invalid email or password')
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

  const handleTelegramContinue = async () => {
    setError(null)
    if (!env.telegramLoginClientId) {
      setError('Telegram Login is not configured (VITE_TELEGRAM_LOGIN_CLIENT_ID).')
      return
    }

    setTelegramLoading(true)
    try {
      const challenge = await apiClient.post<{
        challenge_token: string
        nonce: string
      }>('/auth/telegram/web/challenge', { skipAuth: true })

      if (!challenge.data.nonce || !challenge.data.challenge_token) {
        throw new Error('Invalid Telegram login challenge from server')
      }

      const idToken = await requestTelegramIdToken(
        env.telegramLoginClientId,
        challenge.data.nonce,
      )
      const result = await loginWithTelegramWeb(
        idToken,
        challenge.data.challenge_token,
      )

      if (result.registrationRequired && result.registrationToken) {
        navigate('/register', {
          replace: true,
          state: { registrationToken: result.registrationToken, from },
        })
        return
      }

      navigate(from, { replace: true })
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.message || 'Telegram login failed')
      } else if (err instanceof Error) {
        setError(err.message)
      } else {
        setError('Telegram login failed')
      }
    } finally {
      setTelegramLoading(false)
    }
  }

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
            <Alert
              color="blue"
              variant="filled"
              title="Auto-authentication"
              icon={<IconInfoCircle size={16} />}
            >
              Authentication is handled automatically via Telegram Mini App
              initData.
            </Alert>
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

          {error && (
            <Alert
              color="red"
              variant="filled"
              title="Login Failed"
              icon={<IconInfoCircle size={16} />}
              w="100%"
            >
              {error}
            </Alert>
          )}

          <Button
            fullWidth
            size="lg"
            variant="light"
            leftSection={<IconBrandTelegram size={18} />}
            loading={telegramLoading}
            onClick={handleTelegramContinue}
          >
            Continue with Telegram
          </Button>

          <Divider label="or" labelPosition="center" w="100%" />

          <form onSubmit={handleSubmit} style={{ width: '100%' }}>
            <Stack gap="md">
              <TextInput
                {...form.getInputProps('email')}
                label="Email"
                placeholder="you@example.com"
                autoComplete="email"
                required
                leftSection={<IconMail size={16} />}
              />

              <TextInput
                {...form.getInputProps('password')}
                type="password"
                label="Password"
                placeholder="Enter your password"
                autoComplete="current-password"
                required
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
              New here?
            </Text>
            <Text
              component={Link}
              to="/register"
              size="sm"
              c="blue"
              fw={500}
            >
              Create an account with Telegram
            </Text>
          </Group>
        </Stack>
      </Paper>
    </Center>
  )
}
