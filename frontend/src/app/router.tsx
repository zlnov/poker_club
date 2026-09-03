/**
 * Application router.
 *
 * Defines the client-side routing structure for the Poker Club Frontend.
 *
 * Routing structure from 05_FE_UX.md section 5:
 *   /
 *   ├── /login
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
  LoginPage,
  ClubsPage,
  ClubDashboard,
  ClubGames,
  ClubMembers,
  ClubStatistics,
  ClubSettings,
  GamePage,
  ProfilePage,
} from '../pages'
import {
  RootLayout,
  ClubLayout,
  ProtectedRoute,
  PublicRoute,
} from '../components'

/**
 * Application routes.
 *
 * Route structure:
 * - / (root) — Home page (wrapped in RootLayout)
 * - /login — Login page (public route, redirects if authenticated)
 * - /clubs — Club list (protected)
 * - /clubs/:clubId — Club dashboard (wrapped in ClubLayout with tabs, protected)
 *   - /clubs/:clubId/members — Club members (protected)
 *   - /clubs/:clubId/games — Club games (protected)
 *   - /clubs/:clubId/statistics — Club statistics (protected)
 *   - /clubs/:clubId/settings — Club settings (protected)
 * - /games/:gameId — Game details (protected)
 * - /profile — User profile (protected)
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
        path: 'login',
        element: <PublicRoute />,
        children: [
          {
            index: true,
            element: <LoginPage />,
          },
        ],
      },
      {
        path: 'clubs',
        element: <ProtectedRoute />,
        children: [
          {
            index: true,
            element: <ClubsPage />,
          },
          {
            path: ':clubId',
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
        ],
      },
      {
        path: 'games/:gameId',
        element: <ProtectedRoute />,
        children: [
          {
            index: true,
            element: <GamePage />,
          },
        ],
      },
      {
        path: 'profile',
        element: <ProtectedRoute />,
        children: [
          {
            index: true,
            element: <ProfilePage />,
          },
        ],
      },
    ],
  },
])
