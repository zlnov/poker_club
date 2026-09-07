/**
 * API hooks for Club Member management.
 *
 * Provides TanStack Query hooks for fetching and mutating club members
 * via the Backend API (06_API.md section 4.3).
 *
 * All server state is managed through TanStack Query.
 * Components must not call fetch() directly.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useApiClient } from '../../hooks'
import { clubKeys } from '../clubs/hooks'
import type { ClubMemberRole, InvitationInfo } from '../../types'

// --- Backend response types (snake_case from API) ---

interface BackendPlayer {
  id: number
  first_name: string
  last_name: string
  nickname: string
  created_at: string
  updated_at: string
  phone_number?: string
  email?: string
  tg_user_id?: number
}

interface BackendClubMember {
  id: number
  club_id: number
  player_id: number
  role: string
  status: string
  accepted: boolean
  created_at: string
  updated_at: string
  player: BackendPlayer
}

// --- Mappers ---

function mapClubMember(member: BackendClubMember): ClubMemberRole {
  return {
    clubMemberId: member.id,
    playerId: member.player_id,
    firstName: member.player.first_name,
    lastName: member.player.last_name,
    nickname: member.player.nickname,
    role: member.role as 'owner' | 'admin' | 'member',
    status: member.status as 'pending' | 'active' | 'banned' | 'left',
    canManageParticipants: member.role === 'owner' || member.role === 'admin',
    canAdjustResults: member.role === 'owner' || member.role === 'admin',
    canInvite: member.role === 'owner' || member.role === 'admin',
    canRemove: member.role === 'owner' || member.role === 'admin',
  }
}

function mapInvitation(member: BackendClubMember): InvitationInfo {
  return {
    playerId: member.player_id,
    firstName: member.player.first_name,
    lastName: member.player.last_name,
    nickname: member.player.nickname,
    tgUserId: member.player.tg_user_id,
    status: member.status as 'pending',
    accepted: member.accepted,
    createdAt: member.created_at,
    updatedAt: member.updated_at,
  }
}

// --- Queries ---

/**
 * Fetch all members of a club.
 */
export function useClubMembers(clubId: number) {
  const apiClient = useApiClient()
  return useQuery({
    queryKey: clubKeys.members(clubId),
    queryFn: async () => {
      const response = await apiClient.get<{ members: BackendClubMember[] }>(
        `/clubs/${clubId}/members`,
      )
      return response.data.members.map(mapClubMember)
    },
    enabled: !!clubId,
  })
}

/**
 * Fetch pending membership requests for a club.
 */
export function useMembershipRequests(clubId: number) {
  const apiClient = useApiClient()
  return useQuery({
    queryKey: clubKeys.memberRequests(clubId),
    queryFn: async () => {
      const response = await apiClient.get<{ requests: BackendClubMember[] }>(
        `/clubs/${clubId}/member-requests`,
      )
      return response.data.requests?.map(mapClubMember) ?? []
    },
    enabled: !!clubId,
  })
}

/**
 * Fetch active invitations for a club.
 */
export function useClubInvites(clubId: number) {
  const apiClient = useApiClient()
  return useQuery({
    queryKey: clubKeys.invites(clubId),
    queryFn: async () => {
      const response = await apiClient.get<{ invites: BackendClubMember[] }>(
        `/clubs/${clubId}/invites`,
      )
      return response.data.invites?.map(mapInvitation) ?? []
    },
    enabled: !!clubId,
  })
}

// --- Mutations ---

/**
 * Invite a member to a club by Telegram user ID.
 */
export function useInviteMember() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      clubId,
      tgUserId,
    }: {
      clubId: number
      tgUserId: number
    }) => {
      await apiClient.post(`/clubs/${clubId}/members`, {
        body: { tg_user_id: tgUserId },
      })
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: clubKeys.members(variables.clubId),
      })
      queryClient.invalidateQueries({
        queryKey: clubKeys.invites(variables.clubId),
      })
    },
  })
}

/**
 * Update a member's role or status.
 */
export function useUpdateMember() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      clubId,
      playerId,
      role,
      status,
    }: {
      clubId: number
      playerId: number
      role?: string
      status?: string
    }) => {
      const body: { role?: string; status?: string } = {}
      if (role) body.role = role
      if (status) body.status = status
      await apiClient.patch(`/clubs/${clubId}/members/${playerId}`, {
        body,
      })
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: clubKeys.members(variables.clubId),
      })
    },
  })
}

/**
 * Remove a member from a club.
 */
export function useRemoveMember() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      clubId,
      playerId,
    }: {
      clubId: number
      playerId: number
    }) => {
      await apiClient.delete(`/clubs/${clubId}/members/${playerId}`)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: clubKeys.members(variables.clubId),
      })
    },
  })
}

/**
 * Approve a membership request (pending → active).
 */
export function useApproveMember() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      clubId,
      playerId,
    }: {
      clubId: number
      playerId: number
    }) => {
      await apiClient.post(`/clubs/${clubId}/members/${playerId}/approve`)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: clubKeys.members(variables.clubId),
      })
      queryClient.invalidateQueries({
        queryKey: clubKeys.memberRequests(variables.clubId),
      })
    },
  })
}

/**
 * Reject a membership request (pending → left).
 */
export function useRejectMember() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      clubId,
      playerId,
    }: {
      clubId: number
      playerId: number
    }) => {
      await apiClient.post(`/clubs/${clubId}/members/${playerId}/reject`)
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: clubKeys.members(variables.clubId),
      })
      queryClient.invalidateQueries({
        queryKey: clubKeys.memberRequests(variables.clubId),
      })
    },
  })
}

/**
 * Create an invitation for a player to join the club.
 */
export function useCreateInvite() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      clubId,
      tgUserId,
    }: {
      clubId: number
      tgUserId: number
    }) => {
      await apiClient.post(`/clubs/${clubId}/invites`, {
        body: { tg_user_id: tgUserId },
      })
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({
        queryKey: clubKeys.invites(variables.clubId),
      })
      queryClient.invalidateQueries({
        queryKey: clubKeys.members(variables.clubId),
      })
    },
  })
}
