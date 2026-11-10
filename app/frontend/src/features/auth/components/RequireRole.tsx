import type { ReactNode } from 'react'

import type { Role } from '../auth.type'
import { useAuth } from '../auth.hook'
import { ForbiddenView } from './AuthStatusViews'
import { RequireAuth } from './RequireAuth'

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
