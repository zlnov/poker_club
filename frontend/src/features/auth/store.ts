import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User, AuthState } from '@/entities/auth'
import { setAccessToken, clearAuth } from '@/shared/lib/auth'

interface AuthStore extends AuthState {
  login: (user: User, token: string, refreshToken?: string) => void
  logout: () => void
}

const authStore = create<AuthStore>()(
  persist(
    (set) => ({
      user: null,
      accessToken: null,
      isAuthenticated: false,
      login: (user, token, refreshToken) => {
        setAccessToken(token)
        if (refreshToken) {
          localStorage.setItem('refresh_token', refreshToken)
        }
        set({
          user,
          accessToken: token,
          isAuthenticated: true,
        })
      },
      logout: () => {
        clearAuth() // сбрасывает переменную accessToken в памяти и удаляет из localStorage
        localStorage.removeItem('refresh_token')
        set({
          user: null,
          accessToken: null,
          isAuthenticated: false,
        })
      },
    }),
    {
      name: 'auth-storage',
    }
  )
)

export const useAuthStore = authStore
export const authStoreInstance = authStore
