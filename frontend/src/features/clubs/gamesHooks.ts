/**
 * API hooks for Game management.
 *
 * Provides TanStack Query hooks for fetching game data
 * via the Backend API (06_API.md section 4.5).
 *
 * All server state is managed through TanStack Query.
 * Components must not call fetch() directly.
 */

import { useQuery } from '@tanstack/react-query'
import { useApiClient } from '../../hooks'
import type { GameSummary, GameStatus, GameType } from '../../types'

// --- Backend response types (snake_case from API) ---

interface BackendGame {
  id: number
  club_id: number
  banker_id: number
  game_type: string
  currency: string
  money_model: string
  chip_value: number
  buy_in_amount: number
  rebuy_allowed: boolean
  rebuy_price?: number
  max_rebuys?: number
  duration_seconds?: number
  start_time: string
  end_time?: string
  status: string
  min_players: number
  max_players: number
  ranking_primary: string
  ranking_secondary?: string
  created_at: string
  updated_at: string
  timer_paused_at?: string
  timer_paused_duration?: number
  timer_notified: boolean
}

// --- Query keys ---

export const gameKeys = {
  all: ['games'] as const,
  lists: () => [...gameKeys.all, 'list'] as const,
  list: (clubId: number) => [...gameKeys.lists(), clubId] as const,
  details: () => [...gameKeys.all, 'detail'] as const,
  detail: (id: number) => [...gameKeys.details(), id] as const,
}

// --- Mappers ---

function mapGame(game: BackendGame): GameSummary {
  return {
    id: game.id,
    type: game.game_type as GameType,
    status: game.status as GameStatus,
    bankerName: `Banker #${game.banker_id}`,
    startTime: game.start_time,
    endTime: game.end_time,
    currency: game.currency,
    chipValue: game.chip_value,
    buyInAmount: game.buy_in_amount,
    rebuyAllowed: game.rebuy_allowed,
    rebuyPrice: game.rebuy_price,
    maxRebuys: game.max_rebuys,
    minPlayers: game.min_players,
    maxPlayers: game.max_players,
    currentPlayers: 0,
  }
}

// --- Queries ---

/**
 * Fetch all games for a club.
 */
export function useClubGames(clubId: number) {
  const apiClient = useApiClient()
  return useQuery({
    queryKey: gameKeys.list(clubId),
    queryFn: async () => {
      const response = await apiClient.get<{ games: BackendGame[] }>(
        `/clubs/${clubId}/games`,
      )
      return response.data.games.map(mapGame)
    },
    enabled: !!clubId,
  })
}
