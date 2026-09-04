/**
 * API hooks for Club management.
 *
 * Provides TanStack Query hooks for fetching and mutating club data
 * via the Backend API (06_API.md sections 4.2, 4.3, 4.4).
 *
 * All server state is managed through TanStack Query.
 * Components must not call fetch() directly.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useApiClient } from '../../hooks'
import type { ClubSummary, ClubStatistics } from '../../types'

// --- Backend response types (snake_case from API) ---

interface BackendClub {
  id: number
  name: string
  created_at: string
  updated_at: string
  tg_chat_id?: number
}

interface BackendClubWithRole extends BackendClub {
  role: string
  is_owner: boolean
  is_admin: boolean
}

interface BackendClubStatistics {
  total_members: number
  total_games: number
  cash_games: number
  tournament_games: number
  total_buy_in_amount: number
  total_rebuy_amount: number
  total_bank: number
  average_game_duration: number
}

// --- Mappers ---

function mapClub(club: BackendClub): ClubSummary {
  return {
    id: club.id,
    name: club.name,
    tgChatId: club.tg_chat_id,
    memberCount: 0, // Will be populated from statistics if needed
    isOwner: false,
    isAdmin: false,
  }
}

function mapClubWithRole(club: BackendClubWithRole): ClubSummary {
  return {
    id: club.id,
    name: club.name,
    tgChatId: club.tg_chat_id,
    memberCount: 0,
    isOwner: club.is_owner,
    isAdmin: club.is_admin,
  }
}

function mapClubStatistics(stats: BackendClubStatistics): ClubStatistics {
  return {
    totalMembers: stats.total_members,
    totalGames: stats.total_games,
    cashGames: stats.cash_games,
    tournamentGames: stats.tournament_games,
    totalBuyInAmount: stats.total_buy_in_amount,
    totalRebuyAmount: stats.total_rebuy_amount,
    totalBank: stats.total_bank,
    averageGameDuration: formatDuration(stats.average_game_duration),
  }
}

function formatDuration(seconds: number): string {
  if (isNaN(seconds) || seconds < 0) return '—'
  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = Math.floor(seconds % 60)
  if (h > 0) {
    return `${h}:${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`
  }
  return `${m}:${s.toString().padStart(2, '0')}`
}

// --- Query keys ---

export const clubKeys = {
  all: ['clubs'] as const,
  lists: () => [...clubKeys.all, 'list'] as const,
  list: (filters?: string) => [...clubKeys.lists(), filters] as const,
  details: () => [...clubKeys.all, 'detail'] as const,
  detail: (id: number) => [...clubKeys.details(), id] as const,
  members: (clubId: number) => [...clubKeys.all, 'members', clubId] as const,
  memberRequests: (clubId: number) =>
    [...clubKeys.all, 'member-requests', clubId] as const,
  invites: (clubId: number) => [...clubKeys.all, 'invites', clubId] as const,
  statistics: (clubId: number) =>
    [...clubKeys.all, 'statistics', clubId] as const,
}

// --- Queries ---

/**
 * Fetch all clubs for the authenticated user.
 */
export function useClubs() {
  const apiClient = useApiClient()
  return useQuery({
    queryKey: clubKeys.lists(),
    queryFn: async () => {
      const response = await apiClient.get<{ clubs: BackendClub[] }>('/clubs')
      return response.data.clubs.map(mapClub)
    },
  })
}

/**
 * Fetch a single club by ID.
 * Returns club info along with the current user's role in the club.
 */
export function useClub(clubId: number) {
  const apiClient = useApiClient()
  return useQuery({
    queryKey: clubKeys.detail(clubId),
    queryFn: async () => {
      const response = await apiClient.get<BackendClubWithRole>(
        `/clubs/${clubId}`,
      )
      return mapClubWithRole(response.data)
    },
    enabled: !!clubId,
  })
}

/**
 * Fetch club statistics.
 */
export function useClubStatistics(clubId: number) {
  const apiClient = useApiClient()
  return useQuery({
    queryKey: clubKeys.statistics(clubId),
    queryFn: async () => {
      const response = await apiClient.get<BackendClubStatistics>(
        `/clubs/${clubId}/statistics`,
      )
      return mapClubStatistics(response.data)
    },
    enabled: !!clubId,
  })
}

// --- Mutations ---

/**
 * Create a new club.
 */
export function useCreateClub() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (name: string) => {
      const response = await apiClient.post<BackendClub>('/clubs', {
        body: { name },
      })
      return mapClub(response.data)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: clubKeys.lists() })
    },
  })
}

/**
 * Update club name.
 */
export function useUpdateClub() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({ clubId, name }: { clubId: number; name: string }) => {
      const response = await apiClient.patch<BackendClubWithRole>(
        `/clubs/${clubId}`,
        { body: { name } },
      )
      return mapClubWithRole(response.data)
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: clubKeys.detail(data.id) })
      queryClient.invalidateQueries({ queryKey: clubKeys.lists() })
    },
  })
}

/**
 * Close (delete) a club.
 */
export function useCloseClub() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (clubId: number) => {
      await apiClient.delete(`/clubs/${clubId}`)
      return clubId
    },
    onSuccess: (clubId) => {
      queryClient.removeQueries({ queryKey: clubKeys.detail(clubId) })
      queryClient.invalidateQueries({ queryKey: clubKeys.lists() })
    },
  })
}
