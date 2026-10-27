import { createContext } from 'react'
import type { Credentials, User } from '../lib/auth-api'

export type AuthState =
  | { status: 'loading' }
  | { status: 'anonymous' }
  | { status: 'authenticated'; user: User }

export type AuthContextValue = {
  state: AuthState
  login: (credentials: Credentials) => Promise<User>
  register: (credentials: Credentials) => Promise<User>
  /** Revokes the server session. Rejects and keeps the current state when the session may still be valid. */
  logout: () => Promise<void>
  refresh: () => Promise<void>
}

export const AuthContext = createContext<AuthContextValue | null>(null)
