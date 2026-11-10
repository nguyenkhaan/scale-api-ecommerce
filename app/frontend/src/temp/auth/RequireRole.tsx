import type { ReactNode } from 'react'
import type { Role } from '../lib/auth-api'
import { ForbiddenView } from './AuthStatusViews'
import { RequireAuth } from './RequireAuth'
import { useAuth } from './use-auth'

// Display-only guard: the server still enforces the role on every request.
export function RequireRole({ role, children }: { role: Role; children: ReactNode }) {
  return (
    <RequireAuth>
      <RoleGate role={role}>{children}</RoleGate>
    </RequireAuth>
  )
}

function RoleGate({ role, children }: { role: Role; children: ReactNode }) {
  const { state } = useAuth()
  if (state.status !== 'authenticated' || state.user.role !== role) return <ForbiddenView />
  return <>{children}</>
}
