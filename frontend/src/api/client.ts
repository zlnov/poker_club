/**
 * Centralized HTTP API Client.
 *
 * All HTTP requests to the Backend go through this client.
 * Components must not call fetch() directly — they use feature hooks
 * which in turn use this client (see 04_FE_SPEC.md section 14).
 *
 * Authentication uses JWT Bearer tokens (07_AUTH.md, 07.1_AUTH_SECURITY.md).
 * Access tokens are stored in runtime memory only — never in localStorage/sessionStorage.
 * The client adds `Authorization: Bearer <access_token>` to protected requests.
 *
 * Token refresh is handled automatically: when a request returns 401,
 * the client attempts to refresh the access token using the refresh token.
 * The original request is retried once after a successful refresh.
 */

import { parseApiError } from './errors'

/**
 * HTTP methods supported by the API client.
 */
export type HttpMethod = 'GET' | 'POST' | 'PATCH' | 'DELETE' | 'PUT'

/**
 * Request options for the API client.
 */
export interface ApiRequestOptions {
  method?: HttpMethod
  headers?: Record<string, string>
  body?: unknown
  query?: Record<string, string | number | boolean | undefined>
  /** Skip auth header for this request (e.g., login, refresh). */
  skipAuth?: boolean
}

/**
 * Response from the API client.
 */
export interface ApiResponse<T = unknown> {
  data: T
  status: number
}

/**
 * Token manager interface.
 * Provides access to the current access token and handles refresh.
 */
export interface TokenManager {
  /** Returns the current access token, or null if not available. */
  getAccessToken(): string | null
  /** Returns the current refresh token, or null if not available. */
  getRefreshToken(): string | null
  /** Sets new tokens after login or refresh. */
  setTokens(accessToken: string, refreshToken: string): void
  /** Clears all tokens (logout). */
  clearTokens(): void
  /** Refreshes the access token using the refresh token. Returns new tokens or null on failure. */
  refreshAccessToken(): Promise<{
    accessToken: string
    refreshToken: string
  } | null>
}

/**
 * Configuration for the API client.
 */
export interface ApiClientConfig {
  /** Base URL for the API (includes /api/v1 prefix). */
  baseUrl: string
  /** Token manager for JWT token handling. */
  tokenManager: TokenManager
}

/**
 * Centralized API client for all Backend communication.
 *
 * Usage:
 *   const client = new ApiClient({ baseUrl: 'http://localhost:8080/api/v1', tokenManager })
 *   const response = await client.request('/clubs', { method: 'GET' })
 */
export class ApiClient {
  private readonly baseUrl: string
  private readonly tokenManager: TokenManager

  constructor(config: ApiClientConfig) {
    this.baseUrl = config.baseUrl.replace(/\/+$/, '') // strip trailing slash
    this.tokenManager = config.tokenManager
  }

  /**
   * Builds a full URL with optional query parameters.
   */
  private buildUrl(path: string, query?: ApiRequestOptions['query']): string {
    const normalizedPath = path.startsWith('/') ? path : `/${path}`
    const url = new URL(this.baseUrl + normalizedPath)

    if (query) {
      for (const [key, value] of Object.entries(query)) {
        if (value !== undefined && value !== null) {
          url.searchParams.append(key, String(value))
        }
      }
    }

    return url.toString()
  }

  /**
   * Serializes the request body to JSON if present.
   */
  private serializeBody(body: unknown): BodyInit | undefined {
    if (body === undefined || body === null) {
      return undefined
    }
    return JSON.stringify(body)
  }

  /**
   * Builds headers for the request, including Authorization if available.
   */
  private buildHeaders(options: ApiRequestOptions): Record<string, string> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...options.headers,
    }

    // Add Authorization header if not skipped and token is available
    if (!options.skipAuth) {
      const token = this.tokenManager.getAccessToken()
      if (token) {
        headers['Authorization'] = `Bearer ${token}`
      }
    }

    return headers
  }

  /**
   * Performs an HTTP request to the Backend API.
   *
   * - Adds Authorization: Bearer <access_token> to protected requests.
   * - On 401, attempts token refresh and retries once.
   * - Parses JSON responses.
   * - Throws ApiClientError for non-2xx responses.
   */
  async request<T = unknown>(
    path: string,
    options: ApiRequestOptions = {},
  ): Promise<ApiResponse<T>> {
    const { method = 'GET', body, query, skipAuth = false } = options

    const url = this.buildUrl(path, query)
    const headers = this.buildHeaders(options)

    const response = await fetch(url, {
      method,
      headers,
      body: this.serializeBody(body),
    })

    const status = response.status

    // Parse response body
    let responseBody: unknown
    const contentType = response.headers.get('content-type')
    if (contentType?.includes('application/json')) {
      responseBody = await response.json().catch(() => null)
    }

    if (!response.ok) {
      // Handle 401: attempt token refresh and retry once
      if (status === 401 && !skipAuth) {
        const refreshed = await this.tokenManager.refreshAccessToken()
        if (refreshed) {
          // Retry the original request with new token
          const newHeaders = this.buildHeaders({ ...options, skipAuth: false })
          const retryResponse = await fetch(url, {
            method,
            headers: newHeaders,
            body: this.serializeBody(body),
          })

          const retryStatus = retryResponse.status
          let retryBody: unknown
          const retryContentType = retryResponse.headers.get('content-type')
          if (retryContentType?.includes('application/json')) {
            retryBody = await retryResponse.json().catch(() => null)
          }

          if (!retryResponse.ok) {
            throw parseApiError(retryBody, retryStatus)
          }

          return {
            data: retryBody as T,
            status: retryStatus,
          }
        }
      }

      throw parseApiError(responseBody, status)
    }

    return {
      data: responseBody as T,
      status,
    }
  }

  /**
   * Convenience method for GET requests.
   */
  get<T = unknown>(
    path: string,
    options: Omit<ApiRequestOptions, 'method'> = {},
  ) {
    return this.request<T>(path, { ...options, method: 'GET' })
  }

  /**
   * Convenience method for POST requests.
   */
  post<T = unknown>(
    path: string,
    options: Omit<ApiRequestOptions, 'method'> = {},
  ) {
    return this.request<T>(path, { ...options, method: 'POST' })
  }

  /**
   * Convenience method for PATCH requests.
   */
  patch<T = unknown>(
    path: string,
    options: Omit<ApiRequestOptions, 'method'> = {},
  ) {
    return this.request<T>(path, { ...options, method: 'PATCH' })
  }

  /**
   * Convenience method for DELETE requests.
   */
  delete<T = unknown>(
    path: string,
    options: Omit<ApiRequestOptions, 'method'> = {},
  ) {
    return this.request<T>(path, { ...options, method: 'DELETE' })
  }
}

/**
 * Creates a singleton ApiClient instance from environment configuration.
 * The client is created once and reused across the application.
 */
export function createApiClient(
  baseUrl: string,
  tokenManager: TokenManager,
): ApiClient {
  return new ApiClient({ baseUrl, tokenManager })
}
