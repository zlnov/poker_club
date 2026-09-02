/**
 * API layer barrel exports.
 *
 * All Backend communication goes through the centralized ApiClient.
 * See 04_FE_SPEC.md section 14 (API Communication).
 */

export { ApiClient, createApiClient } from './client'
export type {
  ApiClientConfig,
  ApiRequestOptions,
  ApiResponse,
  HttpMethod,
} from './client'
export { ApiClientError, parseApiError } from './errors'
export type { ApiError, ApiErrorCode, ApiErrorResponse } from './errors'
