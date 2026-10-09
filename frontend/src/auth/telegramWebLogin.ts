/**
 * Official Telegram Login JS library wrapper (OIDC id_token popup).
 * Uses https://telegram.org/js/telegram-login.js — NOT the legacy telegram-widget.js.
 *
 * @see https://core.telegram.org/bots/telegram-login
 */

const TELEGRAM_LOGIN_SRC = 'https://telegram.org/js/telegram-login.js'

export interface TelegramLoginAuthData {
  id_token?: string
  user?: Record<string, unknown>
  error?: string
  [key: string]: unknown
}

export type TelegramLoginScope = 'profile' | 'phone' | 'write'

declare global {
  interface Window {
    Telegram?: {
      Login?: {
        init: (
          options: TelegramLoginInitOptions,
          callback?: (data: TelegramLoginAuthData | false | null) => void,
        ) => void
        open: (
          callback?: (data: TelegramLoginAuthData | false | null) => void,
        ) => void
        auth: (
          options: TelegramLoginInitOptions,
          callback: (data: TelegramLoginAuthData | false | null) => void,
        ) => void
        close: () => void
      }
    }
  }
}

interface TelegramLoginInitOptions {
  /** Client ID from @BotFather → Login Widget (not a secret). */
  client_id: number | string
  /** Optional scopes: profile, phone, write (bot messaging). */
  scope?: TelegramLoginScope[] | string
  lang?: string
  /** OIDC nonce; must match backend challenge nonce embedded in id_token. */
  nonce?: string
}

let scriptLoadPromise: Promise<void> | null = null

function loadTelegramLoginScript(): Promise<void> {
  if (typeof window === 'undefined') {
    return Promise.reject(new Error('Telegram Login requires a browser'))
  }
  if (window.Telegram?.Login?.auth) {
    return Promise.resolve()
  }
  if (scriptLoadPromise) {
    return scriptLoadPromise
  }

  scriptLoadPromise = new Promise((resolve, reject) => {
    const existing = document.querySelector<HTMLScriptElement>(
      `script[src^="https://telegram.org/js/telegram-login.js"]`,
    )
    if (existing) {
      existing.addEventListener('load', () => resolve())
      existing.addEventListener('error', () =>
        reject(new Error('Failed to load Telegram Login script')),
      )
      if (window.Telegram?.Login?.auth) {
        resolve()
      }
      return
    }

    const script = document.createElement('script')
    script.src = TELEGRAM_LOGIN_SRC
    script.async = true
    script.onload = () => resolve()
    script.onerror = () =>
      reject(new Error('Failed to load Telegram Login script'))
    document.head.appendChild(script)
  })

  return scriptLoadPromise
}

/**
 * Opens the official Telegram Login popup and resolves with id_token.
 * `nonce` must be the value from POST /auth/telegram/web/challenge so that
 * the id_token nonce matches the backend challenge binding.
 *
 * BotFather Allowed URLs must include the current page origin
 * (library uses location.origin + location.pathname as redirect_uri for postMessage).
 */
export async function requestTelegramIdToken(
  clientId: string,
  nonce: string,
): Promise<string> {
  if (!clientId) {
    throw new Error('Telegram Login requires a client_id')
  }
  if (!nonce) {
    throw new Error('Telegram Login requires a challenge nonce')
  }

  await loadTelegramLoginScript()

  const auth = window.Telegram?.Login?.auth
  if (!auth) {
    throw new Error('Telegram Login API is unavailable')
  }

  return new Promise((resolve, reject) => {
    auth(
      {
        client_id: clientId,
        scope: ['profile'],
        nonce,
      },
      (data) => {
        if (!data) {
          reject(new Error('Telegram Login was cancelled'))
          return
        }
        if (typeof data.error === 'string' && data.error) {
          reject(new Error(`Telegram Login failed: ${data.error}`))
          return
        }
        const idToken =
          typeof data.id_token === 'string' ? data.id_token : undefined
        if (!idToken) {
          reject(
            new Error(
              'Telegram Login did not return id_token. Ensure Telegram Login / OIDC is configured for this bot and the current origin is in BotFather Allowed URLs.',
            ),
          )
          return
        }
        resolve(idToken)
      },
    )
  })
}
