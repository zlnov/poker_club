/**
 * Club layout — outlet wrapper for club-scoped routes.
 *
 * Desktop and mobile navigation live in RootLayout (sidebar / horizontal nav).
 * This layout only provides the nested route outlet.
 */

import { Outlet } from 'react-router-dom'

/**
 * Club layout component.
 */
export function ClubLayout() {
  return <Outlet />
}
