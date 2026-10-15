/**
 * Demo session persistence (localStorage).
 *
 * Production JWT tokens stay in memory only.
 * Demo uses an opaque bearer token that may be restored across page reloads.
 */

const STORAGE_KEY = 'poker_club_demo_session'

export interface DemoSession {
  accessToken: string
  clubId: number
  expiresAt: string
}

export function loadDemoSession(): DemoSession | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) {
      return null
    }
    const parsed = JSON.parse(raw) as DemoSession
    if (!parsed.accessToken || !parsed.clubId || !parsed.expiresAt) {
      clearDemoSession()
      return null
    }
    if (Date.parse(parsed.expiresAt) <= Date.now()) {
      clearDemoSession()
      return null
    }
    return parsed
  } catch {
    clearDemoSession()
    return null
  }
}

export function saveDemoSession(session: DemoSession): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(session))
}

export function clearDemoSession(): void {
  localStorage.removeItem(STORAGE_KEY)
}

export function isDemoMode(): boolean {
  return loadDemoSession() !== null
}
