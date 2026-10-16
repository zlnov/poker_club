/**
 * Persists the last opened club id across pages outside /clubs/:clubId.
 */

import { useCallback, useEffect, useState } from 'react'

const STORAGE_KEY = 'poker-club:lastClubId'

function readStoredClubId(): number | null {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return null
    const id = Number.parseInt(raw, 10)
    return Number.isFinite(id) && id > 0 ? id : null
  } catch {
    return null
  }
}

/**
 * Hook for reading/writing the last opened club id in localStorage.
 */
export function useLastClubId() {
  const [lastClubId, setLastClubIdState] = useState<number | null>(() =>
    readStoredClubId(),
  )

  useEffect(() => {
    const onStorage = (event: StorageEvent) => {
      if (event.key === STORAGE_KEY) {
        setLastClubIdState(readStoredClubId())
      }
    }
    window.addEventListener('storage', onStorage)
    return () => window.removeEventListener('storage', onStorage)
  }, [])

  const setLastClubId = useCallback((id: number) => {
    if (!Number.isFinite(id) || id <= 0) return
    try {
      localStorage.setItem(STORAGE_KEY, String(id))
    } catch {
      // ignore quota / private mode
    }
    setLastClubIdState(id)
  }, [])

  const clearLastClubId = useCallback(() => {
    try {
      localStorage.removeItem(STORAGE_KEY)
    } catch {
      // ignore
    }
    setLastClubIdState(null)
  }, [])

  return { lastClubId, setLastClubId, clearLastClubId }
}
