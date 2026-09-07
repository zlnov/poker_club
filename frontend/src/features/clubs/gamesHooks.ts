/**
 * API hooks for Game management.
 *
 * Provides TanStack Query hooks for fetching and mutating game data
 * via the Backend API (06_API.md sections 4.5, 4.6, 4.7).
 *
 * All server state is managed through TanStack Query.
 * Components must not call fetch() directly.
 */

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useApiClient } from '../../hooks'
import type {
  GameSummary,
  GameDetails,
  GameConfig,
  GameStatus,
  GameType,
} from '../../types'

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

// --- Game type mapping helpers ---

/**
 * Maps a backend game_type + duration to a frontend GameType.
 * - "cash" with duration → "cash_time"
 * - "cash" without duration → "cash_open"
 * - "tournament" → "tournament"
 */
function mapGameType(
  backendType: string,
  durationSeconds?: number,
): GameType {
  if (backendType === 'tournament') {
    return 'tournament'
  }
  // Cash games: distinguish time-limited vs open-ended
  if (durationSeconds !== undefined && durationSeconds > 0) {
    return 'cash_time'
  }
  return 'cash_open'
}

/**
 * Maps a frontend GameType to the backend game_type string.
 * - "cash_time" and "cash_open" → "cash"
 * - "tournament" → "tournament"
 */
function toBackendGameType(gameType: GameType): string {
  if (gameType === 'tournament') {
    return 'tournament'
  }
  return 'cash'
}

// --- Mappers ---

function mapGame(game: BackendGame): GameSummary {
  return {
    id: game.id,
    clubId: game.club_id,
    type: mapGameType(game.game_type, game.duration_seconds),
    status: game.status as GameStatus,
    bankerId: game.banker_id,
    bankerName: `Banker #${game.banker_id}`,
    startTime: game.start_time,
    endTime: game.end_time,
    currency: game.currency,
    chipValue: game.chip_value,
    buyInAmount: game.buy_in_amount,
    rebuyAllowed: game.rebuy_allowed,
    rebuyPrice: game.rebuy_price,
    maxRebuys: game.max_rebuys,
    durationSeconds: game.duration_seconds,
    minPlayers: game.min_players,
    maxPlayers: game.max_players,
    currentPlayers: 0,
  }
}

function mapGameDetails(game: BackendGame): GameDetails {
  return {
    id: game.id,
    clubId: game.club_id,
    bankerId: game.banker_id,
    gameType: game.game_type,
    currency: game.currency,
    moneyModel: game.money_model,
    chipValue: game.chip_value,
    buyInAmount: game.buy_in_amount,
    rebuyAllowed: game.rebuy_allowed,
    rebuyPrice: game.rebuy_price,
    maxRebuys: game.max_rebuys,
    durationSeconds: game.duration_seconds,
    startTime: game.start_time,
    endTime: game.end_time,
    status: game.status as GameStatus,
    minPlayers: game.min_players,
    maxPlayers: game.max_players,
    rankingPrimary: game.ranking_primary,
    rankingSecondary: game.ranking_secondary,
    createdAt: game.created_at,
    updatedAt: game.updated_at,
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

/**
 * Fetch a single game by ID.
 */
export function useGame(gameId: number) {
  const apiClient = useApiClient()
  return useQuery({
    queryKey: gameKeys.detail(gameId),
    queryFn: async () => {
      const response = await apiClient.get<BackendGame>(`/games/${gameId}`)
      return mapGameDetails(response.data)
    },
    enabled: !!gameId,
  })
}

// --- Mutations ---

/**
 * Create a new game in a club.
 * Only Owner/Admin can create games (PermCreateGame).
 */
export function useCreateGame() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      clubId,
      config,
    }: {
      clubId: number
      config: GameConfig
    }) => {
      const backendType = toBackendGameType(config.gameType)
      const body: Record<string, unknown> = {
        game_type: backendType,
        currency: config.currency,
        money_model: config.moneyModel,
        chip_value: config.chipValue,
        buy_in_amount: config.buyInAmount,
        rebuy_allowed: config.rebuyAllowed,
        min_players: config.minPlayers,
        max_players: config.maxPlayers,
        ranking_primary: config.rankingPrimary,
        banker_id: config.bankerId,
      }

      if (config.rebuyPrice !== undefined) {
        body.rebuy_price = config.rebuyPrice
      }
      if (config.maxRebuys !== undefined) {
        body.max_rebuys = config.maxRebuys
      }
      // Only send duration for cash_time games
      if (config.gameType === 'cash_time' && config.durationSeconds) {
        body.duration_seconds = config.durationSeconds
      }
      if (config.startTime) {
        body.start_time = config.startTime
      }
      if (config.rankingSecondary) {
        body.ranking_secondary = config.rankingSecondary
      }

      const response = await apiClient.post<BackendGame>(
        `/clubs/${clubId}/games`,
        { body },
      )
      return mapGameDetails(response.data)
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: gameKeys.list(data.clubId) })
    },
  })
}

/**
 * Update game parameters.
 * Only allowed in planned status. Only Owner/Admin can edit (PermEditGame).
 */
export function useUpdateGame() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      gameId,
      config,
    }: {
      gameId: number
      config: Partial<GameConfig>
    }) => {
      const body: Record<string, unknown> = {}

      if (config.gameType !== undefined) {
        body.game_type = toBackendGameType(config.gameType)
      }
      if (config.currency !== undefined) {
        body.currency = config.currency
      }
      if (config.moneyModel !== undefined) {
        body.money_model = config.moneyModel
      }
      if (config.chipValue !== undefined) {
        body.chip_value = config.chipValue
      }
      if (config.buyInAmount !== undefined) {
        body.buy_in_amount = config.buyInAmount
      }
      if (config.rebuyAllowed !== undefined) {
        body.rebuy_allowed = config.rebuyAllowed
      }
      if (config.rebuyPrice !== undefined) {
        body.rebuy_price = config.rebuyPrice
      }
      if (config.maxRebuys !== undefined) {
        body.max_rebuys = config.maxRebuys
      }
      if (config.durationSeconds !== undefined) {
        body.duration_seconds = config.durationSeconds
      }
      if (config.startTime !== undefined) {
        body.start_time = config.startTime
      }
      if (config.minPlayers !== undefined) {
        body.min_players = config.minPlayers
      }
      if (config.maxPlayers !== undefined) {
        body.max_players = config.maxPlayers
      }
      if (config.rankingPrimary !== undefined) {
        body.ranking_primary = config.rankingPrimary
      }
      if (config.rankingSecondary !== undefined) {
        body.ranking_secondary = config.rankingSecondary
      }

      const response = await apiClient.patch<BackendGame>(
        `/games/${gameId}`,
        { body },
      )
      return mapGameDetails(response.data)
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: gameKeys.detail(data.id) })
      queryClient.invalidateQueries({ queryKey: gameKeys.list(data.clubId) })
    },
  })
}

/**
 * Start a game (transitions from planned to active).
 * Only Banker, Owner, or Admin can start (checkGameAccess).
 */
export function useStartGame() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (gameId: number) => {
      const response = await apiClient.post<BackendGame>(
        `/games/${gameId}/start`,
      )
      return mapGameDetails(response.data)
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: gameKeys.detail(data.id) })
      queryClient.invalidateQueries({ queryKey: gameKeys.list(data.clubId) })
    },
  })
}

/**
 * Cancel a game (transitions to cancelled status).
 * Only Owner/Admin can cancel (PermCancelGame).
 */
export function useCancelGame() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (gameId: number) => {
      await apiClient.post(`/games/${gameId}/cancel`)
      return gameId
    },
    onSuccess: (gameId) => {
      // Invalidate the specific game query; list invalidation
      // requires knowing the clubId, which we don't have here.
      // The caller can invalidate the list if needed.
      queryClient.invalidateQueries({ queryKey: gameKeys.detail(gameId) })
    },
  })
}

/**
 * Assign or change the banker for a game.
 * Only Owner/Admin can change banker (PermEditGame).
 * The banker_id is the player_id; Backend resolves it to club_members.id.
 */
export function useUpdateBanker() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      gameId,
      bankerId,
    }: {
      gameId: number
      bankerId: number
    }) => {
      const response = await apiClient.patch<BackendGame>(
        `/games/${gameId}/banker`,
        { body: { banker_id: bankerId } },
      )
      return mapGameDetails(response.data)
    },
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: gameKeys.detail(data.id) })
      queryClient.invalidateQueries({ queryKey: gameKeys.list(data.clubId) })
    },
  })
}
