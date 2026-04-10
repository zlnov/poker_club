import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import * as gamesApi from './api'
import type { ListGamesRequest } from '@/entities/game'

export const useGames = (params: ListGamesRequest) => {
  return useQuery({
    queryKey: ['games', params.status, params.limit, params.offset],
    queryFn: () => gamesApi.getGames(params),
  })
}

export const useGame = (gameId: number) => {
  return useQuery({
    queryKey: ['games', gameId],
    queryFn: () => gamesApi.getGame(gameId),
    enabled: !!gameId,
  })
}

export const useCreateGame = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: gamesApi.createGame,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['games'] })
    },
  })
}

export const useBuyIn = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: gamesApi.buyIn,
    onSuccess: (_, { game_id }) => {
      queryClient.invalidateQueries({ queryKey: ['games', game_id] })
    },
  })
}

export const useSetChips = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: gamesApi.setChips,
    onSuccess: (_, { game_id }) => {
      queryClient.invalidateQueries({ queryKey: ['games', game_id] })
    },
  })
}

export const useFinishGame = () => {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: gamesApi.finishGame,
    onSuccess: (_, { game_id }) => {
      queryClient.invalidateQueries({ queryKey: ['games', game_id] })
      queryClient.invalidateQueries({ queryKey: ['games'] })
    },
  })
}
