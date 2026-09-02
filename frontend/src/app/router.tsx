/**
 * Application router.
 *
 * Defines the client-side routing structure for the Poker Club Frontend.
 *
 * Phase 0 (Foundation) only includes a minimal route structure.
 * Full routing per 05_FE_UX.md section 5 will be implemented in
 * subsequent phases (RM_FE_01+).
 *
 * See 04_FE_SPEC.md section 11 (Routing).
 */

import { createBrowserRouter } from 'react-router-dom'
import { HomePage } from '../pages/HomePage'
import { RootLayout } from '../components/RootLayout'

/**
 * Application routes.
 *
 * Current routes (Phase 0 — Foundation only):
 * - / (root) — Home page
 *
 * Full route structure from 05_FE_UX.md section 5 will be added in
 * subsequent phases:
 *   /clubs
 *   /clubs/:clubId
 *   /clubs/:clubId/members
 *   /clubs/:clubId/games
 *   /clubs/:clubId/statistics
 *   /clubs/:clubId/settings
 *   /games/:gameId
 *   /profile
 */
export const router = createBrowserRouter([
  {
    path: '/',
    element: <RootLayout />,
    children: [
      {
        index: true,
        element: <HomePage />,
      },
    ],
  },
])
