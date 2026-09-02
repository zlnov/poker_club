/**
 * Home page component.
 *
 * Phase 0 (Foundation) — displays a minimal landing page that confirms
 * the application is running in Standard Web mode.
 *
 * No business functionality is implemented at this stage.
 * See RM_FE_0.md (Frontend Foundation).
 */

import { useTelegramEnvironment } from '../hooks'

/**
 * Home page.
 *
 * Displays:
 * - Application name
 * - Current environment (Standard Web or Telegram Mini App)
 * - Confirmation that the app is running
 *
 * This is a placeholder page for Phase 0. Business functionality
 * will be added in subsequent phases.
 */
export function HomePage() {
  const telegramEnv = useTelegramEnvironment()

  return (
    <main className="home-page">
      <h1>Poker Club</h1>
      <p>Frontend Foundation — Phase 0</p>
      <div className="environment-info">
        <p>
          <strong>Environment:</strong>{' '}
          {telegramEnv.isTelegram ? 'Telegram Mini App' : 'Standard Web'}
        </p>
        <p>
          <strong>API Base URL:</strong>{' '}
          {import.meta.env.VITE_API_BASE_URL || 'not configured'}
        </p>
      </div>
    </main>
  )
}
