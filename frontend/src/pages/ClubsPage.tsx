/**
 * Clubs page component.
 *
 * Displays the list of clubs the user belongs to.
 * Allows creating a new club (Owner/Admin can create).
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
  Modal,
  TextInput,
  Badge,
} from '@mantine/core'
import { IconPlus, IconChessKing, IconUsers } from '@tabler/icons-react'
import { useNavigate } from 'react-router-dom'
import { useForm } from '@mantine/form'
import {
  PageContainer,
  PageHeader,
  EmptyState,
  LoadingState,
  ErrorState,
} from '../components/ui'
import { useClubs, useCreateClub } from '../features'
import { ApiClientError } from '../api'
import { notifications } from '@mantine/notifications'

interface ClubFormValues {
  name: string
}

/**
 * Clubs page.
 *
 * Displays:
 * - List of clubs the user is a member of
 * - Create club modal (for new club creation)
 * - Loading, empty, and error states
 */
export function ClubsPage() {
  const navigate = useNavigate()
  const [createModalOpen, setCreateModalOpen] = useState(false)

  const { data: clubs, isLoading, isError, error, refetch } = useClubs()

  const {
    mutate: createClub,
    isPending: isCreating,
    isError: isCreateError,
    error: createError,
  } = useCreateClub()

  const form = useForm<ClubFormValues>({
    initialValues: {
      name: '',
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

  const handleCreateClub = form.onSubmit(async (values) => {
    try {
      await createClub(values.name.trim())
      notifications.show({
        title: 'Club created',
        message: 'Your club has been created successfully.',
        color: 'green',
      })
      setCreateModalOpen(false)
      form.reset()
      refetch()
    } catch (err) {
      if (err instanceof ApiClientError) {
        notifications.show({
          title: 'Failed to create club',
          message: err.message,
          color: 'red',
        })
      } else {
        notifications.show({
          title: 'Failed to create club',
          message: 'An unexpected error occurred.',
          color: 'red',
        })
      }
    }
  })

  if (isLoading) {
    return (
      <PageContainer>
        <PageHeader title="Clubs" description="Your poker clubs" />
        <LoadingState message="Loading clubs..." />
      </PageContainer>
    )
  }

  if (isError) {
    return (
      <PageContainer>
        <PageHeader title="Clubs" description="Your poker clubs" />
        <ErrorState
          message={
            error instanceof ApiClientError
              ? error.message
              : 'Failed to load clubs'
          }
          onRetry={() => refetch()}
        />
      </PageContainer>
    )
  }

  return (
    <PageContainer>
      <PageHeader
        title="Clubs"
        description="Your poker clubs"
        action={
          <Button
            leftSection={<IconPlus size={16} />}
            onClick={() => setCreateModalOpen(true)}
          >
            Create Club
          </Button>
        }
      />

      {clubs && clubs.length === 0 ? (
        <EmptyState
          title="No clubs yet"
          description="You haven't joined or created any clubs. Create a club to get started."
          actionLabel="Create Club"
          onAction={() => setCreateModalOpen(true)}
          icon={<IconChessKing size={48} stroke={1.5} />}
        />
      ) : (
        <Stack gap="md" mt="md">
          {clubs?.map((club) => (
            <Card
              key={club.id}
              padding="lg"
              radius="md"
              withBorder
              style={{ cursor: 'pointer' }}
              onClick={() => navigate(`/clubs/${club.id}`)}
            >
              <Group justify="space-between">
                <Stack gap={4}>
                  <Title order={4} mb={0}>
                    {club.name}
                  </Title>
                  <Group gap="xs">
                    <IconUsers size={16} />
                    <Text size="sm" c="dimmed">
                      {club.memberCount} members
                    </Text>
                  </Group>
                </Stack>
                <Group gap="xs">
                  {club.isOwner && (
                    <Badge color="violet" variant="light">
                      Owner
                    </Badge>
                  )}
                  {club.isAdmin && !club.isOwner && (
                    <Badge color="blue" variant="light">
                      Admin
                    </Badge>
                  )}
                </Group>
              </Group>
            </Card>
          ))}
        </Stack>
      )}

      <Modal
        opened={createModalOpen}
        onClose={() => setCreateModalOpen(false)}
        title="Create Club"
        centered
      >
        <form onSubmit={handleCreateClub}>
          <Stack gap="md">
            <TextInput
              {...form.getInputProps('name')}
              label="Club Name"
              placeholder="Enter club name"
              required
              error={form.errors.name}
            />
            {isCreateError && createError instanceof ApiClientError && (
              <Text c="red" size="sm">
                {createError.message}
              </Text>
            )}
            <Group justify="flex-end" gap="sm">
              <Button
                variant="subtle"
                onClick={() => setCreateModalOpen(false)}
                disabled={isCreating}
              >
                Cancel
              </Button>
              <Button type="submit" loading={isCreating}>
                Create
              </Button>
            </Group>
          </Stack>
        </form>
      </Modal>
    </PageContainer>
  )
}
