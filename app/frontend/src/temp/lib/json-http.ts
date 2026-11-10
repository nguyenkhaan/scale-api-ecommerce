import { errorFromResponse, invalidResponseError, networkError } from './api-error'
import { csrfCookieName, readDocumentCookie } from './auth-api'

export type Fetcher = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>

export type JsonHttpOptions = {
  baseURL: string
  fetcher?: Fetcher
  readCookie?: (name: string) => string | null
}

type RequestSpec = { method: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'; body?: unknown }

/** Cookie-session JSON transport: CSRF header on every non-GET, ApiError on non-2xx. */
export function createJsonHttp(options: JsonHttpOptions) {
  const baseURL = options.baseURL.replace(/\/$/, '')
  const fetcher: Fetcher = options.fetcher ?? ((input, init) => globalThis.fetch(input, init))
  const readCookie = options.readCookie ?? readDocumentCookie

  async function send(path: string, spec: RequestSpec): Promise<Response> {
    const headers: Record<string, string> = { Accept: 'application/json' }
    if (spec.body !== undefined) headers['Content-Type'] = 'application/json'
    if (spec.method !== 'GET') headers['X-CSRF-Token'] = readCookie(csrfCookieName) ?? ''
    let response: Response
    try {
      response = await fetcher(`${baseURL}${path}`, {
        method: spec.method,
        credentials: 'include',
        headers,
        body: spec.body === undefined ? undefined : JSON.stringify(spec.body),
      })
    } catch {
      throw networkError()
    }
    if (!response.ok) throw await errorFromResponse(response)
    return response
  }

  return {
    async json<T>(path: string, spec: RequestSpec): Promise<T> {
      const response = await send(path, spec)
      try {
        return (await response.json()) as T
      } catch {
        throw invalidResponseError(response)
      }
    },
    /** For 204 No Content operations: any body is ignored. */
    async noContent(path: string, spec: RequestSpec): Promise<void> {
      await send(path, spec)
    },
  }
}
