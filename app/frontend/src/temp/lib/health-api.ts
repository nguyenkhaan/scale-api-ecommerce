const readinessPath = '/api/health/ready'
const requestIDHeader = 'X-Request-ID'

export type ReadinessResponse = {
  status: 'service up'
}

export type ReadinessResult = ReadinessResponse & {
  requestID: string
}

type HealthClientOptions = {
  baseURL: string
  fetcher?: Fetcher
}

type Fetcher = (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>

export const healthFixture = {
  readiness: { status: 'service up' } satisfies ReadinessResponse,
}

export function createHealthClient(options: HealthClientOptions) {
  const baseURL = options.baseURL.replace(/\/$/, '')
  const fetcher: Fetcher = options.fetcher ?? globalThis.fetch

  return {
    async getReadiness(signal?: AbortSignal): Promise<ReadinessResult> {
      const response = await fetcher(`${baseURL}${readinessPath}`, {
        credentials: 'include',
        headers: { Accept: 'application/json' },
        signal,
      })
      if (!response.ok) {
        throw new Error(`backend readiness check failed (${response.status})`)
      }

      const body: unknown = await response.json()
      if (!isReadinessResponse(body)) {
        throw new Error('invalid readiness response')
      }
      return {
        ...body,
        requestID: response.headers.get(requestIDHeader) ?? 'not returned',
      }
    },
  }
}

export const healthClient = createHealthClient({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? '',
})

function isReadinessResponse(value: unknown): value is ReadinessResponse {
  if (typeof value !== 'object' || value === null) {
    return false
  }
  return 'status' in value && value.status === 'service up'
}
