import type { ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { AuthLoadingView } from './AuthStatusViews'
import { loginRedirectPath } from './redirect-paths'
import { useAuth } from './use-auth'

// Route guards are display-only. The server enforces authentication and
// authorization on every request; never rely on these for security.
export function RequireAuth({ children }: { children: ReactNode }) {
  const { state } = useAuth()
  const location = useLocation()

  if (state.status === 'loading') return <AuthLoadingView />
  if (state.status === 'anonymous') {
    return <Navigate to={loginRedirectPath(location.pathname, location.search)} replace />
  }
  return <>{children}</>
}
