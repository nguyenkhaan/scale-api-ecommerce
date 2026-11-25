import { describe, expect, it } from 'bun:test'

import { ApiError } from '@/lib/api-error'
import { createJsonHttp } from '@/lib/json-http'

describe('json http checkout transport', () => {
  it('merges custom headers and returns response headers', async () => {
    let request: RequestInit | undefined
    const http = createJsonHttp({
      baseURL: '/api',
      readCookie: () => 'csrf-token',
      fetcher: async (_input, init) => {
        request = init
        return new Response(JSON.stringify({ ok: true }), {
          status: 200,
          headers: { 'Content-Type': 'application/json', 'Idempotency-Replayed': 'true' },
        })
      },
    })

    const result = await http.withHeaders<{ ok: boolean }>('/orders', {
      method: 'POST',
      body: { order: true },
      headers: { 'Idempotency-Key': 'intent-1' },
    })

    const sentHeaders = request?.headers as Record<string, string> | undefined
    expect(result.data).toEqual({ ok: true })
    expect(result.headers.get('Idempotency-Replayed')).toBe('true')
    expect(sentHeaders?.['Idempotency-Key']).toBe('intent-1')
    expect(sentHeaders?.['X-CSRF-Token']).toBe('csrf-token')
  })

  it('turns a request timeout into a NETWORK_ERROR ApiError', async () => {
    const http = createJsonHttp({
      baseURL: '/api',
      fetcher: (_input, init) =>
        new Promise<Response>((_resolve, reject) => {
          init?.signal?.addEventListener('abort', () => reject(new DOMException('aborted', 'AbortError')))
        }),
    })

    const error = await http.json('/orders', { method: 'GET', timeoutMs: 5 }).catch((value: unknown) => value)

    expect(error).toBeInstanceOf(ApiError)
    expect((error as ApiError).status).toBe(0)
    expect((error as ApiError).code).toBe('NETWORK_ERROR')
  })
})
