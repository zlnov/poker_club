import { jwtDecode } from 'jwt-decode'
import { User } from '@/entities/auth/types'

const TOKEN_KEY = 'access_token'

let accessToken: string | null = null

export const setAccessToken = (token: string | null) => {
  accessToken = token
  if (token) {
    localStorage.setItem(TOKEN_KEY, token)
  } else {
    localStorage.removeItem(TOKEN_KEY)
  }
}

export const getAccessToken = (): string | null => {
  if (accessToken === null) {
    accessToken = localStorage.getItem(TOKEN_KEY)
  }
  return accessToken
}

export const decodeToken = (token: string): any => {
  try {
    return jwtDecode(token)
  } catch {
    return null
  }
}

export const getUserFromToken = (): User | null => {
  const token = getAccessToken()
  if (!token) return null

  const decoded = decodeToken(token)
  if (!decoded) return null

  return {
    id: decoded.user_id || decoded.sub,
    first_name: decoded.first_name || '',
    last_name: decoded.last_name || '',
    nickname: decoded.nickname || '',
    phone_number: decoded.phone_number || '',
    role: decoded.role === 'admin' ? 'admin' : 'member',
  }
}

export const clearAuth = () => {
  setAccessToken(null)
}
