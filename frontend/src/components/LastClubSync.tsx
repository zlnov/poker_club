/**
 * Keeps lastClubId in sync with the active club route or game club.
 */

import { useEffect } from 'react'
import { useParams } from 'react-router-dom'
import { useAuth } from '../auth'
import { useGame } from '../features'
import { useLastClubId } from '../hooks/useLastClubId'

export function LastClubSync() {
  const params = useParams()
  const { isAuthenticated } = useAuth()
  const { setLastClubId } = useLastClubId()

  const routeClubId = params.clubId
    ? Number.parseInt(params.clubId, 10)
    : 0
  const gameId = params.gameId ? Number.parseInt(params.gameId, 10) : 0

  const { data: game } = useGame(isAuthenticated ? gameId : 0)

  useEffect(() => {
    if (routeClubId > 0) {
      setLastClubId(routeClubId)
    }
  }, [routeClubId, setLastClubId])

  useEffect(() => {
    if (game?.clubId) {
      setLastClubId(game.clubId)
    }
  }, [game?.clubId, setLastClubId])

  return null
}
