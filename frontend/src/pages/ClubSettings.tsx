/**
 * Club settings page component.
 *
 * Displays the settings for a club.
 * Allows Owner to edit club name and close the club.
 *
 * See 05_FE_UX.md section 4 (Application Navigation) and
 * 04_FE_SPEC.md section 11 (Routing).
 */

import { useState } from 'react'
import {
  Button,
  Card,
  Group,
  Stack,
  Text,
  Title,
  TextInput,
  Modal,
  Alert,
} from '@mantine/core'
import { IconEdit, IconTrash, IconInfoCircle } from '@tabler/icons-react'
import { useParams, useNavigate } from 'react-router-dom'
import { useForm } from '@mantine/form'
import {
  PageContainer,
  PageHeader,
  EmptyState,
  LoadingState,
  ErrorState,
} from '../components/ui'
import { useClub, useUpdateClub, useCloseClub } from '../features'
import { ApiClientError } from '../api'
import { notifications } from '@mantine/notifications'

interface ClubSettingsFormValues {
  name: string
}

/**
 * Club settings page.
 *
 * Displays:
 * - Club information (name, creation date)
 * - Edit club name form (Owner only)
 * - Close club action (Owner only)
 * - Loading, empty, and error states
 */
export function ClubSettings() {
  const { clubId } = useParams<{ clubId: string }>()
  const navigate = useNavigate()
  const clubIdNum = clubId ? parseInt(clubId, 10) : 0

  const [closeConfirmOpen, setCloseConfirmOpen] = useState(false)

  const { data: club, isLoading, isError, error, refetch } = useClub(clubIdNum)

  const updateClub = useUpdateClub()
  const closeClub = useCloseClub()

  const isOwner = club?.isOwner ?? false

  const form = useForm<ClubSettingsFormValues>({
    initialValues: {
      name: club?.name ?? '',
    },
    validate: {
      name: (value) =>
        !value || value.trim().length < 1
          ? 'Club name is required'
          : value.trim().length > 100
            ? 'Club name must be 100 characters or less'
            : null,
    },
  })

  // Update form values when club data changes
  if (club && form.values.name !== club.name) {
    form.setFieldValue('name', club.name)
  }

  const handleUpdateClub = form.onSubmit(async (values) => {
    try {
      await updateClub.mutateAsync({
        clubId: clubIdNum,
        name: values.name.trim(),
      })
      notifications.show({
        title: 'Club updated',
        message: 'Club settings have been updated successfully.',
        color: 'green',
      })
      refetch()
    } catch (err) {
      if (err instanceof ApiClientError) {
        notifications.show({
          title: 'Failed to update club',
          message: err.message,
          color: 'red',
        })
      }
    }
  })

  const handleCloseClub = async () => {
    try {
      await closeClub.mutateAsync(clubIdNum)
      notifications.show({
        title: 'Club closed',
        message: 'The club has been closed successfully.',
        color: 'green',
      })
      navigate('/clubs')
    } catch (err) {
      if (err instanceof ApiClientError) {
        notifications.show({
          title: 'Failed to close club',
          message: err.message,
          color: 'red',
        })
      }
    } finally {
      setCloseConfirmOpen(false)
    }
  }

  if (isLoading) {
    return (
      <PageContainer>
        <PageHeader title="Settings" description="Club settings" />
        <LoadingState message="Loading club settings..." />
      </PageContainer>
    )
  }

  if (isError) {
    return (
      <PageContainer>
        <PageHeader title="Settings" description="Club settings" />
        <ErrorState
          message={
            error instanceof ApiClientError
              ? error.message
              : 'Failed to load club settings'
          }
          onRetry={() => refetch()}
        />
      </PageContainer>
    )
  }

  if (!club) {
    return (
      <PageContainer>
        <PageHeader title="Settings" description="Club settings" />
        <EmptyState
          title="Club not found"
          description="The club you're looking for doesn't exist or you don't have access."
        />
      </PageContainer>
    )
  }

  return (
    <PageContainer>
      <PageHeader title="Settings" description="Manage club settings" />

      {/* Club Information */}
      <Card mt="lg" padding="lg" radius="md" withBorder>
        <Title order={4} mb="md">
          Club Information
        </Title>
        <Stack gap="sm">
          <Group gap="sm">
            <Text fw={500} style={{ minWidth: 120 }}>
              Club Name:
            </Text>
            <Text>{club.name}</Text>
          </Group>
          <Group gap="sm">
            <Text fw={500} style={{ minWidth: 120 }}>
              Club ID:
            </Text>
            <Text>{club.id}</Text>
          </Group>
          {club.tgChatId && (
            <Group gap="sm">
              <Text fw={500} style={{ minWidth: 120 }}>
                Telegram Chat ID:
              </Text>
              <Text>{club.tgChatId}</Text>
            </Group>
          )}
          <Group gap="sm">
            <Text fw={500} style={{ minWidth: 120 }}>
              Your Role:
            </Text>
            <Text>{isOwner ? 'Owner' : club.isAdmin ? 'Admin' : 'Member'}</Text>
          </Group>
        </Stack>
      </Card>

      {/* Edit Club Name (Owner/Admin only) */}
      {isOwner && (
        <Card mt="lg" padding="lg" radius="md" withBorder>
          <Title order={4} mb="md">
            Edit Club Name
          </Title>
          <form onSubmit={handleUpdateClub}>
            <Stack gap="md">
              <TextInput
                {...form.getInputProps('name')}
                label="Club Name"
                placeholder="Enter club name"
                required
                error={form.errors.name}
                leftSection={<IconEdit size={16} />}
              />
              {updateClub.isError &&
                updateClub.error instanceof ApiClientError && (
                  <Alert
                    color="red"
                    variant="light"
                    title="Update failed"
                    icon={<IconInfoCircle size={16} />}
                  >
                    {updateClub.error.message}
                  </Alert>
                )}
              <Group justify="flex-end">
                <Button
                  type="submit"
                  loading={updateClub.isPending}
                  leftSection={<IconEdit size={16} />}
                >
                  Save Changes
                </Button>
              </Group>
            </Stack>
          </form>
        </Card>
      )}

      {/* Close Club (Owner only) */}
      {isOwner && (
        <Card mt="lg" padding="lg" radius="md" withBorder>
          <Title order={4} mb="md">
            Danger Zone
          </Title>
          <Stack gap="md">
            <Text size="sm" c="dimmed">
              Closing the club will permanently delete all club data including
              games, members, and statistics. This action cannot be undone.
            </Text>
            <Button
              color="red"
              leftSection={<IconTrash size={16} />}
              onClick={() => setCloseConfirmOpen(true)}
              loading={closeClub.isPending}
            >
              Close Club
            </Button>
          </Stack>
        </Card>
      )}

      {!isOwner && !club.isAdmin && (
        <Card mt="lg" padding="lg" radius="md" withBorder>
          <EmptyState
            title="No settings available"
            description="You don't have permission to modify club settings. Contact the club owner or admin."
          />
        </Card>
      )}

      {/* Close Club Confirmation Modal */}
      <Modal
        opened={closeConfirmOpen}
        onClose={() => setCloseConfirmOpen(false)}
        title="Close Club"
        centered
      >
        <Stack gap="md">
          <Alert
            color="red"
            variant="light"
            title="Warning"
            icon={<IconInfoCircle size={16} />}
          >
            This action is irreversible. All club data will be permanently
            deleted.
          </Alert>
          <Group justify="flex-end" gap="sm">
            <Button
              variant="subtle"
              onClick={() => setCloseConfirmOpen(false)}
              disabled={closeClub.isPending}
            >
              Cancel
            </Button>
            <Button
              color="red"
              onClick={handleCloseClub}
              loading={closeClub.isPending}
            >
              Close Club
            </Button>
          </Group>
        </Stack>
      </Modal>
    </PageContainer>
  )
}
