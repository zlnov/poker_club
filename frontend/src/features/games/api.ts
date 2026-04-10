import { api } from '@/shared/lib'
import type {
  CreateGameRequest,
  BuyInRequest,
  SetChipsRequest,
  FinishGameRequest,
  BuyInResponse,
  FinishGameResponse,
  GetGameDetailsResponse,
  ListGamesRequest,
  ListGamesResponse,
} from '@/entities/game'

export const getGames = async (params: ListGamesRequest): Promise<ListGamesResponse> => {
  const response = await api.get('/games', { params })
  return response.data
}

export const getGame = async (id: number): Promise<GetGameDetailsResponse> => {
  const response = await api.get(`/games/${id}`)
  return response.data
}

export const createGame = async (payload: CreateGameRequest): Promise<{ id: number }> => {
  const response = await api.post('/games', payload)
  return response.data
}

export const buyIn = async (payload: BuyInRequest): Promise<BuyInResponse> => {
  const response = await api.post(`/games/${payload.game_id}/buyin`, {
    player_id: payload.player_id,
  })
  return response.data
}

export const setChips = async (payload: SetChipsRequest): Promise<void> => {
  await api.post(`/games/${payload.game_id}/participants/${payload.player_id}/chips`, {
    chips: payload.chips,
  })
}

export const finishGame = async (payload: FinishGameRequest): Promise<FinishGameResponse> => {
  const response = await api.post(`/games/${payload.game_id}/finish`, {})
  return response.data
}
