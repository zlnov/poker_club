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
  GameParticipantSummary,
  PlayerSummary,
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
    timerPausedAt: game.timer_paused_at,
    timerPausedDuration: game.timer_paused_duration
      ? Number(game.timer_paused_duration)
      : undefined,
    timerNotified: game.timer_notified,
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
 * For active games, polls every 5 seconds (06_API.md section 8).
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
    // Poll active games every 5 seconds (04_FE_SPEC.md section 26)
    refetchInterval: (query) => {
      const game = query.state.data as GameDetails | undefined
      return game?.status === 'active' ? 5000 : false
    },
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

// --- Game Participants ---

// Backend response types for game participants

interface BackendPlayer {
  id: number
  first_name: string
  last_name: string
  nickname: string
  tg_user_id?: number
  phone_number?: string
  email?: string
  created_at: string
  updated_at: string
}

interface BackendGameParticipant {
  id: number
  game_id: number
  player_id: number
  buy_in_count: number
  rebuy_count: number
  chips_end?: number
  current_stack?: number
  payout_amount?: number
  place?: number
  status: string
  created_at: string
  updated_at: string
  player: BackendPlayer
}

function mapPlayer(player: BackendPlayer): PlayerSummary {
  return {
    id: player.id,
    tgUserID: player.tg_user_id ?? 0,
    firstName: player.first_name,
    lastName: player.last_name,
    nickname: player.nickname,
    photoUrl: undefined,
  }
}

function mapGameParticipant(p: BackendGameParticipant): GameParticipantSummary {
  return {
    player: mapPlayer(p.player),
    buyInCount: p.buy_in_count,
    rebuyCount: p.rebuy_count,
    chipsEnd: p.chips_end,
    currentStack: p.current_stack,
    payoutAmount: p.payout_amount,
    place: p.place,
    status: p.status as GameParticipantSummary['status'],
  }
}

// Query keys for participants

export const participantKeys = {
  all: ['participants'] as const,
  list: (gameId: number) => [...participantKeys.all, 'list', gameId] as const,
}

/**
 * Fetch all participants for a game.
 */
export function useGameParticipants(gameId: number) {
  const apiClient = useApiClient()
  return useQuery({
    queryKey: participantKeys.list(gameId),
    queryFn: async () => {
      const response = await apiClient.get<{ participants: BackendGameParticipant[] }>(
        `/games/${gameId}/participants`,
      )
      return response.data.participants.map(mapGameParticipant)
    },
    enabled: !!gameId,
    // Poll active games every 5 seconds (04_FE_SPEC.md section 26)
    refetchInterval: 5000,
  })
}

/**
 * Accept a game invitation.
 * The player must have an invited status for the game.
 */
export function useAcceptGameParticipation() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (gameId: number) => {
      const response = await apiClient.post<{ player: BackendPlayer }>(
        `/games/${gameId}/participants/me/accept`,
      )
      return mapPlayer(response.data.player)
    },
    onSuccess: (_, gameId) => {
      queryClient.invalidateQueries({ queryKey: participantKeys.list(gameId) })
    },
  })
}

/**
 * Decline a game invitation.
 * The player must have an invited or accepted status for the game.
 */
export function useDeclineGameParticipation() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (gameId: number) => {
      await apiClient.post(`/games/${gameId}/participants/me/decline`)
      return gameId
    },
    onSuccess: (gameId) => {
      queryClient.invalidateQueries({ queryKey: participantKeys.list(gameId) })
    },
  })
}

/**
 * Confirm a player's accepted game invitation.
 * Only owner/admin can confirm (PermManageGameParticipants).
 */
export function useConfirmGameParticipation() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      gameId,
      playerId,
    }: {
      gameId: number
      playerId: number
    }) => {
      const response = await apiClient.post<{ player: BackendPlayer }>(
        `/games/${gameId}/participants/${playerId}/confirm`,
      )
      return { gameId, player: mapPlayer(response.data.player) }
    },
    onSuccess: ({ gameId }) => {
      queryClient.invalidateQueries({ queryKey: participantKeys.list(gameId) })
    },
  })
}

// --- Phase 6: Active Game ---

// Backend response types for active game

interface BackendGameEvent {
  id: number
  game_id: number
  player_id: number
  type: string
  old_value?: number
  new_value?: number
  metadata?: Record<string, unknown>
  created_at: string
  created_by: number
}

interface BackendGameResult {
  player_id: number
  buy_in_count: number
  rebuy_count: number
  chips_end?: number
  payout_amount?: number
  place?: number
  status: string
}

// Query keys for active game

export const monitorKeys = {
  all: ['monitor'] as const,
  detail: (gameId: number) => [...monitorKeys.all, gameId] as const,
}

export const eventKeys = {
  all: ['events'] as const,
  list: (gameId: number) => [...eventKeys.all, gameId] as const,
}

export const resultKeys = {
  all: ['results'] as const,
  detail: (gameId: number) => [...resultKeys.all, gameId] as const,
}

/**
 * Fetch game monitor data (game + participants) for banker/owner/admin.
 * Uses GET /games/{gameId}/monitor.
 * Only available for active games.
 */
export function useGameMonitor(gameId: number) {
  const apiClient = useApiClient()
  return useQuery({
    queryKey: monitorKeys.detail(gameId),
    queryFn: async () => {
      const response = await apiClient.get<{
        game: BackendGame
        participants: BackendGameParticipant[]
      }>(`/games/${gameId}/monitor`)
      return {
        game: mapGameDetails(response.data.game),
        participants: response.data.participants.map(mapGameParticipant),
      }
    },
    enabled: !!gameId,
    // Poll active games every 5 seconds (04_FE_SPEC.md section 26)
    refetchInterval: 5000,
  })
}

/**
 * Fetch game events (event log) for a game.
 * Uses GET /games/{gameId}/events.
 * Any club member can view events.
 */
export function useGameEvents(gameId: number) {
  const apiClient = useApiClient()
  return useQuery({
    queryKey: eventKeys.list(gameId),
    queryFn: async () => {
      const response = await apiClient.get<{ events: BackendGameEvent[] }>(
        `/games/${gameId}/events`,
      )
      return response.data.events.map((e) => ({
        id: e.id,
        gameId: e.game_id,
        playerId: e.player_id,
        type: e.type,
        oldValue: e.old_value,
        newValue: e.new_value,
        metadata: e.metadata,
        createdAt: e.created_at,
        createdBy: e.created_by,
      }))
    },
    enabled: !!gameId,
  })
}

/**
 * Fetch game results for a finished game.
 * Uses GET /games/{gameId}/results.
 */
export function useGameResults(gameId: number) {
  const apiClient = useApiClient()
  return useQuery({
    queryKey: resultKeys.detail(gameId),
    queryFn: async () => {
      const response = await apiClient.get<{
        game: BackendGame
        results: BackendGameResult[]
      }>(`/games/${gameId}/results`)
      return {
        game: mapGameDetails(response.data.game),
        results: response.data.results.map((r) => ({
          playerId: r.player_id,
          buyInCount: r.buy_in_count,
          rebuyCount: r.rebuy_count,
          chipsEnd: r.chips_end,
          payoutAmount: r.payout_amount,
          place: r.place,
          status: r.status,
        })),
      }
    },
    enabled: !!gameId,
  })
}

/**
 * Finish a game (transitions from active to finished).
 * Only banker/owner/admin can finish (checkGameAccess).
 * Backend performs all calculations (payout, profit, ROI, place, winner).
 */
export function useFinishGame() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async (gameId: number) => {
      await apiClient.post(`/games/${gameId}/finish`)
      return gameId
    },
    onSuccess: (gameId) => {
      queryClient.invalidateQueries({ queryKey: gameKeys.detail(gameId) })
      queryClient.invalidateQueries({ queryKey: participantKeys.list(gameId) })
      queryClient.invalidateQueries({ queryKey: monitorKeys.detail(gameId) })
      queryClient.invalidateQueries({ queryKey: resultKeys.detail(gameId) })
    },
  })
}

/**
 * Register a buy-in for a participant.
 * Only banker/owner/admin can register buy-in (checkGameAccess).
 */
export function useRegisterBuyIn() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      gameId,
      playerId,
    }: {
      gameId: number
      playerId: number
    }) => {
      await apiClient.post(`/games/${gameId}/participants/${playerId}/buy-in`)
      return { gameId, playerId }
    },
    onSuccess: ({ gameId }) => {
      queryClient.invalidateQueries({ queryKey: participantKeys.list(gameId) })
      queryClient.invalidateQueries({ queryKey: monitorKeys.detail(gameId) })
    },
  })
}

/**
 * Register a rebuy for a participant.
 * Only banker/owner/admin can register rebuy (checkGameAccess).
 */
export function useRegisterRebuy() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      gameId,
      playerId,
    }: {
      gameId: number
      playerId: number
    }) => {
      await apiClient.post(`/games/${gameId}/participants/${playerId}/rebuy`)
      return { gameId, playerId }
    },
    onSuccess: ({ gameId }) => {
      queryClient.invalidateQueries({ queryKey: participantKeys.list(gameId) })
      queryClient.invalidateQueries({ queryKey: monitorKeys.detail(gameId) })
    },
  })
}

/**
 * Fix (correct) rebuy count for a participant.
 * Only banker/owner/admin can fix rebuy (checkGameAccess).
 * Allows setting any rebuy count value (increase or decrease).
 */
export function useFixRebuy() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      gameId,
      playerId,
      rebuyCount,
    }: {
      gameId: number
      playerId: number
      rebuyCount: number
    }) => {
      await apiClient.post(`/games/${gameId}/participants/${playerId}/rebuy/fix`, {
        body: { rebuy_count: rebuyCount },
      })
      return { gameId, playerId }
    },
    onSuccess: ({ gameId }) => {
      queryClient.invalidateQueries({ queryKey: participantKeys.list(gameId) })
      queryClient.invalidateQueries({ queryKey: monitorKeys.detail(gameId) })
    },
  })
}

/**
 * Set chips end for a participant.
 * Only banker/owner/admin can set chips end (checkGameAccess).
 * Chips end can be 0.
 */
export function useSetChipsEnd() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      gameId,
      playerId,
      chipsEnd,
    }: {
      gameId: number
      playerId: number
      chipsEnd: number
    }) => {
      await apiClient.post(`/games/${gameId}/participants/${playerId}/chips`, {
        body: { chips_end: chipsEnd },
      })
      return { gameId, playerId }
    },
    onSuccess: ({ gameId }) => {
      queryClient.invalidateQueries({ queryKey: participantKeys.list(gameId) })
      queryClient.invalidateQueries({ queryKey: monitorKeys.detail(gameId) })
      queryClient.invalidateQueries({ queryKey: resultKeys.detail(gameId) })
    },
  })
}

/**
 * Set current stack for the requesting player.
 * A player can update their own current stack during an active game.
 * The current stack is stored as a chips_set event.
 */
export function useSetCurrentStack() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      gameId,
      stack,
    }: {
      gameId: number
      stack: number
    }) => {
      await apiClient.post(`/games/${gameId}/participants/me/stack`, {
        body: { stack },
      })
      return { gameId }
    },
    onSuccess: ({ gameId }) => {
      queryClient.invalidateQueries({ queryKey: participantKeys.list(gameId) })
      queryClient.invalidateQueries({ queryKey: monitorKeys.detail(gameId) })
    },
  })
}

/**
 * Correct game results (adjust chips_end for a finished game).
 * Only owner/admin can adjust results (PermAdjustGameResults).
 */
export function useAdjustGameResults() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      gameId,
      playerId,
      chipsEnd,
    }: {
      gameId: number
      playerId: number
      chipsEnd: number
    }) => {
      await apiClient.post(`/games/${gameId}/results/correct`, {
        body: { player_id: playerId, chips_end: chipsEnd },
      })
      return { gameId, playerId }
    },
    onSuccess: ({ gameId }) => {
      queryClient.invalidateQueries({ queryKey: resultKeys.detail(gameId) })
      queryClient.invalidateQueries({ queryKey: participantKeys.list(gameId) })
    },
  })
}

/**
 * Remove a player from a game.
 * Only owner/admin can remove participants (PermManageGameParticipants).
 */
export function useRemoveGameParticipant() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      gameId,
      playerId,
    }: {
      gameId: number
      playerId: number
    }) => {
      await apiClient.delete(`/games/${gameId}/participants/${playerId}`)
      return { gameId, playerId }
    },
    onSuccess: ({ gameId }) => {
      queryClient.invalidateQueries({ queryKey: participantKeys.list(gameId) })
      queryClient.invalidateQueries({ queryKey: monitorKeys.detail(gameId) })
    },
  })
}

/**
 * Add a player to a game.
 * Only owner/admin can add participants (PermManageGameParticipants).
 */
export function useAddGameParticipant() {
  const apiClient = useApiClient()
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: async ({
      gameId,
      playerId,
    }: {
      gameId: number
      playerId: number
    }) => {
      await apiClient.post(`/games/${gameId}/participants`, {
        body: { player_id: playerId },
      })
      return { gameId, playerId }
    },
    onSuccess: ({ gameId }) => {
      queryClient.invalidateQueries({ queryKey: participantKeys.list(gameId) })
      queryClient.invalidateQueries({ queryKey: monitorKeys.detail(gameId) })
    },
  })
}
