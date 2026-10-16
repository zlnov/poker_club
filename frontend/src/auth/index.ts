/**
 * Authentication layer for Poker Club Frontend.
 */

export type AuthState = 'unauthenticated' | 'loading' | 'authenticated'

export interface CurrentUser {
  id: number
  firstName: string
  lastName: string
  /** Poker Club nickname; may be null until set. */
  nickname: string | null
  /** Telegram username without @; may be null. */
  tgUserName: string | null
  email: string | null
  phoneNumber: string | null
  tgUserId?: number
  hasPassword: boolean
  createdAt: string
  updatedAt: string
}

export interface LoginCredentials {
  email: string
  password: string
}

export interface TelegramWebAuthResult {
  authenticated: boolean
  registrationRequired?: boolean
  registrationToken?: string
}

export interface CompleteTelegramRegistrationInput {
  registrationToken: string
  email: string
  password: string
  passwordConfirmation: string
  nickname: string
  firstName?: string
  lastName?: string
}

export interface UpdateProfileInput {
  firstName?: string
  lastName?: string
  nickname?: string
  email?: string | null
  phoneNumber?: string | null
}

export interface ChangePasswordInput {
  currentPassword?: string
  newPassword: string
  passwordConfirmation: string
}

export interface AuthContextValue {
  state: AuthState
  user: CurrentUser | null
  isAuthenticated: boolean
  isLoading: boolean
  isDemo: boolean
  login: (credentials: LoginCredentials) => Promise<void>
  loginWithTelegram: (initData: string) => Promise<void>
  loginWithTelegramWeb: (idToken: string, challengeToken: string) => Promise<TelegramWebAuthResult>
  completeTelegramRegistration: (input: CompleteTelegramRegistrationInput) => Promise<void>
  /** Leave demo without deleting the saved club session. */
  pauseDemo: () => void
  startDemo: () => Promise<number>
  updateProfile: (input: UpdateProfileInput) => Promise<void>
  changePassword: (input: ChangePasswordInput) => Promise<void>
  logout: () => Promise<void>
  refreshUser: () => Promise<void>
}

export const AUTH_CONTEXT_PLACEHOLDER: AuthContextValue = {
  state: 'unauthenticated',
  user: null,
  isAuthenticated: false,
  isLoading: false,
  isDemo: false,
  login: async () => {},
  loginWithTelegram: async () => {},
  loginWithTelegramWeb: async () => ({ authenticated: false }),
  completeTelegramRegistration: async () => {},
  pauseDemo: () => {},
  startDemo: async () => 0,
  updateProfile: async () => {},
  changePassword: async () => {},
  logout: async () => {},
  refreshUser: async () => {},
}

export { AuthProvider } from './AuthProvider'
export { useAuth } from './useAuth'
export { useTelegramAuth } from './useTelegramAuth'
export {
  useAuthErrorHandler,
  withAuthErrorHandling,
} from './useAuthErrorHandler'
