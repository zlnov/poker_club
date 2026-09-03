/**
 * Hooks for the Poker Club Frontend.
 *
 * These hooks provide reusable logic for environment detection,
 * API client access, and other cross-cutting concerns.
 *
 * See 04_FE_SPEC.md section 14 (API Communication) and
 * section 6 (Telegram Integration Layer).
 */

import { useMemo, useContext } from 'react'
import { getEnvironment, type EnvironmentConfig } from '../env'
import type { ApiClient } from '../api'
import {
  createTelegramEnvironment,
  getTelegramWebApp,
  type TelegramEnvironmentInfo,
} from '../telegram'
import { ApiClientContext } from '../auth/ApiClientContext'

/**
 * Hook that provides the current environment configuration.
 *
 * Reads Vite environment variables and returns a typed config object.
 * Environment variables are static at build time, so no re-renders are needed.
 */
export function useEnvironment(): EnvironmentConfig {
  return useMemo(() => getEnvironment(), [])
}

/**
 * Hook that provides the API client instance.
 *
 * Returns the application-wide singleton ApiClient that is configured
 * with the current token manager for JWT Bearer authentication.
 *
 * See 06_API.md section 4 (API Client) and 07_AUTH.md section 5 (JWT).
 */
export function useApiClient(): ApiClient {
  const client = useContext(ApiClientContext)
  if (!client) {
    throw new Error('useApiClient must be used within an AuthProvider')
  }
  return client
}

/**
 * Hook that detects the Telegram environment.
 *
 * Returns information about whether the app is running in:
 * - Telegram Mini App mode (window.Telegram.WebApp is available)
 * - Standard Web mode (no Telegram WebApp API)
 *
 * The hook safely handles the absence of the Telegram WebApp API
 * and never throws (see 04_FE_SPEC.md section 7).
 */
export function useTelegramEnvironment(): TelegramEnvironmentInfo {
  return useMemo(() => {
    const webApp = getTelegramWebApp()
    return createTelegramEnvironment(webApp)
  }, [])
}

/**
 * Hook that provides the Telegram WebApp API instance.
 *
 * Returns `undefined` when not in Telegram Mini App mode.
 */
export function useTelegramWebApp() {
  const env = useTelegramEnvironment()
  return env.webApp
}

/**
 * Hook that provides the current environment name.
 */
export function useEnvironmentName(): EnvironmentConfig['envName'] {
  const env = useEnvironment()
  return env.envName
}
