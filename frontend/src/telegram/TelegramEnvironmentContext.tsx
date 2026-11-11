/**
 * Telegram Environment Context.
 *
 * Provides a single source of truth for Telegram environment detection
 * across the entire application. Without this context, each component
 * that calls `useTelegramEnvironment()` would have its own independent
 * state and polling, leading to inconsistent `isTelegram` values.
 *
 * SDK loading is asynchronous and non-blocking: the app always starts in
 * Standard Web mode, then upgrades to Mini App if WebApp + initData appear.
 *
 * See 04_FE_SPEC.md section 6 (Telegram Integration Layer).
 */

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import { createTelegramEnvironment, type TelegramEnvironmentInfo } from './index'
import { getTelegramWebApp } from './webApp'
import {
  loadTelegramWebAppSdk,
  onTelegramWebAppSdkReady,
  retryTelegramWebAppSdk,
  TELEGRAM_WEB_APP_SDK_TIMEOUT_MS,
  type TelegramSdkLoadResult,
  type TelegramSdkStatus,
} from './loadSdk'

const POLL_INTERVAL_MS = 100

export interface TelegramEnvironmentContextValue extends TelegramEnvironmentInfo {
  /** Independent of `isTelegram` — reflects SDK script load lifecycle. */
  sdkStatus: TelegramSdkStatus
  /**
   * Explicitly retry SDK load after `failed` or `timed_out`.
   * Safe to call anytime; no-ops when SDK is already ready.
   */
  retryTelegramSdk: () => Promise<TelegramSdkLoadResult>
}

/**
 * Context for sharing Telegram environment state across the application.
 * Created by `TelegramEnvironmentProvider` and consumed by
 * `useTelegramEnvironmentContext()`.
 */
export const TelegramEnvironmentContext =
  createContext<TelegramEnvironmentContextValue | null>(null)

function readEnvironment(): TelegramEnvironmentInfo {
  return createTelegramEnvironment(getTelegramWebApp())
}

/**
 * Provider that detects the Telegram environment and shares it via context.
 *
 * Boot behaviour:
 * 1. Starts immediately as Standard Web (does not block on SDK).
 * 2. Kicks off async SDK load with timeout.
 * 3. Polls for WebApp + initData for a bounded window (2–5s).
 * 4. After timeout/failure, still accepts late SDK readiness and updates env.
 * 5. Exposes `retryTelegramSdk` for explicit Mini App recovery (no auto-loop).
 */
export function TelegramEnvironmentProvider({
  children,
}: {
  children: ReactNode
}) {
  const [env, setEnv] = useState<TelegramEnvironmentInfo>(() =>
    readEnvironment(),
  )
  // Start as `loading` when SDK is absent so first paint does not flash `idle`
  // and the mount effect does not need a synchronous setState.
  const [sdkStatus, setSdkStatus] = useState<TelegramSdkStatus>(() =>
    getTelegramWebApp() ? 'ready' : 'loading',
  )
  const envRef = useRef(env)

  useEffect(() => {
    envRef.current = env
  }, [env])

  const syncEnvironment = useCallback(() => {
    const next = readEnvironment()
    const prev = envRef.current
    if (
      prev.isTelegram !== next.isTelegram ||
      prev.webApp?.initData !== next.webApp?.initData ||
      prev.environment !== next.environment
    ) {
      setEnv(next)
    }
  }, [])

  const retryTelegramSdk = useCallback(async () => {
    setSdkStatus('loading')
    const result = await retryTelegramWebAppSdk({
      timeoutMs: TELEGRAM_WEB_APP_SDK_TIMEOUT_MS,
    })
    setSdkStatus(result === 'ready' ? 'ready' : result)
    syncEnvironment()
    return result
  }, [syncEnvironment])

  // Async SDK load on mount — never blocks first paint / React boot.
  useEffect(() => {
    if (getTelegramWebApp()) {
      syncEnvironment()
      return
    }

    let cancelled = false

    void loadTelegramWebAppSdk({
      timeoutMs: TELEGRAM_WEB_APP_SDK_TIMEOUT_MS,
      onSettled: (result) => {
        if (cancelled) {
          return
        }
        setSdkStatus(result === 'ready' ? 'ready' : result)
        if (result === 'ready') {
          syncEnvironment()
        }
      },
    })

    return () => {
      cancelled = true
    }
  }, [syncEnvironment])

  // Late readiness after timed_out / failed (script finishes later, or retry).
  useEffect(() => {
    return onTelegramWebAppSdkReady(() => {
      setSdkStatus('ready')
      syncEnvironment()
    })
  }, [syncEnvironment])

  // Bounded polling for WebApp API and initData during the detection window.
  // Also runs a short window after retry when status returns to loading.
  useEffect(() => {
    if (env.isTelegram && env.webApp?.initData) {
      return
    }

    if (sdkStatus !== 'loading' && sdkStatus !== 'ready') {
      // After timeout/failure, rely on onTelegramWebAppSdkReady + explicit retry
      // rather than unbounded polling.
      // One immediate sync covers a race where SDK arrived between status updates.
      if (sdkStatus === 'timed_out' || sdkStatus === 'failed') {
        syncEnvironment()
      }
      return
    }

    const startedAt = Date.now()
    const interval = window.setInterval(() => {
      syncEnvironment()

      const current = envRef.current
      if (current.isTelegram && current.webApp?.initData) {
        window.clearInterval(interval)
        return
      }

      if (Date.now() - startedAt >= TELEGRAM_WEB_APP_SDK_TIMEOUT_MS) {
        window.clearInterval(interval)
      }
    }, POLL_INTERVAL_MS)

    return () => {
      window.clearInterval(interval)
    }
  }, [env.isTelegram, env.webApp?.initData, sdkStatus, syncEnvironment])

  const value = useMemo<TelegramEnvironmentContextValue>(
    () => ({
      ...env,
      sdkStatus,
      retryTelegramSdk,
    }),
    [env, sdkStatus, retryTelegramSdk],
  )

  return (
    <TelegramEnvironmentContext.Provider value={value}>
      {children}
    </TelegramEnvironmentContext.Provider>
  )
}

/**
 * Hook to access the shared Telegram environment context.
 *
 * Must be used within a `TelegramEnvironmentProvider`.
 * Returns the current Telegram environment info plus SDK load status.
 */
export function useTelegramEnvironmentContext(): TelegramEnvironmentContextValue {
  const context = useContext(TelegramEnvironmentContext)
  if (!context) {
    throw new Error(
      'useTelegramEnvironmentContext must be used within a TelegramEnvironmentProvider',
    )
  }
  return context
}
