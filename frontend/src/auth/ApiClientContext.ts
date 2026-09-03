/**
 * API Client context.
 *
 * Provides the application-wide singleton ApiClient instance
 * configured with the token manager for JWT Bearer authentication.
 *
 * See 04_FE_SPEC.md section 14 (API Communication) and
 * 07_AUTH.md section 5 (JWT).
 */

import { createContext } from 'react'
import type { ApiClient } from '../api'

/**
 * Context for the application-wide ApiClient instance.
 * Created by AuthProvider and consumed by useApiClient().
 */
export const ApiClientContext = createContext<ApiClient | null>(null)
