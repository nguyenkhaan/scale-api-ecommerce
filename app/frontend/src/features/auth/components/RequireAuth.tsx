import type { ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'

import { loginRedirectPath, useAuth } from '../auth.hook'
import { AuthLoadingView } from './AuthStatusViews'

export function RequireAuth({ children }: { children: ReactNode }) {
  const { state } = useAuth()
  const location = useLocation()

  if (state.status === 'loading') return <AuthLoadingView />
  if (state.status === 'anonymous') {
    return <Navigate to={loginRedirectPath(location.pathname, location.search)} replace />
  }
  return <>{children}</>
}
