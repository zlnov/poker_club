/**
 * Registration page — Telegram OIDC first, then email/password/nickname.
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
import {
  IconBrandTelegram,
  IconInfoCircle,
  IconLock,
  IconMail,
  IconUser,
} from '@tabler/icons-react'
import { useAuth } from '../auth'
import { requestTelegramIdToken } from '../auth/telegramWebLogin'
import { ApiClientError } from '../api'
import { useApiClient } from '../hooks'
import { getEnvironment } from '../env'

interface RegisterFormValues {
  nickname: string
  email: string
  password: string
  passwordConfirmation: string
}

export function RegisterPage() {
  const navigate = useNavigate()
  const location = useLocation()
  const {
    completeTelegramRegistration,
    loginWithTelegramWeb,
    isLoading: authLoading,
  } = useAuth()
  const apiClient = useApiClient()
  const env = getEnvironment()

  const locationState = location.state as
    | { registrationToken?: string; from?: string }
    | null
  const [registrationToken, setRegistrationToken] = useState<string | null>(
    locationState?.registrationToken ?? null,
  )
  const [error, setError] = useState<string | null>(null)
  const [telegramLoading, setTelegramLoading] = useState(false)
  const from = locationState?.from || '/clubs'

  const form = useForm<RegisterFormValues>({
    initialValues: {
      nickname: '',
      email: '',
      password: '',
      passwordConfirmation: '',
    },
    validate: {
      nickname: (value) =>
        value.trim().length < 2
          ? 'Nickname must be at least 2 characters'
          : value.trim().length > 32
            ? 'Nickname must be at most 32 characters'
            : null,
      email: (value) =>
        /^\S+@\S+\.\S+$/.test(value) ? null : 'Enter a valid email',
      password: (value) =>
        value.length < 12 ? 'Password must be at least 12 characters' : null,
      passwordConfirmation: (value, values) =>
        value !== values.password ? 'Passwords do not match' : null,
    },
  })

  const startTelegram = async () => {
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

      if (result.authenticated) {
        navigate(from, { replace: true })
        return
      }

      if (result.registrationRequired && result.registrationToken) {
        setRegistrationToken(result.registrationToken)
        return
      }

      setError('Unexpected Telegram authentication response')
    } catch (err) {
      if (err instanceof ApiClientError) {
        setError(err.message || 'Telegram authentication failed')
      } else if (err instanceof Error) {
        setError(err.message)
      } else {
        setError('Telegram authentication failed')
      }
    } finally {
      setTelegramLoading(false)
    }
  }

  const handleSubmit = form.onSubmit(async (values) => {
    setError(null)
    if (!registrationToken) {
      setError('Authenticate with Telegram before completing registration')
      return
    }

    try {
      await completeTelegramRegistration({
        registrationToken,
        email: values.email,
        password: values.password,
        passwordConfirmation: values.passwordConfirmation,
        nickname: values.nickname.trim(),
      })
      navigate(from, { replace: true })
    } catch (err) {
      if (err instanceof ApiClientError) {
        switch (err.code) {
          case 'EMAIL_ALREADY_EXISTS':
            setError('This email is already registered')
            break
          case 'INVALID_PASSWORD':
            setError('Invalid password or confirmation mismatch')
            break
          case 'INVALID_NICKNAME':
            setError('Invalid nickname')
            break
          case 'INVALID_REGISTRATION_TOKEN':
          case 'REGISTRATION_TOKEN_EXPIRED':
            setError('Telegram session expired. Please authenticate again.')
            setRegistrationToken(null)
            break
          default:
            setError(err.message || 'Registration failed')
        }
      } else if (err instanceof Error) {
        setError(err.message)
      } else {
        setError('Registration failed')
      }
    }
  })

  return (
    <Center style={{ minHeight: '100vh' }}>
      <Paper
        shadow="md"
        radius="md"
        p="xl"
        withBorder
        style={{ maxWidth: '420px', width: '100%' }}
      >
        <Stack gap="lg" align="center">
          <Title order={1} ta="center">
            Create account
          </Title>
          <Text c="dimmed" ta="center">
            Continue with Telegram, then set your Poker Club nickname, email and
            password.
          </Text>

          {error && (
            <Alert
              color="red"
              variant="filled"
              title="Registration"
              icon={<IconInfoCircle size={16} />}
              w="100%"
            >
              {error}
            </Alert>
          )}

          {!registrationToken ? (
            <>
              <Button
                fullWidth
                size="lg"
                leftSection={<IconBrandTelegram size={18} />}
                loading={telegramLoading}
                onClick={startTelegram}
              >
                Continue with Telegram
              </Button>
              <Group justify="center" gap="xs">
                <Text c="dimmed" size="sm">
                  Already registered?
                </Text>
                <Text component={Link} to="/login" size="sm" c="blue" fw={500}>
                  Sign in
                </Text>
              </Group>
            </>
          ) : (
            <>
              <Alert color="teal" variant="light" w="100%">
                Telegram verified. Complete your Poker Club profile below.
              </Alert>
              <Divider w="100%" />
              <form onSubmit={handleSubmit} style={{ width: '100%' }}>
                <Stack gap="md">
                  <TextInput
                    {...form.getInputProps('nickname')}
                    label="Nickname"
                    description="Your display name in Poker Club"
                    placeholder="Choose a nickname"
                    required
                    leftSection={<IconUser size={16} />}
                  />
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
                    placeholder="At least 12 characters"
                    autoComplete="new-password"
                    required
                    leftSection={<IconLock size={16} />}
                  />
                  <TextInput
                    {...form.getInputProps('passwordConfirmation')}
                    type="password"
                    label="Confirm password"
                    autoComplete="new-password"
                    required
                    leftSection={<IconLock size={16} />}
                  />
                  <Button
                    type="submit"
                    fullWidth
                    size="lg"
                    loading={authLoading || form.submitting}
                  >
                    Create account
                  </Button>
                </Stack>
              </form>
            </>
          )}
        </Stack>
      </Paper>
    </Center>
  )
}
