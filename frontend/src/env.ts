/**
 * Environment configuration access.
 *
 * Reads public, non-secret configuration from Vite environment variables.
 * No secrets (tokens, passwords, keys) are ever stored here.
 *
 * See 04_FE_SPEC.md section 28 (Environment Configuration).
 */

export type Environment = 'development' | 'production'

export interface EnvironmentConfig {
  /** Base URL for the Backend HTTP API (includes /api/v1 prefix). */
  apiBaseUrl: string
  /** Current environment name. */
  envName: Environment
  /** True when running in development mode. */
  isDevelopment: boolean
  /** True when running in production mode. */
  isProduction: boolean
}

/**
 * Returns the parsed environment configuration.
 * Throws if required variables are missing.
 */
export function getEnvironment(): EnvironmentConfig {
  const apiBaseUrl = import.meta.env.VITE_API_BASE_URL
  const envName = (import.meta.env.VITE_ENV_NAME ??
    'development') as Environment

  if (!apiBaseUrl) {
    throw new Error(
      'VITE_API_BASE_URL is not defined. Please check your environment configuration.',
    )
  }

  const isDevelopment = envName === 'development'
  const isProduction = envName === 'production'

  return {
    apiBaseUrl,
    envName,
    isDevelopment,
    isProduction,
  }
}
