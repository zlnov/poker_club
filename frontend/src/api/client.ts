/**
 * Centralized HTTP API Client.
 *
 * All HTTP requests to the Backend go through this client.
 * Components must not call fetch() directly — they use feature hooks
 * which in turn use this client (see 04_FE_SPEC.md section 14).
 *
 * Session management is cookie-based (HTTP-only, Secure, SameSite).
 * The client sends credentials automatically so the Backend session cookie
 * is included on every request (see 07_AUTH.md section 5).
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
}

/**
 * Response from the API client.
 */
export interface ApiResponse<T = unknown> {
  data: T
  status: number
}

/**
 * Configuration for the API client.
 */
export interface ApiClientConfig {
  /** Base URL for the API (includes /api/v1 prefix). */
  baseUrl: string
}

/**
 * Centralized API client for all Backend communication.
 *
 * Usage:
 *   const client = new ApiClient({ baseUrl: 'http://localhost:8080/api/v1' })
 *   const response = await client.request('/clubs', { method: 'GET' })
 */
export class ApiClient {
  private readonly baseUrl: string

  constructor(config: ApiClientConfig) {
    this.baseUrl = config.baseUrl.replace(/\/+$/, '') // strip trailing slash
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
   * Performs an HTTP request to the Backend API.
   *
   * - Automatically includes credentials (cookies) for session-based auth.
   * - Parses JSON responses.
   * - Throws ApiClientError for non-2xx responses.
   */
  async request<T = unknown>(
    path: string,
    options: ApiRequestOptions = {},
  ): Promise<ApiResponse<T>> {
    const { method = 'GET', headers = {}, body, query } = options

    const url = this.buildUrl(path, query)

    const response = await fetch(url, {
      method,
      headers: {
        'Content-Type': 'application/json',
        ...headers,
      },
      body: this.serializeBody(body),
      credentials: 'include', // send HTTP-only session cookie
    })

    const status = response.status

    // Parse response body
    let responseBody: unknown
    const contentType = response.headers.get('content-type')
    if (contentType?.includes('application/json')) {
      responseBody = await response.json().catch(() => null)
    }

    if (!response.ok) {
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
export function createApiClient(baseUrl: string): ApiClient {
  return new ApiClient({ baseUrl })
}
