/**
 * Telegram Environment Context.
 *
 * Provides a single source of truth for Telegram environment detection
 * across the entire application. Without this context, each component
 * that calls `useTelegramEnvironment()` would have its own independent
 * state and polling, leading to inconsistent `isTelegram` values.
 *
 * See 04_FE_SPEC.md section 6 (Telegram Integration Layer).
 */

import {
  createContext,
  useContext,
  useState,
  useEffect,
  ReactNode,
} from 'react'
import {
  createTelegramEnvironment,
  getTelegramWebApp,
  type TelegramEnvironmentInfo,
} from './index'

/**
 * Context for sharing Telegram environment state across the application.
 * Created by `TelegramEnvironmentProvider` and consumed by
 * `useTelegramEnvironmentContext()`.
 */
export const TelegramEnvironmentContext =
  createContext<TelegramEnvironmentInfo | null>(null)

/**
 * Provider that detects the Telegram environment and shares it via context.
 *
 * On mount, it checks for `window.Telegram.WebApp` synchronously.
 * If not available, it polls every 100ms until the API is detected
 * or the component unmounts.
 *
 * Once the WebApp API is detected, it continues polling for `initData`
 * availability, since `initData` may not be populated immediately
 * even after the WebApp API object exists.
 *
 * In Telegram Mini App, the WebApp API and initData may be injected
 * asynchronously after the initial render. Polling ensures detection.
 */
export function TelegramEnvironmentProvider({
  children,
}: {
  children: ReactNode
}) {
  const [env, setEnv] = useState<TelegramEnvironmentInfo>(() => {
    const webApp = getTelegramWebApp()
    return createTelegramEnvironment(webApp)
  })

  useEffect(() => {
    // If already detected Telegram with initData, no need to poll
    if (env.isTelegram && env.webApp?.initData) {
      return
    }

    // Poll for Telegram WebApp API availability and initData
    const interval = setInterval(() => {
      const webApp = getTelegramWebApp()
      if (webApp) {
        const newEnv = createTelegramEnvironment(webApp)
        // Update state if Telegram was just detected, or if initData became available
        if (
          !env.isTelegram ||
          (!env.webApp?.initData && newEnv.webApp?.initData)
        ) {
          setEnv(newEnv)
        }
        // Stop polling once we have both Telegram and initData
        if (newEnv.isTelegram && newEnv.webApp?.initData) {
          clearInterval(interval)
        }
      }
    }, 100)

    // Clean up on unmount
    return () => clearInterval(interval)
  }, [env.isTelegram, env.webApp?.initData])

  return (
    <TelegramEnvironmentContext.Provider value={env}>
      {children}
    </TelegramEnvironmentContext.Provider>
  )
}

/**
 * Hook to access the shared Telegram environment context.
 *
 * Must be used within a `TelegramEnvironmentProvider`.
 * Returns the current Telegram environment info.
 */
export function useTelegramEnvironmentContext(): TelegramEnvironmentInfo {
  const context = useContext(TelegramEnvironmentContext)
  if (!context) {
    throw new Error(
      'useTelegramEnvironmentContext must be used within a TelegramEnvironmentProvider',
    )
  }
  return context
}
