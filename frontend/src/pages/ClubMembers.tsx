/**
 * Club members page component.
 *
 * Displays the list of members for a club with their roles and status.
 * Allows Owner/Admin to manage members: change roles, ban/unban, remove.
 * Shows membership requests for Owner/Admin approval.
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
  Badge,
  Select,
  Modal,
  ActionIcon,
} from '@mantine/core'
import {
  IconCrown,
  IconShield,
  IconUser,
  IconTrash,
  IconCheck,
  IconX,
} from '@tabler/icons-react'
import { useParams } from 'react-router-dom'
import {
  PageContainer,
  PageHeader,
  EmptyState,
  LoadingState,
  ErrorState,
} from '../components/ui'
import {
  useClub,
  useClubMembers,
  useMembershipRequests,
  useUpdateMember,
  useRemoveMember,
  useApproveMember,
  useRejectMember,
} from '../features'
import { ApiClientError } from '../api'
import { notifications } from '@mantine/notifications'
import type { ClubMemberRole } from '../types'

/** Role options for the select dropdown. */
const roleOptions = [
  { value: 'admin', label: 'Admin' },
  { value: 'member', label: 'Member' },
]

/** Status options for the select dropdown. */
const statusOptions = [
  { value: 'active', label: 'Active' },
  { value: 'banned', label: 'Banned' },
]

/**
 * Formats a player's display name.
 */
function formatPlayerName(member: ClubMemberRole): string {
  const parts = [member.firstName, member.lastName].filter(Boolean)
  return parts.length > 0 ? parts.join(' ') : member.nickname || 'Unknown'
}

/**
 * Club members page.
 *
 * Displays:
 * - List of active members with role management
 * - Pending membership requests (for Owner/Admin)
 * - Loading, empty, and error states
 */
export function ClubMembers() {
  const { clubId } = useParams<{ clubId: string }>()
  const clubIdNum = clubId ? parseInt(clubId, 10) : 0

  const [removeConfirmOpen, setRemoveConfirmOpen] = useState(false)
  const [memberToRemove, setMemberToRemove] = useState<{
    playerId: number
    name: string
  } | null>(null)

  const { data: club } = useClub(clubIdNum)
  const isOwner = club?.isOwner ?? false
  const isAdmin = club?.isAdmin ?? false
  const canManage = isOwner || isAdmin

  const {
    data: members,
    isLoading: membersLoading,
    isError: membersError,
    error: membersFetchError,
    refetch: refetchMembers,
  } = useClubMembers(clubIdNum)

  const {
    data: requests,
    isLoading: requestsLoading,
    isError: requestsError,
    refetch: refetchRequests,
  } = useMembershipRequests(clubIdNum)

  const updateMember = useUpdateMember()
  const removeMember = useRemoveMember()
  const approveMember = useApproveMember()
  const rejectMember = useRejectMember()

  const handleRoleChange = async (playerId: number, newRole: string) => {
    try {
      await updateMember.mutateAsync({
        clubId: clubIdNum,
        playerId,
        role: newRole,
      })
      notifications.show({
        title: 'Role updated',
        message: 'Member role has been updated.',
        color: 'green',
      })
    } catch (err) {
      if (err instanceof ApiClientError) {
        notifications.show({
          title: 'Failed to update role',
          message: err.message,
          color: 'red',
        })
      }
    }
  }

  const handleStatusChange = async (playerId: number, newStatus: string) => {
    try {
      await updateMember.mutateAsync({
        clubId: clubIdNum,
        playerId,
        status: newStatus,
      })
      notifications.show({
        title: 'Status updated',
        message: `Member status has been updated to ${newStatus}.`,
        color: 'green',
      })
    } catch (err) {
      if (err instanceof ApiClientError) {
        notifications.show({
          title: 'Failed to update status',
          message: err.message,
          color: 'red',
        })
      }
    }
  }

  const handleRemoveMember = (playerId: number, name: string) => {
    setMemberToRemove({ playerId, name })
    setRemoveConfirmOpen(true)
  }

  const confirmRemove = async () => {
    if (!memberToRemove) return
    try {
      await removeMember.mutateAsync({
        clubId: clubIdNum,
        playerId: memberToRemove.playerId,
      })
      notifications.show({
        title: 'Member removed',
        message: `${memberToRemove.name} has been removed from the club.`,
        color: 'green',
      })
    } catch (err) {
      if (err instanceof ApiClientError) {
        notifications.show({
          title: 'Failed to remove member',
          message: err.message,
          color: 'red',
        })
      }
    } finally {
      setRemoveConfirmOpen(false)
      setMemberToRemove(null)
    }
  }

  const handleApprove = async (playerId: number) => {
    try {
      await approveMember.mutateAsync({
        clubId: clubIdNum,
        playerId,
      })
      notifications.show({
        title: 'Member approved',
        message: 'Membership request has been approved.',
        color: 'green',
      })
    } catch (err) {
      if (err instanceof ApiClientError) {
        notifications.show({
          title: 'Failed to approve member',
          message: err.message,
          color: 'red',
        })
      }
    }
  }

  const handleReject = async (playerId: number) => {
    try {
      await rejectMember.mutateAsync({
        clubId: clubIdNum,
        playerId,
      })
      notifications.show({
        title: 'Membership request rejected',
        message: 'The membership request has been rejected.',
        color: 'green',
      })
    } catch (err) {
      if (err instanceof ApiClientError) {
        notifications.show({
          title: 'Failed to reject member',
          message: err.message,
          color: 'red',
        })
      }
    }
  }

  const isLoading = membersLoading || requestsLoading
  const isError = membersError || requestsError

  if (isLoading) {
    return (
      <PageContainer>
        <PageHeader title="Members" description="Club members" />
        <LoadingState message="Loading members..." />
      </PageContainer>
    )
  }

  if (isError) {
    return (
      <PageContainer>
        <PageHeader title="Members" description="Club members" />
        <ErrorState
          message={
            membersFetchError instanceof ApiClientError
              ? membersFetchError.message
              : 'Failed to load members'
          }
          onRetry={() => {
            refetchMembers()
            refetchRequests()
          }}
        />
      </PageContainer>
    )
  }

  return (
    <PageContainer>
      <PageHeader title="Members" description="Club members" />

      {/* Membership Requests Section (Owner/Admin only) */}
      {canManage && requests && requests.length > 0 && (
        <Card mt="lg" padding="lg" radius="md" withBorder>
          <Title order={4} mb="md">
            Membership Requests
          </Title>
          <Stack gap="sm">
            {requests.map((request) => (
              <Group key={request.playerId} justify="space-between">
                <Group gap="sm">
                  <IconUser size={20} />
                  <Stack gap={2}>
                    <Text fw={500}>{formatPlayerName(request)}</Text>
                    <Text size="sm" c="dimmed">
                      Role: {request.role}
                    </Text>
                  </Stack>
                </Group>
                <Group gap="xs">
                  <ActionIcon
                    color="green"
                    variant="light"
                    onClick={() => handleApprove(request.playerId)}
                    disabled={approveMember.isPending}
                  >
                    <IconCheck size={16} />
                  </ActionIcon>
                  <ActionIcon
                    color="red"
                    variant="light"
                    onClick={() => handleReject(request.playerId)}
                    disabled={rejectMember.isPending}
                  >
                    <IconX size={16} />
                  </ActionIcon>
                </Group>
              </Group>
            ))}
          </Stack>
        </Card>
      )}

      {/* Active Members Section */}
      {members && members.length === 0 ? (
        <EmptyState
          title="No members"
          description="This club has no members yet."
        />
      ) : (
        <Stack gap="sm" mt="md">
          {members?.map((member) => (
            <Card key={member.playerId} padding="md" radius="md" withBorder>
              <Group justify="space-between">
                <Group gap="sm">
                  {member.role === 'owner' ? (
                    <IconCrown size={20} />
                  ) : member.role === 'admin' ? (
                    <IconShield size={20} />
                  ) : (
                    <IconUser size={20} />
                  )}
                  <Stack gap={2}>
                    <Group gap="xs">
                      <Text fw={500}>{formatPlayerName(member)}</Text>
                      <Badge
                        color={
                          member.role === 'owner'
                            ? 'violet'
                            : member.role === 'admin'
                              ? 'blue'
                              : 'gray'
                        }
                        variant="light"
                        size="sm"
                      >
                        {member.role}
                      </Badge>
                      <Badge
                        color={
                          member.status === 'active'
                            ? 'green'
                            : member.status === 'banned'
                              ? 'red'
                              : 'gray'
                        }
                        variant="light"
                        size="sm"
                      >
                        {member.status}
                      </Badge>
                    </Group>
                    <Text size="sm" c="dimmed">
                      Player ID: {member.playerId}
                    </Text>
                  </Stack>
                </Group>

                {canManage && member.role !== 'owner' && (
                  <Group gap="xs">
                    <Select
                      placeholder="Change role"
                      data={roleOptions}
                      value={member.role}
                      onChange={(newRole: string | null) =>
                        newRole && handleRoleChange(member.playerId, newRole)
                      }
                      disabled={
                        updateMember.isPending || member.status === 'banned'
                      }
                      style={{ width: 120 }}
                    />
                    <Select
                      placeholder="Change status"
                      data={statusOptions}
                      value={
                        member.status === 'active' || member.status === 'banned'
                          ? member.status
                          : null
                      }
                      onChange={(newStatus) =>
                        newStatus &&
                        handleStatusChange(member.playerId, newStatus)
                      }
                      disabled={updateMember.isPending}
                      style={{ width: 120 }}
                    />
                    <ActionIcon
                      color="red"
                      variant="light"
                      onClick={() =>
                        handleRemoveMember(
                          member.playerId,
                          formatPlayerName(member),
                        )
                      }
                      disabled={removeMember.isPending}
                    >
                      <IconTrash size={16} />
                    </ActionIcon>
                  </Group>
                )}
              </Group>
            </Card>
          ))}
        </Stack>
      )}

      {/* Remove Member Confirmation Modal */}
      <Modal
        opened={removeConfirmOpen}
        onClose={() => setRemoveConfirmOpen(false)}
        title="Remove Member"
        centered
      >
        <Stack gap="md">
          <Text>
            Are you sure you want to remove{' '}
            <strong>{memberToRemove?.name}</strong> from the club?
          </Text>
          <Group justify="flex-end" gap="sm">
            <Button
              variant="subtle"
              onClick={() => setRemoveConfirmOpen(false)}
            >
              Cancel
            </Button>
            <Button
              color="red"
              onClick={confirmRemove}
              loading={removeMember.isPending}
            >
              Remove
            </Button>
          </Group>
        </Stack>
      </Modal>
    </PageContainer>
  )
}
