/**
 * Authentication context provider.
 *
 * Manages authentication state, JWT tokens, and current user.
 * Tokens stay in runtime memory only (no localStorage/sessionStorage).
 */

import { useState, useEffect, useCallback, useMemo, ReactNode } from 'react'
import { createApiClient, type TokenManager } from '../api'
import { getEnvironment } from '../env'
import { ApiClientContext } from './ApiClientContext'
import { AuthContext } from './AuthContext'
import type {
  AuthContextValue,
  AuthState,
  ChangePasswordInput,
  CompleteTelegramRegistrationInput,
  CurrentUser,
  LoginCredentials,
  TelegramWebAuthResult,
  UpdateProfileInput,
} from './index'

interface AuthProviderProps {
  children: ReactNode
}

class RuntimeTokenManager implements TokenManager {
  private accessToken: string | null = null
  private refreshToken: string | null = null
  private refreshPromise: Promise<{
    accessToken: string
    refreshToken: string
  } | null> | null = null

  getAccessToken(): string | null {
    return this.accessToken
  }

  getRefreshToken(): string | null {
    return this.refreshToken
  }

  setTokens(accessToken: string, refreshToken: string): void {
    this.accessToken = accessToken
    this.refreshToken = refreshToken
  }

  clearTokens(): void {
    this.accessToken = null
    this.refreshToken = null
    this.refreshPromise = null
  }

  async refreshAccessToken(): Promise<{
    accessToken: string
    refreshToken: string
  } | null> {
    if (this.refreshPromise) {
      return this.refreshPromise
    }

    const refreshToken = this.refreshToken
    if (!refreshToken) {
      return null
    }

    this.refreshPromise = this.doRefresh(refreshToken)

    try {
      return await this.refreshPromise
    } finally {
      this.refreshPromise = null
    }
  }

  private async doRefresh(
    refreshToken: string,
  ): Promise<{ accessToken: string; refreshToken: string } | null> {
    try {
      const env = getEnvironment()
      const response = await fetch(`${env.apiBaseUrl}/auth/refresh`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ refresh_token: refreshToken }),
      })

      if (!response.ok) {
        return null
      }

      const data = await response.json()
      const newAccessToken = data.access_token
      const newRefreshToken = data.refresh_token

      if (!newAccessToken || !newRefreshToken) {
        return null
      }

      this.accessToken = newAccessToken
      this.refreshToken = newRefreshToken

      return { accessToken: newAccessToken, refreshToken: newRefreshToken }
    } catch {
      return null
    }
  }
}

type BackendMe = {
  id: number
  first_name: string
  last_name: string
  nickname: string | null
  tg_user_name?: string | null
  email?: string | null
  phone_number?: string | null
  tg_user_id?: number
  has_password?: boolean
  created_at: string
  updated_at: string
}

function mapCurrentUser(backendUser: BackendMe): CurrentUser {
  return {
    id: backendUser.id,
    firstName: backendUser.first_name,
    lastName: backendUser.last_name,
    nickname: backendUser.nickname ?? null,
    tgUserName: backendUser.tg_user_name ?? null,
    email: backendUser.email ?? null,
    phoneNumber: backendUser.phone_number ?? null,
    tgUserId: backendUser.tg_user_id,
    hasPassword: Boolean(backendUser.has_password),
    createdAt: backendUser.created_at,
    updatedAt: backendUser.updated_at,
  }
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [state, setState] = useState<AuthState>('loading')
  const [user, setUser] = useState<CurrentUser | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  const tokenManager = useMemo(() => new RuntimeTokenManager(), [])
  const env = getEnvironment()
  const apiClient = useMemo(
    () => createApiClient(env.apiBaseUrl, tokenManager),
    [env.apiBaseUrl, tokenManager],
  )

  const isAuthenticated = state === 'authenticated' && user !== null

  const applyTokensAndLoadUser = useCallback(
    async (accessToken: string, refreshToken: string) => {
      tokenManager.setTokens(accessToken, refreshToken)
      const response = await apiClient.get<BackendMe>('/me')
      const currentUser = mapCurrentUser(response.data)
      setUser(currentUser)
      setState('authenticated')
    },
    [apiClient, tokenManager],
  )

  const fetchCurrentUser = useCallback(async (): Promise<CurrentUser | null> => {
    try {
      const response = await apiClient.get<BackendMe>('/me')
      return mapCurrentUser(response.data)
    } catch (error) {
      if (
        error instanceof Error &&
        'statusCode' in error &&
        error.statusCode === 401
      ) {
        return null
      }
      throw error
    }
  }, [apiClient])

  useEffect(() => {
    let mounted = true

    async function initializeAuth() {
      setState('loading')
      try {
        // Without persisted tokens, /me will 401 — that is expected.
        if (!tokenManager.getAccessToken()) {
          if (mounted) {
            setUser(null)
            setState('unauthenticated')
          }
          return
        }
        const currentUser = await fetchCurrentUser()
        if (mounted) {
          if (currentUser) {
            setUser(currentUser)
            setState('authenticated')
          } else {
            setUser(null)
            setState('unauthenticated')
          }
        }
      } catch {
        if (mounted) {
          setUser(null)
          setState('unauthenticated')
        }
      }
    }

    initializeAuth()

    return () => {
      mounted = false
    }
  }, [fetchCurrentUser, tokenManager])

  const login = useCallback(
    async (credentials: LoginCredentials): Promise<void> => {
      setIsLoading(true)
      try {
        const response = await apiClient.post<{
          access_token: string
          refresh_token: string
        }>('/auth/login', {
          body: {
            email: credentials.email,
            password: credentials.password,
          },
          skipAuth: true,
        })

        await applyTokensAndLoadUser(
          response.data.access_token,
          response.data.refresh_token,
        )
      } finally {
        setIsLoading(false)
      }
    },
    [apiClient, applyTokensAndLoadUser],
  )

  const loginWithTelegram = useCallback(
    async (initData: string): Promise<void> => {
      setIsLoading(true)
      try {
        const response = await apiClient.post<{
          access_token: string
          refresh_token: string
        }>('/auth/telegram', {
          body: { init_data: initData },
          skipAuth: true,
        })

        await applyTokensAndLoadUser(
          response.data.access_token,
          response.data.refresh_token,
        )
      } finally {
        setIsLoading(false)
      }
    },
    [apiClient, applyTokensAndLoadUser],
  )

  const loginWithTelegramWeb = useCallback(
    async (
      idToken: string,
      challengeToken: string,
    ): Promise<TelegramWebAuthResult> => {
      setIsLoading(true)
      try {
        const response = await apiClient.post<{
          access_token?: string
          refresh_token?: string
          registration_required?: boolean
          registration_token?: string
        }>('/auth/telegram/web', {
          body: {
            id_token: idToken,
            challenge_token: challengeToken,
          },
          skipAuth: true,
        })

        if (response.data.registration_required) {
          return {
            authenticated: false,
            registrationRequired: true,
            registrationToken: response.data.registration_token,
          }
        }

        if (!response.data.access_token || !response.data.refresh_token) {
          throw new Error('Telegram Web login failed: missing tokens')
        }

        await applyTokensAndLoadUser(
          response.data.access_token,
          response.data.refresh_token,
        )
        return { authenticated: true }
      } finally {
        setIsLoading(false)
      }
    },
    [apiClient, applyTokensAndLoadUser],
  )

  const completeTelegramRegistration = useCallback(
    async (input: CompleteTelegramRegistrationInput): Promise<void> => {
      setIsLoading(true)
      try {
        const response = await apiClient.post<{
          access_token: string
          refresh_token: string
        }>('/auth/telegram/web/register', {
          body: {
            registration_token: input.registrationToken,
            email: input.email,
            password: input.password,
            password_confirmation: input.passwordConfirmation,
            nickname: input.nickname,
            first_name: input.firstName,
            last_name: input.lastName,
          },
          skipAuth: true,
        })

        await applyTokensAndLoadUser(
          response.data.access_token,
          response.data.refresh_token,
        )
      } finally {
        setIsLoading(false)
      }
    },
    [apiClient, applyTokensAndLoadUser],
  )

  const updateProfile = useCallback(
    async (input: UpdateProfileInput): Promise<void> => {
      setIsLoading(true)
      try {
        const response = await apiClient.patch<BackendMe>('/me', {
          body: {
            first_name: input.firstName,
            last_name: input.lastName,
            nickname: input.nickname,
            email: input.email,
            phone_number: input.phoneNumber,
          },
        })
        setUser(mapCurrentUser(response.data))
      } finally {
        setIsLoading(false)
      }
    },
    [apiClient],
  )

  const changePassword = useCallback(
    async (input: ChangePasswordInput): Promise<void> => {
      setIsLoading(true)
      try {
        await apiClient.post('/me/password', {
          body: {
            current_password: input.currentPassword,
            new_password: input.newPassword,
            password_confirmation: input.passwordConfirmation,
          },
        })
        const currentUser = await fetchCurrentUser()
        if (currentUser) {
          setUser(currentUser)
        }
      } finally {
        setIsLoading(false)
      }
    },
    [apiClient, fetchCurrentUser],
  )

  const logout = useCallback(async (): Promise<void> => {
    setIsLoading(true)
    try {
      const refreshToken = tokenManager.getRefreshToken()
      if (refreshToken) {
        await apiClient.post('/auth/logout', {
          body: { refresh_token: refreshToken },
          skipAuth: true,
        })
      }
    } catch {
      // Ignore logout errors
    } finally {
      tokenManager.clearTokens()
      setUser(null)
      setState('unauthenticated')
      setIsLoading(false)
    }
  }, [apiClient, tokenManager])

  const refreshUser = useCallback(async (): Promise<void> => {
    const currentUser = await fetchCurrentUser()
    if (currentUser) {
      setUser(currentUser)
    } else {
      setUser(null)
      setState('unauthenticated')
    }
  }, [fetchCurrentUser])

  const value: AuthContextValue = {
    state,
    user,
    isAuthenticated,
    isLoading,
    login,
    loginWithTelegram,
    loginWithTelegramWeb,
    completeTelegramRegistration,
    updateProfile,
    changePassword,
    logout,
    refreshUser,
  }

  return (
    <ApiClientContext.Provider value={apiClient}>
      <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
    </ApiClientContext.Provider>
  )
}
