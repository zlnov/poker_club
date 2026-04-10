import { Navigate } from 'react-router-dom'
import { useAuthStore } from '@/features/auth/store'

interface ProtectedRouteProps {
  children: React.ReactNode
  requireRole?: 'admin' | 'member'
}

export function ProtectedRoute({ children, requireRole }: ProtectedRouteProps) {
  const isAuthenticated = useAuthStore((state) => state.isAuthenticated)
  const user = useAuthStore((state) => state.user)

  if (!isAuthenticated || !user) {
    return <Navigate to="/login" replace />
  }

  if (requireRole && user.role !== requireRole) {
    return <Navigate to="/dashboard" replace />
  }

  return <>{children}</>
}

export function RoleBased({ children, allowedRoles }: { children: React.ReactNode; allowedRoles: string[] }) {
  const user = useAuthStore((state) => state.user)

  if (!user || !allowedRoles.includes(user.role)) {
    return null
  }

  return <>{children}</>
}
