import { describe, expect, it } from 'bun:test'
import { ApiError, createAuthClient, readDocumentCookie } from './auth-api'

const user = { id: 'u1', email: 'a@b.co', role: 'customer' as const }
const authResult = { user, expires_at: '2030-01-01T00:00:00Z', csrf_token: 'tok' }

type Call = { url: string; init: RequestInit }

function recorder(respond: () => Response) {
  const calls: Call[] = []
  const fetcher = async (input: RequestInfo | URL, init?: RequestInit) => {
    calls.push({ url: String(input), init: init ?? {} })
    return respond()
  }
  return { calls, fetcher }
}

function problem(status: number, body: object, headers: Record<string, string> = {}) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/problem+json', ...headers },
  })
}

async function caught(promise: Promise<unknown>): Promise<ApiError> {
  try {
    await promise
  } catch (error) {
    if (error instanceof ApiError) return error
    throw error
  }
  throw new Error('expected rejection')
}

describe('auth client requests', () => {
  it('registers with JSON body, credentials and headers', async () => {
    const { calls, fetcher } = recorder(() => Response.json(authResult, { status: 201 }))
    const client = createAuthClient({ baseURL: 'http://api.test/', fetcher })

    const result = await client.register({ email: 'a@b.co', password: 'password1' })

    expect(result.user.email).toBe('a@b.co')
    expect(calls[0].url).toBe('http://api.test/api/v1/auth/register')
    expect(calls[0].init.method).toBe('POST')
    expect(calls[0].init.credentials).toBe('include')
    expect(calls[0].init.headers).toEqual({ Accept: 'application/json', 'Content-Type': 'application/json' })
    expect(calls[0].init.body).toBe(JSON.stringify({ email: 'a@b.co', password: 'password1' }))
  })

  it('logs in against the login path', async () => {
    const { calls, fetcher } = recorder(() => Response.json(authResult))
    const client = createAuthClient({ baseURL: '', fetcher })

    await client.login({ email: 'a@b.co', password: 'x' })

    expect(calls[0].url).toBe('/api/v1/auth/login')
  })

  it('me sends a GET without a body and returns the user', async () => {
    const { calls, fetcher } = recorder(() => Response.json(user))
    const client = createAuthClient({ baseURL: '', fetcher })

    expect(await client.me()).toEqual(user)
    expect(calls[0].init.method).toBe('GET')
    expect(calls[0].init.credentials).toBe('include')
    expect(calls[0].init.body).toBeUndefined()
    expect(calls[0].init.headers).toEqual({ Accept: 'application/json' })
  })

  it('me resolves to null on 401', async () => {
    const { fetcher } = recorder(() => problem(401, { code: 'UNAUTHENTICATED' }))
    const client = createAuthClient({ baseURL: '', fetcher })

    expect(await client.me()).toBeNull()
  })

  it('logout sends the CSRF token read from the salescale_csrf cookie', async () => {
    const { calls, fetcher } = recorder(() => new Response(null, { status: 204 }))
    const names: string[] = []
    const client = createAuthClient({
      baseURL: '',
      fetcher,
      readCookie: (name) => {
        names.push(name)
        return 'csrf-abc'
      },
    })

    await client.logout()

    expect(names).toEqual(['salescale_csrf'])
    expect(calls[0].url).toBe('/api/v1/auth/logout')
    expect(calls[0].init.method).toBe('POST')
    expect(calls[0].init.body).toBeUndefined()
    expect(calls[0].init.headers).toEqual({ Accept: 'application/json', 'X-CSRF-Token': 'csrf-abc' })
  })

  it('logout surfaces 403 as ApiError', async () => {
    const { fetcher } = recorder(() => problem(403, { code: 'FORBIDDEN', request_id: 'r1' }))
    const client = createAuthClient({ baseURL: '', fetcher, readCookie: () => null })

    const error = await caught(client.logout())

    expect(error.status).toBe(403)
    expect(error.code).toBe('FORBIDDEN')
  })
})

describe('auth client errors', () => {
  it('parses violations, request id and Retry-After from a Problem body', async () => {
    const { fetcher } = recorder(() =>
      problem(
        422,
        {
          type: 't',
          title: 'x',
          status: 422,
          code: 'VALIDATION_ERROR',
          request_id: 'req-9',
          violations: [{ field: 'email', reason: 'invalid' }, { bad: true }],
        },
        { 'Retry-After': '30' },
      ),
    )
    const client = createAuthClient({ baseURL: '', fetcher })

    const error = await caught(client.login({ email: 'x', password: 'y' }))

    expect(error.code).toBe('VALIDATION_ERROR')
    expect(error.requestID).toBe('req-9')
    expect(error.violations).toEqual([{ field: 'email', reason: 'invalid' }])
    expect(error.retryAfterSeconds).toBe(30)
  })

  it('parses Retry-After on RATE_LIMITED', async () => {
    const { fetcher } = recorder(() => problem(429, { code: 'RATE_LIMITED' }, { 'Retry-After': '45' }))
    const client = createAuthClient({ baseURL: '', fetcher })

    const error = await caught(client.login({ email: 'a@b.co', password: 'x' }))

    expect(error.status).toBe(429)
    expect(error.retryAfterSeconds).toBe(45)
  })

  it('maps a non-JSON error body to UNKNOWN_ERROR with the status', async () => {
    const { fetcher } = recorder(() => new Response('boom', { status: 502 }))
    const client = createAuthClient({ baseURL: '', fetcher })

    const error = await caught(client.me())

    expect(error.status).toBe(502)
    expect(error.code).toBe('UNKNOWN_ERROR')
  })

  it('maps a thrown fetch to NETWORK_ERROR', async () => {
    const client = createAuthClient({
      baseURL: '',
      fetcher: async () => {
        throw new TypeError('failed to fetch')
      },
    })

    const error = await caught(client.login({ email: 'a@b.co', password: 'x' }))

    expect(error.code).toBe('NETWORK_ERROR')
    expect(error.status).toBe(0)
  })

  it('rejects an AuthResult with an invalid payload', async () => {
    const { fetcher } = recorder(() => Response.json({ user: { id: 1 } }))
    const client = createAuthClient({ baseURL: '', fetcher })

    const error = await caught(client.login({ email: 'a@b.co', password: 'x' }))

    expect(error.code).toBe('INVALID_RESPONSE')
  })

  it('rejects a User with an unknown role', async () => {
    const { fetcher } = recorder(() => Response.json({ ...user, role: 'root' }))
    const client = createAuthClient({ baseURL: '', fetcher })

    const error = await caught(client.me())

    expect(error.code).toBe('INVALID_RESPONSE')
  })
})

describe('readDocumentCookie', () => {
  it('returns null when document is unavailable', () => {
    expect(readDocumentCookie('salescale_csrf')).toBeNull()
  })
})
