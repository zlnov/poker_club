/**
 * Telegram WebApp API wrapper.
 *
 * Wraps the raw `window.Telegram.WebApp` API with safe fallbacks.
 * The main application should use `createTelegramEnvironment()` to get
 * a typed, safe representation of the Telegram environment.
 *
 * See 04_FE_SPEC.md section 6 (Telegram Integration Layer).
 *
 * Design:
 * - Never throws when Telegram API is unavailable.
 * - All methods are no-ops when called in Standard Web mode.
 * - Provides typed access to WebApp capabilities.
 */

import type { TelegramWebAppAPI } from './index'

export interface WebAppInitData {
  user?: {
    id: number
    first_name: string
    last_name?: string
    username?: string
    language_code?: string
  }
}

/**
 * Safely retrieves the Telegram WebApp API instance.
 *
 * Returns `undefined` when not in Telegram Mini App mode.
 * The caller must check the return value before using Telegram API.
 */
export function getTelegramWebApp(): TelegramWebAppAPI | undefined {
  // Safe access — if Telegram WebApp API is not available, return undefined
  if (typeof window === 'undefined') {
    return undefined
  }

  const tg = (window as any).Telegram?.WebApp
  if (!tg) {
    return undefined
  }

  // Return a typed wrapper with safe method signatures
  return {
    initData: tg.initData ?? undefined,
    initDataUnsafe: tg.initDataUnsafe ?? undefined,
    theme: tg.theme,
    safeArea: tg.safeArea,
    viewport: tg.viewport,
    isFullscreen: tg.isFullscreen ?? false,
    useFullscreen: () => {
      try {
        tg.useFullscreen?.()
      } catch {
        // no-op: fullscreen not supported or not in Telegram mode
      }
    },
    exitFullscreen: () => {
      try {
        tg.exitFullscreen?.()
      } catch {
        // no-op
      }
    },
    mainButton: {
      text: tg.mainButton?.text ?? '',
      color: tg.mainButton?.color ?? '',
      textColor: tg.mainButton?.textColor ?? '',
      show: () => {
        try {
          tg.mainButton?.show?.()
        } catch {
          // no-op
        }
      },
      hide: () => {
        try {
          tg.mainButton?.hide?.()
        } catch {
          // no-op
        }
      },
      setText: (text: string) => {
        try {
          tg.mainButton?.setText?.(text)
        } catch {
          // no-op
        }
      },
      setColor: (color: string) => {
        try {
          tg.mainButton?.setColor?.(color)
        } catch {
          // no-op
        }
      },
      setTextColor: (color: string) => {
        try {
          tg.mainButton?.setTextColor?.(color)
        } catch {
          // no-op
        }
      },
      onClick: (handler: () => void) => {
        try {
          tg.mainButton?.onClick?.(handler)
        } catch {
          // no-op
        }
      },
      offClick: (handler: () => void) => {
        try {
          tg.mainButton?.offClick?.(handler)
        } catch {
          // no-op
        }
      },
      disable: () => {
        try {
          tg.mainButton?.disable?.()
        } catch {
          // no-op
        }
      },
      enable: () => {
        try {
          tg.mainButton?.enable?.()
        } catch {
          // no-op
        }
      },
    },
    backButton: {
      show: () => {
        try {
          tg.backButton?.show?.()
        } catch {
          // no-op
        }
      },
      hide: () => {
        try {
          tg.backButton?.hide?.()
        } catch {
          // no-op
        }
      },
      onClick: (handler: () => void) => {
        try {
          tg.backButton?.onClick?.(handler)
        } catch {
          // no-op
        }
      },
      offClick: (handler: () => void) => {
        try {
          tg.backButton?.offClick?.(handler)
        } catch {
          // no-op
        }
      },
    },
  }
}

/**
 * Initializes the Telegram WebApp.
 *
 * Should be called after the component mounts.
 * In Standard Web mode, this is a no-op.
 */
export function initTelegramWebApp(): void {
  const webApp = getTelegramWebApp()
  if (!webApp) {
    // Not in Telegram mode — no-op
    return
  }

  // WebApp is already available — no additional initialization needed
  // for Phase 0 (Foundation). Telegram-specific initialization
  // (Main Button, Back Button, etc.) will be added in RM_FE_08.
}

/**
 * Safely applies Telegram theme to the document.
 *
 * In Standard Web mode, applies the default theme.
 */
export function applyTelegramTheme(): void {
  const webApp = getTelegramWebApp()
  if (!webApp) {
    // Standard Web — apply default theme from CSS variables
    // The CSS already handles light/dark mode via prefers-color-scheme
    return
  }

  try {
    if (webApp.theme) {
      const { bgColor, textColor } = webApp.theme
      // Apply theme colors to CSS variables
      ;(document.documentElement.style as any).setProperty('--bg', bgColor)
      ;(document.documentElement.style as any).setProperty('--text', textColor)
    }
  } catch {
    // no-op: theme application error
  }
}
