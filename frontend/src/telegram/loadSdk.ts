/**
 * Dynamic loader for the official Telegram Web Apps SDK.
 *
 * Loads https://telegram.org/js/telegram-web-app.js without blocking app boot.
 * Idempotent: concurrent callers share one in-flight promise and at most one
 * <script> tag. Failed / timed-out attempts can be retried explicitly.
 *
 * This is separate from Telegram Login / OIDC (telegram-login.js).
 */

export const TELEGRAM_WEB_APP_SDK_SRC =
  'https://telegram.org/js/telegram-web-app.js'

export const TELEGRAM_WEB_APP_SDK_TIMEOUT_MS = 3000

export type TelegramSdkLoadResult = 'ready' | 'failed' | 'timed_out'

export type TelegramSdkStatus =
  | 'idle'
  | 'loading'
  | 'ready'
  | 'failed'
  | 'timed_out'

const SCRIPT_ATTR = 'data-telegram-web-app-sdk'

type LoadListeners = {
  onSettled?: (result: TelegramSdkLoadResult) => void
}

let inflightPromise: Promise<TelegramSdkLoadResult> | null = null
let sdkReady = false
const readyListeners = new Set<() => void>()

/**
 * Read `window.Telegram.WebApp` without augmenting the global `Window` type
 * (Telegram Login declares `window.Telegram.Login` separately).
 */
function getRawTelegramWebApp(): unknown {
  if (typeof window === 'undefined') {
    return undefined
  }
  return (window as unknown as { Telegram?: { WebApp?: unknown } }).Telegram
    ?.WebApp
}

/**
 * True when `window.Telegram.WebApp` is already present.
 */
export function isTelegramWebAppSdkPresent(): boolean {
  return !!getRawTelegramWebApp()
}

function getExistingScript(): HTMLScriptElement | null {
  if (typeof document === 'undefined') {
    return null
  }
  return document.querySelector<HTMLScriptElement>(
    `script[${SCRIPT_ATTR}="true"], script[src="${TELEGRAM_WEB_APP_SDK_SRC}"]`,
  )
}

function notifyReady(): void {
  sdkReady = true
  for (const listener of readyListeners) {
    try {
      listener()
    } catch {
      // Listeners must not break SDK load lifecycle.
    }
  }
}

/**
 * Subscribe to SDK becoming ready (including late success after a timeout).
 * Returns an unsubscribe function.
 */
export function onTelegramWebAppSdkReady(listener: () => void): () => void {
  if (sdkReady || isTelegramWebAppSdkPresent()) {
    sdkReady = true
    listener()
    return () => {}
  }
  readyListeners.add(listener)
  return () => {
    readyListeners.delete(listener)
  }
}

function markScriptState(
  script: HTMLScriptElement,
  state: 'loading' | 'ready' | 'error',
): void {
  script.dataset.telegramSdkState = state
}

function attachScriptHandlers(
  script: HTMLScriptElement,
  finish: (result: TelegramSdkLoadResult) => void,
): void {
  const handleLoad = () => {
    if (isTelegramWebAppSdkPresent()) {
      markScriptState(script, 'ready')
      finish('ready')
      return
    }
    // Script executed but WebApp API is missing — treat as failure.
    markScriptState(script, 'error')
    finish('failed')
  }

  const handleError = () => {
    markScriptState(script, 'error')
    finish('failed')
  }

  script.addEventListener('load', handleLoad)
  script.addEventListener('error', handleError)
}

/**
 * Loads the Telegram Web Apps SDK asynchronously.
 *
 * - Resolves `ready` when `window.Telegram.WebApp` is available.
 * - Resolves `timed_out` if the timeout elapses first; the script may still
 *   finish later and notify `onTelegramWebAppSdkReady` subscribers.
 * - Resolves `failed` on script error.
 * - Concurrent calls share the same in-flight promise.
 */
export function loadTelegramWebAppSdk(
  options: { timeoutMs?: number } & LoadListeners = {},
): Promise<TelegramSdkLoadResult> {
  const timeoutMs = options.timeoutMs ?? TELEGRAM_WEB_APP_SDK_TIMEOUT_MS

  if (typeof window === 'undefined') {
    const result: TelegramSdkLoadResult = 'failed'
    options.onSettled?.(result)
    return Promise.resolve(result)
  }

  if (isTelegramWebAppSdkPresent()) {
    notifyReady()
    const result: TelegramSdkLoadResult = 'ready'
    options.onSettled?.(result)
    return Promise.resolve(result)
  }

  if (inflightPromise) {
    return inflightPromise.then((result) => {
      options.onSettled?.(result)
      return result
    })
  }

  inflightPromise = new Promise<TelegramSdkLoadResult>((resolve) => {
    let settled = false

    const finish = (result: TelegramSdkLoadResult) => {
      if (settled) {
        // Late ready after timeout: still notify listeners, do not re-resolve.
        if (result === 'ready') {
          notifyReady()
        }
        return
      }
      settled = true
      window.clearTimeout(timeoutId)

      if (result === 'ready') {
        notifyReady()
      } else {
        // Allow an explicit retry after failure / timeout.
        inflightPromise = null
      }

      options.onSettled?.(result)
      resolve(result)
    }

    const timeoutId = window.setTimeout(() => {
      finish('timed_out')
    }, timeoutMs)

    const existing = getExistingScript()
    if (existing) {
      if (existing.dataset.telegramSdkState === 'error') {
        // Stale failed tag — replace below.
        existing.remove()
      } else {
        attachScriptHandlers(existing, finish)
        // Race: script may have completed between query and listeners.
        if (isTelegramWebAppSdkPresent()) {
          markScriptState(existing, 'ready')
          finish('ready')
        }
        return
      }
    }

    const script = document.createElement('script')
    script.src = TELEGRAM_WEB_APP_SDK_SRC
    script.async = true
    script.setAttribute(SCRIPT_ATTR, 'true')
    markScriptState(script, 'loading')
    attachScriptHandlers(script, finish)
    document.head.appendChild(script)
  })

  return inflightPromise
}

/**
 * Explicitly retries SDK load after `failed` or `timed_out`.
 * No-ops (resolves `ready`) when the SDK is already present.
 * Does not auto-loop — callers decide when to invoke.
 */
export function retryTelegramWebAppSdk(
  options: { timeoutMs?: number } & LoadListeners = {},
): Promise<TelegramSdkLoadResult> {
  if (isTelegramWebAppSdkPresent()) {
    notifyReady()
    const result: TelegramSdkLoadResult = 'ready'
    options.onSettled?.(result)
    return Promise.resolve(result)
  }

  const existing = getExistingScript()
  if (existing) {
    const state = existing.dataset.telegramSdkState
    // Drop failed tags so a fresh network request can be made.
    // Keep in-flight / unknown tags and reuse them (single-flight).
    if (state === 'error') {
      existing.remove()
    }
  }

  inflightPromise = null
  sdkReady = false
  return loadTelegramWebAppSdk(options)
}
