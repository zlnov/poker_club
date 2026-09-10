/**
 * Telegram Integration Layer.
 *
 * Provides a safe abstraction over the Telegram WebApp API.
 * The main application must not directly depend on `window.Telegram.WebApp`.
 *
 * See 04_FE_SPEC.md section 6 (Telegram Integration Layer) and
 * section 7 (Telegram and Standard Web compatibility).
 *
 * Key design goals:
 * 1. Never crash when Telegram WebApp API is unavailable (Standard Web mode).
 * 2. Provide environment detection (Telegram Mini App vs Standard Web).
 * 3. Isolate Telegram-specific code from the main application logic.
 * 4. Provide fallbacks for all Telegram-specific capabilities.
 */

export type TelegramEnvironment = 'standard-web' | 'telegram-mini-app'

/**
 * Telegram WebApp initialization data (simplified).
 * Contains user identity information from Telegram.
 */
export interface TelegramInitData {
  /** Telegram user ID. */
  userId: number
  /** First name. */
  firstName: string
  /** Last name. */
  lastName: string
  /** Username. */
  username?: string
  /** Language code. */
  languageCode?: string
}

/**
 * Telegram theme parameters.
 */
export interface TelegramTheme {
  /** Background color. */
  bgColor: string
  /** Text color. */
  textColor: string
  /** Accent color. */
  accentColor: string
  /** Secondary text color. */
  secondaryTextColor: string
  /** Hint color. */
  hintColor: string
  /** Link color. */
  linkColor: string
  /** Button background color. */
  buttonBgColor: string
  /** Button text color. */
  buttonTextColor: string
  /** Button text hover color. */
  buttonTextHoverColor: string
  /** Divider color. */
  dividerColor: string
  /** Input background color. */
  inputBgColor: string
  /** Input text color. */
  inputTextColor: string
  /** Switch background color. */
  switchBgColor: string
  /** Switch text color. */
  switchTextColor: string
}

/**
 * Telegram WebApp API interface (minimal subset needed for Foundation).
 *
 * The full Telegram WebApp API is available at:
 * https://core.telegram.org/bots/webapps
 *
 * This interface only exposes what the Foundation needs and provides
 * safe fallbacks when the API is unavailable.
 */
export interface TelegramWebAppAPI {
  /** Initialization data from the WebApp. */
  initData?: string
  /** User identifier from initData. */
  initDataUnsafe?: {
    userId: string
  }
  /** Current theme parameters. */
  theme?: TelegramTheme
  /** Safe area insets. */
  safeArea?: {
    top: number
    bottom: number
    left: number
    right: number
  }
  /** Viewport dimensions. */
  viewport?: {
    width: number
    height: number
  }
  /** Is the app running in fullscreen mode? */
  isFullscreen?: boolean
  /** Use fullscreen mode. */
  useFullscreen?: () => void
  /** Exit fullscreen mode. */
  exitFullscreen?: () => void
  /** Main button API. */
  mainButton?: {
    text: string
    color: string
    textColor: string
    show(): void
    hide(): void
    setText(text: string): void
    setColor(color: string): void
    setTextColor(color: string): void
    onClick(handler: () => void): void
    offClick(handler: () => void): void
    disable(): void
    enable(): void
  }
  /** Back button API. */
  backButton?: {
    show(): void
    hide(): void
    onClick(handler: () => void): void
    offClick(handler: () => void): void
  }
}

/**
 * Environment detection result.
 */
export interface TelegramEnvironmentInfo {
  /** Current environment. */
  environment: TelegramEnvironment
  /** True when running in Telegram Mini App. */
  isTelegram: boolean
  /** True when running in Standard Web mode. */
  isStandardWeb: boolean
  /** Telegram WebApp API instance (may be undefined in Standard Web). */
  webApp?: TelegramWebAppAPI
}

/**
 * Creates a Telegram environment info object.
 *
 * In Telegram Mini App mode, this returns the actual WebApp API.
 * In Standard Web mode, this returns a safe object with `undefined` webApp,
 * and the environment flags are set accordingly.
 *
 * The main application should use this to detect the environment and
 * provide appropriate fallbacks for Telegram-specific UI.
 */
export function createTelegramEnvironment(
  rawWebApp?: TelegramWebAppAPI,
): TelegramEnvironmentInfo {
  // Telegram Mini App requires both the WebApp API AND initData.
  // In a regular browser, the Telegram SDK script may be loaded (creating
  // window.Telegram.WebApp) but initData will be absent because there is no
  // Telegram WebView context. Only treat it as Mini App when initData is present.
  const isTelegram = !!rawWebApp && !!rawWebApp.initData
  const environment: TelegramEnvironment = isTelegram
    ? 'telegram-mini-app'
    : 'standard-web'

  const isStandardWeb = !isTelegram

  // Safe default when not in Telegram
  const defaultInfo: TelegramEnvironmentInfo = {
    environment,
    isTelegram,
    isStandardWeb,
    webApp: undefined,
  }

  if (!isTelegram) {
    return defaultInfo
  }

  // In Telegram Mini App mode, return the actual WebApp API
  return {
    environment,
    isTelegram,
    isStandardWeb,
    webApp: rawWebApp,
  }
}

/**
 * Default Telegram environment info for Standard Web mode.
 *
 * This is used when the application starts before the Telegram WebApp
 * API is available (e.g., during SSR or Standard Web mode).
 */
export const defaultTelegramEnvironment: TelegramEnvironmentInfo = {
  environment: 'standard-web',
  isTelegram: false,
  isStandardWeb: true,
  webApp: undefined,
}

// Re-export the TelegramMode component (defined in TelegramMode.tsx
// to avoid JSX in this .ts file)
export type { TelegramModeProps } from './TelegramMode'
export { TelegramMode } from './TelegramMode'

// Re-export WebApp API utilities
export {
  getTelegramWebApp,
  initTelegramWebApp,
  applyTelegramTheme,
  type WebAppInitData,
} from './webApp'

// Re-export TelegramEnvironmentContext and Provider
export {
  TelegramEnvironmentContext,
  TelegramEnvironmentProvider,
  useTelegramEnvironmentContext,
} from './TelegramEnvironmentContext'
