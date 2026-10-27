import { useCallback, useEffect, useMemo, useRef, useState, type ReactNode } from 'react'
import { authClient, type AuthClient } from '../lib/auth-api'
import { AuthContext, type AuthContextValue, type AuthState } from './auth-context'
import { readSessionState, signOut } from './session-state'

type AuthProviderProps = {
  children: ReactNode
  client?: AuthClient
}

export function AuthProvider({ children, client = authClient }: AuthProviderProps) {
  const [state, setState] = useState<AuthState>({ status: 'loading' })
  // Bumped by login, register and logout so a slower, earlier session read cannot overwrite their result.
  const generation = useRef(0)

  const refresh = useCallback(async () => {
    const started = generation.current
    // Unknown session state (network/5xx) is shown as signed out; protected routes redirect to login.
    const next = (await readSessionState(client)) ?? { status: 'anonymous' }
    if (generation.current === started) setState(next)
  }, [client])

  useEffect(() => {
    void refresh()
  }, [refresh])

  const value = useMemo<AuthContextValue>(() => {
    const apply = (next: AuthState) => {
      generation.current += 1
      setState(next)
    }
    return {
      state,
      refresh,
      async login(credentials) {
        const result = await client.login(credentials)
        apply({ status: 'authenticated', user: result.user })
        return result.user
      },
      async register(credentials) {
        const result = await client.register(credentials)
        apply({ status: 'authenticated', user: result.user })
        return result.user
      },
      async logout() {
        apply(await signOut(client))
      },
    }
  }, [client, refresh, state])

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
