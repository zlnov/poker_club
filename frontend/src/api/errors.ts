/**
 * API error types and utilities.
 *
 * Error format is defined in 06_API.md section 7:
 * {
 *   "error": {
 *     "code": "GAME_ALREADY_FINISHED",
 *     "message": "Game is already finished"
 *   }
 * }
 *
 * Frontend uses `code` for error handling, not `message`.
 */

/**
 * Stable machine-readable error codes returned by the Backend.
 * Full list defined in 06_API.md sections 7.2–7.13.
 */
export type ApiErrorCode =
  // Auth errors
  | 'AUTHENTICATION_REQUIRED'
  | 'INVALID_CREDENTIALS'
  | 'INVALID_TELEGRAM_INIT_DATA'
  | 'TELEGRAM_INIT_DATA_EXPIRED'
  | 'SESSION_EXPIRED'
  | 'SESSION_INVALID'
  // Club errors
  | 'CLUB_NOT_FOUND'
  | 'CLUB_ALREADY_EXISTS'
  | 'CLUB_ACCESS_DENIED'
  | 'CLUB_OWNER_REQUIRED'
  | 'CLUB_INVALID_STATE'
  | 'CLUB_UPDATE_NOT_ALLOWED'
  // Member errors
  | 'MEMBER_NOT_FOUND'
  | 'MEMBER_ALREADY_EXISTS'
  | 'MEMBER_NOT_IN_CLUB'
  | 'MEMBERSHIP_REQUEST_NOT_FOUND'
  | 'MEMBERSHIP_REQUEST_ALREADY_PROCESSED'
  | 'MEMBERSHIP_REQUEST_NOT_ALLOWED'
  | 'MEMBER_APPROVAL_NOT_ALLOWED'
  | 'MEMBER_REJECTION_NOT_ALLOWED'
  | 'MEMBER_REMOVAL_NOT_ALLOWED'
  | 'ROLE_CHANGE_NOT_ALLOWED'
  | 'INVALID_MEMBER_ROLE'
  // Invitation errors
  | 'INVITATION_NOT_FOUND'
  | 'INVITATION_EXPIRED'
  | 'INVITATION_ALREADY_USED'
  | 'INVITATION_ALREADY_EXISTS'
  | 'INVITATION_NOT_ALLOWED'
  | 'INVALID_INVITATION'
  // Game errors
  | 'GAME_NOT_FOUND'
  | 'GAME_ACCESS_DENIED'
  | 'GAME_INVALID_STATE'
  | 'GAME_ALREADY_STARTED'
  | 'GAME_ALREADY_FINISHED'
  | 'GAME_ALREADY_CANCELLED'
  | 'GAME_START_NOT_ALLOWED'
  | 'GAME_FINISH_NOT_ALLOWED'
  | 'GAME_CANCEL_NOT_ALLOWED'
  | 'GAME_UPDATE_NOT_ALLOWED'
  | 'INVALID_GAME_TYPE'
  | 'INVALID_GAME_CONFIGURATION'
  // Game participant errors
  | 'PARTICIPANT_NOT_FOUND'
  | 'PLAYER_ALREADY_IN_GAME'
  | 'PLAYER_NOT_IN_GAME'
  | 'PARTICIPANT_ADD_NOT_ALLOWED'
  | 'PARTICIPANT_REMOVE_NOT_ALLOWED'
  // Banker errors
  | 'BANKER_NOT_FOUND'
  | 'BANKER_NOT_ALLOWED'
  | 'BANKER_ALREADY_ASSIGNED'
  | 'BANKER_CHANGE_NOT_ALLOWED'
  | 'INVALID_BANKER'
  // Buy-in / Rebuy / Chips errors
  | 'BUY_IN_NOT_ALLOWED'
  | 'REBUY_NOT_ALLOWED'
  | 'CHIPS_UPDATE_NOT_ALLOWED'
  | 'INVALID_BUY_IN_AMOUNT'
  | 'INVALID_REBUY_AMOUNT'
  | 'INVALID_CHIPS_AMOUNT'
  // Result errors
  | 'RESULTS_NOT_FOUND'
  | 'RESULTS_NOT_AVAILABLE'
  | 'RESULT_CORRECTION_NOT_ALLOWED'
  | 'INVALID_GAME_RESULT'
  | 'RESULT_ALREADY_CORRECTED'
  // Statistics errors
  | 'STATISTICS_NOT_AVAILABLE'
  | 'PLAYER_STATISTICS_NOT_FOUND'
  | 'CLUB_STATISTICS_NOT_FOUND'
  // Validation errors
  | 'VALIDATION_ERROR'
  // System errors
  | 'RATE_LIMIT_EXCEEDED'
  | 'INTERNAL_SERVER_ERROR'
  | 'SERVICE_UNAVAILABLE'
  // Fallback for unknown codes
  | string

/**
 * Structured API error returned by the Backend.
 */
export interface ApiError {
  /** Stable machine-readable error code. */
  code: ApiErrorCode
  /** Human-readable message for display/logging. */
  message: string
}

/**
 * Raw error response body from the Backend.
 */
export interface ApiErrorResponse {
  error: ApiError
}

/**
 * Custom error class for API errors.
 * Carries the structured error code so callers can branch on it.
 */
export class ApiClientError extends Error {
  readonly code: ApiErrorCode
  readonly statusCode: number

  constructor(code: ApiErrorCode, message: string, statusCode: number) {
    super(message)
    this.name = 'ApiClientError'
    this.code = code
    this.statusCode = statusCode
  }

  /** Returns true if this error represents an authentication issue. */
  isAuthError(): boolean {
    return (
      this.code === 'AUTHENTICATION_REQUIRED' ||
      this.code === 'INVALID_CREDENTIALS' ||
      this.code === 'SESSION_EXPIRED' ||
      this.code === 'SESSION_INVALID' ||
      this.code === 'INVALID_TELEGRAM_INIT_DATA' ||
      this.code === 'TELEGRAM_INIT_DATA_EXPIRED'
    )
  }

  /** Returns true if this error represents a 401/403 status. */
  isForbidden(): boolean {
    return this.statusCode === 401 || this.statusCode === 403
  }
}

/**
 * Parses a raw API error response body into an ApiClientError.
 * Falls back to a generic error if the body cannot be parsed.
 */
export function parseApiError(
  body: unknown,
  statusCode: number,
): ApiClientError {
  if (
    body !== null &&
    typeof body === 'object' &&
    'error' in body &&
    typeof (body as ApiErrorResponse).error === 'object'
  ) {
    const err = (body as ApiErrorResponse).error
    if (typeof err.code === 'string' && typeof err.message === 'string') {
      return new ApiClientError(err.code, err.message, statusCode)
    }
  }

  // Fallback for unexpected error shapes
  return new ApiClientError(
    'INTERNAL_SERVER_ERROR',
    `Unexpected error (HTTP ${statusCode})`,
    statusCode,
  )
}
