/**
 * Authentication layer foundation.
 *
 * Phase 0 (Foundation) — provides the structure for the authentication layer
 * but does NOT implement authentication, authorization, login, or logout.
 *
 * These will be implemented in RM_FE_02 (Authentication).
 *
 * See 04_FE_SPEC.md section 12 (Authentication) and
 * 07_AUTH.md (Authentication Specification).
 */

/**
 * Authentication state.
 *
 * Phase 0 only defines the type structure.
 * Actual authentication implementation comes in RM_FE_02.
 */
export type AuthState = 'unauthenticated' | 'loading' | 'authenticated'

/**
 * Current user information.
 *
 * Retrieved from `GET /api/v1/me` (06_API.md section 4.1).
 * Phase 0 only defines the type — the actual API call is not implemented.
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
  /** Club memberships with roles. */
  clubMemberships: ClubMembership[]
}

/**
 * Club membership information for the current user.
 */
export interface ClubMembership {
  /** Club ID. */
  clubId: number
  /** Club name. */
  clubName: string
  /** User's role in the club. */
  role: 'owner' | 'admin' | 'member'
  /** Membership status. */
  status: 'pending' | 'active' | 'banned' | 'left'
}

/**
 * Authentication context value.
 *
 * Phase 0 only defines the interface — the actual context provider
 * will be implemented in RM_FE_02.
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
}

/**
 * Placeholder for the authentication context.
 *
 * This will be replaced with a real context provider in RM_FE_02.
 * For Phase 0, it provides the type structure only.
 */
export const AUTH_CONTEXT_PLACEHOLDER: AuthContextValue = {
  state: 'unauthenticated',
  user: null,
  isAuthenticated: false,
  isLoading: false,
}
