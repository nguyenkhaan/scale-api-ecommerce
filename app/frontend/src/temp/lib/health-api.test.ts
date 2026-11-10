import { describe, expect, it } from 'bun:test'
import { createHealthClient, healthFixture } from './health-api'

describe('health client', () => {
  it('uses the configured API base URL and accepts the readiness schema', async () => {
    const requestedURLs: string[] = []
    const client = createHealthClient({
      baseURL: 'http://localhost:4000/',
      fetcher: async (input) => {
        requestedURLs.push(String(input))
        return Response.json(healthFixture.readiness)
      },
    })

    const readiness = await client.getReadiness()

    expect(requestedURLs).toEqual(['http://localhost:4000/api/health/ready'])
    expect(readiness.status).toBe('service up')
  })

  it('rejects a successful response with an invalid payload', async () => {
    const client = createHealthClient({
      baseURL: '',
      fetcher: async () => Response.json({ unexpected: true }),
    })

    expect(client.getReadiness()).rejects.toThrow('invalid readiness response')
  })

  it('returns a stable error without exposing the response body', async () => {
    const client = createHealthClient({
      baseURL: '',
      fetcher: async () => new Response('database hostname is secret', { status: 503 }),
    })

    expect(client.getReadiness()).rejects.toThrow('backend readiness check failed (503)')
  })
})
