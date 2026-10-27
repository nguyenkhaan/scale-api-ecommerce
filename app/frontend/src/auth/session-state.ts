import { ApiError, type AuthClient } from '../lib/auth-api'
import type { AuthState } from './auth-context'

/** Reads the server-side session; `null` means the state is unknown (network or server error). */
export async function readSessionState(client: AuthClient): Promise<AuthState | null> {
  try {
    const user = await client.me()
    return user === null ? { status: 'anonymous' } : { status: 'authenticated', user }
  } catch {
    return null
  }
}

/**
 * Revokes the session and resolves to the signed-out state. Rejects when the server may still honor the
 * HttpOnly session cookie, so the UI never claims a sign-out that did not happen (e.g. on a shared computer).
 */
export async function signOut(client: AuthClient): Promise<AuthState> {
  try {
    await client.logout()
    return { status: 'anonymous' }
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) return { status: 'anonymous' }
    const current = await readSessionState(client)
    if (current?.status === 'anonymous') return current
    throw error
  }
}
