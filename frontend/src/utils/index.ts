/**
 * Utility functions for the Poker Club Frontend.
 *
 * Presentation-oriented logic only (see 04_FE_SPEC.md section 17).
 * Business logic (calculations, business rules) must not live here.
 */

import type { LoadingState, ToastConfig } from '../types'
import type { ApiErrorCode } from '../api'

/**
 * Formats a numeric value as currency.
 *
 * Frontend does not perform currency conversion — it only formats
 * values received from the Backend (04_FE_SPEC.md section 1031).
 *
 * @param value The numeric value to format.
 * @param currencyCode The 3-letter currency code (USD, EUR, RUB, etc.).
 * @returns Formatted currency string.
 */
export function formatCurrency(value: number, currencyCode: string): string {
  if (isNaN(value)) {
    return '—'
  }
  return new Intl.NumberFormat(undefined, {
    style: 'currency',
    currency: currencyCode,
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value)
}

/**
 * Formats a chip value (numeric, without currency symbol).
 *
 * Chip values are formatted separately from monetary values
 * (04_FE_SPEC.md section 1031, 1033).
 *
 * @param value The chip value to format.
 * @returns Formatted chip value string.
 */
export function formatChipValue(value: number): string {
  if (isNaN(value)) {
    return '—'
  }
  return new Intl.NumberFormat(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  }).format(value)
}

/**
 * Formats a percentage value (e.g., ROI).
 *
 * @param value The percentage value (e.g., 25.4 for 25.4%).
 * @returns Formatted percentage string.
 */
export function formatPercentage(value: number): string {
  if (isNaN(value)) {
    return '—'
  }
  return (
    new Intl.NumberFormat(undefined, {
      minimumFractionDigits: 1,
      maximumFractionDigits: 1,
    }).format(value) + '%'
  )
}

/**
 * Formats a duration (time interval) into HH:MM:SS or MM:SS format.
 *
 * @param seconds The duration in seconds.
 * @returns Formatted duration string.
 */
export function formatDuration(seconds: number): string {
  if (isNaN(seconds) || seconds < 0) {
    return '—'
  }

  const h = Math.floor(seconds / 3600)
  const m = Math.floor((seconds % 3600) / 60)
  const s = seconds % 60

  if (h > 0) {
    return `${h}:${m.toString().padStart(2, '0')}:${s
      .toString()
      .padStart(2, '0')}`
  }
  return `${m}:${s.toString().padStart(2, '0')}`
}

/**
 * Formats a date string into a readable format.
 *
 * Uses the user's local timezone (04_FE_SPEC.md section 496–511).
 *
 * @param isoDate ISO 8601 date string from the Backend.
 * @returns Formatted date string.
 */
export function formatDate(isoDate: string): string {
  if (!isoDate) {
    return '—'
  }
  try {
    const date = new Date(isoDate)
    return new Intl.DateTimeFormat(undefined, {
      day: '2-digit',
      month: 'short',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    }).format(date)
  } catch {
    return '—'
  }
}

/**
 * Formats a player name (truncates if too long).
 *
 * @param name The player name to format.
 * @returns Formatted player name.
 */
export function formatPlayerName(name: string): string {
  if (!name) {
    return '—'
  }
  if (name.length <= 20) {
    return name
  }
  return name.substring(0, 17) + '...'
}

/**
 * Checks if a value represents a positive profit.
 *
 * Used for UI purposes only — business logic is on the Backend.
 *
 * @param value The profit value.
 * @returns True if the value is positive.
 */
export function isPositiveProfit(value: number): boolean {
  return value > 0
}

/**
 * Checks if a value represents a negative profit (loss).
 *
 * Used for UI purposes only — business logic is on the Backend.
 *
 * @param value The profit value.
 * @returns True if the value is negative.
 */
export function isNegativeProfit(value: number): boolean {
  return value < 0
}

/**
 * Creates a loading state object.
 *
 * Used for managing local UI state (04_FE_SPEC.md section 569).
 *
 * @param initialState The initial state name.
 * @returns Initialized loading state.
 */
export function createLoadingState(initialState: LoadingState = 'idle'): {
  state: LoadingState
  setState: (newState: LoadingState) => void
  toggle: () => void
} {
  let currentState = initialState

  return {
    state: currentState,
    setState(newState) {
      currentState = newState
    },
    toggle() {
      if (currentState === 'loading') {
        currentState = 'idle'
      } else {
        currentState = 'loading'
      }
    },
  }
}

/**
 * Creates an error state object.
 *
 * Used for managing error state in UI components.
 *
 * @param message The error message.
 * @returns Error state object.
 */
export function createErrorState(message: string): {
  state: 'error'
  message: string
  reset: () => void
} {
  return {
    state: 'error',
    message,
    reset() {
      // Reset logic — clear the error message
    },
  }
}

/**
 * Creates a success state object.
 *
 * Used for managing success state after mutations (04_FE_SPEC.md section 730–733).
 *
 * @param message The success message.
 * @returns Success state object.
 */
export function createSuccessState(message: string): {
  state: 'success'
  message: string
  reset: () => void
} {
  return {
    state: 'success',
    message,
    reset() {
      // Reset logic — clear the success message
    },
  }
}

export type { LoadingState, ToastConfig, ApiErrorCode }
