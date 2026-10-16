/**
 * Demo session persistence (localStorage).
 *
 * Production JWT tokens stay in memory only.
 * Demo uses an opaque bearer token that may be restored across page reloads.
 *
 * A suspended session is kept until it expires so the visitor can return
 * to the same club, but it is not treated as the active API identity.
 */

const STORAGE_KEY = 'poker_club_demo_session'

export interface DemoSession {
  accessToken: string
  clubId: number
  expiresAt: string
  /** Set while the visitor uses real login/registration. The club is kept. */
  suspended?: boolean
}

function readStoredSession(): DemoSession | null {
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

/** Active demo session. Suspended sessions are ignored. */
export function loadDemoSession(): DemoSession | null {
  const session = readStoredSession()
  if (!session || session.suspended) {
    return null
  }
  return session
}

/** Saved demo session, including one paused for real account login. */
export function loadRestorableDemoSession(): DemoSession | null {
  return readStoredSession()
}

export function saveDemoSession(session: DemoSession): void {
  localStorage.setItem(
    STORAGE_KEY,
    JSON.stringify({
      accessToken: session.accessToken,
      clubId: session.clubId,
      expiresAt: session.expiresAt,
    }),
  )
}

/** Keep the club token, but stop using it for API calls. */
export function suspendDemoSession(): void {
  const session = readStoredSession()
  if (!session) {
    return
  }
  localStorage.setItem(
    STORAGE_KEY,
    JSON.stringify({ ...session, suspended: true }),
  )
}

export function clearDemoSession(): void {
  localStorage.removeItem(STORAGE_KEY)
}

export function isDemoMode(): boolean {
  return loadDemoSession() !== null
}
