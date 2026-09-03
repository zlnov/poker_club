/**
 * Protected route component.
 *
 * Wraps routes that require authentication.
 * Redirects to login page if user is not authenticated.
 *
 * See 04_FE_SPEC.md section 11 (Routing) and
 * 07_AUTH.md section 8 (Session Expiration).
 */

import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '../auth'
import { LoadingState } from '../components'

/**
 * ProtectedRoute component.
 *
 * Checks authentication state and either:
 * - Renders child routes (Outlet) if authenticated
 * - Shows loading state while checking authentication
 * - Redirects to /login with return path if not authenticated
 */
export function ProtectedRoute() {
  const { state, isLoading } = useAuth()
  const location = useLocation()

  // Show loading while authentication is being checked
  if (isLoading || state === 'loading') {
    return <LoadingState />
  }

  // Redirect to login if not authenticated
  if (state === 'unauthenticated') {
    return <Navigate to="/login" state={{ from: location.pathname }} replace />
  }

  // Render protected content
  return <Outlet />
}

/**
 * Public route component.
 *
 * Wraps routes that should only be accessible when NOT authenticated
 * (e.g., login page). Redirects authenticated users to /clubs.
 */
export function PublicRoute() {
  const { state, isLoading } = useAuth()

  // Show loading while authentication is being checked
  if (isLoading || state === 'loading') {
    return <LoadingState />
  }

  // Redirect to clubs if already authenticated
  if (state === 'authenticated') {
    return <Navigate to="/clubs" replace />
  }

  // Render public content
  return <Outlet />
}
