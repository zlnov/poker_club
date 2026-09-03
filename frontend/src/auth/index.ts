/**
 * Authentication layer for Poker Club Frontend.
 *
 * Phase 2 (Authentication) — implements authentication, JWT handling,
 * login, logout, and current user management.
 *
 * See 04_FE_SPEC.md section 12 (Authentication),
 * 07_AUTH.md (Authentication Specification),
 * 07.1_AUTH_SECURITY.md (Authentication Security & Token Policy),
 * 06_API.md section 4.0 (Authentication endpoints).
 */

/**
 * Authentication state.
 */
export type AuthState = 'unauthenticated' | 'loading' | 'authenticated'

/**
 * Current user information.
 *
 * Retrieved from `GET /api/v1/me` (06_API.md section 4.1).
 * Backend response uses snake_case:
 * {
 *   "id": 1,
 *   "first_name": "...",
 *   "last_name": "...",
 *   "nickname": "...",
 *   "tg_user_id": 123,
 *   "created_at": "...",
 *   "updated_at": "..."
 * }
 *
 * Note: Backend does NOT return clubMemberships in /me response.
 * Club memberships are fetched separately via club endpoints.
 */
export interface CurrentUser {
  /** Player ID. */
  id: number
  /** First name. */
  firstName: string
  /** Last name. */
  lastName: string
  /** Username/nickname. */
  nickname: string
  /** Telegram user ID (if available). */
  tgUserId?: number
  /** Created at timestamp. */
  createdAt: string
  /** Updated at timestamp. */
  updatedAt: string
}

/**
 * Login credentials for Standard Web authentication.
 */
export interface LoginCredentials {
  /** Login (username or email). */
  login: string
  /** Password. */
  password: string
}

/**
 * Login response from Backend.
 * Backend returns snake_case:
 * {
 *   "access_token": "...",
 *   "refresh_token": "...",
 *   "token_type": "Bearer"
 * }
 */
export interface LoginResponse {
  /** Access JWT token. */
  accessToken: string
  /** Refresh token. */
  refreshToken: string
  /** Token type (always "Bearer"). */
  tokenType: string
}

/**
 * Telegram authentication request.
 * Sends Telegram initData to Backend for validation.
 */
export interface TelegramAuthRequest {
  /** Raw Telegram initData string. */
  initData: string
}

/**
 * Authentication context value.
 */
export interface AuthContextValue {
  /** Current authentication state. */
  state: AuthState
  /** Current user (if authenticated). */
  user: CurrentUser | null
  /** Whether the user is authenticated. */
  isAuthenticated: boolean
  /** Whether authentication is being checked. */
  isLoading: boolean
  /** Login with credentials (Standard Web). */
  login: (credentials: LoginCredentials) => Promise<void>
  /** Login with Telegram initData (Telegram Mini App). */
  loginWithTelegram: (initData: string) => Promise<void>
  /** Logout current user. */
  logout: () => Promise<void>
  /** Refresh current user data from Backend. */
  refreshUser: () => Promise<void>
}

/**
 * Authentication context placeholder.
 * Replaced by AuthProvider in the application.
 */
export const AUTH_CONTEXT_PLACEHOLDER: AuthContextValue = {
  state: 'unauthenticated',
  user: null,
  isAuthenticated: false,
  isLoading: false,
  login: async () => {},
  loginWithTelegram: async () => {},
  logout: async () => {},
  refreshUser: async () => {},
}

// Re-export AuthProvider and useAuth
export { AuthProvider } from './AuthProvider'
export { useAuth } from './useAuth'
export { useTelegramAuth } from './useTelegramAuth'
export {
  useAuthErrorHandler,
  withAuthErrorHandling,
} from './useAuthErrorHandler'
