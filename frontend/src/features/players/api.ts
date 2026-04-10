import { api } from '@/shared/lib'
import type { PlayerStats, GetLeaderboardRequest, LeaderboardResponse } from '@/entities/player'

// В backend нет GET /players - используем leaderboard для получения статистики игроков
export const getPlayerStats = async (playerId: number): Promise<PlayerStats> => {
  const response = await api.get(`/players/${playerId}/stats`)
  return response.data
}

// Leaderboard теперь получается через GET /clubs/:club_id/leaderboard
export const getLeaderboard = async (params: GetLeaderboardRequest): Promise<LeaderboardResponse> => {
  const response = await api.get(`/clubs/${params.club_id}/leaderboard`, { params })
  return response.data
}
