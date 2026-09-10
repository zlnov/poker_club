/**
 * Authentication context provider.
 *
 * Manages authentication state, JWT tokens, and current user.
 * Integrates with Backend authentication API.
 *
 * See 07_AUTH.md, 07.1_AUTH_SECURITY.md, 06_API.md section 4.0.
 */

import { useState, useEffect, useCallback, useMemo, ReactNode } from 'react'
import { createApiClient, type TokenManager } from '../api'
import { getEnvironment } from '../env'
import { ApiClientContext } from './ApiClientContext'
import { AuthContext } from './AuthContext'
import type {
  AuthContextValue,
  AuthState,
  CurrentUser,
  LoginCredentials,
} from './index'

interface AuthProviderProps {
  children: ReactNode
}

/**
 * Runtime token manager.
 *
 * Stores access and refresh tokens in runtime memory only.
 * Per 07.1_AUTH_SECURITY.md:
 * - Access token is NOT stored in localStorage/sessionStorage
 * - Refresh token is NOT stored in localStorage/sessionStorage
 * - Tokens exist only in runtime memory
 */
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

  /**
   * Refreshes the access token using the refresh token.
   * Uses a promise lock to prevent concurrent refresh attempts.
   * Returns new tokens or null on failure.
   */
  async refreshAccessToken(): Promise<{
    accessToken: string
    refreshToken: string
  } | null> {
    // If a refresh is already in progress, return that promise
    if (this.refreshPromise) {
      return this.refreshPromise
    }

    const refreshToken = this.refreshToken
    if (!refreshToken) {
      return null
    }

    // Create the refresh promise
    this.refreshPromise = this.doRefresh(refreshToken)

    try {
      const result = await this.refreshPromise
      return result
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

      // Update tokens with rotated pair
      this.accessToken = newAccessToken
      this.refreshToken = newRefreshToken

      return { accessToken: newAccessToken, refreshToken: newRefreshToken }
    } catch {
      return null
    }
  }
}

/**
 * AuthProvider component.
 *
 * Provides authentication state and methods to the application.
 * Also creates the application-wide ApiClient with the token manager
 * and provides it via ApiClientContext.
 *
 * Handles:
 * - Initial authentication check on app load
 * - Login with credentials (Standard Web)
 * - Login with Telegram initData (Telegram Mini App)
 * - Logout
 * - Current user management
 * - 401 handling and token refresh
 */
export function AuthProvider({ children }: AuthProviderProps) {
  console.log('[DIAG AuthProvider] MOUNT')
  const [state, setState] = useState<AuthState>('loading')
  const [user, setUser] = useState<CurrentUser | null>(null)
  const [isLoading, setIsLoading] = useState(false)

  // Diagnostic: log every state change
  useEffect(() => {
    console.log('[DIAG AuthProvider] state changed:', state, 'user:', user?.id ?? null)
  }, [state, user])

  useEffect(() => {
    return () => {
      console.log('[DIAG AuthProvider] UNMOUNT')
    }
  }, [])

  // Create token manager once and persist across re-renders
  const tokenManager = useMemo(() => new RuntimeTokenManager(), [])

  // Create API client with token manager
  const env = getEnvironment()
  const apiClient = useMemo(
    () => createApiClient(env.apiBaseUrl, tokenManager),
    [env.apiBaseUrl, tokenManager],
  )

  const isAuthenticated = state === 'authenticated' && user !== null

  /**
   * Fetches current user from Backend.
   * Called after successful login and on app initialization.
   * Maps Backend snake_case response to frontend camelCase model.
   */
   const fetchCurrentUser =
     useCallback(async (): Promise<CurrentUser | null> => {
       console.log('[DIAG fetchCurrentUser] called, token:', tokenManager.getAccessToken() ? 'present' : 'null')
       try {
         const response = await apiClient.get<{
           id: number
           first_name: string
           last_name: string
           nickname: string
           tg_user_id?: number
           created_at: string
           updated_at: string
         }>('/me')

         // Map Backend snake_case response to frontend camelCase model
         const backendUser = response.data
         console.log('[DIAG fetchCurrentUser] success:', backendUser.id)
         return {
           id: backendUser.id,
           firstName: backendUser.first_name,
           lastName: backendUser.last_name,
           nickname: backendUser.nickname,
           tgUserId: backendUser.tg_user_id,
           createdAt: backendUser.created_at,
           updatedAt: backendUser.updated_at,
         }
       } catch (error) {
         console.log('[DIAG fetchCurrentUser] error:', error)
         // If 401, user is not authenticated
         if (
           error instanceof Error &&
           'statusCode' in error &&
           error.statusCode === 401
         ) {
           return null
         }
         // For other errors, re-throw
         throw error
       }
     }, [apiClient])

  /**
   * Initializes authentication state on app load.
   * Checks if user has an active session by calling /me.
   */
  useEffect(() => {
    let mounted = true

    async function initializeAuth() {
      console.log('[DIAG initializeAuth] started, tokenManager.accessToken:', tokenManager.getAccessToken() ? 'present' : 'null')
      setState('loading')
      try {
        const currentUser = await fetchCurrentUser()
        console.log('[DIAG initializeAuth] fetchCurrentUser result:', currentUser ? `user ${currentUser.id}` : 'null')
        if (mounted) {
          if (currentUser) {
            setUser(currentUser)
            setState('authenticated')
          } else {
            setUser(null)
            setState('unauthenticated')
          }
        }
      } catch (e) {
        console.log('[DIAG initializeAuth] error:', e)
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
  }, [fetchCurrentUser])

  /**
   * Login with credentials (Standard Web mode).
   * Calls POST /api/v1/auth/login.
   * Backend returns snake_case tokens which are mapped to runtime storage.
   */
  const login = useCallback(
    async (credentials: LoginCredentials): Promise<void> => {
      setIsLoading(true)
      try {
        const response = await apiClient.post<{
          access_token: string
          refresh_token: string
          token_type: string
        }>('/auth/login', {
          body: credentials,
          skipAuth: true,
        })

        // Map Backend snake_case response to runtime token storage
        tokenManager.setTokens(
          response.data.access_token,
          response.data.refresh_token,
        )

        // Fetch current user with the new access token
        const currentUser = await fetchCurrentUser()
        if (currentUser) {
          setUser(currentUser)
          setState('authenticated')
        } else {
          throw new Error('Failed to fetch user after login')
        }
      } finally {
        setIsLoading(false)
      }
    },
    [apiClient, fetchCurrentUser, tokenManager],
  )

  /**
   * Login with Telegram initData (Telegram Mini App mode).
   * Calls POST /api/v1/auth/telegram with initData.
   */
   const loginWithTelegram = useCallback(
     async (initData: string): Promise<void> => {
       console.log('[DIAG loginWithTelegram] started')
       setIsLoading(true)
       try {
         const response = await apiClient.post<{
           access_token: string
           refresh_token: string
           token_type: string
         }>('/auth/telegram', {
           body: { init_data: initData },
           skipAuth: true,
         })

         // Map Backend snake_case response to runtime token storage
         tokenManager.setTokens(
           response.data.access_token,
           response.data.refresh_token,
         )
         console.log('[DIAG loginWithTelegram] tokens set')

         // Fetch current user with the new access token
         const currentUser = await fetchCurrentUser()
         if (currentUser) {
           setUser(currentUser)
           setState('authenticated')
           console.log('[DIAG loginWithTelegram] success, state=authenticated')
         } else {
           throw new Error('Failed to fetch user after Telegram login')
         }
       } finally {
         setIsLoading(false)
       }
     },
     [apiClient, fetchCurrentUser, tokenManager],
   )

  /**
   * Logout current user.
   * Calls POST /api/v1/auth/logout with refresh token.
   * Clears runtime tokens regardless of Backend response.
   */
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
      // Ignore logout errors — always clear local state
    } finally {
      // Always clear tokens and state regardless of Backend response
      tokenManager.clearTokens()
      setUser(null)
      setState('unauthenticated')
      setIsLoading(false)
    }
  }, [apiClient, tokenManager])

  /**
   * Refresh current user data from Backend.
   * Useful after mutations that might affect user data.
   */
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
    logout,
    refreshUser,
  }

  return (
    <ApiClientContext.Provider value={apiClient}>
      <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
    </ApiClientContext.Provider>
  )
}
