/**
 * Hooks for the Poker Club Frontend.
 *
 * These hooks provide reusable logic for environment detection,
 * API client access, and other cross-cutting concerns.
 *
 * See 04_FE_SPEC.md section 14 (API Communication) and
 * section 6 (Telegram Integration Layer).
 */

import { useMemo } from 'react'
import { getEnvironment, type EnvironmentConfig } from '../env'
import { createApiClient, type ApiClient } from '../api'
import {
  createTelegramEnvironment,
  getTelegramWebApp,
  type TelegramEnvironmentInfo,
} from '../telegram'

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
 * Creates a singleton ApiClient from the environment configuration.
 * The client is memoized so it persists across re-renders.
 */
export function useApiClient(): ApiClient {
  const env = useEnvironment()
  return useMemo(() => createApiClient(env.apiBaseUrl), [env.apiBaseUrl])
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
