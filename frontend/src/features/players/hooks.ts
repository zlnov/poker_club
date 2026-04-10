import { useQuery } from '@tanstack/react-query'
import * as playersApi from './api'
import type { GetLeaderboardRequest } from '@/entities/player'

export const usePlayerStats = (playerId: number) => {
  return useQuery({
    queryKey: ['players', playerId, 'stats'],
    queryFn: () => playersApi.getPlayerStats(playerId),
    enabled: !!playerId,
  })
}

export const useLeaderboard = (params: GetLeaderboardRequest) => {
  return useQuery({
    queryKey: ['leaderboard', params.club_id, params.metric, params.period],
    queryFn: () => playersApi.getLeaderboard(params),
  })
}
