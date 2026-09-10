/**
 * Telegram authentication hook.
 *
 * Handles automatic authentication in Telegram Mini App mode
 * by sending initData to the Backend.
 *
 * See 07_AUTH.md section 3 (Telegram Mini App Authentication) and
 * 07.1_AUTH_SECURITY.md section 5 (Telegram Mini App Authentication).
 */

import { useEffect, useCallback, useRef } from 'react'
import { useAuth } from '../auth'
import { useTelegramWebApp } from '../hooks'

/**
 * Hook that handles Telegram Mini App authentication.
 *
 * In Telegram Mini App mode, automatically sends initData to Backend
 * for authentication on app load.
 * In Standard Web mode, does nothing.
 *
 * Ensures authentication is only attempted once per session to prevent
 * duplicate initialization.
 *
 * If the Telegram WebApp API is detected but initData is not yet available,
 * the hook retries on subsequent renders until initData is present.
 */
export function useTelegramAuth() {
  const { loginWithTelegram, state } = useAuth()
  const webApp = useTelegramWebApp()

  // Track whether we've already attempted Telegram authentication
  const authAttemptedRef = useRef(false)

   const authenticateWithTelegram = useCallback(async () => {
    // Prevent duplicate authentication attempts
    if (authAttemptedRef.current) {
      return
    }

    // Only proceed if we have initData
    if (!webApp?.initData) {
      // Don't mark as attempted — initData may become available later
      return
    }

    // Mark as attempted to prevent duplicate calls
    authAttemptedRef.current = true

    try {
      await loginWithTelegram(webApp.initData)
    } catch (error) {
      // Authentication failed - user will need to use Standard Web login
      // Error is logged but not shown to user (they'll see login page)
      // Do not log initData per 07.1_AUTH_SECURITY.md
      console.warn(
        'Telegram authentication failed:',
        error instanceof Error ? error.message : 'Unknown error',
      )
    }
  }, [webApp, loginWithTelegram])

  // Run authentication when webApp or initData changes
  useEffect(() => {
    // Only run in Telegram Mini App mode and when not already authenticated
    if (webApp && state !== 'authenticated') {
      authenticateWithTelegram()
    }
  }, [webApp, state, authenticateWithTelegram])

  return { authenticateWithTelegram }
}
