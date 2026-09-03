/**
 * Authentication context.
 *
 * Provides the React context for authentication state.
 *
 * See 07_AUTH.md and 04_FE_SPEC.md section 12 (Authentication).
 */

import { createContext } from 'react'
import type { AuthContextValue } from './index'

/**
 * Auth context for the application.
 */
export const AuthContext = createContext<AuthContextValue | null>(null)
