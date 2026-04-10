import { api } from '@/shared/lib'
import type { LoginPayload, Tokens, User } from '@/entities/auth'

export const login = async (payload: LoginPayload): Promise<{ user: User; token: string; refresh_token?: string }> => {
  const response = await api.post('/auth/login', payload)
  return {
    user: response.data.user,
    token: response.data.access_token,
    refresh_token: response.data.refresh_token,
  }
}

export const logout = async (): Promise<void> => {
  await api.post('/auth/logout')
}

export const refreshToken = async (): Promise<Tokens> => {
  const response = await api.post('/auth/refresh')
  return response.data
}
