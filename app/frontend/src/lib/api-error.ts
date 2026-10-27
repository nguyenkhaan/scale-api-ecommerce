export type FieldViolation = {
  field: string
  reason: string
}

export type ApiErrorInit = {
  status: number
  code: string
  requestID?: string
  violations?: FieldViolation[]
  retryAfterSeconds?: number
}

/** Error built from an RFC 9457 Problem body, or a client-side NETWORK_ERROR / INVALID_RESPONSE. */
export class ApiError extends Error {
  readonly status: number
  readonly code: string
  readonly requestID?: string
  readonly violations: FieldViolation[]
  readonly retryAfterSeconds?: number

  constructor(init: ApiErrorInit) {
    super(`api error ${init.code} (${init.status})`)
    this.name = 'ApiError'
    this.status = init.status
    this.code = init.code
    this.requestID = init.requestID
    this.violations = init.violations ?? []
    this.retryAfterSeconds = init.retryAfterSeconds
  }
}

export function networkError(): ApiError {
  return new ApiError({ status: 0, code: 'NETWORK_ERROR' })
}

export function invalidResponseError(response: Response): ApiError {
  return new ApiError({ status: response.status, code: 'INVALID_RESPONSE' })
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null
}

function parseViolations(value: unknown): FieldViolation[] {
  if (!Array.isArray(value)) return []
  return value.flatMap((item): FieldViolation[] =>
    isRecord(item) && typeof item.field === 'string' && typeof item.reason === 'string'
      ? [{ field: item.field, reason: item.reason }]
      : [],
  )
}

function parseRetryAfter(header: string | null): number | undefined {
  if (header === null || !/^\d+$/.test(header.trim())) return undefined
  return Number(header.trim())
}

/** Builds an ApiError from a non-2xx response. Unreadable bodies still yield a status-based error. */
export async function errorFromResponse(response: Response): Promise<ApiError> {
  const retryAfterSeconds = parseRetryAfter(response.headers.get('Retry-After'))
  let body: unknown
  try {
    body = await response.json()
  } catch {
    body = undefined
  }
  const problem = isRecord(body) ? body : {}
  return new ApiError({
    status: response.status,
    code: typeof problem.code === 'string' ? problem.code : 'UNKNOWN_ERROR',
    requestID:
      typeof problem.request_id === 'string'
        ? problem.request_id
        : (response.headers.get('X-Request-ID') ?? undefined),
    violations: parseViolations(problem.violations),
    retryAfterSeconds,
  })
}
