/**
 * Authentication context hook.
 *
 * Provides access to the authentication context.
 * Must be used within an AuthProvider.
 *
 * See 07_AUTH.md and 04_FE_SPEC.md section 12 (Authentication).
 */

import { useContext } from 'react'
import { AuthContext } from './AuthContext'
import type { AuthContextValue } from './index'

/**
 * Custom hook to access the authentication context.
 * Must be used within an AuthProvider.
 */
export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext)
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider')
  }
  return context
}
