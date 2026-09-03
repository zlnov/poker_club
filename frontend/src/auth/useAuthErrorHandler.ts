/**
 * Authentication error handling hook.
 *
 * Provides a way to handle 401 Unauthorized errors globally
 * by clearing authentication state and redirecting to login.
 *
 * See 07_AUTH.md section 8 (Session Expiration) and
 * 07.1_AUTH_SECURITY.md section 10 (Authentication Error Contract).
 */

import { useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from './index'
import { ApiClientError } from '../api'

/**
 * Hook that provides a handler for authentication errors.
 *
 * When a 401 error occurs, this handler:
 * 1. Clears the authentication state
 * 2. Redirects to the login page with the current path as return URL
 */
export function useAuthErrorHandler() {
  const { state, logout } = useAuth()
  const navigate = useNavigate()

  const handleAuthError = useCallback(
    (error: unknown) => {
      // Check if it's a 401 authentication error
      if (error instanceof ApiClientError && error.isAuthError()) {
        // Only handle if we were previously authenticated
        if (state === 'authenticated') {
          logout()
          // Navigate to login with current path as return URL
          navigate('/login', {
            replace: true,
            state: { from: window.location.pathname },
          })
        }
        return true // Error was handled
      }
      return false // Error was not handled
    },
    [state, logout, navigate],
  )

  return { handleAuthError }
}

/**
 * Higher-order function to wrap API calls with auth error handling.
 *
 * Usage:
 *   const { handleAuthError } = useAuthErrorHandler()
 *   const wrappedCall = withAuthErrorHandling(apiCall, handleAuthError)
 */
export function withAuthErrorHandling<
  T extends (...args: unknown[]) => Promise<unknown>,
>(apiCall: T, handleAuthError: (error: unknown) => boolean): T {
  return (async (...args: unknown[]) => {
    try {
      return await apiCall(...args)
    } catch (error) {
      if (handleAuthError(error)) {
        // Error was handled (redirected to login)
        // Return a rejected promise that won't be caught by callers
        throw new Error('Authentication required')
      }
      throw error
    }
  }) as T
}
