import { ApiError, errorFromResponse, invalidResponseError, networkError } from './api-error'

export { ApiError } from './api-error'
export type { FieldViolation } from './api-error'

export type Role = 'customer' | 'admin'

export type User = {
  id: string
  email: string
  role: Role
}

export type AuthResult = {
  user: User
  expires_at: string
  csrf_token: string
}

export type Credentials = {
  email: string
  password: string
}

type Fetcher = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>

type AuthClientOptions = {
  baseURL: string
  fetcher?: Fetcher
  /** Returns the value of a cookie, or null when absent. Injectable for tests. */
  readCookie?: (name: string) => string | null
}

export const csrfCookieName = 'salescale_csrf'
const csrfHeader = 'X-CSRF-Token'

export function readDocumentCookie(name: string): string | null {
  if (typeof document === 'undefined') return null
  for (const part of document.cookie.split(';')) {
    const index = part.indexOf('=')
    if (index > 0 && part.slice(0, index).trim() === name) {
      try {
        return decodeURIComponent(part.slice(index + 1).trim())
      } catch {
        return null
      }
    }
  }
  return null
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function isUser(value: unknown): value is User {
  return (
    isRecord(value) &&
    typeof value.id === 'string' &&
    typeof value.email === 'string' &&
    (value.role === 'customer' || value.role === 'admin')
  )
}

function isAuthResult(value: unknown): value is AuthResult {
  return (
    isRecord(value) &&
    isUser(value.user) &&
    typeof value.expires_at === 'string' &&
    typeof value.csrf_token === 'string'
  )
}

export function createAuthClient(options: AuthClientOptions) {
  const baseURL = options.baseURL.replace(/\/$/, '')
  const fetcher: Fetcher = options.fetcher ?? ((input, init) => globalThis.fetch(input, init))
  const readCookie = options.readCookie ?? readDocumentCookie

  async function send(path: string, init: { method: string; body?: unknown; csrf?: string }) {
    const headers: Record<string, string> = { Accept: 'application/json' }
    if (init.body !== undefined) headers['Content-Type'] = 'application/json'
    if (init.csrf !== undefined) headers[csrfHeader] = init.csrf
    try {
      return await fetcher(`${baseURL}${path}`, {
        method: init.method,
        credentials: 'include',
        headers,
        body: init.body === undefined ? undefined : JSON.stringify(init.body),
      })
    } catch {
      throw networkError()
    }
  }

  async function readJSON(response: Response): Promise<unknown> {
    try {
      return await response.json()
    } catch {
      throw invalidResponseError(response)
    }
  }

  async function authenticate(path: string, credentials: Credentials): Promise<AuthResult> {
    const response = await send(path, { method: 'POST', body: credentials })
    if (!response.ok) throw await errorFromResponse(response)
    const body = await readJSON(response)
    if (!isAuthResult(body)) throw invalidResponseError(response)
    return body
  }

  return {
    register: (credentials: Credentials) => authenticate('/api/v1/auth/register', credentials),
    login: (credentials: Credentials) => authenticate('/api/v1/auth/login', credentials),

    /** Resolves to null when there is no valid session (401). */
    async me(): Promise<User | null> {
      const response = await send('/api/v1/auth/me', { method: 'GET' })
      if (response.status === 401) return null
      if (!response.ok) throw await errorFromResponse(response)
      const body = await readJSON(response)
      if (!isUser(body)) throw invalidResponseError(response)
      return body
    },

    async logout(): Promise<void> {
      // Missing cookie: send an empty header; the server answers 403 and the caller still clears local state.
      const csrf = readCookie(csrfCookieName) ?? ''
      const response = await send('/api/v1/auth/logout', { method: 'POST', csrf })
      if (!response.ok) throw await errorFromResponse(response)
    },
  }
}

export type AuthClient = ReturnType<typeof createAuthClient>

export const authClient = createAuthClient({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? '',
})

export function isApiError(error: unknown): error is ApiError {
  return error instanceof ApiError
}
