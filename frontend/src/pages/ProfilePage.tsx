/**
 * Profile page — view/edit personal data, nickname, contacts, password.
 */

import { useEffect, useState } from 'react'
import {
  Alert,
  Button,
  Divider,
  Stack,
  Text,
  TextInput,
  Title,
} from '@mantine/core'
import { useForm } from '@mantine/form'
import { IconInfoCircle } from '@tabler/icons-react'
import { PageContainer, PageHeader, LoadingState } from '../components/ui'
import { useAuth } from '../auth'
import { ApiClientError } from '../api'

interface ProfileFormValues {
  firstName: string
  lastName: string
  nickname: string
  email: string
  phoneNumber: string
}

interface PasswordFormValues {
  currentPassword: string
  newPassword: string
  passwordConfirmation: string
}

export function ProfilePage() {
  const { user, state, updateProfile, changePassword, isLoading, refreshUser } =
    useAuth()
  const [profileError, setProfileError] = useState<string | null>(null)
  const [profileSuccess, setProfileSuccess] = useState<string | null>(null)
  const [passwordError, setPasswordError] = useState<string | null>(null)
  const [passwordSuccess, setPasswordSuccess] = useState<string | null>(null)

  const profileForm = useForm<ProfileFormValues>({
    initialValues: {
      firstName: '',
      lastName: '',
      nickname: '',
      email: '',
      phoneNumber: '',
    },
    validate: {
      firstName: (v) => (v.trim() ? null : 'First name is required'),
      lastName: (v) => (v.trim() ? null : 'Last name is required'),
      nickname: (v) =>
        v.trim().length < 2
          ? 'Nickname must be at least 2 characters'
          : v.trim().length > 32
            ? 'Nickname must be at most 32 characters'
            : null,
      email: (v) =>
        !v.trim()
          ? null
          : /^\S+@\S+\.\S+$/.test(v)
            ? null
            : 'Enter a valid email',
    },
  })

  const passwordForm = useForm<PasswordFormValues>({
    initialValues: {
      currentPassword: '',
      newPassword: '',
      passwordConfirmation: '',
    },
    validate: {
      newPassword: (v) =>
        v.length < 12 ? 'Password must be at least 12 characters' : null,
      passwordConfirmation: (v, values) =>
        v !== values.newPassword ? 'Passwords do not match' : null,
      currentPassword: (v) =>
        user?.hasPassword && !v
          ? 'Current password is required'
          : null,
    },
  })

  useEffect(() => {
    if (user) {
      profileForm.setValues({
        firstName: user.firstName ?? '',
        lastName: user.lastName ?? '',
        nickname: user.nickname ?? '',
        email: user.email ?? '',
        phoneNumber: user.phoneNumber ?? '',
      })
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps -- sync when user identity changes
  }, [user?.id, user?.updatedAt])

  if (state === 'loading' || !user) {
    return (
      <PageContainer>
        <LoadingState />
      </PageContainer>
    )
  }

  const handleProfileSubmit = profileForm.onSubmit(async (values) => {
    setProfileError(null)
    setProfileSuccess(null)
    try {
      await updateProfile({
        firstName: values.firstName.trim(),
        lastName: values.lastName.trim(),
        nickname: values.nickname.trim(),
        email: values.email.trim() || null,
        phoneNumber: values.phoneNumber.trim() || null,
      })
      setProfileSuccess('Profile updated')
      await refreshUser()
    } catch (err) {
      if (err instanceof ApiClientError) {
        setProfileError(err.message || 'Failed to update profile')
      } else if (err instanceof Error) {
        setProfileError(err.message)
      } else {
        setProfileError('Failed to update profile')
      }
    }
  })

  const handlePasswordSubmit = passwordForm.onSubmit(async (values) => {
    setPasswordError(null)
    setPasswordSuccess(null)
    try {
      await changePassword({
        currentPassword: values.currentPassword || undefined,
        newPassword: values.newPassword,
        passwordConfirmation: values.passwordConfirmation,
      })
      passwordForm.reset()
      setPasswordSuccess(
        user.hasPassword ? 'Password changed' : 'Password set',
      )
      await refreshUser()
    } catch (err) {
      if (err instanceof ApiClientError) {
        setPasswordError(err.message || 'Failed to update password')
      } else if (err instanceof Error) {
        setPasswordError(err.message)
      } else {
        setPasswordError('Failed to update password')
      }
    }
  })

  return (
    <PageContainer>
      <PageHeader title="Profile" description="Your Poker Club account" />

      <Stack gap="xl" maw={480}>
        <Stack gap="xs">
          <Text size="sm" c="dimmed">
            Telegram username:{' '}
            {user.tgUserName ? `@${user.tgUserName}` : 'not set'}
          </Text>
          {user.tgUserId != null && (
            <Text size="sm" c="dimmed">
              Telegram ID: {user.tgUserId}
            </Text>
          )}
        </Stack>

        <form onSubmit={handleProfileSubmit}>
          <Stack gap="md">
            <Title order={3}>Personal data</Title>
            {profileError && (
              <Alert color="red" icon={<IconInfoCircle size={16} />}>
                {profileError}
              </Alert>
            )}
            {profileSuccess && (
              <Alert color="teal" icon={<IconInfoCircle size={16} />}>
                {profileSuccess}
              </Alert>
            )}
            <TextInput
              label="First name"
              {...profileForm.getInputProps('firstName')}
              required
            />
            <TextInput
              label="Last name"
              {...profileForm.getInputProps('lastName')}
              required
            />
            <TextInput
              label="Nickname"
              description="Your Poker Club nickname (not Telegram username)"
              {...profileForm.getInputProps('nickname')}
              required
            />
            <TextInput
              label="Email"
              {...profileForm.getInputProps('email')}
              autoComplete="email"
            />
            <TextInput
              label="Phone"
              {...profileForm.getInputProps('phoneNumber')}
              autoComplete="tel"
            />
            <Button type="submit" loading={isLoading || profileForm.submitting}>
              Save profile
            </Button>
          </Stack>
        </form>

        <Divider />

        <form onSubmit={handlePasswordSubmit}>
          <Stack gap="md">
            <Title order={3}>
              {user.hasPassword ? 'Change password' : 'Set password'}
            </Title>
            <Text size="sm" c="dimmed">
              {user.hasPassword
                ? 'Enter your current password to set a new one.'
                : 'You signed in with Telegram. Set a password to enable email login.'}
            </Text>
            {passwordError && (
              <Alert color="red" icon={<IconInfoCircle size={16} />}>
                {passwordError}
              </Alert>
            )}
            {passwordSuccess && (
              <Alert color="teal" icon={<IconInfoCircle size={16} />}>
                {passwordSuccess}
              </Alert>
            )}
            {user.hasPassword && (
              <TextInput
                type="password"
                label="Current password"
                {...passwordForm.getInputProps('currentPassword')}
                autoComplete="current-password"
              />
            )}
            <TextInput
              type="password"
              label="New password"
              {...passwordForm.getInputProps('newPassword')}
              autoComplete="new-password"
              required
            />
            <TextInput
              type="password"
              label="Confirm new password"
              {...passwordForm.getInputProps('passwordConfirmation')}
              autoComplete="new-password"
              required
            />
            <Button type="submit" loading={isLoading || passwordForm.submitting}>
              {user.hasPassword ? 'Change password' : 'Set password'}
            </Button>
          </Stack>
        </form>
      </Stack>
    </PageContainer>
  )
}
