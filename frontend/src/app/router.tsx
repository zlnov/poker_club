/**
 * Application router.
 *
 * Defines the client-side routing structure for the Poker Club Frontend.
 *
 * Routing structure from 05_FE_UX.md section 5:
 *   /
 *   ├── /clubs
 *   ├── /clubs/:clubId
 *   ├── /clubs/:clubId/members
 *   ├── /clubs/:clubId/games
 *   ├── /clubs/:clubId/statistics
 *   ├── /clubs/:clubId/settings
 *   ├── /games/:gameId
 *   └── /profile
 *
 * See 04_FE_SPEC.md section 11 (Routing).
 */

import { createBrowserRouter } from 'react-router-dom'
import {
  HomePage,
  ClubsPage,
  ClubDashboard,
  ClubGames,
  ClubMembers,
  ClubStatistics,
  ClubSettings,
  GamePage,
  ProfilePage,
} from '../pages'
import { RootLayout, ClubLayout } from '../components'

/**
 * Application routes.
 *
 * Route structure:
 * - / (root) — Home page (wrapped in RootLayout)
 * - /clubs — Club list
 * - /clubs/:clubId — Club dashboard (wrapped in ClubLayout with tabs)
 *   - /clubs/:clubId/members — Club members
 *   - /clubs/:clubId/games — Club games
 *   - /clubs/:clubId/statistics — Club statistics
 *   - /clubs/:clubId/settings — Club settings
 * - /games/:gameId — Game details
 * - /profile — User profile
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
      {
        path: 'clubs',
        element: <ClubsPage />,
      },
      {
        path: 'clubs/:clubId',
        element: <ClubLayout />,
        children: [
          {
            index: true,
            element: <ClubDashboard />,
          },
          {
            path: 'games',
            element: <ClubGames />,
          },
          {
            path: 'members',
            element: <ClubMembers />,
          },
          {
            path: 'statistics',
            element: <ClubStatistics />,
          },
          {
            path: 'settings',
            element: <ClubSettings />,
          },
        ],
      },
      {
        path: 'games/:gameId',
        element: <GamePage />,
      },
      {
        path: 'profile',
        element: <ProfilePage />,
      },
    ],
  },
])
