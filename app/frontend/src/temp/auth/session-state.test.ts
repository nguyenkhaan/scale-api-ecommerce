import { describe, expect, it } from 'bun:test'
import { ApiError, createAuthClient } from '../lib/auth-api'
import { readSessionState, signOut } from './session-state'

const user = { id: 'u1', email: 'a@b.co', role: 'customer' as const }

type Route = () => Response | Promise<Response>

/** Builds a real auth client whose fetcher answers per path; a missing route simulates a network failure. */
function clientFor(routes: { logout?: Route; me?: Route }) {
  return createAuthClient({
    baseURL: '',
    readCookie: () => 'csrf-token',
    fetcher: async (input) => {
      const path = String(input)
      const route = path.endsWith('/logout') ? routes.logout : routes.me
      if (route === undefined) throw new TypeError('network down')
      return route()
    },
  })
}

const problem = (status: number, code: string) =>
  new Response(JSON.stringify({ status, code, request_id: 'r1' }), {
    status,
    headers: { 'Content-Type': 'application/problem+json' },
  })

describe('signOut', () => {
  it('signs out when the server revokes the session', async () => {
    const client = clientFor({ logout: () => new Response(null, { status: 204 }) })
    expect(await signOut(client)).toEqual({ status: 'anonymous' })
  })

  it('signs out when the server reports no session', async () => {
    const client = clientFor({ logout: () => problem(401, 'UNAUTHENTICATED') })
    expect(await signOut(client)).toEqual({ status: 'anonymous' })
  })

  it('rejects when the session is still valid after a failed logout', async () => {
    const client = clientFor({ logout: () => problem(403, 'FORBIDDEN'), me: () => Response.json(user) })
    expect(signOut(client)).rejects.toBeInstanceOf(ApiError)
  })

  it('signs out when a failed logout leaves no session behind', async () => {
    const client = clientFor({ logout: () => problem(500, 'INTERNAL_ERROR'), me: () => problem(401, 'UNAUTHENTICATED') })
    expect(await signOut(client)).toEqual({ status: 'anonymous' })
  })

  it('rejects when the network is down and the session state is unknown', async () => {
    expect(signOut(clientFor({}))).rejects.toBeInstanceOf(ApiError)
  })
})

describe('readSessionState', () => {
  it('distinguishes signed in, signed out and unknown', async () => {
    expect(await readSessionState(clientFor({ me: () => Response.json(user) }))).toEqual({ status: 'authenticated', user })
    expect(await readSessionState(clientFor({ me: () => problem(401, 'UNAUTHENTICATED') }))).toEqual({ status: 'anonymous' })
    expect(await readSessionState(clientFor({ me: () => problem(503, 'INTERNAL_ERROR') }))).toBeNull()
  })
})
